package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
Escrow Payment System - ระบบชำระเงินส่วนที่เหลืออย่างปลอดภัย

ปัญหาที่แก้:
1. Client กลัวจ่ายเต็มจำนวนแล้วไม่ได้รับบริการ
2. Provider กลัวให้บริการแล้วไม่ได้เงิน

วิธีการทำงาน:
1. Client จ่ายมัดจำ 10% ก่อน (deposit_paid = true)
2. เมื่อ Provider มาถึง → Client ยืนยันว่า "Provider มาถึงแล้ว"
3. ระบบล็อคเงินส่วนที่เหลือ (remaining_amount) ไว้ใน escrow
4. เมื่อบริการเสร็จ:
   - Provider กด "บริการเสร็จสิ้น" → status = completed
   - Client กด "ยืนยันรับบริการ" → ปลดล็อคเงินให้ Provider
5. ถ้ามีปัญหา → Client ร้องเรียน → Admin ตรวจสอบ → ตัดสินใจ

Flow:
pending → confirmed (deposit paid) → provider_arrived → in_progress → completed → funds_released
*/

// EscrowPayment represents the escrow transaction
type EscrowPayment struct {
	EscrowID            int        `json:"escrow_id"`
	BookingID           int        `json:"booking_id"`
	Amount              float64    `json:"amount"` // Remaining amount
	Status              string     `json:"status"` // locked, released, refunded, disputed
	LockedAt            time.Time  `json:"locked_at"`
	ReleasedAt          *time.Time `json:"released_at,omitempty"`
	ClientConfirmedAt   *time.Time `json:"client_confirmed_at,omitempty"`
	ProviderCompletedAt *time.Time `json:"provider_completed_at,omitempty"`
	DisputeReason       *string    `json:"dispute_reason,omitempty"`
	DisputedAt          *time.Time `json:"disputed_at,omitempty"`
	AdminDecision       *string    `json:"admin_decision,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

// POST /bookings/:id/provider-arrived
// Provider เรียก API นี้เมื่อมาถึงสถานที่แล้ว
func providerArrivedHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		providerID := c.GetInt("user_id")

		// ตรวจสอบว่าเป็น provider ของ booking นี้จริง
		var dbProviderID int
		var status, paymentStatus string
		var remainingAmount float64

		err := dbPool.QueryRow(ctx, `
			SELECT provider_id, status, payment_status, remaining_amount
			FROM bookings
			WHERE booking_id = $1
		`, bookingID).Scan(&dbProviderID, &status, &paymentStatus, &remainingAmount)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		if dbProviderID != providerID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not your booking"})
			return
		}

		// ตรวจสอบว่าจ่ายมัดจำแล้ว
		if paymentStatus != "deposit_paid" && paymentStatus != "paid" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":                  "Deposit must be paid first",
				"current_payment_status": paymentStatus,
			})
			return
		}

		// อัปเดตสถานะ
		_, err = dbPool.Exec(ctx, `
			UPDATE bookings
			SET status = 'provider_arrived',
			    provider_arrived_at = NOW(),
			    updated_at = NOW()
			WHERE booking_id = $1
		`, bookingID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
			return
		}

		// ส่ง notification ให้ client
		bookingIDInt, _ := strconv.Atoi(bookingID)
		sendNotification(dbPool, ctx, bookingIDInt, "provider_arrived",
			"Provider has arrived at the location. Please confirm their arrival to proceed.")

		c.JSON(http.StatusOK, gin.H{
			"message":    "Status updated to provider_arrived",
			"booking_id": bookingID,
			"status":     "provider_arrived",
			"next_step":  "Wait for client to confirm your arrival",
		})
	}
}

// POST /bookings/:id/confirm-arrival
// Client ยืนยันว่า Provider มาถึงจริง → ล็อคเงินส่วนที่เหลือไว้ใน escrow
func confirmProviderArrivalHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		clientID := c.GetInt("user_id")

		// ตรวจสอบว่าเป็น client ของ booking นี้จริง
		var dbClientID int
		var status string
		var remainingAmount float64
		var depositPaid bool

		err := dbPool.QueryRow(ctx, `
			SELECT client_id, status, remaining_amount, deposit_paid
			FROM bookings
			WHERE booking_id = $1
		`, bookingID).Scan(&dbClientID, &status, &remainingAmount, &depositPaid)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		if dbClientID != clientID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not your booking"})
			return
		}

		// ตรวจสอบว่า provider arrived แล้ว
		if status != "provider_arrived" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "Provider must arrive first",
				"current_status": status,
			})
			return
		}

		// ตรวจสอบว่าจ่ายมัดจำแล้ว
		if !depositPaid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Deposit must be paid first"})
			return
		}

		// ล็อคเงินส่วนที่เหลือใน escrow
		var escrowID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO escrow_payments (
				booking_id, amount, status, locked_at
			) VALUES ($1, $2, 'locked', NOW())
			RETURNING escrow_id
		`, bookingID, remainingAmount).Scan(&escrowID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lock escrow"})
			return
		}

		// อัปเดตสถานะ booking
		_, err = dbPool.Exec(ctx, `
			UPDATE bookings
			SET status = 'in_progress',
			    client_confirmed_arrival_at = NOW(),
			    escrow_locked = true,
			    updated_at = NOW()
			WHERE booking_id = $1
		`, bookingID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking"})
			return
		}

		// ส่ง notification ให้ provider
		bookingIDInt, _ := strconv.Atoi(bookingID)
		sendNotification(dbPool, ctx, bookingIDInt, "client_confirmed_arrival",
			"Client confirmed your arrival. The remaining payment is now locked in escrow. You can start the service.")

		c.JSON(http.StatusOK, gin.H{
			"message":       "Provider arrival confirmed. Escrow locked.",
			"escrow_id":     escrowID,
			"locked_amount": remainingAmount,
			"status":        "in_progress",
			"next_step":     "Provider can now start the service",
		})
	}
}

