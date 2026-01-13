package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Handler: POST /photos/upload-base64 ---
// (Direct base64 upload to GCS)
func uploadPhotoBase64Handler(dbPool *pgxpool.Pool, storageClient *storage.Client, gcsBucketName string, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("📸 Base64 photo upload request received")

		userID, exists := c.Get("userID")
		if !exists {
			log.Printf("❌ User not authenticated")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		log.Printf("✅ User ID: %v", userID)

		// Check if user is a verified provider
		var verificationStatus string
		err := dbPool.QueryRow(ctx, "SELECT verification_status FROM users WHERE user_id = $1", userID).Scan(&verificationStatus)
		if err != nil {
			log.Printf("❌ Failed to query user status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user status"})
			return
		}

		log.Printf("✅ User verification status: %s", verificationStatus)

		if verificationStatus != "verified" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only verified providers can upload photos."})
			return
		}

		// Parse request body
		var req struct {
			ImageBase64 string `json:"image_base64" binding:"required"`
			FileName    string `json:"file_name"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("❌ Invalid request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Validate base64 (decode to check)
		_, err = base64.StdEncoding.DecodeString(req.ImageBase64)
		if err != nil {
			log.Printf("❌ Failed to decode base64: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 data"})
			return
		}

		// Store as data URL (can be displayed directly in browser)
		photoURL := "data:image/jpeg;base64," + req.ImageBase64
		log.Printf("✅ Storing base64 as data URL (length: %d)", len(photoURL))

		// Save to database
		var photoID int
		err = dbPool.QueryRow(ctx,
			`INSERT INTO user_photos (user_id, photo_url, sort_order, uploaded_at, is_verified)
			 VALUES ($1, $2, 0, NOW(), false)
			 RETURNING photo_id`,
			userID, photoURL,
		).Scan(&photoID)

		if err != nil {
			log.Printf("❌ Failed to save photo to database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo"})
			return
		}

		log.Printf("✅ Photo saved with ID: %d", photoID)

		c.JSON(http.StatusOK, gin.H{
			"photo_id":  photoID,
			"photo_url": photoURL,
			"message":   "Photo uploaded successfully",
		})
	}
}

// --- Handler: POST /photos/start ---
// (Supports both legacy workflow and direct base64 upload)
func startPhotoUploadHandler(dbPool *pgxpool.Pool, storageClient *storage.Client, gcsBucketName string, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("📸 Photo upload start request received")

		userID, exists := c.Get("userID")
		if !exists {
			log.Printf("❌ User not authenticated")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		log.Printf("✅ User ID: %v", userID)

		// Check if user is a verified provider
		var verificationStatus string
		err := dbPool.QueryRow(ctx, "SELECT verification_status FROM users WHERE user_id = $1", userID).Scan(&verificationStatus)
		if err != nil {
			log.Printf("❌ Failed to query user status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user status"})
			return
		}

		log.Printf("✅ User verification status: %s", verificationStatus)

		// (Security Check) Only 'verified' users can upload to the gallery
		if verificationStatus != "verified" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only verified providers can upload photos."})
			return
		}

		// Parse request body (optional image_base64)
		var req struct {
			ImageBase64 string `json:"image_base64"`
			FileName    string `json:"file_name"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			// Empty body is OK for legacy workflow
			log.Printf("⚠️  Empty request body, using legacy workflow")
		}

		// If base64 provided, upload directly (new workflow)
		if req.ImageBase64 != "" {
			log.Printf("📸 Direct base64 upload detected")

			// Validate base64
			_, err := base64.StdEncoding.DecodeString(req.ImageBase64)
			if err != nil {
				log.Printf("❌ Failed to decode base64: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 data"})
				return
			}

			// Store as data URL (can be displayed directly in browser)
			photoURL := "data:image/jpeg;base64," + req.ImageBase64
			log.Printf("✅ Storing base64 as data URL (length: %d)", len(photoURL))

			// Save to database
			var photoID int
			err = dbPool.QueryRow(ctx,
				`INSERT INTO user_photos (user_id, photo_url, sort_order, uploaded_at, is_verified)
				 VALUES ($1, $2, 0, NOW(), false)
				 RETURNING photo_id`,
				userID, photoURL,
			).Scan(&photoID)

			if err != nil {
				log.Printf("❌ Failed to save photo to database: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo"})
				return
			}

			log.Printf("✅ Photo saved with ID: %d", photoID)

			c.JSON(http.StatusOK, gin.H{
				"photo_id":  photoID,
				"photo_url": photoURL,
				"message":   "Photo uploaded successfully",
			})
			return
		}

		// Legacy workflow: Create placeholder, frontend will use /photos/submit
		log.Printf("📸 Legacy workflow: creating placeholder for /photos/submit")

		// Generate unique photo_key that frontend can use
		photoKey := fmt.Sprintf("photos/gallery/%d/%s.jpg", userID, uuid.NewString())

		// Return photo_key with instruction to use /photos/submit
		c.JSON(http.StatusOK, gin.H{
			"photo_key":  photoKey,
			"upload_url": "DEPRECATED_USE_PHOTOS_SUBMIT_WITH_BASE64",
			"message":    "Please send base64 image to /photos/submit with this photo_key",
		})
	}
}

