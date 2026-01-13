package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

// --- Helper: Generate 6-digit OTP ---
func generateOTP() string {
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}

// --- Helper: Send Email ---
// Send verification email using SMTP
func sendVerificationEmail(email string, otp string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	emailFrom := os.Getenv("EMAIL_FROM")

	// Check if email is configured
	if smtpHost == "" || smtpUser == "" || smtpPassword == "" {
		log.Printf("⚠️  Email service not configured - OTP: %s for %s", otp, email)
		return nil // Silent mode - don't fail if email not configured
	}

	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", emailFrom)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Email Verification - Thai Variety")

	// HTML body
	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px; }
				.container { background-color: white; padding: 30px; border-radius: 10px; max-width: 600px; margin: 0 auto; }
				.header { color: #ff10f0; font-size: 24px; font-weight: bold; margin-bottom: 20px; }
				.otp { font-size: 32px; font-weight: bold; color: #333; background-color: #f0f0f0; padding: 15px; text-align: center; border-radius: 5px; letter-spacing: 5px; }
				.footer { margin-top: 20px; color: #666; font-size: 12px; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">🔐 Email Verification</div>
				<p>Hello,</p>
				<p>Your verification code is:</p>
				<div class="otp">%s</div>
				<p>This code will expire in 10 minutes.</p>
				<p>If you didn't request this code, please ignore this email.</p>
				<div class="footer">
					<p>Thai Variety Platform<br>
					https://thaivariety.app</p>
				</div>
			</div>
		</body>
		</html>
	`, otp)

	m.SetBody("text/html", htmlBody)

	// Connect and send
	port, _ := strconv.Atoi(smtpPort)
	d := gomail.NewDialer(smtpHost, port, smtpUser, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("❌ Failed to send email to %s: %v", email, err)
		return err
	}

	log.Printf("✅ OTP email sent successfully to %s", email)
	return nil
}

// --- Handler: POST /auth/send-verification ---
// Send OTP to email for verification
func sendVerificationHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			return
		}

		// Check if email already registered and verified
		var exists bool
		var verificationStatus string
		err := dbPool.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1), COALESCE((SELECT verification_status FROM users WHERE email = $1), '')",
			req.Email).Scan(&exists, &verificationStatus)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if exists {
			// If already verified, don't send OTP
			if verificationStatus == "verified" {
				c.JSON(http.StatusOK, gin.H{
					"message":  "Email is already verified",
					"verified": true,
				})
				return
			}
			// If registered but not verified, allow sending OTP for re-verification
			// Continue to send OTP below
		}

		// Generate OTP
		otp := generateOTP()
		expiresAt := time.Now().Add(10 * time.Minute)

		// Store OTP in database (temporary table)
		sqlStatement := `
			INSERT INTO email_verifications (email, otp, expires_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (email) DO UPDATE 
			SET otp = $2, expires_at = $3, created_at = NOW()
		`
		_, err = dbPool.Exec(ctx, sqlStatement, req.Email, otp, expiresAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store OTP"})
			return
		}

		// Send email asynchronously (don't wait for email to be sent)
		go func(email string, otpCode string) {
			err := sendVerificationEmail(email, otpCode)
			if err != nil {
				log.Printf("❌ Async email send failed for %s: %v", email, err)
			}
		}(req.Email, otp)

		// DEV MODE: Return OTP in response for testing
		devMode := os.Getenv("DEV_MODE") == "true"
		response := gin.H{
			"message":    "Verification code sent to your email",
			"expires_in": "10 minutes",
		}
		if devMode {
			response["dev_otp"] = otp
		}
		c.JSON(http.StatusOK, response)
	}
}

// --- Handler: POST /auth/verify-email ---
// Verify OTP code
func verifyEmailHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
			OTP   string `json:"otp" binding:"required,len=6"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Check OTP
		var storedOTP string
		var expiresAt time.Time
		sqlStatement := `
			SELECT otp, expires_at 
			FROM email_verifications 
			WHERE email = $1
		`
		err := dbPool.QueryRow(ctx, sqlStatement, req.Email).Scan(&storedOTP, &expiresAt)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No verification code found for this email"})
			return
		}

		// Check expiration
		if time.Now().After(expiresAt) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Verification code expired"})
			return
		}

		// Check OTP match
		if storedOTP != req.OTP {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid verification code"})
			return
		}

		// Delete used OTP
		_, _ = dbPool.Exec(ctx, "DELETE FROM email_verifications WHERE email = $1", req.Email)

		c.JSON(http.StatusOK, gin.H{
			"message":  "Email verified successfully",
			"verified": true,
		})
	}
}