// POST /bookings/:id/provider-complete
// Provider กดว่าให้บริการเสร็จแล้ว (รอ client ยืนยัน)
func providerCompleteServiceHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		providerID := c.GetInt("user_id")

		var req struct {
			Notes string `json:"notes"`
		}
		c.ShouldBindJSON(&req)

		// ตรวจสอบว่าเป็น provider ของ booking นี้จริง
		var dbProviderID int
		var status string

		err := dbPool.QueryRow(ctx, `
			SELECT provider_id, status
			FROM bookings
			WHERE booking_id = $1
		`, bookingID).Scan(&dbProviderID, &status)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		if dbProviderID != providerID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not your booking"})
			return
		}

		if status != "in_progress" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "Service must be in progress",
				"current_status": status,
			})
			return
		}

		// อัปเดตสถานะ
		_, err = dbPool.Exec(ctx, `
			UPDATE bookings
			SET status = 'completed',
			    provider_completed_at = NOW(),
			    provider_completion_notes = $1,
			    updated_at = NOW()
			WHERE booking_id = $2
		`, req.Notes, bookingID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
			return
		}

		// อัปเดต escrow
		_, err = dbPool.Exec(ctx, `
			UPDATE escrow_payments
			SET provider_completed_at = NOW()
			WHERE booking_id = $1 AND status = 'locked'
		`, bookingID)

		// ส่ง notification ให้ client
		bookingIDInt, _ := strconv.Atoi(bookingID)
		sendNotification(dbPool, ctx, bookingIDInt, "service_completed",
			"Provider has completed the service. Please confirm to release the payment.")

		c.JSON(http.StatusOK, gin.H{
			"message":         "Service marked as completed",
			"status":          "completed",
			"next_step":       "Wait for client confirmation to release payment",
			"auto_release_in": "24 hours if no dispute",
		})
	}
}

