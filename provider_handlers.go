package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// The PublicProfile struct is now correctly defined only in models.go

// --- Handler: GET /provider/:userId ---
// (ดึงข้อมูลโปรไฟล์สาธารณะของ Provider)
func getPublicProfileHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. ดึง UserID จาก URL
		userID, err := strconv.Atoi(c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Use the PublicProfile struct from models.go
		var profile PublicProfile

		// 2. (SQL หัวใจหลัก) JOIN 3 ตาราง (users, tiers, user_profiles)
		// Note: Both 'approved' and 'verified' users can be viewed as providers
		// ซ่อนข้อมูลที่บ่งบอกว่าขายบริการ (age, height, weight, ethnicity, languages, working_hours, service_type)
		sqlStatement := `
			SELECT 
				u.user_id, u.username, u.gender_id, t.name,
				p.bio, p.location, COALESCE(p.skills, '{}'), p.profile_image_url,
				u.google_profile_picture, COALESCE(p.is_available, false),
				p.province, p.district, p.sub_district,
				COALESCE(AVG(r.rating), 0) as avg_rating,
				COUNT(DISTINCT r.review_id) as review_count
			FROM users u
			-- Join เพื่อดึงชื่อ Tier (ใช้ provider_level_id)
			LEFT JOIN tiers t ON u.provider_level_id = t.tier_id
			-- Join เพื่อดึงรายละเอียดที่ผู้ใช้กรอกเอง
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			-- Join เพื่อดึงรีวิว
			LEFT JOIN reviews r ON u.user_id = r.provider_id
			-- เงื่อนไขสำคัญ: ต้อง verified หรือ approved
			WHERE u.user_id = $1 AND u.verification_status IN ('verified', 'approved')
			GROUP BY u.user_id, u.username, u.gender_id, t.name, p.bio, p.location, 
			         p.skills, p.profile_image_url, u.google_profile_picture, p.is_available,
			         p.province, p.district, p.sub_district
		`
		err = dbPool.QueryRow(ctx, sqlStatement, userID).Scan(
			&profile.UserID, &profile.Username, &profile.GenderID, &profile.TierName,
			&profile.Bio, &profile.Location, &profile.Skills, &profile.ProfileImageUrl,
			&profile.GoogleProfilePicture, &profile.IsAvailable,
			&profile.Province, &profile.District, &profile.SubDistrict,
			&profile.AverageRating, &profile.ReviewCount,
		)

		if err != nil {
			println("❌ getPublicProfileHandler SQL Error:", err.Error())
			println("🔍 UserID:", userID)
			if err.Error() == "no rows in result set" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found or not verified"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "details": err.Error()})
			return
		}
		if profile.Skills == nil {
			profile.Skills = make([]string, 0)
		}

		c.JSON(http.StatusOK, profile)
	}
}

// --- Handler: GET /provider/:userId (Authenticated - Full Details) ---
// (ดึงข้อมูลโปรไฟล์เต็มรูปแบบสำหรับผู้ใช้ที่ login แล้ว)
func getAuthenticatedProfileHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. ดึง UserID จาก URL
		userID, err := strconv.Atoi(c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Use FullProfile struct from models.go (includes sensitive data)
		var profile FullProfile

		// SQL: ดึงข้อมูลเต็มรูปแบบ (รวม Age, Height, Weight, ServiceType, etc.)
		sqlStatement := `
			SELECT 
				u.user_id, u.username, u.gender_id, t.name,
				p.bio, p.location, COALESCE(p.skills, '{}'), p.profile_image_url,
				u.google_profile_picture, COALESCE(p.is_available, false),
				p.province, p.district, p.sub_district,
				p.age, p.height, p.weight, p.ethnicity,
				COALESCE(p.languages, '{}'), p.working_hours, p.service_type,
				p.address_line1, p.latitude, p.longitude,
				COALESCE(AVG(r.rating), 0) as avg_rating,
				COUNT(DISTINCT r.review_id) as review_count
			FROM users u
			LEFT JOIN tiers t ON u.provider_level_id = t.tier_id
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			LEFT JOIN reviews r ON u.user_id = r.provider_id
			WHERE u.user_id = $1 AND u.verification_status IN ('verified', 'approved')
			GROUP BY u.user_id, u.username, u.gender_id, t.name, p.bio, p.location, 
			         p.skills, p.profile_image_url, u.google_profile_picture, p.is_available,
			         p.province, p.district, p.sub_district, p.age, p.height, p.weight,
			         p.ethnicity, p.languages, p.working_hours, p.service_type,
			         p.address_line1, p.latitude, p.longitude
		`
		err = dbPool.QueryRow(ctx, sqlStatement, userID).Scan(
			&profile.UserID, &profile.Username, &profile.GenderID, &profile.TierName,
			&profile.Bio, &profile.Location, &profile.Skills, &profile.ProfileImageUrl,
			&profile.GoogleProfilePicture, &profile.IsAvailable,
			&profile.Province, &profile.District, &profile.SubDistrict,
			&profile.Age, &profile.Height, &profile.Weight, &profile.Ethnicity,
			&profile.Languages, &profile.WorkingHours, &profile.ServiceType,
			&profile.AddressLine1, &profile.Latitude, &profile.Longitude,
			&profile.AverageRating, &profile.ReviewCount,
		)

		if err != nil {
			println("❌ getAuthenticatedProfileHandler SQL Error:", err.Error())
			if err.Error() == "no rows in result set" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found or not verified"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "details": err.Error()})
			return
		}
		if profile.Skills == nil {
			profile.Skills = make([]string, 0)
		}
		if profile.Languages == nil {
			profile.Languages = make([]string, 0)
		}

		c.JSON(http.StatusOK, profile)
	}
}

// --- Handler: GET /provider/:userId/photos ---
// (ดึงแกลเลอรีรูปภาพของ Provider)
func getProviderPhotosHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// UserPhoto is assumed to be available from models.go
		photos := make([]UserPhoto, 0)
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
