package main

import (
	"context"
	"log" // (เพิ่ม)
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt" // (เพิ่ม)
)

// --- (getPendingUsersHandler - เหมือนเดิม) ---
func getPendingUsersHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		pendingUsers := make([]User, 0)

		sqlStatement := `
			SELECT 
				u.user_id, u.username, u.email, u.gender_id, u.registration_date, 
				u.first_name, u.last_name,
				u.verification_status, u.tier_id, u.provider_level_id, u.phone_number, u.is_admin,
				u.google_profile_picture, p.profile_image_url, p.age
			FROM users u
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			WHERE u.verification_status = 'pending'
			ORDER BY u.registration_date ASC
		`
		rows, err := dbPool.Query(ctx, sqlStatement)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "details": err.Error()})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var u User
			if err := rows.Scan(
				&u.UserID, &u.Username, &u.Email, &u.GenderID, &u.RegistrationDate,
				&u.FirstName, &u.LastName, &u.VerificationStatus, &u.TierID, &u.ProviderLevelID,
				&u.PhoneNumber, &u.IsAdmin, &u.GoogleProfilePicture, &u.ProfileImageUrl, &u.Age,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user row", "details": err.Error()})
				return
			}
			pendingUsers = append(pendingUsers, u)
		}

		c.JSON(http.StatusOK, pendingUsers)
	}
}

// --- (getKycDetailsHandler - เหมือนเดิม) ---
func getKycDetailsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var kyc UserVerification
		err = dbPool.QueryRow(ctx,
			"SELECT verification_id, user_id, national_id_url, health_cert_url, face_scan_url, submitted_at FROM user_verifications WHERE user_id = $1",
			userID,
		).Scan(&kyc.VerificationID, &kyc.UserID, &kyc.NationalIDUrl, &kyc.HealthCertUrl, &kyc.FaceScanUrl, &kyc.SubmittedAt)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No KYC data found for this user"})
			return
		}

		c.JSON(http.StatusOK, kyc)
	}
}

// --- (getKycFileUrlHandler - เหมือนเดิม) ---
func getKycFileUrlHandler(storageClient *storage.Client, gcsBucketName string, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileKey := c.Query("key")
		if fileKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File key is required"})
			return
		}

		expires := time.Now().Add(10 * time.Minute)
		opts := &storage.SignedURLOptions{
			Scheme:  storage.SigningSchemeV4,
			Method:  "GET",
			Expires: expires,
		}

		url, err := storage.SignedURL(gcsBucketName, fileKey, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create signed URL", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

// --- (approveUserHandler - เหมือนเดิม) ---
func approveUserHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		_, err = dbPool.Exec(ctx, "UPDATE users SET verification_status = 'verified' WHERE user_id = $1", userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve user", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User approved successfully"})
	}
}

// --- (rejectUserHandler - เหมือนเดิม) ---
func rejectUserHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		_, err = dbPool.Exec(ctx, "UPDATE users SET verification_status = 'rejected' WHERE user_id = $1", userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject user", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User rejected successfully"})
	}
}

// --- (adminUpdateUserRolesHandler - เหมือนเดิม) ---
func adminUpdateUserRolesHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var updates struct {
			TierID          *int `json:"tier_id"`
			ProviderLevelID *int `json:"provider_level_id"`
		}

		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if updates.TierID != nil {
			_, err = dbPool.Exec(ctx, "UPDATE users SET tier_id = $1 WHERE user_id = $2", *updates.TierID, userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription tier"})
				return
			}
		}

		if updates.ProviderLevelID != nil {
			_, err = dbPool.Exec(ctx, "UPDATE users SET provider_level_id = $1 WHERE user_id = $2", *updates.ProviderLevelID, userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider level"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "User roles updated successfully"})
	}
}