// POST /bookings/:id/confirm-completion
// Client ยืนยันว่าได้รับบริการแล้ว → ปลดล็อคเงินให้ provider
func confirmServiceCompletionHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		clientID := c.GetInt("user_id")

		var req struct {
			Rating int    `json:"rating"` // 1-5
			Review string `json:"review"`
		}
		c.ShouldBindJSON(&req)

		// ตรวจสอบว่าเป็น client ของ booking นี้จริง
		var dbClientID, providerID int
		var status string
		var remainingAmount float64

		err := dbPool.QueryRow(ctx, `
			SELECT b.client_id, b.provider_id, b.status, b.remaining_amount
			FROM bookings b
			WHERE b.booking_id = $1
		`, bookingID).Scan(&dbClientID, &providerID, &status, &remainingAmount)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		if dbClientID != clientID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not your booking"})
			return
		}

		if status != "completed" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "Service must be completed first",
				"current_status": status,
			})
			return
		}

		// Release escrow → เพิ่มเงินเข้า wallet ของ provider
		tx, err := dbPool.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
			return
		}
		defer tx.Rollback(ctx)

		// 1. อัปเดต escrow status
		_, err = tx.Exec(ctx, `
			UPDATE escrow_payments
			SET status = 'released',
			    client_confirmed_at = NOW(),
			    released_at = NOW()
			WHERE booking_id = $1 AND status = 'locked'
		`, bookingID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to release escrow"})
			return
		}

		// 2. คำนวณ platform fee (แนะนำ 8% ของยอดรวม)
		const platformFeePercentage = 0.08 // 8% platform commission
		platformFee := remainingAmount * platformFeePercentage
		providerReceives := remainingAmount - platformFee

		// 3. เพิ่มเงินเข้า wallet ของ provider (หักค่าคอมมิชชั่นแล้ว)
		_, err = tx.Exec(ctx, `
			UPDATE users
			SET wallet_balance = wallet_balance + $1,
			    updated_at = NOW()
			WHERE user_id = $2
		`, providerReceives, providerID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wallet"})
			return
		}

		// 4. บันทึก platform fee เข้าระบบ
		_, err = tx.Exec(ctx, `
			INSERT INTO wallet_transactions (
				user_id, transaction_type, amount,
				description, reference_id, reference_type
			) VALUES (0, 'platform_commission', $1, $2, $3, 'booking')
		`, platformFee,
			fmt.Sprintf("Platform commission (8%%) from booking #%s", bookingID),
			bookingID)

		// 5. บันทึก transaction ให้ provider
		_, err = tx.Exec(ctx, `
			INSERT INTO wallet_transactions (
				user_id, transaction_type, amount,
				description, reference_id, reference_type
			) VALUES ($1, 'escrow_release', $2, $3, $4, 'booking')
		`, providerID, providerReceives,
			fmt.Sprintf("Remaining payment for booking #%s (after 8%% platform fee)", bookingID),
			bookingID)

		// 6. อัปเดต booking status
		_, err = tx.Exec(ctx, `
			UPDATE bookings
			SET status = 'funds_released',
			    payment_status = 'fully_paid',
			    client_confirmed_at = NOW(),
			    updated_at = NOW()
			WHERE booking_id = $1
		`, bookingID)

		// 7. บันทึก review (ถ้ามี)
		if req.Rating > 0 {
			_, err = tx.Exec(ctx, `
				INSERT INTO reviews (
					booking_id, provider_id, client_id,
					rating, review_text
				) VALUES ($1, $2, $3, $4, $5)
			`, bookingID, providerID, clientID, req.Rating, req.Review)
		}

		if err := tx.Commit(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		// ส่ง notification ให้ provider
		bookingIDInt, _ := strconv.Atoi(bookingID)
		sendNotification(dbPool, ctx, bookingIDInt, "payment_released",
			fmt.Sprintf("Client confirmed service completion. ฿%.2f has been added to your wallet (after platform fee).", providerReceives))

		c.JSON(http.StatusOK, gin.H{
			"message":           "Service confirmed. Payment released to provider.",
			"amount_released":   providerReceives,
			"platform_fee":      platformFee,
			"platform_fee_rate": "8%",
			"total_amount":      remainingAmount,
			"status":            "funds_released",
		})
	}
}

// POST /bookings/:id/dispute
// Client ร้องเรียนว่ามีปัญหากับการให้บริการ
func disputeBookingHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		clientID := c.GetInt("user_id")

		var req struct {
			Reason      string   `json:"reason" binding:"required"`
			Description string   `json:"description" binding:"required"`
			Evidence    []string `json:"evidence"` // URLs to images/videos
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ตรวจสอบว่าเป็น client ของ booking นี้จริง
		var dbClientID int
		var status string

		err := dbPool.QueryRow(ctx, `
			SELECT client_id, status
			FROM bookings
			WHERE booking_id = $1
		`, bookingID).Scan(&dbClientID, &status)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		if dbClientID != clientID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not your booking"})
			return
		}

		// อัปเดต escrow status → disputed
		_, err = dbPool.Exec(ctx, `
			UPDATE escrow_payments
			SET status = 'disputed',
			    dispute_reason = $1,
			    disputed_at = NOW()
			WHERE booking_id = $2 AND status = 'locked'
		`, req.Reason, bookingID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dispute"})
			return
		}

		// อัปเดต booking
		_, err = dbPool.Exec(ctx, `
			UPDATE bookings
			SET status = 'disputed',
			    dispute_reason = $1,
			    dispute_description = $2,
			    disputed_at = NOW(),
			    updated_at = NOW()
			WHERE booking_id = $3
		`, req.Reason, req.Description, bookingID)

		// ส่ง notification ให้ admin และ provider
		bookingIDInt, _ := strconv.Atoi(bookingID)
		sendNotification(dbPool, ctx, bookingIDInt, "dispute_created",
			"A dispute has been raised. Admin will review and make a decision.")

		c.JSON(http.StatusOK, gin.H{
			"message":    "Dispute created successfully",
			"booking_id": bookingID,
			"status":     "disputed",
			"next_step":  "Admin will review within 24-48 hours",
		})
	}
}

