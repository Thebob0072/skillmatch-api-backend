package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

// (Helper: ดึงข้อมูล Profile หรือสร้างถ้ายังไม่มี)
func getProfileByUserID(dbPool *pgxpool.Pool, ctx context.Context, userID int) (User, error) {
	var user User

	// (INSERT ON CONFLICT ก่อน เพื่อให้แน่ใจว่ามีแถวข้อมูล)
	_, err := dbPool.Exec(ctx, "INSERT INTO user_profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)
	if err != nil {
		return user, err
	}

	err = dbPool.QueryRow(ctx,
		`SELECT 
			u.user_id,
			u.username,
			u.email,
			u.gender_id,
			u.tier_id,
			COALESCE(t.name, 'General') as tier_name,
			u.provider_level_id,
			u.registration_date,
			u.phone_number,
			u.verification_status,
			u.is_admin,
			u.first_name,
			u.last_name,
			u.google_profile_picture,
			COALESCE(u.profile_picture_url, u.google_profile_picture) as profile_picture_url,
			p.profile_image_url,
			p.bio,
			p.location,
			p.age
		FROM users u
		LEFT JOIN user_profiles p ON u.user_id = p.user_id
		LEFT JOIN tiers t ON u.tier_id = t.tier_id
		WHERE u.user_id = $1`,
		userID,
	).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.GenderID,
		&user.TierID,
		&user.TierName,
		&user.ProviderLevelID,
		&user.RegistrationDate,
		&user.PhoneNumber,
		&user.VerificationStatus,
		&user.IsAdmin,
		&user.FirstName,
		&user.LastName,
		&user.GoogleProfilePicture,
		&user.ProfilePictureURL,
		&user.ProfileImageUrl,
		&user.Bio,
		&user.Location,
		&user.Age,
	)

	if err != nil {
		log.Printf("❌ getProfileByUserID SQL Error for UserID %d: %v", userID, err)
		return user, err
	}

	// Initialize empty skills array (workaround for pq array scanning issue)
	user.Skills = []string{}

	// Debug logging
	log.Printf("✅ Profile fetched - UserID: %d, Username: %s, Email: %s, TierID: %d, IsAdmin: %v",
		user.UserID, user.Username, user.Email, user.TierID, user.IsAdmin)

	if user.Skills == nil {
		user.Skills = make([]string, 0)
	}

	return user, nil
}

// --- Handler: GET /profile/me ---
func getMyProfileHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		profile, err := getProfileByUserID(dbPool, ctx, userID.(int))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, profile)
	}
}

// --- Handler: PUT /profile/me ---
func updateMyProfileHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var updatedProfile struct {
			Bio          *string  `json:"bio"`
			Location     *string  `json:"location"` // Legacy
			Skills       []string `json:"skills"`
			Province     *string  `json:"province"`
			District     *string  `json:"district"`
			SubDistrict  *string  `json:"sub_district"`
			PostalCode   *string  `json:"postal_code"`
			AddressLine1 *string  `json:"address_line1"`
			Latitude     *float64 `json:"latitude"`
			Longitude    *float64 `json:"longitude"`
			ServiceType  *string  `json:"service_type"` // incall, outcall
		}
		if err := c.ShouldBindJSON(&updatedProfile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate service_type
		if updatedProfile.ServiceType != nil {
			validTypes := map[string]bool{"incall": true, "outcall": true}
			if !validTypes[*updatedProfile.ServiceType] {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid service_type",
					"message": "service_type must be 'incall' or 'outcall'",
				})
				return
			}
		}

		sqlStatement := `
			UPDATE user_profiles
			SET bio = $1, location = $2, skills = $3, 
			    province = $4, district = $5, sub_district = $6, 
			    postal_code = $7, address_line1 = $8,
			    latitude = $9, longitude = $10, service_type = $11,
			    updated_at = NOW()
			WHERE user_id = $12
		`
		_, err := dbPool.Exec(ctx, sqlStatement,
			updatedProfile.Bio,
			updatedProfile.Location,
			(*pq.StringArray)(&updatedProfile.Skills),
			updatedProfile.Province,
			updatedProfile.District,
			updatedProfile.SubDistrict,
			updatedProfile.PostalCode,
			updatedProfile.AddressLine1,
			updatedProfile.Latitude,
			updatedProfile.Longitude,
			updatedProfile.ServiceType,
			userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
	}
}