// --- Handler: POST /register (with email verification) ---
// Register user after email is verified
func registerWithVerificationHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newUser struct {
			Username  string  `json:"username" binding:"required"`
			Email     string  `json:"email" binding:"required,email"`
			Password  string  `json:"password" binding:"required,min=6"`
			GenderID  int     `json:"gender_id" binding:"required"`
			FirstName *string `json:"first_name"`
			LastName  *string `json:"last_name"`
			Phone     *string `json:"phone"`
			OTP       string  `json:"otp" binding:"required,len=6"`
		}

		if err := c.ShouldBindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 1. Verify OTP first
		var storedOTP string
		var expiresAt time.Time
		err := dbPool.QueryRow(ctx,
			"SELECT otp, expires_at FROM email_verifications WHERE email = $1",
			newUser.Email,
		).Scan(&storedOTP, &expiresAt)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Please verify your email first"})
			return
		}

		if time.Now().After(expiresAt) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Verification code expired"})
			return
		}

		if storedOTP != newUser.OTP {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid verification code"})
			return
		}

		// 2. Create user
		hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		hashedPassword := string(hashedPasswordBytes)

		sqlStatement := `
			INSERT INTO users (username, email, password_hash, gender_id, first_name, last_name, phone_number, verification_status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'verified')
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
			newUser.Phone,
		).Scan(&userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
			return
		}

		// 3. Create empty profile
		_, _ = dbPool.Exec(ctx, "INSERT INTO user_profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)

		// 4. Delete used OTP
		_, _ = dbPool.Exec(ctx, "DELETE FROM email_verifications WHERE email = $1", newUser.Email)

		// 5. Create JWT token
		tokenString, err := createJWT(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Registration successful",
			"user_id": userID,
			"token":   tokenString,
		})
	}
}

// --- Handler: POST /auth/send-otp (For authenticated users) ---
// Send OTP to authenticated user's email
func sendOTPHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from JWT token
		userIDInterface, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDInterface.(int)

		// Get user's email
		var email string
		err := dbPool.QueryRow(ctx, "SELECT email FROM users WHERE user_id = $1", userID).Scan(&email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		// Generate OTP
		otp := generateOTP()
		expiresAt := time.Now().Add(10 * time.Minute)

		// Store OTP in database
		sqlStatement := `
			INSERT INTO email_verifications (email, otp, expires_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (email) DO UPDATE 
			SET otp = $2, expires_at = $3, created_at = NOW()
		`
		_, err = dbPool.Exec(ctx, sqlStatement, email, otp, expiresAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store OTP"})
			return
		}

		// Send email asynchronously (don't wait for email to be sent)
		go func(userEmail string, otpCode string) {
			err := sendVerificationEmail(userEmail, otpCode)
			if err != nil {
				log.Printf("❌ Async email send failed for %s: %v", userEmail, err)
			}
		}(email, otp)

		// For development: return OTP in response (REMOVE IN PRODUCTION)
		c.JSON(http.StatusOK, gin.H{
			"message":    "Verification code sent to your email",
			"expires_in": "10 minutes",
			"dev_otp":    otp, // TODO: Remove in production
		})
	}
}

// --- Handler: POST /auth/verify-otp (For authenticated users) ---
// Verify OTP for authenticated user
func verifyOTPHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from JWT token
		userIDInterface, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDInterface.(int)

		var req struct {
			OTP string `json:"otp" binding:"required,len=6"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP format"})
			return
		}

		// Get user's email
		var email string
		err := dbPool.QueryRow(ctx, "SELECT email FROM users WHERE user_id = $1", userID).Scan(&email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		// Check OTP
		var storedOTP string
		var expiresAt time.Time
		sqlStatement := `
			SELECT otp, expires_at 
			FROM email_verifications 
			WHERE email = $1
		`
		err = dbPool.QueryRow(ctx, sqlStatement, email).Scan(&storedOTP, &expiresAt)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No verification code found"})
			return
		}

		// Check expiration
		if time.Now().After(expiresAt) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Verification code expired. Please request a new one."})
			return
		}

		// Check OTP match
		if storedOTP != req.OTP {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid verification code"})
			return
		}

		// Mark user as verified
		_, err = dbPool.Exec(ctx, "UPDATE users SET verification_status = 'verified' WHERE user_id = $1", userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update verification status"})
			return
		}

		// Delete used OTP
		_, _ = dbPool.Exec(ctx, "DELETE FROM email_verifications WHERE email = $1", email)

		c.JSON(http.StatusOK, gin.H{
			"message":  "Email verified successfully",
			"verified": true,
		})
	}
}

// --- Handler: GET /auth/verification-status ---
// Check if current user needs email verification
func checkVerificationStatusHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from JWT token
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var verificationStatus string
		var email string
		err := dbPool.QueryRow(ctx,
			"SELECT verification_status, email FROM users WHERE user_id = $1",
			userID).Scan(&verificationStatus, &email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get verification status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"email":               email,
			"verification_status": verificationStatus,
			"is_verified":         verificationStatus == "verified",
			"needs_verification":  verificationStatus != "verified",
		})
	}
}
