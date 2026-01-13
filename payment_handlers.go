package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/webhook"
)

// (!!! 1. กรอก PRICE ID (price_...) ที่คุณหามาได้ที่นี่ !!!)
// (TierID มาจาก Database: 2=Silver, 3=Diamond, 4=Premium)
// สำหรับ Subscription (ลูกค้าทั่วไป)
var stripePriceToTierID = map[string]int{
	// "price_..." : TierID
	"price_1SSWjxIHzi9KfpzCjomHhTuq": 2,
	"price_1SSWkQIHzi9KfpzCP0Fp6LMv": 3,
	"price_1SSWklIHzi9KfpzCQhYoguEx": 4,
}

// (!!! 2. กรอก PRICE ID (price_...) ที่นี่อีกครั้ง !!!)
var tierIDToStripePrice = map[int]string{
	// TierID : "price_..."
	2: "price_1SSWjxIHzi9KfpzCjomHhTuq",
	3: "price_1SSWkQIHzi9KfpzCP0Fp6LMv",
	4: "price_1SSWklIHzi9KfpzCQhYoguEx",
}

// (!!! 3. กรอก PRICE ID สำหรับ Provider Tier Upgrade !!!)
// (ProviderTierID มาจาก Database: 2=Silver, 3=Diamond, 4=Premium)
var stripeProviderPriceToTierID = map[string]int{
	// "price_..." : ProviderTierID (ใช้ price เดียวกันหรือต่างก็ได้)
	"price_1SSWjxIHzi9KfpzCjomHhTuq": 2, // Silver Provider
	"price_1SSWkQIHzi9KfpzCP0Fp6LMv": 3, // Diamond Provider
	"price_1SSWklIHzi9KfpzCQhYoguEx": 4, // Premium Provider
}

var providerTierIDToStripePrice = map[int]string{
	// ProviderTierID : "price_..."
	2: "price_1SSWjxIHzi9KfpzCjomHhTuq",
	3: "price_1SSWkQIHzi9KfpzCP0Fp6LMv",
	4: "price_1SSWklIHzi9KfpzCQhYoguEx",
}

func setupStripe() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

// --- Handler: POST /subscription/create-checkout ---
func createCheckoutSessionHandler(_ *pgxpool.Pool, _ context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var requestBody struct {
			TierID int `json:"tier_id"`
		}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// (ใช้ Price ID จาก map)
		stripePriceID, ok := tierIDToStripePrice[requestBody.TierID]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or non-purchasable Tier ID"})
			return
		}

		params := &stripe.CheckoutSessionParams{
			Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					Price:    stripe.String(stripePriceID),
					Quantity: stripe.Int64(1),
				},
			},
			SuccessURL:        stripe.String("http://localhost:5174/dashboard?payment=success"),
			CancelURL:         stripe.String("http://localhost:5174/pricing?payment=cancelled"),
			ClientReferenceID: stripe.String(fmt.Sprintf("%d", userID)),
		}

		s, err := session.New(params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"checkout_url": s.URL})
	}
}

