package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthCheckResponse represents the health status of the system
type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
	Version   string            `json:"version"`
}

// HealthCheck performs a comprehensive health check of all services
func healthCheckHandler(dbPool *pgxpool.Pool, rdb *redis.Client, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := HealthCheckResponse{
			Status:    "healthy",
			Timestamp: time.Now(),
			Services:  make(map[string]string),
			Version:   "1.0.0",
		}

		// Check PostgreSQL
		var pgTime time.Time
		err := dbPool.QueryRow(ctx, "SELECT NOW()").Scan(&pgTime)
		if err != nil {
			response.Status = "unhealthy"
			response.Services["postgres"] = "down"
		} else {
			response.Services["postgres"] = "up"
		}

		// Check Redis
		_, err = rdb.Ping(ctx).Result()
		if err != nil {
			response.Status = "unhealthy"
			response.Services["redis"] = "down"
		} else {
			response.Services["redis"] = "up"
		}

		// Check database connection pool
		stats := dbPool.Stat()
		if stats.TotalConns() == 0 {
			response.Status = "unhealthy"
			response.Services["db_pool"] = "no connections"
		} else {
			response.Services["db_pool"] = "up"
		}

		// Determine HTTP status code
		statusCode := http.StatusOK
		if response.Status == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, response)
	}
}

// ReadinessCheck checks if the service is ready to serve requests
func readinessCheckHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Quick check - can we query the database?
		var count int
		err := dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM users LIMIT 1").Scan(&count)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	}
}

// LivenessCheck checks if the service is alive
func livenessCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
			"time":   time.Now(),
		})
	}
}
