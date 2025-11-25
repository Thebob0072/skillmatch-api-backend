package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func browseUsersHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// 1. หา Access Level ของ "ผู้ใช้ที่กำลังดู" (จาก Subscription Tier)
		var myAccessLevel int
		err := dbPool.QueryRow(ctx,
			`SELECT t.access_level FROM users u
			 JOIN tiers t ON u.tier_id = t.tier_id 
			 WHERE u.user_id = $1`,
			userID,
		).Scan(&myAccessLevel)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find user access level"})
			return
		}

		// 2. ดึง Filters จาก Query Params
		genderFilter, _ := strconv.Atoi(c.DefaultQuery("gender", "0")) // 0 = All
		locationFilter := c.DefaultQuery("location", "")               // "" = All

		// 3. สร้าง Query แบบไดนามิก
		sqlArgs := []interface{}{myAccessLevel, userID}
		sqlStatement := `
			SELECT 
				u.user_id, u.username, t_provider.name, u.gender_id,
				p.profile_image_url, u.google_profile_picture
			FROM users u
			JOIN tiers t_provider ON u.provider_level_id = t_provider.tier_id
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			WHERE 
				t_provider.access_level <= $1     -- (สิทธิ์ของ Provider ต้องน้อยกว่าหรือเท่ากับสิทธิ์ของ Client)
				AND u.user_id != $2                    -- (ไม่แสดงตัวเอง)
				AND u.verification_status = 'verified' -- (ต้องเป็น Provider ที่อนุมัติแล้ว)
		`

		paramCount := 3 // (เพราะ $1 และ $2 ถูกใช้ไปแล้ว)

		if genderFilter != 0 {
			sqlStatement += fmt.Sprintf(" AND u.gender_id = $%d", paramCount)
			sqlArgs = append(sqlArgs, genderFilter)
			paramCount++
		}

		if locationFilter != "" {
			// (ใช้ ILIKE เพื่อค้นหาแบบ case-insensitive และ partial match)
			sqlStatement += fmt.Sprintf(" AND p.location ILIKE $%d", paramCount)
			sqlArgs = append(sqlArgs, "%"+locationFilter+"%")
			paramCount++
		}

		sqlStatement += " ORDER BY t_provider.access_level DESC, u.registration_date DESC"

		// 4. Execute query
		rows, err := dbPool.Query(ctx, sqlStatement, sqlArgs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "details": err.Error()})
			return
		}
		defer rows.Close()

		// 5. Scan results
		browsableUsers := make([]BrowsableUser, 0)
		for rows.Next() {
			var u BrowsableUser
			if err := rows.Scan(
				&u.UserID, &u.Username, &u.TierName, &u.GenderID,
				&u.ProfileImageUrl, &u.GoogleProfilePicture,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user row", "details": err.Error()})
				return
			}
			browsableUsers = append(browsableUsers, u)
		}

		c.JSON(http.StatusOK, browsableUsers)
	}
}
