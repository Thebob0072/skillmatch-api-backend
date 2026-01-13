package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	google_oauth "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

// --- Google OAuth Config ---
// Reads configuration from environment variables (called at runtime after .env is loaded)
func getGoogleOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "postmessage", // For client-side OAuth flow (Google Sign-In button)
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

// --- Helper: Create JWT ---
// Uses 'jwtKey' which is defined in middleware.go
func createJWT(userID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   strconv.Itoa(userID),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// --- Handler: POST /register ---
// Handles standard email/password registration (supports both client and provider roles)
func createUserHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newUser struct {
			Username  string  `json:"username" binding:"required"`
			Email     string  `json:"email" binding:"required"`
			Password  string  `json:"password" binding:"required"`
			Role      string  `json:"role"`      // "client" or "provider", defaults to "client"
			GenderID  *int    `json:"gender_id"` // Optional, defaults to 4 (Prefer not to say)
			FirstName *string `json:"first_name"`
			LastName  *string `json:"last_name"`
			Birthdate *string `json:"birthdate"` // Optional birthdate in YYYY-MM-DD format
		}

		if err := c.ShouldBindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "error_code": "VALIDATION_ERROR"})
			return
		}

		// Validate and set role (default to "client" if not provided)
		role := "client"
		if newUser.Role != "" {
			if newUser.Role != "client" && newUser.Role != "provider" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'client' or 'provider'", "error_code": "INVALID_ROLE"})
				return
			}
			role = newUser.Role
		}

		// Hash the password
		hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		hashedPassword := string(hashedPasswordBytes)

		// Use default gender_id if not provided
		genderID := 4 // Default to 'Prefer not to say'
		if newUser.GenderID != nil {
			genderID = *newUser.GenderID
		}

		// Determine is_provider flag and user_type
		isProvider := (role == "provider")
		userType := role // "client" or "provider"

		// SQL for creating a new user (support both client and provider)
		// Note: tier_id and provider_level_id use the database DEFAULT (1)
		sqlStatement := `
			INSERT INTO users (username, email, password_hash, gender_id, first_name, last_name, birthdate, is_provider, provider_verification_status, user_type)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::user_type_enum)
			RETURNING user_id
		`

		// Set provider_verification_status
		// For clients, use "unverified" (they don't need provider verification)
		// For providers, use "pending" (waiting for verification)
		providerStatus := "unverified"
		if isProvider {
			providerStatus = "pending"
		}

		var userID int
		err = dbPool.QueryRow(ctx, sqlStatement,
			newUser.Username,
			newUser.Email,
			&hashedPassword,
			genderID,
			newUser.FirstName,
			newUser.LastName,
			newUser.Birthdate,
			isProvider,
			providerStatus,
			userType,
		).Scan(&userID)

		if err != nil {
			// Check for duplicate email
			if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "unique constraint") {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already registered", "error_code": "DUPLICATE_EMAIL"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error(), "error_code": "USER_CREATE_ERROR"})
			return
		}

		// Also create an empty profile row for them
		_, err = dbPool.Exec(ctx, "INSERT INTO user_profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)
		if err != nil {
			log.Printf("Warning: Could not create empty profile for new user %d: %v\n", userID, err)
		}

		// If provider, create provider profile
		if isProvider {
			_, err = dbPool.Exec(ctx, `
				INSERT INTO provider_profiles (user_id, service_type, bio, available)
				VALUES ($1, 'both', '', true)
				ON CONFLICT (user_id) DO NOTHING
			`, userID)
			if err != nil {
				log.Printf("Warning: Could not create provider profile for new user %d: %v\n", userID, err)
			}

			// Create provider wallet
			_, err = dbPool.Exec(ctx, `
				INSERT INTO provider_wallets (user_id, balance, total_earned)
				VALUES ($1, 0.00, 0.00)
				ON CONFLICT (user_id) DO NOTHING
			`, userID)
			if err != nil {
				log.Printf("Warning: Could not create provider wallet for new user %d: %v\n", userID, err)
			}
		}

		// Create JWT token for automatic login after registration
		tokenString, err := createJWT(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User created but failed to generate token", "user_id": userID})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User created successfully",
			"user_id": userID,
			"role":    role,
			"token":   tokenString,
		})
	}
}

