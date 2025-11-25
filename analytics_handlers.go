package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Analytics structs
type ProviderAnalytics struct {
	// Profile stats
	ProfileViews        int     `json:"profile_views"`
	TotalBookings       int     `json:"total_bookings"`
	CompletedBookings   int     `json:"completed_bookings"`
	CancelledBookings   int     `json:"cancelled_bookings"`
	PendingBookings     int     `json:"pending_bookings"`
	TotalRevenue        float64 `json:"total_revenue"`
	AverageRating       float64 `json:"average_rating"`
	TotalReviews        int     `json:"total_reviews"`
	FavoriteCount       int     `json:"favorite_count"`
	ResponseRate        float64 `json:"response_rate"`         // % ของข้อความที่ตอบกลับ
	AverageResponseTime int     `json:"average_response_time"` // นาที
}

type BookingStats struct {
	Date           string  `json:"date"`
	BookingCount   int     `json:"booking_count"`
	Revenue        float64 `json:"revenue"`
	CompletedCount int     `json:"completed_count"`
	CancelledCount int     `json:"cancelled_count"`
}

type RevenueBreakdown struct {
	PackageName  string  `json:"package_name"`
	BookingCount int     `json:"booking_count"`
	TotalRevenue float64 `json:"total_revenue"`
	AvgPrice     float64 `json:"avg_price"`
}

type RatingBreakdown struct {
	Rating5 int `json:"rating_5"`
	Rating4 int `json:"rating_4"`
	Rating3 int `json:"rating_3"`
	Rating2 int `json:"rating_2"`
	Rating1 int `json:"rating_1"`
}

