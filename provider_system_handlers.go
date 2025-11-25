package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// Provider Registration (ลงทะเบียนเป็น Provider)
// ============================================================================

type RegisterProviderRequest struct {
	// ข้อมูลพื้นฐาน (จาก user registration)
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	GenderID  int    `json:"gender_id" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
	OTP       string `json:"otp" binding:"required"` // Email verification OTP

	// ข้อมูล Provider
	CategoryIDs []int  `json:"category_ids" binding:"required,min=1"` // หมวดหมู่บริการที่ให้บริการ
	ServiceType string `json:"service_type"`                          // "Incall", "Outcall", "Both"
	Bio         string `json:"bio"`
	Province    string `json:"province"`
	District    string `json:"district"`
}

// POST /register/provider - ลงทะเบียนเป็น Provider (ต้องส่งเอกสารภายหลัง)
func registerProviderHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterProviderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 1. ตรวจสอบ OTP
		var storedOTP string
		var expiresAt time.Time
		err := dbPool.QueryRow(ctx,
			"SELECT otp, expires_at FROM email_verifications WHERE email = $1",
			req.Email,
		).Scan(&storedOTP, &expiresAt)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification code"})
			return
		}

		if time.Now().After(expiresAt) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code has expired"})
			return
		}

		if storedOTP != req.OTP {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
			return
		}

		// 2. เช็คว่า email ซ้ำหรือไม่
		var existingUserID int
		err = dbPool.QueryRow(ctx, "SELECT user_id FROM users WHERE email = $1", req.Email).Scan(&existingUserID)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}

		// 3. Hash password
		hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		hashedPassword := string(hashedPasswordBytes)

		// 4. สร้าง user (is_provider = true)
		var userID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO users (
				username, email, password_hash, gender_id, first_name, last_name, 
				phone_number, verification_status, is_provider, provider_verification_status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, 'verified', true, 'pending')
			RETURNING user_id
		`, req.Username, req.Email, hashedPassword, req.GenderID, req.FirstName, req.LastName, req.Phone).Scan(&userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// 5. สร้าง user_profile
		_, err = dbPool.Exec(ctx, `
			INSERT INTO user_profiles (
				user_id, bio, service_type, province, district, is_available
			) VALUES ($1, $2, $3, $4, $5, false)
		`, userID, req.Bio, req.ServiceType, req.Province, req.District)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
			return
		}

		// 6. เพิ่ม provider_categories
		for i, categoryID := range req.CategoryIDs {
			isPrimary := (i == 0) // หมวดหมู่แรกเป็น primary
			_, err = dbPool.Exec(ctx, `
				INSERT INTO provider_categories (provider_id, category_id, is_primary)
				VALUES ($1, $2, $3)
			`, userID, categoryID, isPrimary)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add category"})
				return
			}
		}

		// 7. สร้าง provider_stats
		_, err = dbPool.Exec(ctx, `
			INSERT INTO provider_stats (user_id)
			VALUES ($1)
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create provider stats", "details": err.Error()})
			return
		}

		// 8. ลบ OTP ที่ใช้แล้ว
		_, _ = dbPool.Exec(ctx, "DELETE FROM email_verifications WHERE email = $1", req.Email)

		// 9. สร้าง JWT token
		token, err := createJWT(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":   "Provider registration successful. Please upload required documents to complete verification.",
			"user_id":   userID,
			"token":     token,
			"next_step": "Upload documents: National ID, Health Certificate",
		})
	}
}

// ============================================================================
// Provider Document Upload
// ============================================================================

type UploadDocumentRequest struct {
	DocumentType string `json:"document_type" binding:"required"` // national_id, health_certificate, etc.
	FileURL      string `json:"file_url" binding:"required"`
	FileName     string `json:"file_name"`
}

// POST /provider/documents - อัปโหลดเอกสาร (ต้อง login)
func uploadProviderDocumentHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		var req UploadDocumentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ตรวจสอบว่าเป็น provider หรือไม่
		var isProvider bool
		err := dbPool.QueryRow(ctx, "SELECT is_provider FROM users WHERE user_id = $1", userID).Scan(&isProvider)
		if err != nil || !isProvider {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only providers can upload documents"})
			return
		}

		// บันทึกเอกสาร
		var documentID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO provider_documents (
				user_id, document_type, file_url, file_name, verification_status
			) VALUES ($1, $2, $3, $4, 'pending')
			RETURNING document_id
		`, userID, req.DocumentType, req.FileURL, req.FileName).Scan(&documentID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload document"})
			return
		}

		// อัปเดต provider_verification_status
		_, _ = dbPool.Exec(ctx, `
			UPDATE users 
			SET provider_verification_status = 'documents_submitted'
			WHERE user_id = $1
		`, userID)

		c.JSON(http.StatusCreated, gin.H{
			"message":     "Document uploaded successfully",
			"document_id": documentID,
			"status":      "pending",
		})
	}
}

