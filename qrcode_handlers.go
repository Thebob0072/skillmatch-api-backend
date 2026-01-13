package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PromptPay QR Code Generator
// อ้างอิง: EMVCo QR Code Specification for Payment Systems
func generatePromptPayQR(phoneNumber string, amount float64) string {
	// Format: 00020101021129370016A000000677010111021[phone]5303764540[amount]6304[checksum]

	// Merchant Account Information (Tag 29)
	merchantID := fmt.Sprintf("0016A000000677010111%02d%s", len(phoneNumber), phoneNumber)
	merchantInfo := fmt.Sprintf("29%02d%s", len(merchantID), merchantID)

	// Amount (Tag 54)
	amountStr := fmt.Sprintf("%.2f", amount)
	amountTag := fmt.Sprintf("54%02d%s", len(amountStr), amountStr)

	// Country Code (Tag 58)
	countryCode := "5802TH"

	// Transaction Currency (Tag 53) - THB = 764
	currencyCode := "5303764"

	// Build QR payload (without checksum)
	payload := fmt.Sprintf("00020101021130%02d%s%s%s%s6304", len(merchantInfo), merchantInfo, currencyCode, amountTag, countryCode)

	// Calculate CRC16-CCITT checksum
	checksum := calculateCRC16(payload)

	return payload + checksum
}

// CRC16-CCITT Checksum calculation for PromptPay QR
func calculateCRC16(data string) string {
	crc := uint16(0xFFFF)
	polynomial := uint16(0x1021)

	for i := 0; i < len(data); i++ {
		crc ^= uint16(data[i]) << 8
		for j := 0; j < 8; j++ {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ polynomial
			} else {
				crc = crc << 1
			}
		}
	}

	return fmt.Sprintf("%04X", crc&0xFFFF)
}

