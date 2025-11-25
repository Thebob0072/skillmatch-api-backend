package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- POST /reviews (สร้างรีวิวหลังใช้บริการ) ---
func createReviewHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, _ := c.Get("userID")

		var input struct {
			BookingID int     `json:"booking_id" binding:"required"`
			Rating    int     `json:"rating" binding:"required,min=1,max=5"`
			Comment   *string `json:"comment"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ตรวจสอบว่า booking นี้เป็นของ client คนนี้และเสร็จสมบูรณ์แล้ว
		var providerID int
		var status string
		err := dbPool.QueryRow(ctx, `
			SELECT provider_id, status FROM bookings 
			WHERE booking_id = $1 AND client_id = $2
		`, input.BookingID, clientID).Scan(&providerID, &status)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		if status != "completed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Can only review completed bookings"})
			return
		}

		// สร้างรีวิว
		var reviewID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO reviews (booking_id, client_id, provider_id, rating, comment, is_verified)
			VALUES ($1, $2, $3, $4, $5, true)
			RETURNING review_id
		`, input.BookingID, clientID, providerID, input.Rating, input.Comment).Scan(&reviewID)

		if err != nil {
			if err.Error() == "duplicate key value violates unique constraint" {
				c.JSON(http.StatusConflict, gin.H{"error": "You have already reviewed this booking"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review", "details": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"review_id": reviewID, "message": "Review created successfully"})
	}
}

// --- GET /reviews/:providerId (ดูรีวิวของ Provider) ---
func getProviderReviewsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerID, err := strconv.Atoi(c.Param("providerId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
			return
		}

		rows, err := dbPool.Query(ctx, `
			SELECT r.review_id, u.username, r.rating, r.comment, r.is_verified, r.created_at
			FROM reviews r
			JOIN users u ON r.client_id = u.user_id
			WHERE r.provider_id = $1
			ORDER BY r.created_at DESC
		`, providerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		defer rows.Close()

		reviews := make([]ReviewWithDetails, 0)
		for rows.Next() {
			var review ReviewWithDetails
			if err := rows.Scan(&review.ReviewID, &review.ClientUsername, &review.Rating,
				&review.Comment, &review.IsVerified, &review.CreatedAt); err != nil {
				continue
			}
			reviews = append(reviews, review)
		}

		c.JSON(http.StatusOK, reviews)
	}
}

// --- GET /reviews/stats/:providerId (สถิติรีวิวของ Provider) ---
func getProviderReviewStatsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerID, err := strconv.Atoi(c.Param("providerId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
			return
		}

		var stats struct {
			TotalReviews  int     `json:"total_reviews"`
			AverageRating float64 `json:"average_rating"`
			Rating5       int     `json:"rating_5"`
			Rating4       int     `json:"rating_4"`
			Rating3       int     `json:"rating_3"`
			Rating2       int     `json:"rating_2"`
			Rating1       int     `json:"rating_1"`
		}

		err = dbPool.QueryRow(ctx, `
			SELECT 
				COUNT(*) as total_reviews,
				COALESCE(AVG(rating), 0) as average_rating,
				COUNT(CASE WHEN rating = 5 THEN 1 END) as rating_5,
				COUNT(CASE WHEN rating = 4 THEN 1 END) as rating_4,
				COUNT(CASE WHEN rating = 3 THEN 1 END) as rating_3,
				COUNT(CASE WHEN rating = 2 THEN 1 END) as rating_2,
				COUNT(CASE WHEN rating = 1 THEN 1 END) as rating_1
			FROM reviews
			WHERE provider_id = $1
		`, providerID).Scan(&stats.TotalReviews, &stats.AverageRating, &stats.Rating5,
			&stats.Rating4, &stats.Rating3, &stats.Rating2, &stats.Rating1)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}
