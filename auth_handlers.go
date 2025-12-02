package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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
// Handles standard email/password registration
func createUserHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newUser struct {
			Username  string  `json:"username" binding:"required"`
			Email     string  `json:"email" binding:"required"`
			Password  string  `json:"password" binding:"required"`
			GenderID  int     `json:"gender_id" binding:"required"`
			FirstName *string `json:"first_name"`
			LastName  *string `json:"last_name"`
		}

		if err := c.ShouldBindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Hash the password
		hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		hashedPassword := string(hashedPasswordBytes)

		// SQL for creating a new user
		// Note: tier_id and provider_level_id use the database DEFAULT (1)
		sqlStatement := `
			INSERT INTO users (username, email, password_hash, gender_id, first_name, last_name)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING user_id
		`
		var userID int
		err = dbPool.QueryRow(ctx, sqlStatement,
			newUser.Username,
			newUser.Email,
			&hashedPassword,
			newUser.GenderID,
			newUser.FirstName,
			newUser.LastName,
		).Scan(&userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
			return
		}

		// Also create an empty profile row for them
		_, err = dbPool.Exec(ctx, "INSERT INTO user_profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)
		if err != nil {
			log.Printf("Warning: Could not create empty profile for new user %d: %v\n", userID, err)
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user_id": userID})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email/Password cannot be empty"})
			return
		}

		var storedPasswordHash *string // Can be NULL
		var userID int
		sqlStatement := `SELECT user_id, password_hash FROM users WHERE email = $1`

		err := dbPool.QueryRow(ctx, sqlStatement, loginDetails.Email).Scan(&userID, &storedPasswordHash)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Check if user has a password (they might be Google-only)
		if storedPasswordHash == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "This account uses Google login. Please use the 'Login with Google' button."})
			return
		}

		// Compare password with hash
		err = bcrypt.CompareHashAndPassword([]byte(*storedPasswordHash), []byte(loginDetails.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Login Success: Create JWT
		tokenString, err := createJWT(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   tokenString,
		})
	}
}

// --- Handler: POST /auth/google ---
// Handles login/registration via Google OAuth
func handleGoogleCallback(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody struct {
			Code string `json:"code"`
		}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request, missing code."})
			return
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
			// Send 401 so frontend knows to redirect back to home
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Google authentication failed. Please try again.",
				"details": err.Error(),
			})
			return
		}

		// 2. Use token to get user info from Google
		client := googleOauthConfig.Client(ctx, token)
		oauth2Service, err := google_oauth.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			log.Printf("Google Service Creation Error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Failed to connect to Google. Please try again.",
				"details": err.Error(),
			})
			return
		}
		userInfo, err := oauth2Service.Userinfo.Get().Do()
		if err != nil {
			log.Printf("Google UserInfo Error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Failed to retrieve user information from Google.",
				"details": err.Error(),
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

			sqlStatement := `
				INSERT INTO users (username, email, gender_id, first_name, last_name, google_id, google_profile_picture, profile_picture_url)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING user_id
			`
			err = dbPool.QueryRow(ctx, sqlStatement,
				userInfo.Name,
				userEmail,
				4, // (Gender 'Prefer not to say')
				userInfo.GivenName,
				userInfo.FamilyName,
				userInfo.Id,
				profilePictureURL, // (Save to old column for compatibility)
				profilePictureURL, // ⬅️ Save to NEW profile_picture_url column
			).Scan(&userID)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new user", "details": err.Error()})
				return
			}

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
				u.verification_status
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
