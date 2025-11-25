package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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
