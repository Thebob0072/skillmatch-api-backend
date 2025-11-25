package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// --- GOD Statistics Handler ---
func getGodStatsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stats struct {
			TotalUsers          int     `json:"total_users"`
			TotalProviders      int     `json:"total_providers"`
			TotalAdmins         int     `json:"total_admins"`
			PendingVerification int     `json:"pending_verification"`
			TotalBookings       int     `json:"total_bookings"`
			TotalRevenue        float64 `json:"total_revenue"`
			ActiveUsers24h      int     `json:"active_users_24h"`
			NewUsersToday       int     `json:"new_users_today"`
		}

		// Total Users (non-admin)
		err := dbPool.QueryRow(ctx,
			"SELECT COUNT(*) FROM users WHERE is_admin = false",
		).Scan(&stats.TotalUsers)
		if err != nil {
			stats.TotalUsers = 0
		}

		// Total Providers (users with service packages or provider_level_id > 1)
		err = dbPool.QueryRow(ctx,
			`SELECT COUNT(DISTINCT u.user_id) 
			 FROM users u 
			 WHERE u.provider_level_id > 1 
			    OR EXISTS (SELECT 1 FROM service_packages sp WHERE sp.provider_id = u.user_id)`,
		).Scan(&stats.TotalProviders)
		if err != nil {
			stats.TotalProviders = 0
		}

		// Total Admins
		err = dbPool.QueryRow(ctx,
			"SELECT COUNT(*) FROM users WHERE is_admin = true",
		).Scan(&stats.TotalAdmins)
		if err != nil {
			stats.TotalAdmins = 0
		}

		// Pending Verification
		err = dbPool.QueryRow(ctx,
			"SELECT COUNT(*) FROM users WHERE verification_status = 'pending'",
		).Scan(&stats.PendingVerification)
		if err != nil {
			stats.PendingVerification = 0
		}

		// Total Bookings
		err = dbPool.QueryRow(ctx,
			"SELECT COUNT(*) FROM bookings",
		).Scan(&stats.TotalBookings)
		if err != nil {
			stats.TotalBookings = 0
		}

		// Total Revenue (sum of completed bookings)
		err = dbPool.QueryRow(ctx,
			`SELECT COALESCE(SUM(price), 0) 
			 FROM bookings 
			 WHERE status IN ('completed', 'confirmed')`,
		).Scan(&stats.TotalRevenue)
		if err != nil {
			stats.TotalRevenue = 0
		}

		// Active Users in last 24 hours (users who logged in or made actions)
		// For now, we'll count users registered in last 24h as proxy
		err = dbPool.QueryRow(ctx,
			`SELECT COUNT(*) 
			 FROM users 
			 WHERE registration_date >= NOW() - INTERVAL '24 hours'`,
		).Scan(&stats.ActiveUsers24h)
		if err != nil {
			stats.ActiveUsers24h = 0
		}

		// New Users Today
		err = dbPool.QueryRow(ctx,
			`SELECT COUNT(*) 
			 FROM users 
			 WHERE DATE(registration_date) = CURRENT_DATE`,
		).Scan(&stats.NewUsersToday)
		if err != nil {
			stats.NewUsersToday = 0
		}

		c.JSON(http.StatusOK, stats)
	}
}