// POST /admin/bookings/:id/resolve-dispute
// Admin ตัดสินข้อพิพาท
func adminResolveDisputeHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		adminID := c.GetInt("user_id")

		var req struct {
			Decision         string  `json:"decision" binding:"required"` // refund_client, pay_provider, split
			RefundPercentage float64 `json:"refund_percentage"`           // 0-100
			Notes            string  `json:"notes" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ดึงข้อมูล booking
		var clientID, providerID int
		var remainingAmount float64

		err := dbPool.QueryRow(ctx, `
			SELECT client_id, provider_id, remaining_amount
			FROM bookings
			WHERE booking_id = $1 AND status = 'disputed'
		`, bookingID).Scan(&clientID, &providerID, &remainingAmount)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Disputed booking not found"})
			return
		}

		tx, err := dbPool.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
			return
		}
		defer tx.Rollback(ctx)

		var refundAmount, providerAmount float64

		switch req.Decision {
		case "refund_client":
			// คืนเงินให้ client เต็มจำนวน
			refundAmount = remainingAmount
			providerAmount = 0

		case "pay_provider":
			// จ่ายให้ provider เต็มจำนวน
			refundAmount = 0
			providerAmount = remainingAmount

		case "split":
			// แบ่งตาม percentage
			refundAmount = remainingAmount * (req.RefundPercentage / 100)
			providerAmount = remainingAmount - refundAmount

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid decision"})
			return
		}

		// 1. คืนเงินให้ client (ถ้ามี)
		if refundAmount > 0 {
			_, err = tx.Exec(ctx, `
				UPDATE users
				SET wallet_balance = wallet_balance + $1
				WHERE user_id = $2
			`, refundAmount, clientID)

			_, err = tx.Exec(ctx, `
				INSERT INTO wallet_transactions (
					user_id, transaction_type, amount,
					description, reference_id, reference_type
				) VALUES ($1, 'refund', $2, $3, $4, 'booking')
			`, clientID, refundAmount,
				fmt.Sprintf("Refund for disputed booking #%s", bookingID),
				bookingID)
		}

		// 2. จ่ายเงินให้ provider (ถ้ามี)
		if providerAmount > 0 {
			_, err = tx.Exec(ctx, `
				UPDATE users
				SET wallet_balance = wallet_balance + $1
				WHERE user_id = $2
			`, providerAmount, providerID)

			_, err = tx.Exec(ctx, `
				INSERT INTO wallet_transactions (
					user_id, transaction_type, amount,
					description, reference_id, reference_type
				) VALUES ($1, 'escrow_release', $2, $3, $4, 'booking')
			`, providerID, providerAmount,
				fmt.Sprintf("Partial payment for disputed booking #%s", bookingID),
				bookingID)
		}

		// 3. อัปเดต escrow
		escrowStatus := "refunded"
		if req.Decision == "pay_provider" {
			escrowStatus = "released"
		} else if req.Decision == "split" {
			escrowStatus = "partially_refunded"
		}

		_, err = tx.Exec(ctx, `
			UPDATE escrow_payments
			SET status = $1,
			    admin_decision = $2,
			    released_at = NOW()
			WHERE booking_id = $3
		`, escrowStatus, req.Decision, bookingID)

		// 4. อัปเดต booking
		_, err = tx.Exec(ctx, `
			UPDATE bookings
			SET status = 'dispute_resolved',
			    admin_decision = $1,
			    admin_decision_notes = $2,
			    resolved_by_admin_id = $3,
			    resolved_at = NOW(),
			    updated_at = NOW()
			WHERE booking_id = $4
		`, req.Decision, req.Notes, adminID, bookingID)

		if err := tx.Commit(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve dispute"})
			return
		}

		// ส่ง notification
		bookingIDInt, _ := strconv.Atoi(bookingID)
		sendNotification(dbPool, ctx, bookingIDInt, "dispute_resolved",
			fmt.Sprintf("Dispute resolved. Decision: %s", req.Decision))

		c.JSON(http.StatusOK, gin.H{
			"message":         "Dispute resolved successfully",
			"decision":        req.Decision,
			"refund_amount":   refundAmount,
			"provider_amount": providerAmount,
		})
	}
}

// Helper function
func sendNotification(dbPool *pgxpool.Pool, ctx context.Context, bookingID int, notifType, message string) {
	// Implementation for sending notifications
	// This would integrate with your notification system
}
