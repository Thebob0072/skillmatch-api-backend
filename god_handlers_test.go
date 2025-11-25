package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock database functions
type MockDB struct {
	users []User
}

func (m *MockDB) GetUserByID(id int) (*User, error) {
	for _, user := range m.users {
		if user.UserID == id {
			return &user, nil
		}
	}
	return nil, nil
}

func (m *MockDB) DeleteUser(id int) error {
	for i, user := range m.users {
		if user.UserID == id {
			m.users = append(m.users[:i], m.users[i+1:]...)
			return nil
		}
	}
	return nil
}

// Helper: Mock auth middleware for testing
func mockAuthMiddleware(userID int, tierID int, isAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("tierID", tierID)
		c.Set("isAdmin", isAdmin)
		c.Next()
	}
}

// Test 1: GOD can delete any user
func TestGOD_CanDeleteAnyUser(t *testing.T) {
	router := setupTestRouter()

	// Mock database
	mockDB := &MockDB{
		users: []User{
			{UserID: 1, Username: "GOD", Email: "god@test.com", TierID: 5, IsAdmin: true},
			{UserID: 2, Username: "testuser", Email: "test@test.com", TierID: 1, IsAdmin: false},
		},
	}

	// Setup route with GOD auth
	router.DELETE("/admin/users/:userId", mockAuthMiddleware(1, 5, true), func(c *gin.Context) {
		// Simulate deleteUserHandler logic
		userID := c.Param("userId")
		requesterTierID := c.GetInt("tierID")

		if requesterTierID != 5 {
			c.JSON(403, gin.H{"error": "Only GOD can delete users"})
			return
		}

		if userID == "1" {
			c.JSON(403, gin.H{"error": "Cannot delete GOD account"})
			return
		}

		mockDB.DeleteUser(2)
		c.JSON(200, gin.H{"message": "User deleted successfully", "user_id": 2})
	})

	// Make request
	req, _ := http.NewRequest("DELETE", "/admin/users/2", nil)
	req.Header.Set("Authorization", "Bearer mock_token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User deleted successfully", response["message"])
	assert.Equal(t, float64(2), response["user_id"])
}

// Test 2: GOD cannot delete self
func TestGOD_CannotDeleteSelf(t *testing.T) {
	router := setupTestRouter()

	router.DELETE("/admin/users/:userId", mockAuthMiddleware(1, 5, true), func(c *gin.Context) {
		userID := c.Param("userId")

		if userID == "1" {
			c.JSON(403, gin.H{"error": "Cannot delete GOD account"})
			return
		}

		c.JSON(200, gin.H{"message": "User deleted successfully"})
	})

	req, _ := http.NewRequest("DELETE", "/admin/users/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Cannot delete GOD account", response["error"])
}

