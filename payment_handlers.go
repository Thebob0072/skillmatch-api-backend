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
