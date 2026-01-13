package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================================
// Provider Tier Auto-Assignment Algorithm
// ============================================================================
// Tier Points Calculation:
// - Rating: (average_rating * 20) = 0-100 points
// - Completed Bookings: (completed_bookings * 5) = max 250 points (50 bookings)
// - Total Reviews: (total_reviews * 3) = max 150 points (50 reviews)
// - Response Rate: (response_rate * 0.5) = max 50 points (100%)
// - Acceptance Rate: (acceptance_rate * 0.5) = max 50 points (100%)
// - Total: max 600 points
//
// Tier Assignment:
// - General (tier 1): 0-99 points
// - Silver (tier 2): 100-249 points
// - Diamond (tier 3): 250-399 points
// - Premium (tier 4): 400+ points
// ============================================================================

// GET /provider/available-tiers - ดู Tiers ทั้งหมดที่สามารถอัพเกรดได้
func getAvailableTiersHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := dbPool.Query(ctx, `
			SELECT tier_id, name, access_level, price_monthly
			FROM tiers
			WHERE tier_id BETWEEN 2 AND 4  -- Silver, Diamond, Premium only
			ORDER BY tier_id ASC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tiers"})
			return
		}
		defer rows.Close()

		type TierInfo struct {
			TierID       int     `json:"tier_id"`
			Name         string  `json:"name"`
			AccessLevel  int     `json:"access_level"`
			PriceMonthly float64 `json:"price_monthly"`
		}

		tiers := []TierInfo{}
		for rows.Next() {
			var tier TierInfo
			if err := rows.Scan(&tier.TierID, &tier.Name, &tier.AccessLevel, &tier.PriceMonthly); err != nil {
				continue
			}
			tiers = append(tiers, tier)
		}

		c.JSON(http.StatusOK, gin.H{"tiers": tiers})
	}
}

// POST /provider/request-upgrade - ส่งคำขออัพเกรด Tier (ต้องรอแอดมินอนุมัติ)
func requestProviderTierUpgradeHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		var req struct {
			TierID int `json:"tier_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// ตรวจสอบว่าเป็น provider และยืนยันแล้ว
		var isProvider bool
		var verificationStatus string
		var currentTierID int
		err := dbPool.QueryRow(ctx, `
			SELECT is_provider, provider_verification_status, provider_level_id
			FROM users
			WHERE user_id = $1
		`, userID).Scan(&isProvider, &verificationStatus, &currentTierID)

		if err != nil || !isProvider {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not a provider"})
			return
		}

		if verificationStatus != "approved" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Provider not approved"})
			return
		}

		// ตรวจสอบว่า tier ที่ต้องการสูงกว่า tier ปัจจุบัน
		if req.TierID <= currentTierID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Can only upgrade to higher tier"})
			return
		}

		// ตรวจสอบว่า tier ที่ต้องการมีอยู่จริง (2-4 เท่านั้น)
		if req.TierID < 2 || req.TierID > 4 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tier for upgrade"})
			return
		}

		// ตรวจสอบว่ามีคำขออยู่แล้วหรือไม่
		var existingRequestID int
		err = dbPool.QueryRow(ctx, `
			SELECT request_id
			FROM provider_tier_upgrade_requests
			WHERE user_id = $1 AND status = 'pending'
			LIMIT 1
		`, userID).Scan(&existingRequestID)

		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "You already have a pending upgrade request"})
			return
		}

		// ดึงข้อมูล tier
		var tierName string
		var priceMonthly float64
		err = dbPool.QueryRow(ctx, `
			SELECT name, price_monthly
			FROM tiers
			WHERE tier_id = $1
		`, req.TierID).Scan(&tierName, &priceMonthly)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tier not found"})
			return
		}

		// สร้างคำขออัพเกรด
		var requestID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO provider_tier_upgrade_requests (
				user_id, current_tier_id, requested_tier_id, status, payment_status
			) VALUES ($1, $2, $3, 'pending', 'unpaid')
			RETURNING request_id
		`, userID, currentTierID, req.TierID).Scan(&requestID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upgrade request"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"request_id":    requestID,
			"tier_id":       req.TierID,
			"tier_name":     tierName,
			"price_monthly": priceMonthly,
			"status":        "pending",
			"message":       "Upgrade request submitted. Waiting for admin approval.",
		})
	}
}

// GET /provider/my-upgrade-requests - ดูคำขออัพเกรดของตัวเอง
func getMyUpgradeRequestsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		rows, err := dbPool.Query(ctx, `
			SELECT 
				r.request_id,
				r.current_tier_id,
				ct.name as current_tier_name,
				r.requested_tier_id,
				rt.name as requested_tier_name,
				rt.price_monthly,
				r.status,
				r.payment_status,
				r.requested_at,
				r.reviewed_at,
				r.rejection_reason
			FROM provider_tier_upgrade_requests r
			LEFT JOIN tiers ct ON r.current_tier_id = ct.tier_id
			LEFT JOIN tiers rt ON r.requested_tier_id = rt.tier_id
			WHERE r.user_id = $1
			ORDER BY r.requested_at DESC
		`, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
			return
		}
		defer rows.Close()

		type UpgradeRequest struct {
			RequestID         int     `json:"request_id"`
			CurrentTierID     *int    `json:"current_tier_id"`
			CurrentTierName   *string `json:"current_tier_name"`
			RequestedTierID   int     `json:"requested_tier_id"`
			RequestedTierName string  `json:"requested_tier_name"`
			PriceMonthly      float64 `json:"price_monthly"`
			Status            string  `json:"status"`
			PaymentStatus     string  `json:"payment_status"`
			RequestedAt       string  `json:"requested_at"`
			ReviewedAt        *string `json:"reviewed_at"`
			RejectionReason   *string `json:"rejection_reason"`
		}

		requests := []UpgradeRequest{}
		for rows.Next() {
			var req UpgradeRequest
			if err := rows.Scan(
				&req.RequestID, &req.CurrentTierID, &req.CurrentTierName,
				&req.RequestedTierID, &req.RequestedTierName, &req.PriceMonthly,
				&req.Status, &req.PaymentStatus, &req.RequestedAt,
				&req.ReviewedAt, &req.RejectionReason,
			); err != nil {
				continue
			}
			requests = append(requests, req)
		}

		c.JSON(http.StatusOK, gin.H{"requests": requests})
	}
}

// GET /provider/my-tier - ดู Tier ปัจจุบันของตัวเอง
func getMyProviderTierHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		var result struct {
			CurrentTierID     int     `json:"current_tier_id"`
			CurrentTierName   string  `json:"current_tier_name"`
			TierPoints        int     `json:"tier_points"`
			AverageRating     float64 `json:"average_rating"`
			TotalReviews      int     `json:"total_reviews"`
			CompletedBookings int     `json:"completed_bookings"`
			ResponseRate      float64 `json:"response_rate"`
			AcceptanceRate    float64 `json:"acceptance_rate"`
			NextTierID        *int    `json:"next_tier_id"`
			NextTierName      *string `json:"next_tier_name"`
			PointsToNextTier  *int    `json:"points_to_next_tier"`
		}

		err := dbPool.QueryRow(ctx, `
			SELECT 
				u.provider_level_id,
				t.name as tier_name,
				ps.tier_points,
				ps.average_rating,
				ps.total_reviews,
				ps.completed_bookings,
				ps.response_rate,
				ps.acceptance_rate
			FROM users u
			JOIN tiers t ON u.provider_level_id = t.tier_id
			LEFT JOIN provider_stats ps ON u.user_id = ps.user_id
			WHERE u.user_id = $1 AND u.is_provider = true
		`, userID).Scan(
			&result.CurrentTierID, &result.CurrentTierName,
			&result.TierPoints, &result.AverageRating, &result.TotalReviews,
			&result.CompletedBookings, &result.ResponseRate, &result.AcceptanceRate,
		)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
			return
		}

		// คำนวณ tier ถัดไป
		var nextTierID int
		var nextTierName string
		var minPoints int
		err = dbPool.QueryRow(ctx, `
			SELECT tier_id, name
			FROM tiers
			WHERE tier_id > $1
			  AND tier_id < 5 -- ไม่รวม GOD tier
			ORDER BY tier_id ASC
			LIMIT 1
		`, result.CurrentTierID).Scan(&nextTierID, &nextTierName)

		if err == nil {
			result.NextTierID = &nextTierID
			result.NextTierName = &nextTierName

			// คำนวณคะแนนขั้นต่ำสำหรับ tier ถัดไป
			switch nextTierID {
			case 2: // Silver
				minPoints = 100
			case 3: // Diamond
				minPoints = 250
			case 4: // Premium
				minPoints = 400
			}
			pointsNeeded := minPoints - result.TierPoints
			if pointsNeeded > 0 {
				result.PointsToNextTier = &pointsNeeded
			}
		}

		c.JSON(http.StatusOK, result)
	}
}

// GET /admin/upgrade-requests - ดูคำขออัพเกรดทั้งหมด (Admin only)
func adminGetUpgradeRequestsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.DefaultQuery("status", "pending") // pending, approved, rejected, all

		query := `
			SELECT 
				r.request_id,
				r.user_id,
				u.username,
				u.email,
				r.current_tier_id,
				ct.name as current_tier_name,
				r.requested_tier_id,
				rt.name as requested_tier_name,
				rt.price_monthly,
				r.status,
				r.payment_status,
				r.requested_at,
				r.reviewed_at,
				r.reviewed_by,
				r.admin_notes,
				r.rejection_reason
			FROM provider_tier_upgrade_requests r
			LEFT JOIN users u ON r.user_id = u.user_id
			LEFT JOIN tiers ct ON r.current_tier_id = ct.tier_id
			LEFT JOIN tiers rt ON r.requested_tier_id = rt.tier_id
		`

		var rows pgx.Rows
		var err error

		if status == "all" {
			query += ` ORDER BY r.requested_at DESC`
			rows, err = dbPool.Query(ctx, query)
		} else {
			query += ` WHERE r.status = $1 ORDER BY r.requested_at DESC`
			rows, err = dbPool.Query(ctx, query, status)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
			return
		}
		defer rows.Close()

		type UpgradeRequest struct {
			RequestID         int     `json:"request_id"`
			UserID            int     `json:"user_id"`
			Username          string  `json:"username"`
			Email             string  `json:"email"`
			CurrentTierID     *int    `json:"current_tier_id"`
			CurrentTierName   *string `json:"current_tier_name"`
			RequestedTierID   int     `json:"requested_tier_id"`
			RequestedTierName string  `json:"requested_tier_name"`
			PriceMonthly      float64 `json:"price_monthly"`
			Status            string  `json:"status"`
			PaymentStatus     string  `json:"payment_status"`
			RequestedAt       string  `json:"requested_at"`
			ReviewedAt        *string `json:"reviewed_at"`
			ReviewedBy        *int    `json:"reviewed_by"`
			AdminNotes        *string `json:"admin_notes"`
			RejectionReason   *string `json:"rejection_reason"`
		}

		requests := []UpgradeRequest{}
		for rows.Next() {
			var req UpgradeRequest
			if err := rows.Scan(
				&req.RequestID, &req.UserID, &req.Username, &req.Email,
				&req.CurrentTierID, &req.CurrentTierName,
				&req.RequestedTierID, &req.RequestedTierName, &req.PriceMonthly,
				&req.Status, &req.PaymentStatus, &req.RequestedAt,
				&req.ReviewedAt, &req.ReviewedBy, &req.AdminNotes, &req.RejectionReason,
			); err != nil {
				continue
			}
			requests = append(requests, req)
		}

		c.JSON(http.StatusOK, gin.H{"requests": requests})
	}
}

// POST /admin/upgrade-requests/:requestId/approve - อนุมัติคำขออัพเกรด (Admin only)
func adminApproveUpgradeRequestHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID := c.GetInt("user_id")
		requestID := c.Param("requestId")

		var req struct {
			AdminNotes string `json:"admin_notes"`
		}
		c.ShouldBindJSON(&req)

		// ดึงข้อมูลคำขอ
		var userID, requestedTierID int
		var status, paymentStatus string
		err := dbPool.QueryRow(ctx, `
			SELECT user_id, requested_tier_id, status, payment_status
			FROM provider_tier_upgrade_requests
			WHERE request_id = $1
		`, requestID).Scan(&userID, &requestedTierID, &status, &paymentStatus)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
			return
		}

		if status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request already processed"})
			return
		}

		// อัปเดตสถานะเป็น approved
		_, err = dbPool.Exec(ctx, `
			UPDATE provider_tier_upgrade_requests
			SET status = 'approved',
				reviewed_at = NOW(),
				reviewed_by = $1,
				admin_notes = $2
			WHERE request_id = $3
		`, adminID, req.AdminNotes, requestID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve request"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Upgrade request approved. User can now proceed to payment.",
			"request_id": requestID,
			"user_id":    userID,
		})
	}
}

// POST /admin/upgrade-requests/:requestId/reject - ปฏิเสธคำขออัพเกรด (Admin only)
func adminRejectUpgradeRequestHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID := c.GetInt("user_id")
		requestID := c.Param("requestId")

		var req struct {
			RejectionReason string `json:"rejection_reason" binding:"required"`
			AdminNotes      string `json:"admin_notes"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Rejection reason is required"})
			return
		}

		// ดึงข้อมูลคำขอ
		var status string
		err := dbPool.QueryRow(ctx, `
			SELECT status
			FROM provider_tier_upgrade_requests
			WHERE request_id = $1
		`, requestID).Scan(&status)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
			return
		}

		if status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request already processed"})
			return
		}

		// อัปเดตสถานะเป็น rejected
		_, err = dbPool.Exec(ctx, `
			UPDATE provider_tier_upgrade_requests
			SET status = 'rejected',
				reviewed_at = NOW(),
				reviewed_by = $1,
				rejection_reason = $2,
				admin_notes = $3
			WHERE request_id = $4
		`, adminID, req.RejectionReason, req.AdminNotes, requestID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject request"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Upgrade request rejected",
			"request_id": requestID,
		})
	}
}

