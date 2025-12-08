package main

import (
	"context"
	"fmt"
	"io"
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

// --- Handler: POST /verification/provider-submit ---
// Accepts multipart form with:
// - id_document (1 file) - ID card or passport
// - face_scan (1 file) - Face verification photo
// - profile_photo_0, profile_photo_1, ... (minimum 3 files)
// - profile_photo_count (string number)
// - birth_date (YYYY-MM-DD)
func providerSubmitVerificationHandler(dbPool *pgxpool.Pool, storageClient *storage.Client, gcsBucketName string, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// 1. Check user status
		var currentStatus string
		var providerLevelID int
		err := dbPool.QueryRow(ctx,
			`SELECT verification_status, provider_level_id FROM users WHERE user_id = $1`,
			userID,
		).Scan(&currentStatus, &providerLevelID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user status"})
			return
		}

		// GOD users don't need verification
		if providerLevelID == 5 {
			c.JSON(http.StatusForbidden, gin.H{"error": "GOD tier users do not need verification."})
			return
		}

		// Already verified or pending
		if currentStatus == "verified" || currentStatus == "pending" {
			c.JSON(http.StatusConflict, gin.H{"error": "You are already verified or your application is pending review."})
			return
		}

		// 2. Parse birth date and validate age
		birthDateStr := c.PostForm("birth_date")
		if birthDateStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "birth_date is required"})
			return
		}

		birthDate, err := time.Parse("2006-01-02", birthDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid birth date format. Use YYYY-MM-DD"})
			return
		}

		age := time.Now().Year() - birthDate.Year()
		if time.Now().YearDay() < birthDate.YearDay() {
			age--
		}

		if age < 20 {
			c.JSON(http.StatusForbidden, gin.H{"error": "You must be at least 20 years old"})
			return
		}

		// 3. Get ID document file
		idDocFile, err := c.FormFile("id_document")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id_document is required"})
			return
		}

		// 4. Get face scan file
		faceScanFile, err := c.FormFile("face_scan")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "face_scan is required"})
			return
		}

		// 5. Get profile photos (minimum 3)
		profilePhotoCountStr := c.PostForm("profile_photo_count")
		if profilePhotoCountStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "profile_photo_count is required"})
			return
		}

		var profilePhotoCount int
		fmt.Sscanf(profilePhotoCountStr, "%d", &profilePhotoCount)

		if profilePhotoCount < 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "At least 3 profile photos are required"})
			return
		}

		// 6. Upload files to GCS
		basePath := fmt.Sprintf("kyc/%d", userID)

		// Upload ID document
		idDocKey := fmt.Sprintf("%s/id_document_%s.jpg", basePath, uuid.NewString())
		idDocSrc, _ := idDocFile.Open()
		defer idDocSrc.Close()

		wc := storageClient.Bucket(gcsBucketName).Object(idDocKey).NewWriter(ctx)
		if _, err := io.Copy(wc, idDocSrc); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload ID document"})
			return
		}
		wc.Close()

		// Upload face scan
		faceScanKey := fmt.Sprintf("%s/face_scan_%s.jpg", basePath, uuid.NewString())
		faceScanSrc, _ := faceScanFile.Open()
		defer faceScanSrc.Close()

		wc = storageClient.Bucket(gcsBucketName).Object(faceScanKey).NewWriter(ctx)
		if _, err := io.Copy(wc, faceScanSrc); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload face scan"})
			return
		}
		wc.Close()

		// Upload profile photos
		var profilePhotoKeys []string
		for i := 0; i < profilePhotoCount; i++ {
			fieldName := fmt.Sprintf("profile_photo_%d", i)
			file, err := c.FormFile(fieldName)
			if err != nil {
				continue // Skip missing photos
			}

			key := fmt.Sprintf("%s/profile_%d_%s.jpg", basePath, i, uuid.NewString())
			src, _ := file.Open()
			defer src.Close()

			wc := storageClient.Bucket(gcsBucketName).Object(key).NewWriter(ctx)
			if _, err := io.Copy(wc, src); err != nil {
				log.Printf("Failed to upload profile photo %d: %v", i, err)
				continue
			}
			wc.Close()

			profilePhotoKeys = append(profilePhotoKeys, key)
		}

		if len(profilePhotoKeys) < 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload minimum 3 profile photos"})
			return
		}

		// 7. Update age in user_profiles
		_, err = dbPool.Exec(ctx,
			`UPDATE user_profiles SET age = $1 WHERE user_id = $2`,
			age, userID,
		)
		if err != nil {
			log.Printf("Failed to update age: %v", err)
		}

		// 8. Save verification data to database
		// Convert profile photo keys to JSON array
		profileKeysJSON := "["
		for i, key := range profilePhotoKeys {
			if i > 0 {
				profileKeysJSON += ","
			}
			profileKeysJSON += fmt.Sprintf(`"%s"`, key)
		}
		profileKeysJSON += "]"

		sqlStatement := `
			INSERT INTO user_verifications (user_id, national_id_url, face_scan_url, profile_photos, submitted_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (user_id) DO UPDATE SET
				national_id_url = EXCLUDED.national_id_url,
				face_scan_url = EXCLUDED.face_scan_url,
				profile_photos = EXCLUDED.profile_photos,
				submitted_at = NOW();
		`
		_, err = dbPool.Exec(ctx, sqlStatement,
			userID,
			idDocKey,
			faceScanKey,
			profileKeysJSON,
		)
		if err != nil {
			log.Printf("Failed to save verification data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save verification data"})
			return
		}

		// 9. Update user status to pending
		_, err = dbPool.Exec(ctx,
			`UPDATE users SET verification_status = 'pending' WHERE user_id = $1`,
			userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":             "Verification submitted successfully. Please wait for review.",
			"age":                 age,
			"profile_photo_count": len(profilePhotoKeys),
		})
	}
}
