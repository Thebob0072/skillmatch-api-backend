# 💳 Payment System Guide - SkillMatch

## 📋 ภาพรวมระบบการชำระเงิน

SkillMatch มีระบบการชำระเงิน **2 ประเภท**:

### 1. 💎 Subscription Payment (ค่าสมาชิก)
- **วัตถุประสงค์**: อัปเกรด tier ของ user (Silver, Diamond, Premium)
- **Payment Mode**: Stripe Subscription
- **สถานะ**: ✅ Implemented

### 2. 📅 Booking Payment (จ่ายค่าบริการ)
- **วัตถุประสงค์**: จ่ายค่าจองบริการของ provider
- **Payment Mode**: Stripe One-time Payment
- **สถานะ**: ⚠️ Needs Implementation

---

## 💎 Subscription Payment Flow

### User Journey:
```
1. User คลิก "Upgrade to Silver/Diamond/Premium"
   ↓
2. Frontend เรียก POST /subscription/create-checkout
   Body: { "tier_id": 2 }  // 2=Silver, 3=Diamond, 4=Premium
   ↓
3. Backend สร้าง Stripe Checkout Session (subscription mode)
   - Line Item: price_xxx (Price ID from Stripe)
   - Success URL: /dashboard?payment=success
   - Cancel URL: /pricing?payment=cancelled
   - ClientReferenceID: userID
   ↓
4. Backend ส่ง { "checkout_url": "https://checkout.stripe.com/..." }
   ↓
5. Frontend redirect user ไป Stripe Checkout
   ↓
6. User กรอกบัตรเครดิต + ชำระเงิน
   ↓
7. Stripe ส่ง webhook event "checkout.session.completed"
   ↓
8. Backend รับ webhook → validate signature → อัปเดต tier
   UPDATE users SET tier_id = $1 WHERE user_id = $2
   ↓
9. User redirect กลับมา /dashboard?payment=success
   ↓
10. Frontend แสดง "Payment successful! Your tier has been upgraded."
```

### Implementation:

**1. Create Checkout Session:**
```go
// POST /subscription/create-checkout
func createCheckoutSessionHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, _ := c.Get("userID")
        
        var req struct {
            TierID int `json:"tier_id"`
        }
        c.ShouldBindJSON(&req)
        
        // Map tier_id to Stripe Price ID
        stripePriceID := tierIDToStripePrice[req.TierID]
        
        // Create Stripe Checkout Session
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
        
        session, _ := session.New(params)
        c.JSON(200, gin.H{"checkout_url": session.URL})
    }
}
```

**2. Handle Webhook:**
```go
// POST /payment/webhook
func paymentWebhookHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Verify webhook signature
        payload, _ := ioutil.ReadAll(c.Request.Body)
        event, _ := webhook.ConstructEvent(
            payload, 
            c.GetHeader("Stripe-Signature"), 
            os.Getenv("STRIPE_WEBHOOK_SECRET")
        )
        
        if event.Type == "checkout.session.completed" {
            var session stripe.CheckoutSession
            json.Unmarshal(event.Data.Raw, &session)
            
            // Get user_id from ClientReferenceID
            userID, _ := strconv.Atoi(session.ClientReferenceID)
            
            // Get Price ID from line items
            sessionWithItems, _ := session.Get(session.ID, &stripe.CheckoutSessionParams{
                Expand: []*string{stripe.String("line_items")},
            })
            purchasedPriceID := sessionWithItems.LineItems.Data[0].Price.ID
            
            // Map Price ID to tier_id
            newTierID := stripePriceToTierID[purchasedPriceID]
            
            // Update user tier
            dbPool.Exec(ctx, "UPDATE users SET tier_id = $1 WHERE user_id = $2", newTierID, userID)
        }
        
        c.JSON(200, gin.H{"status": "received"})
    }
}
```

---

## 📅 Booking Payment Flow (TO BE IMPLEMENTED)

### User Journey:
```
1. Client เลือก Provider + Package
   ↓
2. Client กรอกข้อมูล booking (วัน, เวลา, location)
   ↓
3. Frontend เรียก POST /bookings/create-with-payment
   Body: {
     "provider_id": 123,
     "package_id": 456,
     "booking_date": "2025-11-20",
     "start_time": "14:00",
     "location": "123 ถ.สุขุมวิท"
   }
   ↓
4. Backend:
   - สร้าง booking (status = "pending_payment")
   - สร้าง Stripe Checkout Session (payment mode)
   - Line Item: Package name + price
   - Metadata: booking_id, provider_id, client_id
   ↓
5. Backend ส่ง { "checkout_url": "...", "booking_id": xxx }
   ↓
6. Frontend redirect client ไป Stripe Checkout
   ↓
7. Client กรอกบัตรเครดิต + ชำระเงิน
   ↓
8. Stripe ส่ง webhook event "checkout.session.completed"
   ↓
9. Backend รับ webhook:
   - อัปเดต booking status = "paid"
   - คำนวณค่าธรรมเนียม 12.75%:
     * Stripe Fee: 2.75%
     * Platform Commission: 10%
   - สร้าง transaction (booking_payment)
   - เพิ่มเงินเข้า provider pending_balance (87.25%)
   - บันทึก GOD commission (10%)
   ↓
10. Client redirect กลับมา /bookings/:id?payment=success
    ↓
11. Frontend แสดง "Payment successful! Your booking is confirmed."
    ↓
12. ส่ง notification ไป Provider (WebSocket + Email)
```

