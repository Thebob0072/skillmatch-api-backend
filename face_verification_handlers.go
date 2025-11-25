package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FaceVerification model
type FaceVerification struct {
	VerificationID     int       `json:"verification_id"`
	UserID             int       `json:"user_id"`
	SelfieURL          string    `json:"selfie_url"`
	LivenessVideoURL   *string   `json:"liveness_video_url,omitempty"`
	MatchConfidence    *float64  `json:"match_confidence,omitempty"`
	IsMatch            bool      `json:"is_match"`
	NationalIDPhotoURL *string   `json:"national_id_photo_url,omitempty"`
	LivenessPassed     bool      `json:"liveness_passed"`
	LivenessConfidence *float64  `json:"liveness_confidence,omitempty"`
	VerificationStatus string    `json:"verification_status"`
	APIProvider        *string   `json:"api_provider,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	VerifiedBy         *int      `json:"verified_by,omitempty"`
	RejectionReason    *string   `json:"rejection_reason,omitempty"`
	RetryCount         int       `json:"retry_count"`
	DocumentType       string    `json:"document_type"`        // "national_id" or "passport"
	DocumentID         *int      `json:"document_id,omitempty"` // References provider_documents.document_id
}

// --- 1. Submit Face Verification (Provider uploads selfie) ---
func submitFaceVerificationHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("userID")

	var req struct {
		SelfieURL        string  `json:"selfie_url" binding:"required"`
		LivenessVideoURL *string `json:"liveness_video_url"`
		DocumentID       int     `json:"document_id" binding:"required"`       // ID ของเอกสารที่อัปโหลดไว้แล้ว (บัตรประชาชนหรือพาสปอร์ต)
		DocumentType     string  `json:"document_type" binding:"required,oneof=national_id passport"` // ประเภทเอกสาร: national_id หรือ passport
	}

	if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
			return
		}

	// ดึง URL ของรูปเอกสาร (บัตรประชาชนหรือพาสปอร์ต) จาก provider_documents
	var documentURL string
	var dbDocumentType string
	err := dbPool.QueryRow(ctx, `
		SELECT file_url, document_type
		FROM provider_documents 
		WHERE document_id = $1 AND user_id = $2 AND document_type = $3
	`, req.DocumentID, userID, req.DocumentType).Scan(&documentURL, &dbDocumentType)

	if err != nil {
		docTypeThai := "บัตรประชาชน"
		if req.DocumentType == "passport" {
			docTypeThai = "พาสปอร์ต"
		}
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("เอกสาร%sไม่พบ", docTypeThai)})
		return
	}	// บันทึกข้อมูล face verification (ยังไม่ทำการตรวจสอบจริง)
	var verificationID int
	err = dbPool.QueryRow(ctx, `
		INSERT INTO face_verifications (
			user_id, selfie_url, liveness_video_url, 
			national_id_photo_url, document_type, document_id,
		verification_status
	) VALUES ($1, $2, $3, $4, $5, $6, 'pending')
	RETURNING verification_id
`, userID, req.SelfieURL, req.LivenessVideoURL, documentURL, req.DocumentType, req.DocumentID).Scan(&verificationID)

	if err != nil {
			log.Printf("Error inserting face verification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit face verification"})
			return
		}

		// TODO: เรียก face matching API ที่นี่ (AWS Rekognition, Azure Face, etc.)
		// ตอนนี้จะ return pending status ก่อน
		// Admin จะต้องมา approve ภายหลัง

		c.JSON(http.StatusCreated, gin.H{
			"message":         "Face verification submitted successfully",
			"verification_id": verificationID,
			"status":          "pending",
			"next_step":       "Admin will review your face verification",
		})
	}
}

// --- 2. Get My Face Verification Status ---
func getMyFaceVerificationHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("userID")

	var verification FaceVerification
	err := dbPool.QueryRow(ctx, `
		SELECT 
			verification_id, user_id, selfie_url, liveness_video_url,
			match_confidence, is_match, national_id_photo_url,
			liveness_passed, liveness_confidence, verification_status,
			api_provider, created_at, verified_at, verified_by,
			rejection_reason, retry_count, document_type, document_id
		FROM face_verifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
		`, userID).Scan(
		&verification.VerificationID,
		&verification.UserID,
		&verification.SelfieURL,
		&verification.LivenessVideoURL,
		&verification.MatchConfidence,
		&verification.IsMatch,
		&verification.NationalIDPhotoURL,
		&verification.LivenessPassed,
		&verification.LivenessConfidence,
		&verification.VerificationStatus,
		&verification.APIProvider,
		&verification.CreatedAt,
		&verification.VerifiedAt,
		&verification.VerifiedBy,
		&verification.RejectionReason,
		&verification.RetryCount,
		&verification.DocumentType,
		&verification.DocumentID,
		)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No face verification found"})
			return
		}

		c.JSON(http.StatusOK, verification)
	}
}

// --- 3. Admin: List All Face Verifications ---
func adminListFaceVerificationsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.DefaultQuery("status", "pending")

	rows, err := dbPool.Query(ctx, `
		SELECT 
			fv.verification_id, fv.user_id, u.username, u.email,
			fv.selfie_url, fv.national_id_photo_url,
			fv.match_confidence, fv.is_match,
			fv.liveness_passed, fv.liveness_confidence,
			fv.verification_status, fv.created_at, fv.retry_count,
			fv.document_type, fv.document_id
	FROM face_verifications fv
	JOIN users u ON fv.user_id = u.user_id
	WHERE fv.verification_status = $1
	ORDER BY fv.created_at DESC
`, status)

	if err != nil {
			log.Printf("Error querying face verifications: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch verifications"})
			return
		}
		defer rows.Close()

		var verifications []map[string]interface{}
		for rows.Next() {
		var v struct {
			VerificationID     int
			UserID             int
			Username           string
			Email              string
			SelfieURL          string
			NationalIDPhotoURL *string
			MatchConfidence    *float64
			IsMatch            bool
			LivenessPassed     bool
			LivenessConfidence *float64
			VerificationStatus string
			CreatedAt          time.Time
		RetryCount         int
		DocumentType       string
		DocumentID         *int
	}

	err := rows.Scan(
			&v.VerificationID, &v.UserID, &v.Username, &v.Email,
			&v.SelfieURL, &v.NationalIDPhotoURL,
			&v.MatchConfidence, &v.IsMatch,
			&v.LivenessPassed, &v.LivenessConfidence,
		&v.VerificationStatus, &v.CreatedAt, &v.RetryCount,
		&v.DocumentType, &v.DocumentID,
	)

	if err != nil {
				log.Printf("Error scanning face verification: %v", err)
				continue
			}

		verifications = append(verifications, map[string]interface{}{
			"verification_id":       v.VerificationID,
			"user_id":               v.UserID,
			"username":              v.Username,
			"email":                 v.Email,
			"selfie_url":            v.SelfieURL,
			"national_id_photo_url": v.NationalIDPhotoURL,
			"match_confidence":      v.MatchConfidence,
			"is_match":              v.IsMatch,
			"liveness_passed":       v.LivenessPassed,
			"liveness_confidence":   v.LivenessConfidence,
			"verification_status":   v.VerificationStatus,
			"document_type":         v.DocumentType,
			"document_id":           v.DocumentID,
				"created_at":            v.CreatedAt,
				"retry_count":           v.RetryCount,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"verifications": verifications,
			"total":         len(verifications),
		})
	}
}

// --- 4. Admin: Approve/Reject Face Verification ---
func adminReviewFaceVerificationHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID := c.GetInt("userID")
		verificationID := c.Param("verificationId")

		var req struct {
			Action          string   `json:"action" binding:"required,oneof=approve reject retry"`
			RejectionReason *string  `json:"rejection_reason"`
			MatchConfidence *float64 `json:"match_confidence"` // Admin สามารถกำหนดเองได้
			IsMatch         *bool    `json:"is_match"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
			return
		}

		if req.Action == "reject" && req.RejectionReason == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rejection_reason is required when rejecting"})
			return
		}

		var newStatus string
		switch req.Action {
		case "approve":
			newStatus = "approved"
		case "reject":
			newStatus = "rejected"
		case "retry":
			newStatus = "needs_retry"
		}

		// อัพเดทสถานะ
		_, err := dbPool.Exec(ctx, `
			UPDATE face_verifications
			SET 
				verification_status = $1,
				verified_at = NOW(),
				verified_by = $2,
				rejection_reason = $3,
				match_confidence = COALESCE($4, match_confidence),
				is_match = COALESCE($5, is_match),
				retry_count = CASE WHEN $1 = 'needs_retry' THEN retry_count + 1 ELSE retry_count END
			WHERE verification_id = $6
		`, newStatus, adminID, req.RejectionReason, req.MatchConfidence, req.IsMatch, verificationID)

		if err != nil {
			log.Printf("Error updating face verification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update verification"})
			return
		}

		// ถ้า approve ให้อัพเดท user.face_verified = true (trigger จะทำให้)
		c.JSON(http.StatusOK, gin.H{
			"message": "Face verification " + req.Action + "d successfully",
			"status":  newStatus,
		})
	}
}