// GET /provider/documents - ดูเอกสารของตัวเอง
func getMyDocumentsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		rows, err := dbPool.Query(ctx, `
			SELECT 
				document_id, document_type, file_url, file_name, 
				verification_status, uploaded_at, verified_at, rejection_reason
			FROM provider_documents
			WHERE user_id = $1
			ORDER BY uploaded_at DESC
		`, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch documents"})
			return
		}
		defer rows.Close()

		type Document struct {
			DocumentID         int        `json:"document_id"`
			DocumentType       string     `json:"document_type"`
			FileURL            string     `json:"file_url"`
			FileName           *string    `json:"file_name"`
			VerificationStatus string     `json:"verification_status"`
			UploadedAt         time.Time  `json:"uploaded_at"`
			VerifiedAt         *time.Time `json:"verified_at"`
			RejectionReason    *string    `json:"rejection_reason"`
		}

		documents := []Document{}
		for rows.Next() {
			var doc Document
			err := rows.Scan(
				&doc.DocumentID, &doc.DocumentType, &doc.FileURL, &doc.FileName,
				&doc.VerificationStatus, &doc.UploadedAt, &doc.VerifiedAt, &doc.RejectionReason,
			)
			if err != nil {
				continue
			}
			documents = append(documents, doc)
		}

		c.JSON(http.StatusOK, gin.H{
			"documents": documents,
			"total":     len(documents),
		})
	}
}

// GET /provider/categories - ดูหมวดหมู่บริการของตัวเอง
func getMyProviderCategoriesHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		rows, err := dbPool.Query(ctx, `
			SELECT 
				pc.category_id, sc.category_name, sc.icon_url, pc.is_primary
			FROM provider_categories pc
			JOIN service_categories sc ON pc.category_id = sc.category_id
			WHERE pc.provider_id = $1
			ORDER BY pc.is_primary DESC, sc.category_name
		`, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
			return
		}
		defer rows.Close()

		type Category struct {
			CategoryID   int     `json:"category_id"`
			CategoryName string  `json:"category_name"`
			IconURL      *string `json:"icon_url"`
			IsPrimary    bool    `json:"is_primary"`
		}

		categories := []Category{}
		for rows.Next() {
			var cat Category
			err := rows.Scan(&cat.CategoryID, &cat.CategoryName, &cat.IconURL, &cat.IsPrimary)
			if err != nil {
				continue
			}
			categories = append(categories, cat)
		}

		c.JSON(http.StatusOK, gin.H{
			"categories": categories,
			"total":      len(categories),
		})
	}
}

// ============================================================================
// Admin: Provider Verification
// ============================================================================