// POST /admin/recalculate-provider-tiers - คำนวณ Tier ของ providers ทั้งหมดใหม่
func adminRecalculateProviderTiersHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. อัปเดต provider_stats ทั้งหมด
		_, err := dbPool.Exec(ctx, `
			UPDATE provider_stats ps
			SET 
				tier_points = calculate_provider_tier_points(ps.user_id),
				updated_at = NOW()
		`)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalculate tier points"})
			return
		}

		// 2. จัดอันดับ Tier อัตโนมัติ
		type ProviderUpdate struct {
			UserID     int
			OldTierID  int
			NewTierID  int
			TierPoints int
		}

		rows, err := dbPool.Query(ctx, `
			SELECT 
				u.user_id,
				u.provider_level_id as old_tier_id,
				ps.tier_points,
				CASE 
					WHEN ps.tier_points >= 400 THEN 4  -- Premium
					WHEN ps.tier_points >= 250 THEN 3  -- Diamond
					WHEN ps.tier_points >= 100 THEN 2  -- Silver
					ELSE 1                              -- General
				END as new_tier_id
			FROM users u
			JOIN provider_stats ps ON u.user_id = ps.user_id
			WHERE u.is_provider = true
			  AND u.provider_verification_status = 'approved'
		`)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch provider tiers"})
			return
		}
		defer rows.Close()

		updates := []ProviderUpdate{}
		for rows.Next() {
			var pu ProviderUpdate
			err := rows.Scan(&pu.UserID, &pu.OldTierID, &pu.TierPoints, &pu.NewTierID)
			if err != nil {
				continue
			}
			if pu.OldTierID != pu.NewTierID {
				updates = append(updates, pu)
			}
		}

		// 3. อัปเดต Tier ที่เปลี่ยนแปลง
		updatedCount := 0
		for _, pu := range updates {
			// อัปเดต tier
			_, err := dbPool.Exec(ctx, `
				UPDATE users
				SET provider_level_id = $1
				WHERE user_id = $2
			`, pu.NewTierID, pu.UserID)

			if err != nil {
				continue
			}

			// บันทึกประวัติ
			_, _ = dbPool.Exec(ctx, `
				INSERT INTO provider_tier_history (
					user_id, old_tier_id, new_tier_id, change_type, reason
				) VALUES ($1, $2, $3, 'auto', $4)
			`, pu.UserID, pu.OldTierID, pu.NewTierID,
				fmt.Sprintf("Auto tier update based on points: %d", pu.TierPoints))

			updatedCount++
		}

		c.JSON(http.StatusOK, gin.H{
			"message":         "Provider tiers recalculated successfully",
			"total_providers": updatedCount,
			"updates":         updates,
		})
	}
}