// --- 5. Mock Face Matching API (สำหรับทดสอบ - ใช้จริงต้องเรียก AWS/Azure) ---
func mockFaceMatchingAPI(selfieURL, idPhotoURL string) (matchConfidence float64, isMatch bool, livenessPassed bool, livenessConfidence float64) {
	// TODO: Replace with real API call
	// Example: AWS Rekognition CompareFaces
	// https://docs.aws.amazon.com/rekognition/latest/dg/API_CompareFaces.html

	// Mock response
	matchConfidence = 85.5      // 85.5% match
	isMatch = matchConfidence > 80.0
	livenessPassed = true
	livenessConfidence = 92.3

	return
}

// --- 6. Trigger Face Matching (Optional - can be called by cron or manually) ---
func triggerFaceMatchingHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		verificationID := c.Param("verificationId")

		// ดึงข้อมูล
		var selfieURL, idPhotoURL string
		err := dbPool.QueryRow(ctx, `
			SELECT selfie_url, national_id_photo_url
			FROM face_verifications
			WHERE verification_id = $1
		`, verificationID).Scan(&selfieURL, &idPhotoURL)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Verification not found"})
			return
		}

		// เรียก face matching API
		matchConf, isMatch, livenessPass, livenessConf := mockFaceMatchingAPI(selfieURL, idPhotoURL)

		// อัพเดทผลลัพธ์
		_, err = dbPool.Exec(ctx, `
			UPDATE face_verifications
			SET 
				match_confidence = $1,
				is_match = $2,
				liveness_passed = $3,
				liveness_confidence = $4,
				api_provider = 'mock_api'
			WHERE verification_id = $5
		`, matchConf, isMatch, livenessPass, livenessConf, verificationID)

		if err != nil {
			log.Printf("Error updating face matching results: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update results"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":             "Face matching completed",
			"match_confidence":    matchConf,
			"is_match":            isMatch,
			"liveness_passed":     livenessPass,
			"liveness_confidence": livenessConf,
		})
	}
}

// Helper: Parse API Response (for real implementations)
func parseFaceAPIResponse(provider string, response []byte) (matchConfidence float64, isMatch bool, livenessPassed bool, err error) {
	switch provider {
	case "aws_rekognition":
		// Parse AWS Rekognition response
		var awsResp struct {
			FaceMatches []struct {
				Similarity float64 `json:"Similarity"`
			} `json:"FaceMatches"`
		}
		err = json.Unmarshal(response, &awsResp)
		if err == nil && len(awsResp.FaceMatches) > 0 {
			matchConfidence = awsResp.FaceMatches[0].Similarity
			isMatch = matchConfidence > 80.0
		}

	case "azure_face":
		// Parse Azure Face API response
		var azureResp struct {
			IsIdentical bool    `json:"isIdentical"`
			Confidence  float64 `json:"confidence"`
		}
		err = json.Unmarshal(response, &azureResp)
		if err == nil {
			matchConfidence = azureResp.Confidence * 100
			isMatch = azureResp.IsIdentical
		}

	default:
		// Generic parser
	}

	// Liveness detection would be separate API call
	livenessPassed = true // Default

	return
}