// --- (adminCreateUserHandler - เหมือนเดิม) ---
func adminCreateUserHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {

		var newUser struct {
			Username           string `json:"username" binding:"required"`
			Email              string `json:"email" binding:"required"`
			Password           string `json:"password" binding:"required"`
			GenderID           int    `json:"gender_id"`
			TierID             int    `json:"tier_id"`
			ProviderLevelID    int    `json:"provider_level_id"`
			VerificationStatus string `json:"verification_status"`
			IsAdmin            bool   `json:"is_admin"`
		}

		if err := c.ShouldBindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		hashedPassword := string(hashedPasswordBytes)

		if newUser.GenderID == 0 {
			newUser.GenderID = 4
		}
		if newUser.TierID == 0 {
			newUser.TierID = 1
		}
		if newUser.ProviderLevelID == 0 {
			newUser.ProviderLevelID = 1
		}
		if newUser.VerificationStatus == "" {
			newUser.VerificationStatus = "unverified"
		}

		var createdUser User
		sqlStatement := `
			INSERT INTO users (
				username, email, password_hash, gender_id, 
				tier_id, provider_level_id, verification_status, is_admin
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING 
				user_id, username, email, gender_id, registration_date, 
				first_name, last_name, tier_id, provider_level_id, 
				verification_status, is_admin, phone_number, google_profile_picture
		`
		err = dbPool.QueryRow(ctx, sqlStatement,
			newUser.Username,
			newUser.Email,
			hashedPassword,
			newUser.GenderID,
			newUser.TierID,
			newUser.ProviderLevelID,
			newUser.VerificationStatus,
			newUser.IsAdmin,
		).Scan(
			&createdUser.UserID, &createdUser.Username, &createdUser.Email,
			&createdUser.GenderID, &createdUser.RegistrationDate,
			&createdUser.FirstName, &createdUser.LastName,
			&createdUser.TierID, &createdUser.ProviderLevelID,
			&createdUser.VerificationStatus, &createdUser.IsAdmin,
			&createdUser.PhoneNumber, &createdUser.GoogleProfilePicture,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
			return
		}

		_, err = dbPool.Exec(ctx, "INSERT INTO user_profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", createdUser.UserID)
		if err != nil {
			log.Printf("Warning: Could not create empty profile for new user %d: %v\n", createdUser.UserID, err)
		}

		c.JSON(http.StatusCreated, createdUser)
	}
}

// --- Provider Queue Info (Admin/GOD only) ---
type ProviderQueueInfo struct {
	ProviderID      int               `json:"provider_id"`
	ActiveBookings  int               `json:"active_bookings"`
	PendingBookings int               `json:"pending_bookings"`
	TotalQueue      int               `json:"total_queue"`
	CurrentLocation *ProviderLocation `json:"current_location,omitempty"`
	IsOnline        bool              `json:"is_online"`
	LastActive      *time.Time        `json:"last_active,omitempty"`
}

type ProviderLocation struct {
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	Province    *string    `json:"province,omitempty"`
	District    *string    `json:"district,omitempty"`
	LastUpdated *time.Time `json:"last_updated,omitempty"`
}

func getProviderQueueInfoHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerID, err := strconv.Atoi(c.Param("providerId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
			return
		}

		var queueInfo ProviderQueueInfo
		queueInfo.ProviderID = providerID

		// Get active bookings count (status = 'confirmed' or 'in_progress')
		err = dbPool.QueryRow(ctx, `
			SELECT COUNT(*) FROM bookings 
			WHERE provider_id = $1 
			AND status IN ('confirmed', 'in_progress', 'accepted')
			AND booking_date >= CURRENT_DATE
		`, providerID).Scan(&queueInfo.ActiveBookings)
		if err != nil {
			log.Printf("Error getting active bookings for provider %d: %v", providerID, err)
			queueInfo.ActiveBookings = 0
		}

		// Get pending bookings count
		err = dbPool.QueryRow(ctx, `
			SELECT COUNT(*) FROM bookings 
			WHERE provider_id = $1 
			AND status = 'pending'
			AND booking_date >= CURRENT_DATE
		`, providerID).Scan(&queueInfo.PendingBookings)
		if err != nil {
			log.Printf("Error getting pending bookings for provider %d: %v", providerID, err)
			queueInfo.PendingBookings = 0
		}

		queueInfo.TotalQueue = queueInfo.ActiveBookings + queueInfo.PendingBookings

		// Get provider location from user_profiles
		var lat, lng *float64
		var province, district *string
		var lastLogin *time.Time
		err = dbPool.QueryRow(ctx, `
			SELECT 
				p.latitude, 
				p.longitude, 
				p.province,
				p.district,
				u.registration_date as last_login
			FROM users u
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			WHERE u.user_id = $1
		`, providerID).Scan(&lat, &lng, &province, &district, &lastLogin)

		if err == nil && lat != nil && lng != nil {
			queueInfo.CurrentLocation = &ProviderLocation{
				Latitude:    *lat,
				Longitude:   *lng,
				Province:    province,
				District:    district,
				LastUpdated: lastLogin,
			}
		}

		// Check if provider is "online" (has bookings today or active recently)
		var recentActivity int
		err = dbPool.QueryRow(ctx, `
			SELECT COUNT(*) FROM bookings 
			WHERE provider_id = $1 
			AND booking_date = CURRENT_DATE
		`, providerID).Scan(&recentActivity)
		queueInfo.IsOnline = recentActivity > 0 || queueInfo.ActiveBookings > 0

		queueInfo.LastActive = lastLogin

		c.JSON(http.StatusOK, queueInfo)
	}
}
