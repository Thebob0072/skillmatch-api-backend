package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func browseUsersHandlerV2(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// 1. หา Access Level ของ "ผู้ใช้ที่กำลังดู" (จาก Subscription Tier)
		var myAccessLevel int
		var myLat, myLon *float64
		err := dbPool.QueryRow(ctx,
			`SELECT t.access_level, p.latitude, p.longitude
			 FROM users u
			 JOIN tiers t ON u.tier_id = t.tier_id 
			 LEFT JOIN user_profiles p ON u.user_id = p.user_id
			 WHERE u.user_id = $1`,
			userID,
		).Scan(&myAccessLevel, &myLat, &myLon)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find user access level"})
			return
		}

		// 2. ดึง Filters จาก Query Params
		genderFilter, _ := strconv.Atoi(c.DefaultQuery("gender", "0")) // 0 = All
		locationFilter := c.DefaultQuery("location", "")               // "" = All (legacy)
		availableOnly := c.DefaultQuery("available", "") == "true"     // เฉพาะที่ว่าง
		minAge, _ := strconv.Atoi(c.DefaultQuery("min_age", "0"))
		maxAge, _ := strconv.Atoi(c.DefaultQuery("max_age", "0"))
		minPrice, _ := strconv.ParseFloat(c.DefaultQuery("min_price", "0"), 64)
		maxPrice, _ := strconv.ParseFloat(c.DefaultQuery("max_price", "0"), 64)
		minRating, _ := strconv.ParseFloat(c.DefaultQuery("min_rating", "0"), 64)
		ethnicity := c.DefaultQuery("ethnicity", "")
		serviceType := c.DefaultQuery("service_type", "") // incall, outcall

		// ฟิลเตอร์ที่อยู่แบบละเอียด
		provinceFilter := c.DefaultQuery("province", "")        // จังหวัด
		districtFilter := c.DefaultQuery("district", "")        // เขต/อำเภอ
		subDistrictFilter := c.DefaultQuery("sub_district", "") // แขวง/ตำบล

		// ฟิลเตอร์ระยะทาง (กิโลเมตร)
		maxDistance, _ := strconv.ParseFloat(c.DefaultQuery("max_distance", "0"), 64) // 0 = ไม่จำกัด

		// 3. สร้าง Query แบบไดนามิก
		sqlArgs := []interface{}{myAccessLevel, userID}
		sqlStatement := `
			SELECT 
				u.user_id, u.username, t_provider.name, u.gender_id,
				p.profile_image_url, u.google_profile_picture,
				p.age, p.location, COALESCE(p.is_available, false) as is_available,
				t_provider.access_level,
				COALESCE(AVG(r.rating), 0) as avg_rating,
				COUNT(DISTINCT r.review_id) as review_count,
				MIN(sp.price) as min_price,
				p.province, p.district, p.sub_district,
				p.latitude, p.longitude, p.service_type
			FROM users u
			JOIN tiers t_provider ON u.provider_level_id = t_provider.tier_id
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			LEFT JOIN reviews r ON u.user_id = r.provider_id
		LEFT JOIN service_packages sp ON u.user_id = sp.provider_id AND sp.is_active = true
		WHERE 
			t_provider.access_level <= $1
			AND u.user_id != $2
			AND u.verification_status IN ('verified', 'approved')
		`

		paramCount := 3

		if genderFilter != 0 {
			sqlStatement += fmt.Sprintf(" AND u.gender_id = $%d", paramCount)
			sqlArgs = append(sqlArgs, genderFilter)
			paramCount++
		}

		if locationFilter != "" {
			sqlStatement += fmt.Sprintf(" AND p.location ILIKE $%d", paramCount)
			sqlArgs = append(sqlArgs, "%"+locationFilter+"%")
			paramCount++
		}

		// ฟิลเตอร์จังหวัด
		if provinceFilter != "" {
			sqlStatement += fmt.Sprintf(" AND p.province = $%d", paramCount)
			sqlArgs = append(sqlArgs, provinceFilter)
			paramCount++
		}

		// ฟิลเตอร์เขต/อำเภอ
		if districtFilter != "" {
			sqlStatement += fmt.Sprintf(" AND p.district = $%d", paramCount)
			sqlArgs = append(sqlArgs, districtFilter)
			paramCount++
		}

		// ฟิลเตอร์แขวง/ตำบล
		if subDistrictFilter != "" {
			sqlStatement += fmt.Sprintf(" AND p.sub_district = $%d", paramCount)
			sqlArgs = append(sqlArgs, subDistrictFilter)
			paramCount++
		}

		if availableOnly {
			sqlStatement += " AND p.is_available = true"
		}

		if minAge > 0 {
			sqlStatement += fmt.Sprintf(" AND p.age >= $%d", paramCount)
			sqlArgs = append(sqlArgs, minAge)
			paramCount++
		}

		if maxAge > 0 {
			sqlStatement += fmt.Sprintf(" AND p.age <= $%d", paramCount)
			sqlArgs = append(sqlArgs, maxAge)
			paramCount++
		}

		if ethnicity != "" {
			sqlStatement += fmt.Sprintf(" AND p.ethnicity = $%d", paramCount)
			sqlArgs = append(sqlArgs, ethnicity)
			paramCount++
		}

		// ฟิลเตอร์ service_type (incall หรือ outcall เท่านั้น)
		if serviceType != "" {
			sqlStatement += fmt.Sprintf(" AND p.service_type = $%d", paramCount)
			sqlArgs = append(sqlArgs, serviceType)
			paramCount++
		}

		sqlStatement += ` 
			GROUP BY u.user_id, u.username, t_provider.name, u.gender_id, 
					 p.profile_image_url, u.google_profile_picture, p.age, 
					 p.location, p.is_available, t_provider.access_level,
					 p.province, p.district, p.sub_district, p.latitude, p.longitude, p.service_type
		`

		// Filter หลัง GROUP BY
		havingClauses := []string{}
		if minRating > 0 {
			havingClauses = append(havingClauses, fmt.Sprintf("AVG(r.rating) >= $%d", paramCount))
			sqlArgs = append(sqlArgs, minRating)
			paramCount++
		}
		if minPrice > 0 {
			havingClauses = append(havingClauses, fmt.Sprintf("MIN(sp.price) >= $%d", paramCount))
			sqlArgs = append(sqlArgs, minPrice)
			paramCount++
		}
		if maxPrice > 0 {
			havingClauses = append(havingClauses, fmt.Sprintf("MIN(sp.price) <= $%d", paramCount))
			sqlArgs = append(sqlArgs, maxPrice)
			paramCount++
		}

		if len(havingClauses) > 0 {
			sqlStatement += " HAVING " + havingClauses[0]
			for i := 1; i < len(havingClauses); i++ {
				sqlStatement += " AND " + havingClauses[i]
			}
		}

		sqlStatement += " ORDER BY p.is_available DESC, avg_rating DESC, t_provider.access_level DESC"

		// 4. Execute query
		rows, err := dbPool.Query(ctx, sqlStatement, sqlArgs...)
		if err != nil {
			println("❌ browseUsersHandlerV2 SQL Error:", err.Error())
			println("🔍 SQL:", sqlStatement)
			println("🔍 Args:", sqlArgs)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "details": err.Error()})
			return
		}
		defer rows.Close()

		// 5. Scan results และคำนวณระยะทาง
		browsableUsers := make([]BrowsableUser, 0)
		for rows.Next() {
			var u BrowsableUser
			var accessLevel int // temporary variable to discard access_level
			if err := rows.Scan(
				&u.UserID, &u.Username, &u.TierName, &u.GenderID,
				&u.ProfileImageUrl, &u.GoogleProfilePicture,
				&u.Age, &u.Location, &u.IsAvailable,
				&accessLevel, // scan but don't use
				&u.AverageRating, &u.ReviewCount, &u.MinPrice,
				&u.Province, &u.District, &u.SubDistrict,
				&u.Latitude, &u.Longitude, &u.ServiceType,
			); err != nil {
				continue
			}

			// คำนวณระยะทาง (ถ้ามีพิกัดครบทั้งสองฝ่าย)
			if hasValidCoordinates(myLat, myLon) && hasValidCoordinates(u.Latitude, u.Longitude) {
				distance := calculateDistance(*myLat, *myLon, *u.Latitude, *u.Longitude)
				u.Distance = &distance

				// ฟิลเตอร์ตามระยะทาง (ถ้ากำหนดมา)
				if maxDistance > 0 && distance > maxDistance {
					continue // ข้ามถ้าไกลเกินกำหนด
				}
			}

			browsableUsers = append(browsableUsers, u)
		}

		c.JSON(http.StatusOK, browsableUsers)
	}
}