### Implementation Plan:

**1. Create Booking with Payment Checkout:**
```go
// POST /bookings/create-with-payment
func createBookingWithPaymentHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientID, _ := c.Get("userID")
        
        var req struct {
            ProviderID   int     `json:"provider_id"`
            PackageID    int     `json:"package_id"`
            BookingDate  string  `json:"booking_date"`
            StartTime    string  `json:"start_time"`
            Location     *string `json:"location"`
        }
        c.ShouldBindJSON(&req)
        
        // 1. Get package details
        var packageName string
        var price float64
        var duration int
        dbPool.QueryRow(ctx, `
            SELECT package_name, price, duration 
            FROM service_packages 
            WHERE package_id = $1
        `, req.PackageID).Scan(&packageName, &price, &duration)
        
        // 2. Create booking (pending_payment)
        var bookingID int
        dbPool.QueryRow(ctx, `
            INSERT INTO bookings (
                client_id, provider_id, package_id, 
                booking_date, start_time, end_time,
                total_price, status
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending_payment')
            RETURNING booking_id
        `, clientID, req.ProviderID, req.PackageID, 
           req.BookingDate, req.StartTime, calculateEndTime(req.StartTime, duration),
           price).Scan(&bookingID)
        
        // 3. Create Stripe Checkout Session
        params := &stripe.CheckoutSessionParams{
            Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
            LineItems: []*stripe.CheckoutSessionLineItemParams{
                {
                    PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
                        Currency: stripe.String("thb"),
                        ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
                            Name: stripe.String(packageName),
                            Description: stripe.String(fmt.Sprintf("Booking #%d", bookingID)),
                        },
                        UnitAmount: stripe.Int64(int64(price * 100)), // Convert to satang
                    },
                    Quantity: stripe.Int64(1),
                },
            },
            SuccessURL: stripe.String(fmt.Sprintf(
                "http://localhost:5174/bookings/%d?payment=success", bookingID
            )),
            CancelURL: stripe.String(fmt.Sprintf(
                "http://localhost:5174/bookings/%d?payment=cancelled", bookingID
            )),
            ClientReferenceID: stripe.String(fmt.Sprintf("%d", clientID)),
            Metadata: map[string]string{
                "booking_id":  fmt.Sprintf("%d", bookingID),
                "provider_id": fmt.Sprintf("%d", req.ProviderID),
                "client_id":   fmt.Sprintf("%d", clientID),
                "type":        "booking_payment",
            },
        }
        
        session, _ := session.New(params)
        
        c.JSON(200, gin.H{
            "checkout_url": session.URL,
            "booking_id":   bookingID,
        })
    }
}
```

**2. Update Webhook Handler for Booking Payments:**
```go
func paymentWebhookHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... verify signature ...
        
        if event.Type == "checkout.session.completed" {
            var session stripe.CheckoutSession
            json.Unmarshal(event.Data.Raw, &session)
            
            // Check payment type
            paymentType := session.Metadata["type"]
            
            if paymentType == "booking_payment" {
                handleBookingPayment(dbPool, ctx, session)
            } else {
                handleSubscriptionPayment(dbPool, ctx, session)
            }
        }
        
        c.JSON(200, gin.H{"status": "received"})
    }
}

func handleBookingPayment(dbPool *pgxpool.Pool, ctx context.Context, session stripe.CheckoutSession) {
    bookingID, _ := strconv.Atoi(session.Metadata["booking_id"])
    providerID, _ := strconv.Atoi(session.Metadata["provider_id"])
    clientID, _ := strconv.Atoi(session.Metadata["client_id"])
    
    // Get booking amount
    var totalPrice float64
    dbPool.QueryRow(ctx, 
        "SELECT total_price FROM bookings WHERE booking_id = $1", 
        bookingID
    ).Scan(&totalPrice)
    
    // Calculate fees (12.75% total)
    stripeFee := totalPrice * 0.0275      // 2.75%
    platformFee := totalPrice * 0.1000    // 10%
    totalFee := totalPrice * 0.1275       // 12.75%
    netAmount := totalPrice * 0.8725      // 87.25% to provider
    
    // Begin transaction
    tx, _ := dbPool.Begin(ctx)
    defer tx.Rollback(ctx)
    
    // 1. Update booking status
    tx.Exec(ctx, `
        UPDATE bookings 
        SET status = 'paid', 
            payment_intent_id = $1, 
            updated_at = CURRENT_TIMESTAMP
        WHERE booking_id = $2
    `, session.PaymentIntent.ID, bookingID)
    
    // 2. Create transaction record (booking_payment)
    tx.Exec(ctx, `
        INSERT INTO transactions (
            user_id, related_user_id, type, status,
            amount, stripe_fee, platform_commission,
            commission_amount, net_amount, total_fee_percentage,
            booking_id, payment_intent_id, payment_method,
            description
        ) VALUES ($1, $2, 'booking_payment', 'completed',
                  $3, $4, $5, $6, $7, 0.1275,
                  $8, $9, 'stripe',
                  $10)
    `, clientID, providerID, totalPrice, stripeFee, platformFee,
       totalFee, netAmount, bookingID, session.PaymentIntent.ID,
       fmt.Sprintf("Booking payment #%d", bookingID))
    
    // 3. Update provider wallet (pending_balance)
    tx.Exec(ctx, `
        INSERT INTO wallets (user_id, pending_balance, total_earned)
        VALUES ($1, $2, $2)
        ON CONFLICT (user_id) DO UPDATE
        SET pending_balance = wallets.pending_balance + $2,
            total_earned = wallets.total_earned + $2,
            last_updated = CURRENT_TIMESTAMP
    `, providerID, netAmount)
    
    // 4. Create provider earning transaction
    tx.Exec(ctx, `
        INSERT INTO transactions (
            user_id, type, status,
            amount, net_amount, booking_id,
            description
        ) VALUES ($1, 'provider_earning', 'pending',
                  $2, $2, $3, $4)
    `, providerID, netAmount, bookingID,
       fmt.Sprintf("Earning from booking #%d", bookingID))
    
    // 5. Update GOD commission balance
    tx.Exec(ctx, `
        UPDATE god_commission_balance
        SET total_commission_collected = total_commission_collected + $1,
            current_balance = current_balance + $1,
            last_updated = CURRENT_TIMESTAMP
        WHERE god_user_id = 1
    `, platformFee)
    
    // 6. Create commission transaction record
    tx.Exec(ctx, `
        INSERT INTO commission_transactions (
            booking_id, booking_amount, commission_rate,
            commission_amount, provider_amount, provider_id,
            status
        ) VALUES ($1, $2, 0.1000, $3, $4, $5, 'collected')
    `, bookingID, totalPrice, platformFee, netAmount, providerID)
    
    tx.Commit(ctx)
    
    // 7. Send notifications (WebSocket + Email)
    // TODO: Implement notification sending
}
```

---

## 🔐 Security Considerations

### Webhook Signature Verification
```go
webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
event, err := webhook.ConstructEvent(
    payload, 
    c.GetHeader("Stripe-Signature"), 
    webhookSecret
)
if err != nil {
    c.JSON(400, gin.H{"error": "Invalid signature"})
    return
}
```

### Idempotency
- บันทึก `payment_intent_id` ใน bookings table
- ตรวจสอบว่า booking ถูก process แล้วหรือยัง
- ป้องกันการ process ซ้ำ

```go
var alreadyProcessed bool
dbPool.QueryRow(ctx, `
    SELECT EXISTS(
        SELECT 1 FROM bookings 
        WHERE booking_id = $1 AND payment_intent_id = $2
    )
`, bookingID, paymentIntentID).Scan(&alreadyProcessed)

if alreadyProcessed {
    return // Skip processing
}
```

---

## 📊 Payment Status Flow

### Subscription Payment:
```
(no database record) → Stripe Checkout → users.tier_id updated
```

### Booking Payment:
```
pending_payment → Stripe Checkout → paid → confirmed → completed
                                     ↓
                              Create transactions
                              Update wallets
                              Send notifications
```

---

## 🧪 Testing

### Test Stripe Cards:
- **Success**: `4242 4242 4242 4242`
- **Decline**: `4000 0000 0000 0002`
- **3D Secure**: `4000 0025 0000 3155`

### Testing Webhook Locally:
```bash
# Install Stripe CLI
brew install stripe/stripe-cli/stripe

# Login
stripe login

# Forward webhooks to local server
stripe listen --forward-to localhost:8080/payment/webhook

# Test webhook
stripe trigger checkout.session.completed
```

---

## 📝 API Endpoints Summary

### Current (Subscription):
- `POST /subscription/create-checkout` - สร้าง checkout session สำหรับอัปเกรด tier
- `POST /payment/webhook` - รับ webhook จาก Stripe

### To Be Implemented (Booking):
- `POST /bookings/create-with-payment` - สร้าง booking + checkout session
- Update `POST /payment/webhook` - เพิ่ม handler สำหรับ booking payments

---

**Last Updated:** November 14, 2025  
**Status:** 
- ✅ Subscription Payment: Implemented
- ⚠️ Booking Payment: Documentation Ready, Needs Implementation
