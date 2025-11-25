package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Handler: POST /verification/start (GCS Version) ---
// (Generates 3 Pre-signed URLs for KYC document upload)
func startVerificationHandler(dbPool *pgxpool.Pool, storageClient *storage.Client, gcsBucketName string, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get UserID from the authenticated session
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// 2. Check the user's current status and provider level
		var currentStatus string
		var providerLevelID int
		err := dbPool.QueryRow(ctx,
			`SELECT verification_status, provider_level_id FROM users WHERE user_id = $1`,
			userID,
		).Scan(&currentStatus, &providerLevelID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user status", "details": err.Error()})
			return
		}

		// --- (GOD BYPASS LOGIC) ---
		if providerLevelID == 5 { // 5 is the Tier ID for "GOD"
			if currentStatus != "verified" {
				// Auto-verify GOD users if they aren't already
				_, err := dbPool.Exec(ctx, `
					UPDATE users SET verification_status = 'verified' WHERE user_id = $1
				`, userID)
				if err != nil {
					log.Printf("Failed to auto-verify GOD user %v: %v", userID, err)
				}
			}
			// Block GOD users from this flow; they don't need to submit KYC
			c.JSON(http.StatusForbidden, gin.H{"error": "GOD tier users do not need verification."})
			return
		}

		// 3. Prevent already verified or pending users from resubmitting
		if currentStatus == "verified" || currentStatus == "pending" {
			c.JSON(http.StatusConflict, gin.H{"error": "You are already verified or your application is pending review."})
			return
		}

		// 4. Generate unique file keys (paths) for GCS
		basePath := fmt.Sprintf("kyc/%d", userID) // Store files in a user-specific folder
		idKey := fmt.Sprintf("%s/national_id_%s.jpg", basePath, uuid.NewString())
		healthKey := fmt.Sprintf("%s/health_cert_%s.jpg", basePath, uuid.NewString())
		faceKey := fmt.Sprintf("%s/face_scan_%s.jpg", basePath, uuid.NewString())

		// 5. Set options for the signed URL (15 minute expiry, PUT method)
		expires := time.Now().Add(15 * time.Minute)
		opts := &storage.SignedURLOptions{
			Scheme:  storage.SigningSchemeV4,
			Method:  "PUT",
			Expires: expires,
			Headers: []string{"Content-Type:application/octet-stream"}, // Allow any binary file
		}

		// 6. Generate the 3 URLs
		urlID, err := storage.SignedURL(gcsBucketName, idKey, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ID upload URL"})
			return
		}
		urlHealth, err := storage.SignedURL(gcsBucketName, healthKey, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Health Cert upload URL"})
			return
		}
		urlFace, err := storage.SignedURL(gcsBucketName, faceKey, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Face Scan upload URL"})
			return
		}

		// 7. Send the URLs and keys back to the React client
		c.JSON(http.StatusOK, gin.H{
			"national_id_upload_url": urlID,
			"national_id_key":        idKey,
			"health_cert_upload_url": urlHealth,
			"health_cert_key":        healthKey,
			"face_scan_upload_url":   urlFace,
			"face_scan_key":          faceKey,
		})
	}
}

// --- Handler: POST /verification/submit ---
// (Confirms KYC files were uploaded and updates user status to 'pending')
func submitVerificationHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// 1. Get the file keys and birth date from the React client
		var submission struct {
			NationalIDKey string `json:"national_id_key" binding:"required"`
			HealthCertKey string `json:"health_cert_key" binding:"required"`
			FaceScanKey   string `json:"face_scan_key" binding:"required"`
			BirthDate     string `json:"birth_date" binding:"required"` // YYYY-MM-DD
		}
		if err := c.ShouldBindJSON(&submission); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 1.5. Validate age (must be 20+ years old)
		birthDate, err := time.Parse("2006-01-02", submission.BirthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid birth date format. Use YYYY-MM-DD"})
			return
		}

		age := time.Now().Year() - birthDate.Year()
		if time.Now().YearDay() < birthDate.YearDay() {
			age-- // Haven't had birthday this year yet
		}

		if age < 20 {
			c.JSON(http.StatusForbidden, gin.H{"error": "You must be at least 20 years old to verify your account", "age": age})
			return
		}

		// 2. Calculate and save age to user_profiles
		_, err = dbPool.Exec(ctx,
			`UPDATE user_profiles SET age = $1 WHERE user_id = $2`,
			age, userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update age", "details": err.Error()})
			return
		}

		// 3. Save the GCS keys (file paths) to the user_verifications table
		//    'ON CONFLICT' handles resubmissions (e.g., if user was 'rejected' and tries again)
		sqlStatement := `
			INSERT INTO user_verifications (user_id, national_id_url, health_cert_url, face_scan_url, submitted_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (user_id) DO UPDATE SET
				national_id_url = EXCLUDED.national_id_url,
				health_cert_url = EXCLUDED.health_cert_url,
				face_scan_url = EXCLUDED.face_scan_url,
				submitted_at = NOW();
		`
		_, err = dbPool.Exec(ctx, sqlStatement,
			userID,
			submission.NationalIDKey,
			submission.HealthCertKey,
			submission.FaceScanKey,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save verification data", "details": err.Error()})
			return
		}

		// 4. Update the user's status to 'pending'
		_, err = dbPool.Exec(ctx,
			`UPDATE users SET verification_status = 'pending' WHERE user_id = $1`,
			userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Verification submitted successfully. Please wait for review.", "age": age})
	}
}