// --- List All Users (GOD/Admin) ---
func listAllUsersHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 50
		}
		offset := (page - 1) * limit

		// Filters
		isAdmin := c.Query("is_admin")
		verificationStatus := c.Query("verification_status")
		searchQuery := c.Query("search") // username or email

		// Build query
		sqlStatement := `
			SELECT 
				u.user_id, u.username, u.email, u.gender_id, u.registration_date, 
				u.first_name, u.last_name,
				u.verification_status, u.tier_id, u.provider_level_id, 
				u.phone_number, u.is_admin,
				u.google_profile_picture, p.profile_image_url, p.age,
				t.name as tier_name
			FROM users u
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			LEFT JOIN tiers t ON u.tier_id = t.tier_id
			WHERE 1=1
		`
		args := []interface{}{}
		argCount := 1

		if isAdmin != "" {
			sqlStatement += " AND u.is_admin = $" + strconv.Itoa(argCount)
			args = append(args, isAdmin == "true")
			argCount++
		}

		if verificationStatus != "" {
			sqlStatement += " AND u.verification_status = $" + strconv.Itoa(argCount)
			args = append(args, verificationStatus)
			argCount++
		}

		if searchQuery != "" {
			sqlStatement += " AND (u.username ILIKE $" + strconv.Itoa(argCount) +
				" OR u.email ILIKE $" + strconv.Itoa(argCount) + ")"
			args = append(args, "%"+searchQuery+"%")
			argCount++
		}

		sqlStatement += " ORDER BY u.registration_date DESC LIMIT $" +
			strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
		args = append(args, limit, offset)

		rows, err := dbPool.Query(ctx, sqlStatement, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Database query failed",
				"details": err.Error(),
			})
			return
		}
		defer rows.Close()

		users := make([]map[string]interface{}, 0)
		for rows.Next() {
			var u User
			var tierName *string

			if err := rows.Scan(
				&u.UserID, &u.Username, &u.Email, &u.GenderID, &u.RegistrationDate,
				&u.FirstName, &u.LastName, &u.VerificationStatus, &u.TierID,
				&u.ProviderLevelID, &u.PhoneNumber, &u.IsAdmin,
				&u.GoogleProfilePicture, &u.ProfileImageUrl, &u.Age,
				&tierName,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to scan user row",
					"details": err.Error(),
				})
				return
			}

			userMap := map[string]interface{}{
				"user_id":                u.UserID,
				"username":               u.Username,
				"email":                  u.Email,
				"gender_id":              u.GenderID,
				"subscription_tier_id":   u.TierID,
				"provider_level_id":      u.ProviderLevelID,
				"is_admin":               u.IsAdmin,
				"verification_status":    u.VerificationStatus,
				"first_name":             u.FirstName,
				"last_name":              u.LastName,
				"phone_number":           u.PhoneNumber,
				"registration_date":      u.RegistrationDate,
				"google_profile_picture": u.GoogleProfilePicture,
				"profile_image_url":      u.ProfileImageUrl,
				"age":                    u.Age,
				"tier_name":              tierName,
			}
			users = append(users, userMap)
		}

		// Get total count
		var totalCount int
		countSQL := "SELECT COUNT(*) FROM users WHERE 1=1"
		countArgs := []interface{}{}
		countArgNum := 1

		if isAdmin != "" {
			countSQL += " AND is_admin = $" + strconv.Itoa(countArgNum)
			countArgs = append(countArgs, isAdmin == "true")
			countArgNum++
		}

		if verificationStatus != "" {
			countSQL += " AND verification_status = $" + strconv.Itoa(countArgNum)
			countArgs = append(countArgs, verificationStatus)
			countArgNum++
		}

		if searchQuery != "" {
			countSQL += " AND (username ILIKE $" + strconv.Itoa(countArgNum) +
				" OR email ILIKE $" + strconv.Itoa(countArgNum) + ")"
			countArgs = append(countArgs, "%"+searchQuery+"%")
		}

		err = dbPool.QueryRow(ctx, countSQL, countArgs...).Scan(&totalCount)
		if err != nil {
			totalCount = 0
		}

		c.JSON(http.StatusOK, gin.H{
			"users":       users,
			"total":       totalCount,
			"page":        page,
			"limit":       limit,
			"total_pages": (totalCount + limit - 1) / limit,
		})
	}
}

