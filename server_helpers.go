package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// GracefulShutdown handles graceful shutdown of the server
func setupGracefulShutdown(server *http.Server, dbPool *pgxpool.Pool, rdb *redis.Client) {
	// Create channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-quit
	log.Println("🛑 Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v\n", err)
	}

	// Close database connections
	if dbPool != nil {
		dbPool.Close()
		log.Println("✅ Database connections closed")
	}

	// Close Redis connection
	if rdb != nil {
		if err := rdb.Close(); err != nil {
			log.Printf("⚠️  Error closing Redis: %v\n", err)
		} else {
			log.Println("✅ Redis connection closed")
		}
	}

	// Close global database connection
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("⚠️  Error closing global database: %v\n", err)
		} else {
			log.Println("✅ Global database connection closed")
		}
	}

	log.Println("✅ Server exited gracefully")
}

// RateLimitMiddleware implements basic rate limiting
func rateLimitMiddleware(rdb *redis.Client, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// Check current count
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// If Redis fails, allow the request
			log.Printf("Rate limit check failed: %v", err)
			c.Next()
			return
		}

		// Set expiry on first request
		if count == 1 {
			rdb.Expire(ctx, key, time.Minute)
		}

		// Rate limit: 100 requests per minute
		if count > 100 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers to responses
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (relaxed for development)
		if os.Getenv("GIN_MODE") == "release" {
			c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https:; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")
		}

		c.Next()
	}
}

// LoggingMiddleware provides detailed request logging
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		// Log request details
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// Color code based on status
		var statusColor string
		switch {
		case statusCode >= 200 && statusCode < 300:
			statusColor = "\033[32m" // Green
		case statusCode >= 300 && statusCode < 400:
			statusColor = "\033[36m" // Cyan
		case statusCode >= 400 && statusCode < 500:
			statusColor = "\033[33m" // Yellow
		default:
			statusColor = "\033[31m" // Red
		}
		resetColor := "\033[0m"

		log.Printf("%s%d%s | %13v | %15s | %s %s",
			statusColor, statusCode, resetColor,
			duration,
			clientIP,
			method,
			path,
		)
	}
}

// ErrorRecoveryMiddleware recovers from panics and logs them
func errorRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("🚨 PANIC RECOVERED: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error. Please try again later.",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