// GET /admin/providers/pending - ดู providers ที่รอการตรวจสอบ
func getAdminPendingProvidersHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := dbPool.Query(ctx, `
			SELECT 
				u.user_id, u.username, u.email, u.first_name, u.last_name,
				u.provider_verification_status, u.registration_date,
				COUNT(DISTINCT pd.document_id) as total_documents,
				COUNT(DISTINCT CASE WHEN pd.verification_status = 'approved' THEN pd.document_id END) as approved_docs,
				COUNT(DISTINCT CASE WHEN pd.verification_status = 'pending' THEN pd.document_id END) as pending_docs
			FROM users u
			LEFT JOIN provider_documents pd ON u.user_id = pd.user_id
			WHERE u.is_provider = true 
			  AND u.provider_verification_status IN ('pending', 'documents_submitted')
			GROUP BY u.user_id
			ORDER BY u.registration_date DESC
		`)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending providers"})
			return
		}
		defer rows.Close()

		type PendingProvider struct {
			UserID                     int       `json:"user_id"`
			Username                   string    `json:"username"`
			Email                      string    `json:"email"`
			FirstName                  *string   `json:"first_name"`
			LastName                   *string   `json:"last_name"`
			ProviderVerificationStatus string    `json:"provider_verification_status"`
			RegistrationDate           time.Time `json:"registration_date"`
			TotalDocuments             int       `json:"total_documents"`
			ApprovedDocuments          int       `json:"approved_documents"`
			PendingDocuments           int       `json:"pending_documents"`
		}

		providers := []PendingProvider{}
		for rows.Next() {
			var p PendingProvider
			err := rows.Scan(
				&p.UserID, &p.Username, &p.Email, &p.FirstName, &p.LastName,
				&p.ProviderVerificationStatus, &p.RegistrationDate,
				&p.TotalDocuments, &p.ApprovedDocuments, &p.PendingDocuments,
			)
			if err != nil {
				continue
			}
			providers = append(providers, p)
		}

		c.JSON(http.StatusOK, gin.H{
			"providers": providers,
			"total":     len(providers),
		})
	}
}

// PATCH /admin/verify-document/:documentId - อนุมัติ/ปฏิเสธเอกสาร
func adminVerifyDocumentHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID := c.GetInt("user_id")
		documentID := c.Param("documentId")

		var req struct {
			Status          string  `json:"status" binding:"required"` // "approved" or "rejected"
			RejectionReason *string `json:"rejection_reason"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Status != "approved" && req.Status != "rejected" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'approved' or 'rejected'"})
			return
		}

		// อัปเดตสถานะเอกสาร
		_, err := dbPool.Exec(ctx, `
			UPDATE provider_documents
			SET 
				verification_status = $1,
				verified_by = $2,
				verified_at = NOW(),
				rejection_reason = $3
			WHERE document_id = $4
		`, req.Status, adminID, req.RejectionReason, documentID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Document %s successfully", req.Status),
		})
	}
}

// PATCH /admin/approve-provider/:userId - อนุมัติ provider (เมื่อเอกสารครบถ้วน)
func adminApproveProviderHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDParam := c.Param("userId")
		userID, err := strconv.Atoi(userIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var req struct {
			Approve bool    `json:"approve" binding:"required"`
			Reason  *string `json:"reason"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		newStatus := "rejected"
		if req.Approve {
			newStatus = "approved"
		}

		// อัปเดตสถานะ provider
		_, err = dbPool.Exec(ctx, `
			UPDATE users
			SET 
				provider_verification_status = $1,
				provider_verified_at = CASE WHEN $1 = 'approved' THEN NOW() ELSE NULL END
			WHERE user_id = $2 AND is_provider = true
		`, newStatus, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider status"})
			return
		}

		message := "Provider approved successfully"
		if !req.Approve {
			message = "Provider rejected"
		}

		c.JSON(http.StatusOK, gin.H{
			"message": message,
			"user_id": userID,
			"status":  newStatus,
		})
	}
}

// GET /admin/provider-stats - สถิติ providers ทั้งหมด
func getAdminProviderStatsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stats struct {
			TotalProviders    int `json:"total_providers"`
			ApprovedProviders int `json:"approved_providers"`
			PendingProviders  int `json:"pending_providers"`
			RejectedProviders int `json:"rejected_providers"`
		}

		err := dbPool.QueryRow(ctx, `
			SELECT 
				COUNT(*) as total,
				COUNT(CASE WHEN provider_verification_status = 'approved' THEN 1 END) as approved,
				COUNT(CASE WHEN provider_verification_status IN ('pending', 'documents_submitted') THEN 1 END) as pending,
				COUNT(CASE WHEN provider_verification_status = 'rejected' THEN 1 END) as rejected
			FROM users
			WHERE is_provider = true
		`).Scan(&stats.TotalProviders, &stats.ApprovedProviders, &stats.PendingProviders, &stats.RejectedProviders)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}
