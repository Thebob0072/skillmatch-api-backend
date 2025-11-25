package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Handler: POST /photos/start ---
// (Generates a Pre-signed URL for one photo upload)
func startPhotoUploadHandler(dbPool *pgxpool.Pool, storageClient *storage.Client, gcsBucketName string, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Check if user is a verified provider
		var verificationStatus string
		err := dbPool.QueryRow(ctx, "SELECT verification_status FROM users WHERE user_id = $1", userID).Scan(&verificationStatus)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user status"})
			return
		}

		// (Security Check) Only 'verified' users can upload to the gallery
		if verificationStatus != "verified" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only verified providers can upload photos."})
			return
		}

		// Generate a unique path for the photo
		fileName := fmt.Sprintf("photos/gallery/%d/%s.jpg", userID, uuid.NewString())

		expires := time.Now().Add(10 * time.Minute)
		opts := &storage.SignedURLOptions{
			Scheme:  storage.SigningSchemeV4,
			Method:  "PUT",
			Expires: expires,
			Headers: []string{"Content-Type:image/jpeg", "Content-Type:image/png"}, // Allow image types
		}

		uploadURL, err := storage.SignedURL(gcsBucketName, fileName, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload URL", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"upload_url": uploadURL,
			"photo_key":  fileName,
		})
	}
}

// --- Handler: POST /photos/submit ---
// (Saves the photo_key (path) to the user_photos table)
func submitPhotoUploadHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var submission struct {
			PhotoKey string `json:"photo_key" binding:"required"`
		}
		if err := c.ShouldBindJSON(&submission); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var newPhoto UserPhoto
		sqlStatement := `
			INSERT INTO user_photos (user_id, photo_url)
			VALUES ($1, $2)
			RETURNING photo_id, user_id, photo_url, sort_order, uploaded_at
		`
		err := dbPool.QueryRow(ctx, sqlStatement,
			userID,
			submission.PhotoKey,
		).Scan(
			&newPhoto.PhotoID,
			&newPhoto.UserID,
			&newPhoto.PhotoURL,
			&newPhoto.SortOrder,
			&newPhoto.UploadedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo to database", "details": err.Error()})
			return
		}

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

		photos := make([]UserPhoto, 0) // Initialize as empty array

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
			var p UserPhoto
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
