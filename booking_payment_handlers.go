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

// --- GET /bookings/:id/extension-packages (ดูแพ็คเกจต่อเวลา) ---
func getExtensionPackagesHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
			return
		}

		// Get booking to calculate extension prices
		var basePrice float64
		var duration int
		err = dbPool.QueryRow(ctx, `
			SELECT b.total_price, sp.duration
			FROM bookings b
			JOIN service_packages sp ON b.package_id = sp.package_id
			WHERE b.booking_id = $1
		`, bookingID).Scan(&basePrice, &duration)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		// Calculate per-minute rate
		perMinuteRate := basePrice / float64(duration)

		// Generate extension packages
		packages := []gin.H{
			{"minutes": 30, "price": perMinuteRate * 30 * 0.9, "label": "30 minutes"}, // 10% discount
			{"minutes": 60, "price": perMinuteRate * 60 * 0.85, "label": "1 hour"},    // 15% discount
			{"minutes": 120, "price": perMinuteRate * 120 * 0.8, "label": "2 hours"},  // 20% discount
		}

		c.JSON(http.StatusOK, gin.H{
			"booking_id":      bookingID,
			"base_price":      basePrice,
			"per_minute_rate": perMinuteRate,
			"packages":        packages,
		})
	}
}

// --- POST /bookings/extend (ต่อเวลา Booking) ---
func extendBookingHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req struct {
			BookingID         int     `json:"booking_id" binding:"required"`
			AdditionalMinutes int     `json:"additional_minutes" binding:"required"`
			Price             float64 `json:"price" binding:"required"`
			SuccessURL        string  `json:"success_url"`
			CancelURL         string  `json:"cancel_url"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify booking exists and is active
		var providerID, clientID int
		var status string
		err := dbPool.QueryRow(ctx, `
			SELECT provider_id, client_id, status
			FROM bookings
			WHERE booking_id = $1
		`, req.BookingID).Scan(&providerID, &clientID, &status)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		// Only client or provider can extend
		if userID != providerID && userID != clientID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			return
		}

		// Verify booking is in progress
		if status != "confirmed" && status != "paid" && status != "in_progress" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Booking must be active to extend"})
			return
		}

		// Create Stripe checkout session for extension
		successURL := req.SuccessURL
		cancelURL := req.CancelURL
		if successURL == "" {
			successURL = fmt.Sprintf("http://localhost:5174/booking/extend-success?booking_id=%d", req.BookingID)
		}
		if cancelURL == "" {
			cancelURL = "http://localhost:5174/booking/extend-cancel"
		}

		priceInCents := int64(req.Price * 100)

		params := &stripe.CheckoutSessionParams{
			Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency: stripe.String("thb"),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name:        stripe.String(fmt.Sprintf("Session Extension +%d minutes", req.AdditionalMinutes)),
							Description: stripe.String(fmt.Sprintf("Extend booking #%d by %d minutes", req.BookingID, req.AdditionalMinutes)),
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
				"payment_type":       "booking_extension",
				"booking_id":         fmt.Sprintf("%d", req.BookingID),
				"provider_id":        fmt.Sprintf("%d", providerID),
				"additional_minutes": fmt.Sprintf("%d", req.AdditionalMinutes),
			},
		}

		stripeSession, err := session.New(params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"checkout_url":       stripeSession.URL,
			"session_id":         stripeSession.ID,
			"booking_id":         req.BookingID,
			"additional_minutes": req.AdditionalMinutes,
			"amount":             req.Price,
		})
	}
}

// Helper: Handle Booking Extension Payment from Webhook
func handleBookingExtension(dbPool *pgxpool.Pool, ctx context.Context, checkoutSession stripe.CheckoutSession) error {
	bookingID := checkoutSession.Metadata["booking_id"]
	providerIDStr := checkoutSession.Metadata["provider_id"]
	additionalMinutesStr := checkoutSession.Metadata["additional_minutes"]

	if bookingID == "" || providerIDStr == "" || additionalMinutesStr == "" {
		return fmt.Errorf("missing metadata for booking extension")
	}

	additionalMinutes, _ := strconv.Atoi(additionalMinutesStr)
	totalAmount := float64(checkoutSession.AmountTotal) / 100
	stripeFee := totalAmount * 0.0275
	platformCommission := totalAmount * 0.1000
	providerEarnings := totalAmount * 0.8725

	// Update booking end_time
	_, err := dbPool.Exec(ctx, `
		UPDATE bookings 
		SET end_time = end_time + INTERVAL '1 minute' * $1,
		    total_price = total_price + $2,
		    updated_at = NOW()
		WHERE booking_id = $3
	`, additionalMinutes, totalAmount, bookingID)

	if err != nil {
		return fmt.Errorf("failed to update booking: %v", err)
	}

	// Update check-in expected_end_time
	dbPool.Exec(ctx, `
		UPDATE booking_check_ins 
		SET expected_end_time = expected_end_time + INTERVAL '1 minute' * $1,
		    updated_at = NOW()
		WHERE booking_id = $2 AND status = 'active'
	`, additionalMinutes, bookingID)

	// Create transaction record
	var transactionID int
	err = dbPool.QueryRow(ctx, `
		INSERT INTO transactions (
			user_id, type, amount, status, booking_id,
			stripe_fee, platform_commission, total_fee_percentage, net_amount
		)
		VALUES ($1, 'booking_extension', $2, 'completed', $3, $4, $5, 0.1275, $6)
		RETURNING transaction_id
	`, providerIDStr, totalAmount, bookingID, stripeFee, platformCommission, providerEarnings).Scan(&transactionID)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	// Add to provider's pending balance
	dbPool.Exec(ctx, `
		UPDATE wallets 
		SET pending_balance = pending_balance + $1,
		    total_earned = total_earned + $1,
		    updated_at = NOW()
		WHERE user_id = $2
	`, providerEarnings, providerIDStr)

	// Notify provider
	providerIDInt, _ := strconv.Atoi(providerIDStr)
	CreateNotification(providerIDInt, "booking_extended",
		fmt.Sprintf("Booking extended by %d minutes. Additional payment: ฿%.0f", additionalMinutes, providerEarnings),
		map[string]interface{}{
			"booking_id":         bookingID,
			"additional_minutes": additionalMinutes,
			"amount":             providerEarnings,
		})

	// WebSocket notification
	if wsManager != nil {
		wsManager.BroadcastToUser(providerIDInt, WebSocketMessage{
			Type: "booking_extended",
			Payload: map[string]interface{}{
				"booking_id":         bookingID,
				"additional_minutes": additionalMinutes,
				"amount":             providerEarnings,
			},
		})
	}

	fmt.Printf("✅ Booking extended: BookingID=%s, +%d minutes, Amount=฿%.2f\n",
		bookingID, additionalMinutes, totalAmount)

	return nil
}