// PATCH /admin/set-provider-tier/:userId - เปลี่ยน Tier แบบ Manual (Admin only)
func adminSetProviderTierHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID := c.GetInt("user_id")
		userID := c.Param("userId")

		var req struct {
			NewTierID int    `json:"new_tier_id" binding:"required"`
			Reason    string `json:"reason"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ดึง tier เดิม
		var oldTierID int
		err := dbPool.QueryRow(ctx, `
			SELECT provider_level_id 
			FROM users 
			WHERE user_id = $1 AND is_provider = true
		`, userID).Scan(&oldTierID)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
			return
		}

		// อัปเดต tier
		_, err = dbPool.Exec(ctx, `
			UPDATE users
			SET provider_level_id = $1
			WHERE user_id = $2
		`, req.NewTierID, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tier"})
			return
		}

		// บันทึกประวัติ
		_, _ = dbPool.Exec(ctx, `
			INSERT INTO provider_tier_history (
				user_id, old_tier_id, new_tier_id, change_type, reason, changed_by
			) VALUES ($1, $2, $3, 'manual', $4, $5)
		`, userID, oldTierID, req.NewTierID, req.Reason, adminID)

		c.JSON(http.StatusOK, gin.H{
			"message":     "Provider tier updated successfully",
			"user_id":     userID,
			"old_tier_id": oldTierID,
			"new_tier_id": req.NewTierID,
		})
	}
}

// GET /provider/tier-history - ดูประวัติการเปลี่ยน Tier ของตัวเอง
func getMyTierHistoryHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		rows, err := dbPool.Query(ctx, `
			SELECT 
				pth.history_id,
				t1.name as old_tier_name,
				t2.name as new_tier_name,
				pth.change_type,
				pth.reason,
				pth.changed_at
			FROM provider_tier_history pth
			LEFT JOIN tiers t1 ON pth.old_tier_id = t1.tier_id
			JOIN tiers t2 ON pth.new_tier_id = t2.tier_id
			WHERE pth.user_id = $1
			ORDER BY pth.changed_at DESC
		`, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tier history"})
			return
		}
		defer rows.Close()

		type TierHistory struct {
			HistoryID   int     `json:"history_id"`
			OldTierName *string `json:"old_tier_name"`
			NewTierName string  `json:"new_tier_name"`
			ChangeType  string  `json:"change_type"`
			Reason      *string `json:"reason"`
			ChangedAt   string  `json:"changed_at"`
		}

		history := []TierHistory{}
		for rows.Next() {
			var th TierHistory
			err := rows.Scan(
				&th.HistoryID, &th.OldTierName, &th.NewTierName,
				&th.ChangeType, &th.Reason, &th.ChangedAt,
			)
			if err != nil {
				continue
			}
			history = append(history, th)
		}

		c.JSON(http.StatusOK, gin.H{
			"history": history,
			"total":   len(history),
		})
	}
}

// GET /admin/provider/:userId/tier-details - ดูรายละเอียด Tier ของ provider (Admin)
func adminGetProviderTierDetailsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userId")

		type TierDetails struct {
			UserID              int     `json:"user_id"`
			Username            string  `json:"username"`
			Email               string  `json:"email"`
			CurrentTierID       int     `json:"current_tier_id"`
			CurrentTierName     string  `json:"current_tier_name"`
			TierPoints          int     `json:"tier_points"`
			AverageRating       float64 `json:"average_rating"`
			TotalReviews        int     `json:"total_reviews"`
			CompletedBookings   int     `json:"completed_bookings"`
			TotalBookings       int     `json:"total_bookings"`
			ResponseRate        float64 `json:"response_rate"`
			AcceptanceRate      float64 `json:"acceptance_rate"`
			TotalEarnings       float64 `json:"total_earnings"`
			RecommendedTierID   int     `json:"recommended_tier_id"`
			RecommendedTierName string  `json:"recommended_tier_name"`
		}

		var details TierDetails
		err := dbPool.QueryRow(ctx, `
			SELECT 
				u.user_id,
				u.username,
				u.email,
				u.provider_level_id,
				t.name as tier_name,
				ps.tier_points,
				ps.average_rating,
				ps.total_reviews,
				ps.completed_bookings,
				ps.total_bookings,
				ps.response_rate,
				ps.acceptance_rate,
				ps.total_earnings,
				CASE 
					WHEN ps.tier_points >= 400 THEN 4  -- Premium
					WHEN ps.tier_points >= 250 THEN 3  -- Diamond
					WHEN ps.tier_points >= 100 THEN 2  -- Silver
					ELSE 1                              -- General
				END as recommended_tier_id,
				CASE 
					WHEN ps.tier_points >= 400 THEN 'Premium'
					WHEN ps.tier_points >= 250 THEN 'Diamond'
					WHEN ps.tier_points >= 100 THEN 'Silver'
					ELSE 'General'
				END as recommended_tier_name
			FROM users u
			JOIN tiers t ON u.provider_level_id = t.tier_id
			LEFT JOIN provider_stats ps ON u.user_id = ps.user_id
			WHERE u.user_id = $1 AND u.is_provider = true
		`, userID).Scan(
			&details.UserID, &details.Username, &details.Email,
			&details.CurrentTierID, &details.CurrentTierName,
			&details.TierPoints, &details.AverageRating, &details.TotalReviews,
			&details.CompletedBookings, &details.TotalBookings,
			&details.ResponseRate, &details.AcceptanceRate, &details.TotalEarnings,
			&details.RecommendedTierID, &details.RecommendedTierName,
		)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
			return
		}

		c.JSON(http.StatusOK, details)
	}
}