// --- Handler: POST /photos/submit ---
// (Saves the photo (base64 or photo_key) to the user_photos table and uploads to GCS)
func submitPhotoUploadHandler(dbPool *pgxpool.Pool, storageClient *storage.Client, gcsBucketName string, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("📸 Photo submit request received")

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var submission struct {
			PhotoKey    string `json:"photo_key"`    // Legacy support
			ImageBase64 string `json:"image_base64"` // New base64 support
		}
		if err := c.ShouldBindJSON(&submission); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Debug log
		log.Printf("📸 Received: photo_key=%s, image_base64_length=%d", submission.PhotoKey, len(submission.ImageBase64))

		var photoURL string

		// If base64 is provided, store as data URL
		if submission.ImageBase64 != "" {
			log.Printf("📸 Processing base64 image - storing in database")

			// Validate base64
			_, err := base64.StdEncoding.DecodeString(submission.ImageBase64)
			if err != nil {
				log.Printf("❌ Failed to decode base64: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 data"})
				return
			}

			// Store as data URL (can be displayed directly in browser)
			photoURL = "data:image/jpeg;base64," + submission.ImageBase64
			log.Printf("✅ Storing base64 as data URL (length: %d)", len(photoURL))
		} else if submission.PhotoKey != "" {
			// Legacy support for photo_key
			photoURL = submission.PhotoKey
			log.Printf("✅ Using photo_key: %s", photoURL)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Either image_base64 or photo_key is required"})
			return
		}

		var newPhoto struct {
			PhotoID    int       `json:"photo_id"`
			UserID     int       `json:"user_id"`
			PhotoURL   string    `json:"photo_url"`
			SortOrder  int       `json:"sort_order"`
			UploadedAt time.Time `json:"uploaded_at"`
		}
		sqlStatement := `
			INSERT INTO user_photos (user_id, photo_url, sort_order, uploaded_at, is_verified)
			VALUES ($1, $2, 0, NOW(), false)
			RETURNING photo_id, user_id, photo_url, sort_order, uploaded_at
		`
		err := dbPool.QueryRow(ctx, sqlStatement,
			userID,
			photoURL,
		).Scan(
			&newPhoto.PhotoID,
			&newPhoto.UserID,
			&newPhoto.PhotoURL,
			&newPhoto.SortOrder,
			&newPhoto.UploadedAt,
		)

		if err != nil {
			log.Printf("❌ Failed to save photo to database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo to database", "details": err.Error()})
			return
		}

		log.Printf("✅ Photo saved with ID: %d", newPhoto.PhotoID)

		c.JSON(http.StatusCreated, newPhoto)
	}
}

// --- Handler: GET /photos/me ---
// (Gets all photos for the currently logged-in user)
func getMyPhotosHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		type PhotoResponse struct {
			PhotoID    int       `json:"photo_id"`
			UserID     int       `json:"user_id"`
			PhotoURL   string    `json:"photo_url"`
			SortOrder  int       `json:"sort_order"`
			UploadedAt time.Time `json:"uploaded_at"`
		}
		photos := make([]PhotoResponse, 0) // Initialize as empty array

		sqlStatement := `
			SELECT photo_id, user_id, photo_url, sort_order, uploaded_at
			FROM user_photos
			WHERE user_id = $1
			ORDER BY sort_order ASC, uploaded_at ASC
		`
		rows, err := dbPool.Query(ctx, sqlStatement, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "details": err.Error()})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var p PhotoResponse
			if err := rows.Scan(
				&p.PhotoID, &p.UserID, &p.PhotoURL,
				&p.SortOrder, &p.UploadedAt,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan photo row"})
				return
			}
			photos = append(photos, p)
		}

		c.JSON(http.StatusOK, photos)
	}
}

// --- Handler: DELETE /photos/:photoId ---
// (Deletes a photo from the database)
func deletePhotoHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get userID from token (the owner)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// 2. Get photoId from URL param
		photoID, err := strconv.Atoi(c.Param("photoId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo ID"})
			return
		}

		// 3. Execute delete, ensuring the user_id matches
		sqlStatement := `
			DELETE FROM user_photos
			WHERE photo_id = $1 AND user_id = $2
		`
		cmdTag, err := dbPool.Exec(ctx, sqlStatement, photoID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete photo", "details": err.Error()})
			return
		}

		// 4. Check if a row was actually deleted
		if cmdTag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Photo not found or you do not have permission to delete it"})
			return
		}

		// (Note: This does not delete the file from GCS, only the DB record)
		c.JSON(http.StatusOK, gin.H{"message": "Photo deleted successfully"})
	}
}