// --- Handler: POST /provider/create-upgrade-checkout ---
func createProviderUpgradeCheckoutHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var requestBody struct {
			RequestID int `json:"request_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. request_id is required"})
			return
		}

		// ตรวจสอบว่าคำขอได้รับการอนุมัติแล้ว
		var requestedTierID int
		var status, paymentStatus string
		err := dbPool.QueryRow(ctx, `
			SELECT requested_tier_id, status, payment_status
			FROM provider_tier_upgrade_requests
			WHERE request_id = $1 AND user_id = $2
		`, requestBody.RequestID, userID).Scan(&requestedTierID, &status, &paymentStatus)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Upgrade request not found"})
			return
		}

		if status != "approved" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Upgrade request not approved yet"})
			return
		}

		if paymentStatus == "paid" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "This request has already been paid"})
			return
		}

		// ใช้ Price ID สำหรับ Provider Tier จาก map
		stripePriceID, ok := providerTierIDToStripePrice[requestedTierID]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or non-purchasable Provider Tier ID"})
			return
		}

		params := &stripe.CheckoutSessionParams{
			Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					Price:    stripe.String(stripePriceID),
					Quantity: stripe.Int64(1),
				},
			},
			SuccessURL:        stripe.String("http://localhost:5174/provider/dashboard?upgrade=success"),
			CancelURL:         stripe.String("http://localhost:5174/provider/tier?upgrade=cancelled"),
			ClientReferenceID: stripe.String(fmt.Sprintf("%d", userID)),
			Metadata: map[string]string{
				"payment_type": "provider_tier_upgrade",
				"request_id":   fmt.Sprintf("%d", requestBody.RequestID),
			},
		}

		s, err := session.New(params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session", "details": err.Error()})
			return
		}

		// บันทึก Stripe Session ID
		_, _ = dbPool.Exec(ctx, `
			UPDATE provider_tier_upgrade_requests
			SET stripe_subscription_id = $1
			WHERE request_id = $2
		`, s.ID, requestBody.RequestID)

		c.JSON(http.StatusOK, gin.H{"checkout_url": s.URL})
	}
}

// --- Handler: POST /payment/webhook ---
func paymentWebhookHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		const MaxBodyBytes = int64(65536)
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

		payload, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to read request body"})
			return
		}

		webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
		event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), webhookSecret)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook signature verification failed"})
			return
		}

		if event.Type == "checkout.session.completed" {
			var checkoutSession stripe.CheckoutSession
			err := json.Unmarshal(event.Data.Raw, &checkoutSession)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse webhook JSON"})
				return
			}

			// ตรวจสอบว่าเป็น payment type ไหน (subscription หรือ booking)
			paymentType := checkoutSession.Metadata["payment_type"]

			if paymentType == "booking" {
				// --- Handle Booking Payment ---
				err := handleBookingPayment(dbPool, ctx, checkoutSession)
				if err != nil {
					fmt.Printf("❌ Error processing booking payment: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process booking payment", "details": err.Error()})
					return
				}
				fmt.Println("✅ Booking payment webhook processed successfully")
			} else if paymentType == "booking_extension" {
				// --- Handle Booking Extension Payment ---
				err := handleBookingExtension(dbPool, ctx, checkoutSession)
				if err != nil {
					fmt.Printf("❌ Error processing booking extension: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process booking extension", "details": err.Error()})
					return
				}
				fmt.Println("✅ Booking extension webhook processed successfully")
			} else if paymentType == "provider_tier_upgrade" {
				// --- Handle Provider Tier Upgrade Payment ---
				err := handleProviderTierUpgrade(dbPool, ctx, checkoutSession)
				if err != nil {
					fmt.Printf("❌ Error processing provider tier upgrade: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process provider tier upgrade", "details": err.Error()})
					return
				}
				fmt.Println("✅ Provider tier upgrade webhook processed successfully")
			} else {
				// --- Handle Subscription Payment (Original Logic) ---
				userID, err := strconv.Atoi(checkoutSession.ClientReferenceID)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ClientReferenceID"})
					return
				}

				// (ดึงข้อมูล LineItems เพื่อหา Price ID)
				sessionWithLineItems, err := session.Get(checkoutSession.ID, &stripe.CheckoutSessionParams{
					Expand: []*string{stripe.String("line_items")},
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to expand line items"})
					return
				}

				if len(sessionWithLineItems.LineItems.Data) == 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "No line items in session"})
					return
				}

				// (Price ID ที่ซื้อ)
				purchasedPriceID := sessionWithLineItems.LineItems.Data[0].Price.ID

				// (หา TierID จาก map)
				newTierID, ok := stripePriceToTierID[purchasedPriceID]
				if !ok {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Unrecognized Price ID from Stripe: " + purchasedPriceID})
					return
				}

				// (อัปเดต Tier ใน DB)
				_, err = dbPool.Exec(ctx, "UPDATE users SET tier_id = $1 WHERE user_id = $2", newTierID, userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user tier"})
					return
				}

				fmt.Printf("✅ Subscription payment successful. Upgraded UserID %d to TierID %d\n", userID, newTierID)
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "received"})
	}
}

// handleProviderTierUpgrade - ฟังก์ชันช่วยในการอัพเกรด Provider Tier
func handleProviderTierUpgrade(dbPool *pgxpool.Pool, ctx context.Context, checkoutSession stripe.CheckoutSession) error {
	userID, err := strconv.Atoi(checkoutSession.ClientReferenceID)
	if err != nil {
		return fmt.Errorf("invalid ClientReferenceID: %v", err)
	}

	// ดึง request_id จาก metadata
	requestIDStr := checkoutSession.Metadata["request_id"]
	if requestIDStr == "" {
		return fmt.Errorf("missing request_id in metadata")
	}

	requestID, err := strconv.Atoi(requestIDStr)
	if err != nil {
		return fmt.Errorf("invalid request_id: %v", err)
	}

	// ตรวจสอบว่าคำขอได้รับการอนุมัติแล้ว
	var newProviderTierID int
	var status string
	err = dbPool.QueryRow(ctx, `
		SELECT requested_tier_id, status
		FROM provider_tier_upgrade_requests
		WHERE request_id = $1 AND user_id = $2
	`, requestID, userID).Scan(&newProviderTierID, &status)

	if err != nil {
		return fmt.Errorf("upgrade request not found: %v", err)
	}

	if status != "approved" {
		return fmt.Errorf("upgrade request not approved")
	}

	// ดึงข้อมูล LineItems เพื่อหา Price ID และยอดเงิน
	sessionWithLineItems, err := session.Get(checkoutSession.ID, &stripe.CheckoutSessionParams{
		Expand: []*string{stripe.String("line_items")},
	})
	if err != nil {
		return fmt.Errorf("failed to expand line items: %v", err)
	}

	if len(sessionWithLineItems.LineItems.Data) == 0 {
		return fmt.Errorf("no line items in session")
	}

	amountPaid := float64(sessionWithLineItems.LineItems.Data[0].AmountTotal) / 100.0

	// ดึงข้อมูล tier เก่าเพื่อบันทึก
	var oldProviderTierID int
	err = dbPool.QueryRow(ctx, "SELECT provider_level_id FROM users WHERE user_id = $1", userID).Scan(&oldProviderTierID)
	if err != nil {
		return fmt.Errorf("failed to get old provider tier: %v", err)
	}

	// อัปเดต Provider Tier ใน DB
	_, err = dbPool.Exec(ctx, `
		UPDATE users 
		SET provider_level_id = $1 
		WHERE user_id = $2
	`, newProviderTierID, userID)
	if err != nil {
		return fmt.Errorf("failed to update provider tier: %v", err)
	}

	// อัปเดตสถานะคำขอเป็น paid
	_, err = dbPool.Exec(ctx, `
		UPDATE provider_tier_upgrade_requests
		SET payment_status = 'paid',
			stripe_subscription_id = $1
		WHERE request_id = $2
	`, checkoutSession.Subscription.ID, requestID)
	if err != nil {
		fmt.Printf("⚠️ Warning: Failed to update request payment status: %v\n", err)
	}

	// บันทึกประวัติการเปลี่ยน Tier
	_, err = dbPool.Exec(ctx, `
		INSERT INTO provider_tier_history (
			provider_id, old_tier_id, new_tier_id, change_type, reason
		) VALUES ($1, $2, $3, 'upgrade', $4)
	`, userID, oldProviderTierID, newProviderTierID,
		fmt.Sprintf("Paid subscription upgrade via Stripe - Amount: %.2f THB/month - Request ID: %d - Subscription ID: %s",
			amountPaid, requestID, checkoutSession.Subscription.ID))
	if err != nil {
		fmt.Printf("⚠️ Warning: Failed to record tier history: %v\n", err)
	}

	// บันทึกธุรกรรมทางการเงิน
	_, err = dbPool.Exec(ctx, `
		INSERT INTO financial_transactions (
			user_id, type, amount, status, payment_method, description
		) VALUES ($1, 'subscription_fee', $2, 'completed', 'stripe', $3)
	`, userID, amountPaid,
		fmt.Sprintf("Provider Tier Upgrade to Tier %d (Request #%d)", newProviderTierID, requestID))
	if err != nil {
		fmt.Printf("⚠️ Warning: Failed to record financial transaction: %v\n", err)
	}

	fmt.Printf("✅ Provider tier upgrade successful. UserID %d upgraded from Tier %d to Tier %d (Amount: %.2f THB, Request #%d)\n",
		userID, oldProviderTierID, newProviderTierID, amountPaid, requestID)

	return nil
}
