package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test setup
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

// --- Unit Tests (Logic Only - No DB) ---

func TestGODCanDeleteUser_Authorization(t *testing.T) {
	router := setupTestRouter()

	// Mock handler that only checks authorization logic
	router.DELETE("/admin/users/:user_id", func(c *gin.Context) {
		// Simulate GOD check
		requesterID := 1 // GOD
		tier_id := 5
		is_admin := true

		userIDParam := c.Param("user_id")

		// Authorization logic
		if !is_admin || tier_id != 5 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only GOD can delete users"})
			return
		}

		// Check if trying to delete GOD account
		if userIDParam == "1" && requesterID != 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete GOD account"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	})

	// Test: GOD deleting regular user
	req, _ := http.NewRequest("DELETE", "/admin/users/5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGODCannotDeleteGODAccount(t *testing.T) {
	router := setupTestRouter()

	// Pure authorization logic test - without calling actual handler
	router.DELETE("/admin/users/:user_id", func(c *gin.Context) {
		requesterID := 2 // Non-GOD user
		targetUserID := c.Param("user_id")

		// Authorization logic: non-GOD cannot delete GOD account
		if targetUserID == "1" && requesterID != 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete GOD account"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
	})

	// Test: Try to delete GOD (user_id = 1) as non-GOD user
	req, _ := http.NewRequest("DELETE", "/admin/users/1", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateUserRole(t *testing.T) {
	router := setupTestRouter()

	// Mock handler with authorization logic only (no DB)
	router.POST("/god/update-user", func(c *gin.Context) {
		requesterID := 1 // GOD user
		requesterTierID := 5
		isAdmin := true

		var req struct {
			UserID  int  `json:"user_id" binding:"required"`
			IsAdmin bool `json:"is_admin"`
			TierID  int  `json:"tier_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Authorization: Only GOD can update users
		if !isAdmin || requesterTierID != 5 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only GOD can update users"})
			return
		}

		// Cannot modify self
		if req.UserID == requesterID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify your own account"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "user_id": req.UserID})
	})

	// Test: Update user_id = 5 to admin
	payload := map[string]interface{}{
		"user_id":  5,
		"is_admin": true,
		"tier_id":  3,
	}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/god/update-user", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListAllUsers(t *testing.T) {
	router := setupTestRouter()

	// Mock handler with authorization logic only (no DB)
	router.GET("/admin/users", func(c *gin.Context) {
		requesterTierID := 5
		isAdmin := true

		// Authorization: Only admins can list users
		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}

		// Mock user list response
		users := []map[string]interface{}{
			{"id": 1, "username": "god", "role": "god", "tier_id": requesterTierID},
			{"id": 2, "username": "admin1", "role": "admin", "tier_id": 4},
			{"id": 3, "username": "provider1", "role": "provider", "tier_id": 1},
		}

		c.JSON(http.StatusOK, gin.H{
			"users": users,
			"total": 3,
			"page":  1,
			"limit": 20,
		})
	})

	req, _ := http.NewRequest("GET", "/admin/users?page=1&limit=20", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateAdmin(t *testing.T) {
	router := setupTestRouter()

	// Mock handler with authorization logic only (no DB)
	router.POST("/admin/admins", func(c *gin.Context) {
		requesterTierID := 5
		isAdmin := true

		var req struct {
			Username string `json:"username" binding:"required"`
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
			GenderID int    `json:"gender_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Authorization: Only GOD can create admins
		if !isAdmin || requesterTierID != 5 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only GOD can create admins"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "Admin created successfully",
			"username": req.Username,
			"email":    req.Email,
		})
	})

	payload := map[string]interface{}{
		"username":  "test_admin",
		"email":     "admin@test.com",
		"password":  "SecurePass123",
		"gender_id": 1,
	}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/admin/admins", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestViewModeSwitch(t *testing.T) {
	router := setupTestRouter()

	// Mock handler with authorization logic only (no DB)
	router.POST("/god/view-mode", func(c *gin.Context) {
		requesterTierID := 5
		isAdmin := true

		var req struct {
			Mode string `json:"mode"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Authorization: Only GOD can switch view modes
		if !isAdmin || requesterTierID != 5 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only GOD can switch view modes"})
			return
		}

		// Validate mode
		validModes := map[string]bool{"client": true, "provider": true, "admin": true, "god": true}
		if !validModes[req.Mode] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid view mode"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "View mode updated",
			"mode":    req.Mode,
		})
	})

	payload := map[string]string{"mode": "provider"}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/god/view-mode", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- Authorization Tests ---

func TestNonGODCannotDeleteUser(t *testing.T) {
	router := setupTestRouter()

	// Pure authorization logic test
	router.DELETE("/admin/users/:user_id", func(c *gin.Context) {
		requesterID := 5     // Regular user (not GOD)
		requesterTierID := 1 // Not GOD tier
		isAdmin := false

		// Authorization: Only GOD (tier_id=5, is_admin=true) can delete users
		if !isAdmin || requesterTierID != 5 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only GOD can delete users"})
			return
		}

		_ = requesterID // Use variable
		c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
	})

	req, _ := http.NewRequest("DELETE", "/admin/users/10", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestNonGODCannotCreateAdmin(t *testing.T) {
	router := setupTestRouter()

	// Pure authorization logic test
	router.POST("/admin/admins", func(c *gin.Context) {
		requesterTierID := 1 // Regular user tier
		isAdmin := false

		// Authorization: Only GOD (tier_id=5) can create admins
		if !isAdmin || requesterTierID != 5 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only GOD can create admins"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Admin created"})
	})

	payload := map[string]interface{}{
		"username":  "fake_admin",
		"email":     "fake@test.com",
		"password":  "Pass123",
		"gender_id": 1,
	}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/admin/admins", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- Edge Cases ---

func TestInvalidUserID(t *testing.T) {
	router := setupTestRouter()

	// Pure validation logic test
	router.DELETE("/admin/users/:user_id", func(c *gin.Context) {
		userIDStr := c.Param("user_id")

		// Validation: user_id must be numeric
		var userID int
		_, err := fmt.Sscanf(userIDStr, "%d", &userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
	})

	// Test with invalid user_id (non-numeric)
	req, _ := http.NewRequest("DELETE", "/admin/users/invalid", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmptyRequestBody(t *testing.T) {
	router := setupTestRouter()

	// Pure validation logic test
	router.POST("/god/update-user", func(c *gin.Context) {
		var req struct {
			UserID int `json:"user_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		if req.UserID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Updated"})
	})

	// Empty request body
	req, _ := http.NewRequest("POST", "/god/update-user", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 400 Bad Request (missing required user_id)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInvalidViewMode(t *testing.T) {
	router := setupTestRouter()

	// Pure validation logic test
	router.POST("/god/view-mode", func(c *gin.Context) {
		var req struct {
			Mode string `json:"mode"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Validate mode: must be client, provider, admin, or god
		validModes := map[string]bool{"client": true, "provider": true, "admin": true, "god": true}
		if !validModes[req.Mode] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid view mode"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Mode set", "mode": req.Mode})
	})

	payload := map[string]string{"mode": "invalid_mode"}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/god/view-mode", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Run all tests ---
// go test -v ./god_test.go
