package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalUsers           int     `json:"total_users"`
	TotalProviders       int     `json:"total_providers"`
	TotalClients         int     `json:"total_clients"`
	TotalBookings        int     `json:"total_bookings"`
	CompletedBookings    int     `json:"completed_bookings"`
	TotalRevenue         float64 `json:"total_revenue"`
	ActiveUsers24h       int     `json:"active_users_24h"`
	NewUsersToday        int     `json:"new_users_today"`
	BookingsToday        int     `json:"bookings_today"`
	RevenueToday         float64 `json:"revenue_today"`
	PendingVerifications int     `json:"pending_verifications"`
	AvgRating            float64 `json:"average_rating"`
}

// getSystemStatsHandler returns comprehensive system statistics
func getSystemStatsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stats SystemStats

		// Query 1: Total users by role
		err := dbPool.QueryRow(ctx, `
			SELECT 
				COUNT(*) FILTER (WHERE tier_id > 1) as total_providers,
				COUNT(*) FILTER (WHERE tier_id = 1) as total_clients,
				COUNT(*) as total_users
			FROM users
			WHERE deleted_at IS NULL
		`).Scan(&stats.TotalProviders, &stats.TotalClients, &stats.TotalUsers)

		if err != nil {
			log.Printf("Error fetching user stats: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
			return
		}

		// Query 2: Booking statistics
		err = dbPool.QueryRow(ctx, `
			SELECT 
				COUNT(*) as total_bookings,
				COUNT(*) FILTER (WHERE status = 'completed') as completed_bookings,
				COALESCE(SUM(total_price) FILTER (WHERE status = 'completed'), 0) as total_revenue,
				COUNT(*) FILTER (WHERE booking_date >= CURRENT_DATE) as bookings_today,
				COALESCE(SUM(total_price) FILTER (WHERE booking_date >= CURRENT_DATE), 0) as revenue_today
			FROM bookings
		`).Scan(&stats.TotalBookings, &stats.CompletedBookings, &stats.TotalRevenue,
			&stats.BookingsToday, &stats.RevenueToday)

		if err != nil {
			log.Printf("Error fetching booking stats: %v", err)
		}

		// Query 3: Active users (last 24 hours)
		err = dbPool.QueryRow(ctx, `
			SELECT COUNT(DISTINCT user_id)
			FROM bookings
			WHERE created_at >= NOW() - INTERVAL '24 hours'
		`).Scan(&stats.ActiveUsers24h)

		if err != nil {
			log.Printf("Error fetching active users: %v", err)
		}

		// Query 4: New users today
		err = dbPool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM users
			WHERE created_at >= CURRENT_DATE
		`).Scan(&stats.NewUsersToday)

		if err != nil {
			log.Printf("Error fetching new users: %v", err)
		}

		// Query 5: Pending verifications
		err = dbPool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM users
			WHERE verification_status = 'pending'
		`).Scan(&stats.PendingVerifications)

		if err != nil {
			log.Printf("Error fetching pending verifications: %v", err)
		}

		// Query 6: Average rating
		err = dbPool.QueryRow(ctx, `
			SELECT COALESCE(AVG(rating), 0)
			FROM reviews
		`).Scan(&stats.AvgRating)

		if err != nil {
			log.Printf("Error fetching average rating: %v", err)
		}

		c.JSON(http.StatusOK, stats)
	}
}

// DatabaseStatsResponse represents database connection pool statistics
type DatabaseStatsResponse struct {
	TotalConns           int32 `json:"total_connections"`
	AcquiredConns        int32 `json:"acquired_connections"`
	IdleConns            int32 `json:"idle_connections"`
	MaxConns             int32 `json:"max_connections"`
	AcquireCount         int64 `json:"acquire_count"`
	EmptyAcquireCount    int64 `json:"empty_acquire_count"`
	CanceledAcquireCount int64 `json:"canceled_acquire_count"`
}

// getDatabaseStatsHandler returns database connection pool statistics
func getDatabaseStatsHandler(dbPool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := dbPool.Stat()

		response := DatabaseStatsResponse{
			TotalConns:           stats.TotalConns(),
			AcquiredConns:        stats.AcquiredConns(),
			IdleConns:            stats.IdleConns(),
			MaxConns:             stats.MaxConns(),
			AcquireCount:         stats.AcquireCount(),
			EmptyAcquireCount:    stats.EmptyAcquireCount(),
			CanceledAcquireCount: stats.CanceledAcquireCount(),
		}

		c.JSON(http.StatusOK, response)
	}
}

// getServerInfoHandler returns server configuration and runtime information
func getServerInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		info := gin.H{
			"version":     "1.0.0",
			"environment": gin.Mode(),
			"go_version":  fmt.Sprintf("%s", "go1.24.4"),
			"endpoints": gin.H{
				"public":    120,
				"protected": 80,
				"admin":     40,
				"god":       4,
			},
			"features": []string{
				"Authentication (JWT + Google OAuth)",
				"Provider System with KYC",
				"Booking & Payment (Stripe)",
				"Real-time Messaging (WebSocket)",
				"Notifications",
				"Reviews & Ratings",
				"Financial System (Wallet + Withdrawals)",
				"Analytics Dashboard",
				"Multi-language Support",
				"Face Verification",
			},
		}

		c.JSON(http.StatusOK, info)
	}
}
