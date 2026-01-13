package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// (ต้องตั้งค่าใน Environment Variable)
var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

// (ยามตัวที่ 1: ตรวจสอบว่า Login หรือยัง)
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		userID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token subject"})
			c.Abort()
			return
		}

		// (สำคัญ) ส่ง userID ต่อไปให้ Handler ตัวถัดไป
		c.Set("userID", userID)
		c.Next()
	}
}

// (ยามตัวที่ 2: ตรวจสอบว่าเป็น Admin หรือไม่)
func adminAuthMiddleware(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {

		userIDInterface, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context (auth middleware error)"})
			c.Abort()
			return
		}

		userID, ok := userIDInterface.(int)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format in context"})
			c.Abort()
			return
		}

		// (ตรวจสอบใน Database)
		var isAdmin bool
		err := dbPool.QueryRow(ctx, "SELECT is_admin FROM users WHERE user_id = $1", userID).Scan(&isAdmin)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check admin status"})
			c.Abort()
			return
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this resource"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// (ยามตัวที่ 3: ตรวจสอบว่าเป็น GOD หรือไม่)
func godAuthMiddleware(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {

		userIDInterface, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			c.Abort()
			return
		}

		userID, ok := userIDInterface.(int)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format"})
			c.Abort()
			return
		}

		// (ตรวจสอบใน Database ว่าเป็น GOD หรือไม่)
		var userType string
		err := dbPool.QueryRow(ctx, "SELECT user_type FROM users WHERE user_id = $1", userID).Scan(&userType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user type"})
			c.Abort()
			return
		}

		if userType != "god" {
			c.JSON(http.StatusForbidden, gin.H{"error": "This action requires GOD privileges. Only the supreme administrator can perform this operation."})
			c.Abort()
			return
		}

		c.Next()
	}
}

// (ยามตัวที่ 4: ตรวจสอบว่าเป็น Admin หรือ GOD)
func adminOrGodAuthMiddleware(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {

		userIDInterface, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			c.Abort()
			return
		}

		userID, ok := userIDInterface.(int)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format"})
			c.Abort()
			return
		}

		// (ตรวจสอบว่าเป็น Admin หรือ GOD)
		var userType string
		var isAdmin bool
		err := dbPool.QueryRow(ctx,
			"SELECT user_type, is_admin FROM users WHERE user_id = $1",
			userID,
		).Scan(&userType, &isAdmin)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		// Allow if GOD or Admin
		if userType != "god" && !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin or GOD privileges required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
