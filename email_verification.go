package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// --- Helper: Generate 6-digit OTP ---
func generateOTP() string {
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}

// --- Helper: Send Email (Placeholder) ---
// TODO: Integrate with actual email service (SendGrid, AWS SES, etc.)
func sendVerificationEmail(email string, otp string) error {
	// For now, just log it (in production, use real email service)
	fmt.Printf("📧 Sending OTP to %s: %s\n", email, otp)
	fmt.Printf("   OTP expires in 10 minutes\n")

	// TODO: Replace with actual email service
	// Example with SendGrid:
	// from := mail.NewEmail("SkillMatch", "noreply@skillmatch.com")
	// to := mail.NewEmail("User", email)
	// subject := "Email Verification - SkillMatch"
	// content := mail.NewContent("text/html", fmt.Sprintf("Your verification code is: <b>%s</b>", otp))
	// message := mail.NewV3MailInit(from, subject, to, content)
	// client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	// response, err := client.Send(message)

	return nil // Simulate success
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

		// Check if email already registered
		var exists bool
		err := dbPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
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

		// Send email
		err = sendVerificationEmail(req.Email, otp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Verification code sent to your email",
			"expires_in": "10 minutes",
		})
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
