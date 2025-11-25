package main

import (
"context"
"net/http"
"strconv"

"github.com/gin-gonic/gin"
"github.com/jackc/pgx/v5/pgxpool"
"github.com/lib/pq" // (จำเป็นสำหรับ .Scan() array)
)

// --- Handler: GET /users/:id (Used by getMeHandler) ---
// (ดึงข้อมูลทั้งหมดของ User เพื่อใช้ใน AuthContext)
func getUserHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		userID, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		println("🔍 getUserHandler - Querying for UserID:", userID)

		var foundUser User // (ใช้ User struct จาก models.go)

		// SQL นี้จะ JOIN ตาราง users และ user_profiles
		sqlStatement := `
			SELECT 
				u.user_id, u.username, u.email, u.gender_id, u.registration_date, 
				u.first_name, u.last_name,
				u.verification_status, 
				u.tier_id,           -- (Subscription Tier)
				u.provider_level_id, -- (Provider Level)
				u.is_admin, 
				u.phone_number, 
				u.google_profile_picture,
				p.bio,
				p.location,
				p.skills,
				p.profile_image_url   -- (รูปโปรไฟล์ที่อัปโหลดเอง)
				
			FROM users u
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			WHERE u.user_id = $1
		`

		// .Scan() ต้องตรงกับลำดับของ SELECT (17 fields)
		err = dbPool.QueryRow(ctx, sqlStatement, userID).Scan(
&foundUser.UserID, &foundUser.Username, &foundUser.Email,
			&foundUser.GenderID, &foundUser.RegistrationDate,
			&foundUser.FirstName, &foundUser.LastName,
			&foundUser.VerificationStatus,
			&foundUser.TierID,
			&foundUser.ProviderLevelID,
			&foundUser.IsAdmin,
			&foundUser.PhoneNumber,
			&foundUser.GoogleProfilePicture,
			&foundUser.Bio,                       // (from user_profiles)
			&foundUser.Location,                  // (from user_profiles)
			(*pq.StringArray)(&foundUser.Skills), // (from user_profiles)
			&foundUser.ProfileImageUrl,           // (from user_profiles)
		)

		if err != nil {
			// Log the full error for debugging
			println("❌ getUserHandler SQL Error:", err.Error())
			println("🔍 UserID:", userID)
			println("🔍 SQL:", sqlStatement)

			if err.Error() == "no rows in result set" {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "details": err.Error()})
			return
		}

		println("✅ getUserHandler - Successfully fetched user:", userID)

		if foundUser.Skills == nil {
			foundUser.Skills = make([]string, 0)
		}

		c.JSON(http.StatusOK, foundUser)
	}
}

// --- Handler: GET /users/me ---
func getMeHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {

		userIDInterface, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context (auth middleware error)"})
			return
		}

		userID, ok := userIDInterface.(int)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format in context"})
			return
		}

		println("🔍 getMeHandler - UserID from token:", userID)

		c.Params = append(c.Params, gin.Param{Key: "id", Value: strconv.Itoa(userID)})
		getUserHandler(dbPool, ctx)(c)
	}
}