// --- Update User Role/Tier (GOD only) ---
// Renamed from switchUserRoleHandler to updateUserHandler for clarity
func updateUserHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if requester is GOD
		requesterID, _ := c.Get("userID")
		var requesterTierID int
		var requesterIsAdmin bool
		err := dbPool.QueryRow(ctx,
			"SELECT tier_id, is_admin FROM users WHERE user_id = $1",
			requesterID,
		).Scan(&requesterTierID, &requesterIsAdmin)

		if err != nil || !requesterIsAdmin || requesterTierID != 5 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Only GOD can switch user roles",
			})
			return
		}

		var req struct {
			UserID             int     `json:"user_id" binding:"required"`
			IsAdmin            *bool   `json:"is_admin"`
			TierID             *int    `json:"tier_id"`
			ProviderLevelID    *int    `json:"provider_level_id"`
			VerificationStatus *string `json:"verification_status"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Prevent modifying GOD account (user_id = 1)
		if req.UserID == 1 && requesterID.(int) != 1 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Cannot modify GOD account",
			})
			return
		}

		// Build update query dynamically
		updates := []string{}
		args := []interface{}{}
		argCount := 1

		if req.IsAdmin != nil {
			updates = append(updates, "is_admin = $"+strconv.Itoa(argCount))
			args = append(args, *req.IsAdmin)
			argCount++
		}

		if req.TierID != nil {
			updates = append(updates, "tier_id = $"+strconv.Itoa(argCount))
			args = append(args, *req.TierID)
			argCount++
		}

		if req.ProviderLevelID != nil {
			updates = append(updates, "provider_level_id = $"+strconv.Itoa(argCount))
			args = append(args, *req.ProviderLevelID)
			argCount++
		}

		if req.VerificationStatus != nil {
			updates = append(updates, "verification_status = $"+strconv.Itoa(argCount))
			args = append(args, *req.VerificationStatus)
			argCount++
		}

		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No fields to update",
			})
			return
		}

		// Add user_id to args
		args = append(args, req.UserID)

		sqlStatement := "UPDATE users SET " +
			updates[0]
		for i := 1; i < len(updates); i++ {
			sqlStatement += ", " + updates[i]
		}
		sqlStatement += " WHERE user_id = $" + strconv.Itoa(argCount)

		_, err = dbPool.Exec(ctx, sqlStatement, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update user",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User role updated successfully",
			"user_id": req.UserID,
		})
	}
}

// --- Create Admin (GOD only) ---
func createAdminHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if requester is GOD
		requesterID, _ := c.Get("userID")
		var requesterTierID int
		var requesterIsAdmin bool
		err := dbPool.QueryRow(ctx,
			"SELECT tier_id, is_admin FROM users WHERE user_id = $1",
			requesterID,
		).Scan(&requesterTierID, &requesterIsAdmin)

		if err != nil || !requesterIsAdmin || requesterTierID != 5 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Only GOD can create admins",
			})
			return
		}

		var req struct {
			Username  string `json:"username" binding:"required"`
			Email     string `json:"email" binding:"required,email"`
			Password  string `json:"password" binding:"required,min=8"`
			AdminType string `json:"admin_type"` // user_manager, provider_manager
			TierID    int    `json:"tier_id"`
			GenderID  int    `json:"gender_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Default tier for admins
		if req.TierID == 0 {
			req.TierID = 2 // Silver
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to hash password",
			})
			return
		}

		// Insert new admin
		var newUserID int
		err = dbPool.QueryRow(ctx,
			`INSERT INTO users 
			 (username, email, password_hash, gender_id, tier_id, is_admin, verification_status) 
			 VALUES ($1, $2, $3, $4, $5, true, 'verified') 
			 RETURNING user_id`,
			req.Username, req.Email, string(hashedPassword), req.GenderID, req.TierID,
		).Scan(&newUserID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create admin",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Admin created successfully",
			"user_id":    newUserID,
			"admin_type": req.AdminType,
		})
	}
}

// --- Delete Any User (GOD only) ---
func deleteUserHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if requester is GOD
		requesterID, _ := c.Get("userID")
		var requesterTierID int
		var requesterIsAdmin bool
		err := dbPool.QueryRow(ctx,
			"SELECT tier_id, is_admin FROM users WHERE user_id = $1",
			requesterID,
		).Scan(&requesterTierID, &requesterIsAdmin)

		if err != nil || !requesterIsAdmin || requesterTierID != 5 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Only GOD can delete users",
			})
			return
		}

		userIDStr := c.Param("user_id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Cannot delete GOD account (unless self-deletion)
		if userID == 1 && requesterID.(int) != 1 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Cannot delete GOD account",
			})
			return
		}

		// Get user info before deletion
		var username, email string
		var isAdmin bool
		err = dbPool.QueryRow(ctx,
			"SELECT username, email, is_admin FROM users WHERE user_id = $1",
			userID,
		).Scan(&username, &email, &isAdmin)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		// Delete user (CASCADE will handle related data)
		result, err := dbPool.Exec(ctx,
			"DELETE FROM users WHERE user_id = $1",
			userID,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete user",
				"details": err.Error(),
			})
			return
		}

		if result.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "User deleted successfully",
			"user_id":  userID,
			"username": username,
			"email":    email,
			"was_admin": isAdmin,
		})
	}
}

