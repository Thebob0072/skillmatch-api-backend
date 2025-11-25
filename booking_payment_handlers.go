package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
)

// --- POST /bookings/create-with-payment (สร้าง Booking พร้อมชำระเงิน) ---
func createBookingWithPaymentHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req struct {
			ProviderID  int    `json:"provider_id" binding:"required"`
			PackageID   int    `json:"package_id" binding:"required"`
			BookingDate string `json:"booking_date" binding:"required"` // ISO 8601
			Notes       string `json:"notes"`
			SuccessURL  string `json:"success_url"` // Optional
			CancelURL   string `json:"cancel_url"`  // Optional
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
			return
		}

		// 1. ตรวจสอบว่า Package มีอยู่จริง
		var packagePrice float64
		var packageName string
		var providerID int
		err := dbPool.QueryRow(ctx, `
			SELECT provider_id, package_name, price 
			FROM service_packages 
			WHERE package_id = $1 AND is_active = true
		`, req.PackageID).Scan(&providerID, &packageName, &packagePrice)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Package not found or inactive"})
			return
		}

		// ตรวจสอบว่า provider_id ตรงกับที่ระบุไหม
		if providerID != req.ProviderID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Package does not belong to specified provider"})
			return
		}

		// 2. สร้าง Booking ใน DB (สถานะ pending)
		var bookingID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO bookings (client_id, provider_id, package_id, booking_date, notes, status, total_price)
			VALUES ($1, $2, $3, $4, $5, 'pending', $6)
			RETURNING booking_id
		`, userID, req.ProviderID, req.PackageID, req.BookingDate, req.Notes, packagePrice).Scan(&bookingID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking", "details": err.Error()})
			return
		}

		// 3. สร้าง Stripe Checkout Session (payment mode)
		successURL := req.SuccessURL
		cancelURL := req.CancelURL
		if successURL == "" {
			successURL = "http://localhost:5174/booking/success?session_id={CHECKOUT_SESSION_ID}"
		}
		if cancelURL == "" {
			cancelURL = "http://localhost:5174/booking/cancel"
		}

		// แปลงราคาเป็น cents (Stripe ใช้หน่วยเล็กที่สุด)
		priceInCents := int64(packagePrice * 100)

		params := &stripe.CheckoutSessionParams{
			Mode: stripe.String(string(stripe.CheckoutSessionModePayment)), // One-time payment
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency: stripe.String("thb"), // Thai Baht
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name:        stripe.String(packageName),
							Description: stripe.String(fmt.Sprintf("Booking with Provider #%d", req.ProviderID)),
						},
						UnitAmount: stripe.Int64(priceInCents),
					},
					Quantity: stripe.Int64(1),
				},
			},
			SuccessURL:        stripe.String(successURL),
			CancelURL:         stripe.String(cancelURL),
			ClientReferenceID: stripe.String(fmt.Sprintf("%d", userID)),
			Metadata: map[string]string{
				"payment_type": "booking",
				"booking_id":   fmt.Sprintf("%d", bookingID),
				"provider_id":  fmt.Sprintf("%d", req.ProviderID),
				"package_id":   fmt.Sprintf("%d", req.PackageID),
			},
		}

		stripeSession, err := session.New(params)
		if err != nil {
			// หากสร้าง Stripe session ไม่สำเร็จ ลบ booking ที่สร้างไว้
			dbPool.Exec(ctx, "DELETE FROM bookings WHERE booking_id = $1", bookingID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment session", "details": err.Error()})
			return
		}

		// 4. อัปเดต booking ด้วย payment_intent_id (optional - ถ้าจะใช้)
		if stripeSession.PaymentIntent != nil {
			dbPool.Exec(ctx, `
				UPDATE bookings 
				SET payment_intent_id = $1 
				WHERE booking_id = $2
			`, stripeSession.PaymentIntent.ID, bookingID)
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Booking created. Please complete payment.",
			"checkout_url": stripeSession.URL,
			"session_id":   stripeSession.ID,
			"booking_id":   bookingID,
			"total_amount": packagePrice,
		})
	}
}

// --- Helper: Handle Booking Payment from Webhook ---
// เรียกใช้จาก paymentWebhookHandler() เมื่อ metadata.payment_type == "booking"
func handleBookingPayment(dbPool *pgxpool.Pool, ctx context.Context, checkoutSession stripe.CheckoutSession) error {
	// 1. ดึงข้อมูลจาก metadata
	bookingID := checkoutSession.Metadata["booking_id"]
	providerIDStr := checkoutSession.Metadata["provider_id"]

	if bookingID == "" || providerIDStr == "" {
		return fmt.Errorf("missing booking_id or provider_id in metadata")
	}

	// 2. คำนวณค่าธรรมเนียม 12.75%
	totalAmount := float64(checkoutSession.AmountTotal) / 100 // Convert from cents to THB
	stripeFee := totalAmount * 0.0275                         // 2.75%
	platformCommission := totalAmount * 0.1000                // 10%
	// totalFee := stripeFee + platformCommission             // 12.75% (not used directly)
	providerEarnings := totalAmount * 0.8725 // 87.25%

	// 3. อัปเดตสถานะ booking เป็น "paid"
	_, err := dbPool.Exec(ctx, `
		UPDATE bookings 
		SET status = 'paid', 
		    payment_intent_id = $1,
		    updated_at = NOW()
		WHERE booking_id = $2
	`, checkoutSession.PaymentIntent.ID, bookingID)

	if err != nil {
		return fmt.Errorf("failed to update booking status: %v", err)
	}

	// 4. สร้าง transaction record
	var transactionID int
	err = dbPool.QueryRow(ctx, `
		INSERT INTO transactions (
			user_id, type, amount, status, booking_id, 
			stripe_fee, platform_commission, total_fee_percentage, net_amount
		)
		VALUES ($1, 'booking_payment', $2, 'completed', $3, $4, $5, 0.1275, $6)
		RETURNING transaction_id
	`, providerIDStr, totalAmount, bookingID, stripeFee, platformCommission, providerEarnings).Scan(&transactionID)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	// 5. อัปเดต provider wallet (pending_balance - 7 day hold)
	_, err = dbPool.Exec(ctx, `
		INSERT INTO wallets (user_id, available_balance, pending_balance, total_earned)
		VALUES ($1, 0, $2, $2)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			pending_balance = wallets.pending_balance + $2,
			total_earned = wallets.total_earned + $2,
			updated_at = NOW()
	`, providerIDStr, providerEarnings)

	if err != nil {
		return fmt.Errorf("failed to update provider wallet: %v", err)
	}

	// 6. บันทึก commission transaction
	_, err = dbPool.Exec(ctx, `
		INSERT INTO commission_transactions (
			booking_id, transaction_id, booking_amount, 
			commission_rate, commission_amount, provider_amount, provider_id
		)
		VALUES ($1, $2, $3, 0.1000, $4, $5, $6)
	`, bookingID, transactionID, totalAmount, platformCommission, providerEarnings, providerIDStr)

	if err != nil {
		return fmt.Errorf("failed to record commission transaction: %v", err)
	}

	// 7. ส่ง notification ให้ provider
	_, err = dbPool.Exec(ctx, `
		INSERT INTO notifications (user_id, type, title, message, is_read)
		VALUES ($1, 'booking_payment', 'New Booking Payment', 
			'You received a new booking payment of ฿' || $2 || ' (net: ฿' || $3 || ')', false)
	`, providerIDStr, totalAmount, providerEarnings)

	if err != nil {
		fmt.Printf("Warning: Failed to create notification: %v\n", err)
	}

	// 8. Broadcast WebSocket notification
	if wsManager != nil {
		// Convert providerIDStr to int
		providerIDInt, _ := strconv.Atoi(providerIDStr)
		wsManager.BroadcastToUser(providerIDInt, WebSocketMessage{
			Type: "booking_payment",
			Payload: map[string]interface{}{
				"booking_id":        bookingID,
				"amount":            totalAmount,
				"provider_earnings": providerEarnings,
				"message":           "New booking payment received!",
			},
		})
	}

	fmt.Printf("✅ Booking payment processed: BookingID=%s, Amount=฿%.2f, Provider Earnings=฿%.2f\n",
		bookingID, totalAmount, providerEarnings)

	return nil
}
