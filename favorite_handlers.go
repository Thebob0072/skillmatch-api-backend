package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- POST /favorites (เพิ่มรายการโปรด) ---
func addFavoriteHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, _ := c.Get("userID")

		var input struct {
			ProviderID int `json:"provider_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var favoriteID int
		err := dbPool.QueryRow(ctx, `
			INSERT INTO favorites (client_id, provider_id)
			VALUES ($1, $2)
			ON CONFLICT (client_id, provider_id) DO NOTHING
			RETURNING favorite_id
		`, clientID, input.ProviderID).Scan(&favoriteID)

		if err != nil {
			// ถ้า conflict แล้วไม่มี RETURNING ก็แสดงว่ามีอยู่แล้ว
			c.JSON(http.StatusOK, gin.H{"message": "Already in favorites"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"favorite_id": favoriteID, "message": "Added to favorites"})
	}
}

// --- DELETE /favorites/:providerId (ลบรายการโปรด) ---
func removeFavoriteHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, _ := c.Get("userID")
		providerID, err := strconv.Atoi(c.Param("providerId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
			return
		}

		_, err = dbPool.Exec(ctx, `
			DELETE FROM favorites WHERE client_id = $1 AND provider_id = $2
		`, clientID, providerID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove favorite"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Removed from favorites"})
	}
}

// --- GET /favorites (ดูรายการโปรด) ---
func getMyFavoritesHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, _ := c.Get("userID")

		rows, err := dbPool.Query(ctx, `
			SELECT u.user_id, u.username, t.name, u.gender_id, p.profile_image_url, u.google_profile_picture,
				   COALESCE(AVG(r.rating), 0) as avg_rating, COUNT(r.review_id) as review_count,
				   MAX(f.created_at) as latest_favorite
			FROM favorites f
			JOIN users u ON f.provider_id = u.user_id
			JOIN tiers t ON u.provider_level_id = t.tier_id
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			LEFT JOIN reviews r ON u.user_id = r.provider_id
			WHERE f.client_id = $1
			GROUP BY u.user_id, u.username, t.name, u.gender_id, p.profile_image_url, u.google_profile_picture
			ORDER BY latest_favorite DESC
		`, clientID)
		if err != nil {
			println("❌ getMyFavoritesHandler SQL Error:", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "details": err.Error()})
			return
		}
		defer rows.Close()

		type FavoriteProvider struct {
			UserID               int     `json:"user_id"`
			Username             string  `json:"username"`
			TierName             string  `json:"tier_name"`
			GenderID             int     `json:"gender_id"`
			ProfileImageUrl      *string `json:"profile_image_url"`
			GoogleProfilePicture *string `json:"google_profile_picture"`
			AverageRating        float64 `json:"average_rating"`
			ReviewCount          int     `json:"review_count"`
		}

		favorites := make([]FavoriteProvider, 0)
		for rows.Next() {
			var fav FavoriteProvider
			var latestFavorite time.Time // discard this field
			if err := rows.Scan(&fav.UserID, &fav.Username, &fav.TierName, &fav.GenderID,
				&fav.ProfileImageUrl, &fav.GoogleProfilePicture, &fav.AverageRating, &fav.ReviewCount, &latestFavorite); err != nil {
				continue
			}
			favorites = append(favorites, fav)
		}

		c.JSON(http.StatusOK, favorites)
	}
}

// --- GET /favorites/check/:providerId (เช็คว่าอยู่ในรายการโปรดหรือไม่) ---
// Supports optional authentication - returns false if no token provided
func checkFavoriteHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerID, err := strconv.Atoi(c.Param("providerId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
			return
		}

		// Try to get userID (may not exist if no token)
		userIDInterface, exists := c.Get("userID")
		if !exists {
			// No token provided -> return false
			c.JSON(http.StatusOK, gin.H{"is_favorite": false})
			return
		}

		clientID, ok := userIDInterface.(int)
		if !ok {
			// Invalid userID type -> return false
			c.JSON(http.StatusOK, gin.H{"is_favorite": false})
			return
		}

		var isFavorite bool
		err = dbPool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM favorites WHERE client_id = $1 AND provider_id = $2)
		`, clientID, providerID).Scan(&isFavorite)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"is_favorite": isFavorite})
	}
}