// --- POST /bookings/create-with-qr (สร้างการจองพร้อม QR Code PromptPay) ---
func createBookingWithQRHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req struct {
			ProviderID   int     `json:"provider_id" binding:"required"`
			PackageID    int     `json:"package_id" binding:"required"`
			BookingDate  string  `json:"booking_date" binding:"required"` // YYYY-MM-DD
			StartTime    string  `json:"start_time" binding:"required"`   // HH:MM
			Location     *string `json:"location"`
			SpecialNotes *string `json:"special_notes"`
			PhoneNumber  string  `json:"phone_number" binding:"required"` // เบอร์โทร PromptPay ของผู้รับเงิน
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 1. ดึงราคาแพ็คเกจ
		var packagePrice float64
		var packageName string
		err := dbPool.QueryRow(ctx, `
			SELECT price, package_name FROM service_packages WHERE package_id = $1 AND is_active = true
		`, req.PackageID).Scan(&packagePrice, &packageName)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Package not found or inactive"})
			return
		}

		// 2. สร้าง Booking ใน DB (สถานะ pending_payment)
		var bookingID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO bookings (
				client_id, provider_id, package_id, booking_date, start_time, 
				location, special_notes, status, total_price, payment_method
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending_payment', $8, 'promptpay')
			RETURNING booking_id
		`, clientID, req.ProviderID, req.PackageID, req.BookingDate, req.StartTime,
			req.Location, req.SpecialNotes, packagePrice).Scan(&bookingID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking", "details": err.Error()})
			return
		}

		// 3. สร้าง QR Code PromptPay
		qrCode := generatePromptPayQR(req.PhoneNumber, packagePrice)

		// 4. สร้าง Payment Record
		paymentReference := generatePaymentReference(bookingID)
		_, err = dbPool.Exec(ctx, `
			INSERT INTO payments (
				booking_id, amount, payment_method, payment_status, 
				payment_reference, qr_code, expires_at
			)
			VALUES ($1, $2, 'promptpay', 'pending', $3, $4, $5)
		`, bookingID, packagePrice, paymentReference, qrCode, time.Now().Add(15*time.Minute))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment record"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"booking_id":        bookingID,
			"qr_code":           qrCode,
			"amount":            packagePrice,
			"payment_reference": paymentReference,
			"expires_at":        time.Now().Add(15 * time.Minute).Format(time.RFC3339),
			"package_name":      packageName,
			"message":           "Scan QR code to pay within 15 minutes",
		})
	}
}

// --- POST /payments/:payment_reference/confirm (ยืนยันการชำระเงินแบบแมนนวล) ---
func confirmPaymentHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		paymentRef := c.Param("payment_reference")

		var req struct {
			TransactionID string  `json:"transaction_id"` // เลข Ref จากธนาคาร (optional)
			SlipImage     *string `json:"slip_image"`     // URL รูปสลิป (optional)
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 1. ค้นหา Payment และ Booking
		var bookingID int
		var paymentID int
		var amount float64
		var providerID int

		err := dbPool.QueryRow(ctx, `
			SELECT p.payment_id, p.booking_id, p.amount, b.provider_id
			FROM payments p
			JOIN bookings b ON p.booking_id = b.booking_id
			WHERE p.payment_reference = $1 AND p.payment_status = 'pending'
		`, paymentRef).Scan(&paymentID, &bookingID, &amount, &providerID)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found or already completed"})
			return
		}

		// 2. อัพเดทสถานะ Payment
		_, err = dbPool.Exec(ctx, `
			UPDATE payments 
			SET payment_status = 'completed', 
			    transaction_id = $1,
			    slip_image = $2,
			    paid_at = NOW()
			WHERE payment_id = $3
		`, req.TransactionID, req.SlipImage, paymentID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment"})
			return
		}

		// 3. อัพเดทสถานะ Booking เป็น confirmed
		_, err = dbPool.Exec(ctx, `
			UPDATE bookings 
			SET status = 'confirmed', payment_status = 'paid'
			WHERE booking_id = $1
		`, bookingID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking"})
			return
		}

		// 4. เพิ่มเงินเข้า Wallet ของ Provider (หักค่าคอมฯ 12.75%)
		commission := amount * 0.1275
		netAmount := amount - commission

		_, err = dbPool.Exec(ctx, `
			INSERT INTO wallets (user_id, balance)
			VALUES ($1, 0)
			ON CONFLICT (user_id) DO NOTHING
		`, providerID)

		_, err = dbPool.Exec(ctx, `
			UPDATE wallets 
			SET balance = balance + $1
			WHERE user_id = $2
		`, netAmount, providerID)

		// 5. บันทึก Transaction
		_, err = dbPool.Exec(ctx, `
			INSERT INTO transactions (
				user_id, transaction_type, amount, booking_id, 
				description, balance_after
			)
			VALUES (
				$1, 'booking_payment', $2, $3, 
				'Payment from booking', 
				(SELECT balance FROM wallets WHERE user_id = $1)
			)
		`, providerID, netAmount, bookingID)

		c.JSON(http.StatusOK, gin.H{
			"message":     "Payment confirmed successfully",
			"booking_id":  bookingID,
			"amount_paid": amount,
			"net_amount":  netAmount,
			"commission":  commission,
		})
	}
}

// --- GET /payments/:payment_reference/status (ตรวจสอบสถานะการชำระเงิน) ---
func checkPaymentStatusHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		paymentRef := c.Param("payment_reference")

		var payment struct {
			PaymentID     int        `json:"payment_id"`
			BookingID     int        `json:"booking_id"`
			Amount        float64    `json:"amount"`
			PaymentStatus string     `json:"payment_status"`
			QRCode        string     `json:"qr_code"`
			ExpiresAt     time.Time  `json:"expires_at"`
			PaidAt        *time.Time `json:"paid_at"`
		}

		err := dbPool.QueryRow(ctx, `
			SELECT payment_id, booking_id, amount, payment_status, qr_code, expires_at, paid_at
			FROM payments
			WHERE payment_reference = $1
		`, paymentRef).Scan(
			&payment.PaymentID, &payment.BookingID, &payment.Amount,
			&payment.PaymentStatus, &payment.QRCode, &payment.ExpiresAt, &payment.PaidAt,
		)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
			return
		}

		// ตรวจสอบว่าหมดอายุหรือยัง
		isExpired := time.Now().After(payment.ExpiresAt) && payment.PaymentStatus == "pending"
		if isExpired {
			// อัพเดทสถานะเป็น expired
			dbPool.Exec(ctx, `
				UPDATE payments SET payment_status = 'expired' WHERE payment_id = $1
			`, payment.PaymentID)
			payment.PaymentStatus = "expired"
		}

		c.JSON(http.StatusOK, payment)
	}
}

// สร้าง Payment Reference (unique)
func generatePaymentReference(bookingID int) string {
	timestamp := time.Now().Unix()
	data := fmt.Sprintf("%d-%d", bookingID, timestamp)
	hash := sha256.Sum256([]byte(data))
	return "PAY" + hex.EncodeToString(hash[:])[:12]
}

// --- GET /bookings/:booking_id/payment (ดูข้อมูลการชำระเงินของ Booking) ---
func getBookingPaymentHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID, err := strconv.Atoi(c.Param("booking_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
			return
		}

		userID, _ := c.Get("userID")

		// ตรวจสอบว่าเป็นเจ้าของ Booking หรือ Provider
		var isAuthorized bool
		err = dbPool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM bookings 
				WHERE booking_id = $1 AND (client_id = $2 OR provider_id = $2)
			)
		`, bookingID, userID).Scan(&isAuthorized)

		if err != nil || !isAuthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		var payment struct {
			PaymentID        int        `json:"payment_id"`
			Amount           float64    `json:"amount"`
			PaymentMethod    string     `json:"payment_method"`
			PaymentStatus    string     `json:"payment_status"`
			PaymentReference string     `json:"payment_reference"`
			QRCode           *string    `json:"qr_code"`
			TransactionID    *string    `json:"transaction_id"`
			SlipImage        *string    `json:"slip_image"`
			PaidAt           *time.Time `json:"paid_at"`
			ExpiresAt        *time.Time `json:"expires_at"`
			CreatedAt        time.Time  `json:"created_at"`
		}

		err = dbPool.QueryRow(ctx, `
			SELECT 
				payment_id, amount, payment_method, payment_status, payment_reference,
				qr_code, transaction_id, slip_image, paid_at, expires_at, created_at
			FROM payments
			WHERE booking_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		`, bookingID).Scan(
			&payment.PaymentID, &payment.Amount, &payment.PaymentMethod, &payment.PaymentStatus,
			&payment.PaymentReference, &payment.QRCode, &payment.TransactionID, &payment.SlipImage,
			&payment.PaidAt, &payment.ExpiresAt, &payment.CreatedAt,
		)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
			return
		}

		c.JSON(http.StatusOK, payment)
	}
}