// --- Delete Admin (GOD only) ---
func deleteAdminHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if requester is GOD
		requesterID, _ := c.Get("userID")
		var requesterTierID int
		var requesterIsAdmin bool
		err := dbPool.QueryRow(ctx,
			"SELECT tier_id, is_admin FROM users WHERE user_id = $1",
			requesterID,
		).Scan(&requesterTierID, &requesterIsAdmin)

		if err != nil || !requesterIsAdmin || requesterTierID != 5 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Only GOD can delete admins",
			})
			return
		}

		userIDStr := c.Param("user_id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Cannot delete GOD account
		if userID == 1 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Cannot delete GOD account",
			})
			return
		}

		// Delete user
		result, err := dbPool.Exec(ctx,
			"DELETE FROM users WHERE user_id = $1 AND is_admin = true",
			userID,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete admin",
				"details": err.Error(),
			})
			return
		}

		if result.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Admin not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Admin deleted successfully",
			"user_id": userID,
		})
	}
}

// --- List All Admins (GOD only) ---
func listAdminsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := dbPool.Query(ctx,
			`SELECT 
				u.user_id, u.username, u.email, u.is_admin, u.tier_id, 
				u.registration_date, t.name as tier_name
			 FROM users u
			 LEFT JOIN tiers t ON u.tier_id = t.tier_id
			 WHERE u.is_admin = true
			 ORDER BY u.tier_id DESC, u.registration_date ASC`,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Database query failed",
				"details": err.Error(),
			})
			return
		}
		defer rows.Close()

		admins := make([]map[string]interface{}, 0)
		for rows.Next() {
			var userID, tierID int
			var username, email string
			var isAdmin bool
			var registrationDate time.Time
			var tierName *string

			if err := rows.Scan(
				&userID, &username, &email, &isAdmin, &tierID,
				&registrationDate, &tierName,
			); err != nil {
				continue
			}

			adminType := "admin"
			if tierID == 5 {
				adminType = "god"
			}

			admins = append(admins, map[string]interface{}{
				"user_id":    userID,
				"username":   username,
				"email":      email,
				"is_admin":   isAdmin,
				"tier_id":    tierID,
				"tier_name":  tierName,
				"admin_type": adminType,
				"created_at": registrationDate,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"admins": admins,
			"total":  len(admins),
		})
	}
}

// --- GOD View Mode Switching (UI Simulation) ---
// This does NOT modify the database or user's actual role
// It's for GOD to preview UI as different roles while remaining GOD
var godViewModes = make(map[int]string) // userID -> mode (user/provider/admin)

func setGodViewModeHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if requester is GOD
		requesterID, _ := c.Get("userID")
		var requesterTierID int
		var requesterIsAdmin bool
		err := dbPool.QueryRow(ctx,
			"SELECT tier_id, is_admin FROM users WHERE user_id = $1",
			requesterID,
		).Scan(&requesterTierID, &requesterIsAdmin)

		if err != nil || !requesterIsAdmin || requesterTierID != 5 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Only GOD can switch view modes",
			})
			return
		}

		var req struct {
			Mode string `json:"mode" binding:"required"` // user, provider, admin, god
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate mode
		validModes := map[string]bool{
			"user":     true,
			"provider": true,
			"admin":    true,
			"god":      true,
		}

		if !validModes[req.Mode] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid mode. Must be: user, provider, admin, or god",
			})
			return
		}

		// Store view mode in memory (or could use Redis for production)
		userID := requesterID.(int)
		godViewModes[userID] = req.Mode

		c.JSON(http.StatusOK, gin.H{
			"message":      "View mode updated successfully",
			"current_mode": req.Mode,
			"note":         "You are still GOD. This only affects UI display.",
			"actual_role": gin.H{
				"is_admin": true,
				"tier_id":  5,
			},
		})
	}
}

// --- Get Current GOD View Mode ---
func getGodViewModeHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if requester is GOD
		requesterID, _ := c.Get("userID")
		var requesterTierID int
		var requesterIsAdmin bool
		err := dbPool.QueryRow(ctx,
			"SELECT tier_id, is_admin FROM users WHERE user_id = $1",
			requesterID,
		).Scan(&requesterTierID, &requesterIsAdmin)

		if err != nil || !requesterIsAdmin || requesterTierID != 5 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Only GOD can access view modes",
			})
			return
		}

		userID := requesterID.(int)
		currentMode := godViewModes[userID]
		if currentMode == "" {
			currentMode = "god" // Default to god view
		}

		c.JSON(http.StatusOK, gin.H{
			"current_mode": currentMode,
			"actual_role": gin.H{
				"is_admin": true,
				"tier_id":  5,
			},
			"available_modes": []string{"user", "provider", "admin", "god"},
		})
	}
}