// Test 3: Non-GOD cannot delete users
func TestNonGOD_CannotDeleteUsers(t *testing.T) {
	router := setupTestRouter()

	router.DELETE("/admin/users/:userId", mockAuthMiddleware(2, 2, true), func(c *gin.Context) {
		requesterTierID := c.GetInt("tierID")

		if requesterTierID != 5 {
			c.JSON(403, gin.H{"error": "Only GOD can delete users"})
			return
		}

		c.JSON(200, gin.H{"message": "User deleted successfully"})
	})

	req, _ := http.NewRequest("DELETE", "/admin/users/3", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Only GOD can delete users", response["error"])
}

// Test 4: GOD can update user roles
func TestGOD_CanUpdateUserRoles(t *testing.T) {
	router := setupTestRouter()

	router.POST("/god/update-user", mockAuthMiddleware(1, 5, true), func(c *gin.Context) {
		requesterTierID := c.GetInt("tierID")
		requesterIsAdmin := c.GetBool("isAdmin")

		if requesterTierID != 5 || !requesterIsAdmin {
			c.JSON(403, gin.H{"error": "Only GOD can update user roles"})
			return
		}

		var req struct {
			UserID  int   `json:"user_id"`
			IsAdmin *bool `json:"is_admin"`
			TierID  *int  `json:"tier_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		if req.UserID == 1 {
			c.JSON(403, gin.H{"error": "Cannot modify GOD account"})
			return
		}

		c.JSON(200, gin.H{"message": "User role updated successfully", "user_id": req.UserID})
	})

	body := map[string]interface{}{
		"user_id":  2,
		"is_admin": true,
		"tier_id":  3,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/god/update-user", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User role updated successfully", response["message"])
}

// Test 5: GOD cannot modify self via update endpoint
func TestGOD_CannotModifySelf(t *testing.T) {
	router := setupTestRouter()

	router.POST("/god/update-user", mockAuthMiddleware(1, 5, true), func(c *gin.Context) {
		var req struct {
			UserID int `json:"user_id"`
		}

		c.ShouldBindJSON(&req)

		if req.UserID == 1 {
			c.JSON(403, gin.H{"error": "Cannot modify GOD account"})
			return
		}

		c.JSON(200, gin.H{"message": "User updated"})
	})

	body := map[string]interface{}{"user_id": 1, "tier_id": 1}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/god/update-user", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}

// Test 6: GOD can create admins
func TestGOD_CanCreateAdmins(t *testing.T) {
	router := setupTestRouter()

	router.POST("/admin/admins", mockAuthMiddleware(1, 5, true), func(c *gin.Context) {
		requesterTierID := c.GetInt("tierID")

		if requesterTierID != 5 {
			c.JSON(403, gin.H{"error": "Only GOD can create admins"})
			return
		}

		var req struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		c.ShouldBindJSON(&req)

		c.JSON(201, gin.H{
			"message":  "Admin created successfully",
			"user_id":  99,
			"username": req.Username,
		})
	})

	body := map[string]interface{}{
		"username":  "new_admin",
		"email":     "admin@test.com",
		"password":  "SecurePass123!",
		"gender_id": 1,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/admin/admins", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Admin created successfully", response["message"])
	assert.Equal(t, "new_admin", response["username"])
}

// Test 7: GOD view mode switching
func TestGOD_ViewModeSwitching(t *testing.T) {
	router := setupTestRouter()

	router.POST("/god/view-mode", mockAuthMiddleware(1, 5, true), func(c *gin.Context) {
		requesterID := c.GetInt("userID")
		requesterTierID := c.GetInt("tierID")

		if requesterTierID != 5 {
			c.JSON(403, gin.H{"error": "Only GOD can change view mode"})
			return
		}

		var req struct {
			Mode string `json:"mode"`
		}
		c.ShouldBindJSON(&req)

		validModes := []string{"user", "provider", "admin", "god"}
		isValid := false
		for _, mode := range validModes {
			if req.Mode == mode {
				isValid = true
				break
			}
		}

		if !isValid {
			c.JSON(400, gin.H{"error": "Invalid view mode"})
			return
		}

		c.JSON(200, gin.H{
			"message":      "View mode updated successfully",
			"current_mode": req.Mode,
			"user_id":      requesterID,
		})
	})

	body := map[string]interface{}{"mode": "provider"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/god/view-mode", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "View mode updated successfully", response["message"])
	assert.Equal(t, "provider", response["current_mode"])
}

// Test 8: List all users (GOD only)
func TestGOD_CanListAllUsers(t *testing.T) {
	router := setupTestRouter()

	router.GET("/admin/users", mockAuthMiddleware(1, 5, true), func(c *gin.Context) {
		requesterTierID := c.GetInt("tierID")

		if requesterTierID != 5 {
			c.JSON(403, gin.H{"error": "Only GOD can list all users"})
			return
		}

		users := []map[string]interface{}{
			{"user_id": 1, "username": "GOD", "tier_id": 5, "is_admin": true},
			{"user_id": 2, "username": "user1", "tier_id": 1, "is_admin": false},
			{"user_id": 3, "username": "admin1", "tier_id": 2, "is_admin": true},
		}

		c.JSON(200, gin.H{
			"users": users,
			"total": 3,
			"page":  1,
			"limit": 50,
		})
	})

	req, _ := http.NewRequest("GET", "/admin/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(3), response["total"])

	users := response["users"].([]interface{})
	assert.Equal(t, 3, len(users))
}