// --- Handler: POST /login ---
// Handles standard email/password login
func loginHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginDetails struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&loginDetails); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email/Password cannot be empty", "error_code": "MISSING_CREDENTIALS"})
			return
		}

		var storedPasswordHash *string // Can be NULL
		var userID int
		sqlStatement := `SELECT user_id, password_hash FROM users WHERE email = $1`

		err := dbPool.QueryRow(ctx, sqlStatement, loginDetails.Email).Scan(&userID, &storedPasswordHash)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password", "error_code": "INVALID_CREDENTIALS"})
			return
		}

		// Check if user has a password (they might be Google-only)
		if storedPasswordHash == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "This account uses Google login. Please use the 'Login with Google' button.", "error_code": "GOOGLE_LOGIN_REQUIRED"})
			return
		}

		// Compare password with hash
		err = bcrypt.CompareHashAndPassword([]byte(*storedPasswordHash), []byte(loginDetails.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password", "error_code": "INVALID_CREDENTIALS"})
			return
		}

		// Login Success: Create JWT
		tokenString, err := createJWT(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token", "error_code": "TOKEN_CREATE_ERROR"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   tokenString,
			"user_id": userID,
		})
	}
}

// --- Handler: POST /auth/set-password ---
// Allows users (especially Google OAuth users) to set/update their password
func setPasswordHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from JWT token
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req struct {
			NewPassword string `json:"new_password" binding:"required,min=8"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long"})
			return
		}

		// Hash the new password
		hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		hashedPassword := string(hashedPasswordBytes)

		// Update password in database
		sqlStatement := `UPDATE users SET password_hash = $1 WHERE user_id = $2`
		_, err = dbPool.Exec(ctx, sqlStatement, hashedPassword, userID)
		if err != nil {
			log.Printf("Failed to update password for user %v: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Password set successfully. You can now login with email and password.",
		})
	}
}

// --- Handler: POST /auth/google ---
// Handles login/registration via Google OAuth
func handleGoogleCallback(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody struct {
			Code string `json:"code"`
			Role string `json:"role"` // NEW: Accept role from frontend
		}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request, missing code."})
			return
		}

		// DEBUG: Log received role
		log.Printf("🔍 Google OAuth - Received role from frontend: '%s'", requestBody.Role)

		// Validate and default role
		role := requestBody.Role
		if role != "client" && role != "provider" {
			log.Printf("⚠️  Invalid or empty role '%s', defaulting to 'client'", role)
			role = "client" // Default to client if not specified or invalid
		} else {
			log.Printf("✅ Using role: '%s'", role)
		}

		googleOauthConfig := getGoogleOauthConfig()
		if googleOauthConfig.ClientID == "" || googleOauthConfig.ClientSecret == "" {
			log.Println("ERROR: GOOGLE_CLIENT_ID or GOOGLE_CLIENT_SECRET is not set in environment variables")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Google Auth is not configured on server"})
			return
		}

		// 1. Exchange auth code for a token
		token, err := googleOauthConfig.Exchange(ctx, requestBody.Code)
		if err != nil {
			log.Printf("Google OAuth Exchange Error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Google authentication failed. Please try again.",
				"error_code": "GOOGLE_AUTH_FAILED",
			})
			return
		}

		// 2. Use token to get user info from Google
		client := googleOauthConfig.Client(ctx, token)
		oauth2Service, err := google_oauth.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			log.Printf("Google Service Creation Error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Failed to connect to Google. Please try again.",
			})
			return
		}
		userInfo, err := oauth2Service.Userinfo.Get().Do()
		if err != nil {
			log.Printf("Google UserInfo Error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Failed to retrieve user information from Google.",
			})
			return
		}

		// 3. Find or Create User in our DB
		var userID int
		var userEmail string = userInfo.Email
		var profilePictureURL string = userInfo.Picture // ⬅️ Google profile picture URL

		// Check if user exists by email
		err = dbPool.QueryRow(ctx, "SELECT user_id FROM users WHERE email = $1", userEmail).Scan(&userID)

		if err != nil { // User does not exist, create them
			fmt.Println("User not found, creating new user...")

			// Set user_type based on role from frontend
			userType := role
			if userType == "" {
				userType = "client"
			}

			sqlStatement := `
				INSERT INTO users (username, email, gender_id, first_name, last_name, google_id, google_profile_picture, profile_picture_url, verification_status, user_type)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'verified', $9)
				RETURNING user_id
			`
			err = dbPool.QueryRow(ctx, sqlStatement,
				userInfo.Name,
				userEmail,
				4, // (Gender 'Prefer not to say')
				userInfo.GivenName,
				userInfo.FamilyName,
				userInfo.Id,
				profilePictureURL, // google_profile_picture
				profilePictureURL, // profile_picture_url
				userType,          // user_type (client/provider)
			).Scan(&userID)

			if err != nil {
				log.Printf("❌ Failed to create user in database: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new user"})
				return
			}
			log.Println("✅ Successfully created new user")

			// Also create an empty profile row for them
			_, _ = dbPool.Exec(ctx, "INSERT INTO user_profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)

		} else { // User exists, update their Google info
			_, err = dbPool.Exec(ctx,
				`UPDATE users SET 
					google_id = $1, 
					google_profile_picture = $2,
					profile_picture_url = $3,
					first_name = COALESCE($4, first_name),
					last_name = COALESCE($5, last_name),
					username = COALESCE($6, username)
				WHERE user_id = $7`,
				userInfo.Id,
				profilePictureURL,   // Keep old column
				profilePictureURL,   // ⬅️ Update NEW profile_picture_url
				userInfo.GivenName,  // first_name
				userInfo.FamilyName, // last_name
				userInfo.Name,       // username (use Google name if username is empty)
				userID)
			if err != nil {
				log.Printf("Warning: Failed to update Google data for user %d: %v\n", userID, err)
			}
		}

		// 4. Create our own JWT
		tokenString, err := createJWT(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session token"})
			return
		}

		// 5. Fetch complete user data with tier name to send to frontend
		var userData struct {
			UserID             int     `json:"user_id"`
			Username           string  `json:"username"`
			Email              string  `json:"email"`
			FirstName          *string `json:"first_name"`
			LastName           *string `json:"last_name"`
			TierID             int     `json:"tier_id"`
			TierName           string  `json:"tier_name"`
			IsAdmin            bool    `json:"is_admin"`
			ProfilePictureURL  *string `json:"profile_picture_url"` // ⬅️ Changed from google_profile_picture
			VerificationStatus string  `json:"verification_status"`
			UserType           string  `json:"user_type"` // NEW: Add user_type
		}

		err = dbPool.QueryRow(ctx, `
			SELECT 
				u.user_id, 
				u.username, 
				u.email, 
				u.first_name, 
				u.last_name, 
				u.tier_id, 
				COALESCE(t.name, 'General') as tier_name,
				u.is_admin, 
				u.profile_picture_url,
				u.verification_status,
				u.user_type
			FROM users u
			LEFT JOIN tiers t ON u.tier_id = t.tier_id
			WHERE u.user_id = $1
		`, userID).Scan(
			&userData.UserID,
			&userData.Username,
			&userData.Email,
			&userData.FirstName,
			&userData.LastName,
			&userData.TierID,
			&userData.TierName,
			&userData.IsAdmin,
			&userData.ProfilePictureURL, // ⬅️ Now using profile_picture_url
			&userData.VerificationStatus,
			&userData.UserType, // NEW: Scan user_type
		)

		if err != nil {
			log.Printf("Warning: Failed to fetch user data after Google login: %v\n", err)
			// Still send token, but without user data
			c.JSON(http.StatusOK, gin.H{
				"message": "Login successful",
				"token":   tokenString,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   tokenString,
			"user":    userData,
		})
	}
}