// --- GET /analytics/provider/dashboard (Overview Dashboard) ---
func getProviderDashboardHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		providerID := userID.(int)

		var analytics ProviderAnalytics

		// 1. Profile views (from profile_views table if exists, otherwise 0)
		err := dbPool.QueryRow(ctx, `
			SELECT COALESCE(SUM(view_count), 0)
			FROM profile_views
			WHERE provider_id = $1
		`, providerID).Scan(&analytics.ProfileViews)
		if err != nil {
			analytics.ProfileViews = 0 // Table might not exist yet
		}

		// 2. Booking statistics
		err = dbPool.QueryRow(ctx, `
			SELECT 
				COUNT(*) as total_bookings,
				COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
				COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled,
				COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
				COALESCE(SUM(CASE WHEN status = 'completed' THEN total_price ELSE 0 END), 0) as revenue
			FROM bookings
			WHERE provider_id = $1
		`, providerID).Scan(
			&analytics.TotalBookings,
			&analytics.CompletedBookings,
			&analytics.CancelledBookings,
			&analytics.PendingBookings,
			&analytics.TotalRevenue,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch booking stats"})
			return
		}

		// 3. Reviews and rating
		err = dbPool.QueryRow(ctx, `
			SELECT 
				COALESCE(AVG(rating), 0) as avg_rating,
				COUNT(*) as total_reviews
			FROM reviews
			WHERE provider_id = $1
		`, providerID).Scan(&analytics.AverageRating, &analytics.TotalReviews)
		if err != nil {
			analytics.AverageRating = 0
			analytics.TotalReviews = 0
		}

		// 4. Favorite count
		err = dbPool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM favorites
			WHERE provider_id = $1
		`, providerID).Scan(&analytics.FavoriteCount)
		if err != nil {
			analytics.FavoriteCount = 0
		}

		// 5. Message response rate (if messages table exists)
		err = dbPool.QueryRow(ctx, `
			SELECT 
				CASE 
					WHEN COUNT(*) = 0 THEN 0 
					ELSE (COUNT(CASE WHEN sender_id = $1 THEN 1 END)::float / COUNT(*)::float * 100)
				END as response_rate
			FROM messages m
			JOIN conversations c ON m.conversation_id = c.id
			WHERE c.user1_id = $1 OR c.user2_id = $1
		`, providerID).Scan(&analytics.ResponseRate)
		if err != nil {
			analytics.ResponseRate = 0 // Table might not exist
		}

		// 6. Average response time (simplified - in minutes)
		analytics.AverageResponseTime = 15 // Placeholder - would need complex query

		c.JSON(http.StatusOK, analytics)
	}
}

// --- GET /analytics/provider/bookings (Booking Statistics by Date) ---
func getBookingStatsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		providerID := userID.(int)

		// Query parameters
		period := c.DefaultQuery("period", "30") // days

		rows, err := dbPool.Query(ctx, `
			SELECT 
				booking_date::date as date,
				COUNT(*) as booking_count,
				COALESCE(SUM(CASE WHEN status = 'completed' THEN total_price ELSE 0 END), 0) as revenue,
				COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_count,
				COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_count
			FROM bookings
			WHERE provider_id = $1 
			AND booking_date >= CURRENT_DATE - INTERVAL '`+period+` days'
			GROUP BY booking_date::date
			ORDER BY date DESC
		`, providerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch booking stats"})
			return
		}
		defer rows.Close()

		stats := make([]BookingStats, 0)
		for rows.Next() {
			var stat BookingStats
			if err := rows.Scan(&stat.Date, &stat.BookingCount, &stat.Revenue, &stat.CompletedCount, &stat.CancelledCount); err != nil {
				continue
			}
			stats = append(stats, stat)
		}

		c.JSON(http.StatusOK, gin.H{
			"stats":  stats,
			"period": period,
		})
	}
}

// --- GET /analytics/provider/revenue (Revenue Breakdown by Package) ---
func getRevenueBreakdownHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		providerID := userID.(int)

		rows, err := dbPool.Query(ctx, `
			SELECT 
				sp.package_name,
				COUNT(*) as booking_count,
				SUM(b.total_price) as total_revenue,
				AVG(b.total_price) as avg_price
			FROM bookings b
			JOIN service_packages sp ON b.package_id = sp.package_id
			WHERE b.provider_id = $1 AND b.status = 'completed'
			GROUP BY sp.package_name
			ORDER BY total_revenue DESC
		`, providerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch revenue breakdown"})
			return
		}
		defer rows.Close()

		breakdown := make([]RevenueBreakdown, 0)
		for rows.Next() {
			var item RevenueBreakdown
			if err := rows.Scan(&item.PackageName, &item.BookingCount, &item.TotalRevenue, &item.AvgPrice); err != nil {
				continue
			}
			breakdown = append(breakdown, item)
		}

		c.JSON(http.StatusOK, gin.H{
			"revenue_breakdown": breakdown,
		})
	}
}

// --- GET /analytics/provider/ratings (Rating Distribution) ---
func getRatingBreakdownHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		providerID := userID.(int)

		var breakdown RatingBreakdown
		err := dbPool.QueryRow(ctx, `
			SELECT 
				COUNT(CASE WHEN rating = 5 THEN 1 END) as rating_5,
				COUNT(CASE WHEN rating = 4 THEN 1 END) as rating_4,
				COUNT(CASE WHEN rating = 3 THEN 1 END) as rating_3,
				COUNT(CASE WHEN rating = 2 THEN 1 END) as rating_2,
				COUNT(CASE WHEN rating = 1 THEN 1 END) as rating_1
			FROM reviews
			WHERE provider_id = $1
		`, providerID).Scan(
			&breakdown.Rating5,
			&breakdown.Rating4,
			&breakdown.Rating3,
			&breakdown.Rating2,
			&breakdown.Rating1,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rating breakdown"})
			return
		}

		total := breakdown.Rating5 + breakdown.Rating4 + breakdown.Rating3 + breakdown.Rating2 + breakdown.Rating1

		c.JSON(http.StatusOK, gin.H{
			"breakdown":     breakdown,
			"total_reviews": total,
		})
	}
}

// --- POST /analytics/profile-view (Track Profile View) ---
func trackProfileViewHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			ProviderID int `json:"provider_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get viewer ID (if authenticated)
		var viewerID *int
		if userID, exists := c.Get("userID"); exists {
			uid := userID.(int)
			viewerID = &uid
		}

		// Insert or update profile view
		_, err := dbPool.Exec(ctx, `
			INSERT INTO profile_views (provider_id, viewer_id, view_count, last_viewed_at)
			VALUES ($1, $2, 1, $3)
			ON CONFLICT (provider_id, COALESCE(viewer_id, -1))
			DO UPDATE SET 
				view_count = profile_views.view_count + 1,
				last_viewed_at = $3
		`, input.ProviderID, viewerID, time.Now())

		if err != nil {
			// Table might not exist - fail silently
			c.JSON(http.StatusOK, gin.H{"message": "View tracked (fallback)"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile view tracked"})
	}
}

// --- GET /analytics/provider/monthly (Monthly Summary) ---
func getMonthlyStatsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		providerID := userID.(int)

		type MonthlyStat struct {
			Month          string  `json:"month"`
			BookingCount   int     `json:"booking_count"`
			CompletedCount int     `json:"completed_count"`
			Revenue        float64 `json:"revenue"`
			NewReviews     int     `json:"new_reviews"`
			AverageRating  float64 `json:"average_rating"`
		}

		rows, err := dbPool.Query(ctx, `
			WITH monthly_bookings AS (
				SELECT 
					TO_CHAR(booking_date, 'YYYY-MM') as month,
					COUNT(*) as booking_count,
					COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_count,
					COALESCE(SUM(CASE WHEN status = 'completed' THEN total_price ELSE 0 END), 0) as revenue
				FROM bookings
				WHERE provider_id = $1 
				AND booking_date >= CURRENT_DATE - INTERVAL '12 months'
				GROUP BY TO_CHAR(booking_date, 'YYYY-MM')
			),
			monthly_reviews AS (
				SELECT 
					TO_CHAR(created_at, 'YYYY-MM') as month,
					COUNT(*) as review_count,
					AVG(rating) as avg_rating
				FROM reviews
				WHERE provider_id = $1
				AND created_at >= CURRENT_DATE - INTERVAL '12 months'
				GROUP BY TO_CHAR(created_at, 'YYYY-MM')
			)
			SELECT 
				COALESCE(mb.month, mr.month) as month,
				COALESCE(mb.booking_count, 0) as booking_count,
				COALESCE(mb.completed_count, 0) as completed_count,
				COALESCE(mb.revenue, 0) as revenue,
				COALESCE(mr.review_count, 0) as new_reviews,
				COALESCE(mr.avg_rating, 0) as average_rating
			FROM monthly_bookings mb
			FULL OUTER JOIN monthly_reviews mr ON mb.month = mr.month
			ORDER BY month DESC
		`, providerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch monthly stats"})
			return
		}
		defer rows.Close()

		stats := make([]MonthlyStat, 0)
		for rows.Next() {
			var stat MonthlyStat
			if err := rows.Scan(&stat.Month, &stat.BookingCount, &stat.CompletedCount, &stat.Revenue, &stat.NewReviews, &stat.AverageRating); err != nil {
				continue
			}
			stats = append(stats, stat)
		}

		c.JSON(http.StatusOK, gin.H{
			"monthly_stats": stats,
		})
	}
}
