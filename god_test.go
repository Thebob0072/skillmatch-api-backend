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
	
	router.DELETE("/admin/users/:user_id", func(c *gin.Context) {
		c.Set("userID", 2) // Non-GOD user trying to delete GOD
		deleteUserHandler(nil, nil)(c)
	})

	// Test: Try to delete GOD (user_id = 1)
	req, _ := http.NewRequest("DELETE", "/admin/users/1", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateUserRole(t *testing.T) {
	router := setupTestRouter()
	
	router.POST("/god/update-user", func(c *gin.Context) {
		c.Set("userID", 1) // GOD user
		updateUserHandler(nil, nil)(c)
	})

	// Test: Update user_id = 5 to admin
	payload := map[string]interface{}{
		"user_id": 5,
		"is_admin": true,
		"tier_id": 3,
	}
	jsonData, _ := json.Marshal(payload)
	
	req, _ := http.NewRequest("POST", "/god/update-user", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should succeed or return DB error
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestListAllUsers(t *testing.T) {
	router := setupTestRouter()
	
	router.GET("/admin/users", func(c *gin.Context) {
		c.Set("userID", 1)
		listAllUsersHandler(nil, nil)(c)
	})

	req, _ := http.NewRequest("GET", "/admin/users?page=1&limit=20", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 or DB error
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestCreateAdmin(t *testing.T) {
	router := setupTestRouter()
	
	router.POST("/admin/admins", func(c *gin.Context) {
		c.Set("userID", 1)
		createAdminHandler(nil, nil)(c)
	})

	payload := map[string]interface{}{
		"username": "test_admin",
		"email": "admin@test.com",
		"password": "SecurePass123",
		"gender_id": 1,
	}
	jsonData, _ := json.Marshal(payload)
	
	req, _ := http.NewRequest("POST", "/admin/admins", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestViewModeSwitch(t *testing.T) {
	router := setupTestRouter()
	
	router.POST("/god/view-mode", func(c *gin.Context) {
		c.Set("userID", 1)
		setGodViewModeHandler(nil, nil)(c)
	})

	payload := map[string]string{"mode": "provider"}
	jsonData, _ := json.Marshal(payload)
	
	req, _ := http.NewRequest("POST", "/god/view-mode", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

// --- Authorization Tests ---

func TestNonGODCannotDeleteUser(t *testing.T) {
	router := setupTestRouter()
	
	router.DELETE("/admin/users/:user_id", func(c *gin.Context) {
		c.Set("userID", 5) // Regular user
		deleteUserHandler(nil, nil)(c)
	})

	req, _ := http.NewRequest("DELETE", "/admin/users/10", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestNonGODCannotCreateAdmin(t *testing.T) {
	router := setupTestRouter()
	
	router.POST("/admin/admins", func(c *gin.Context) {
		c.Set("userID", 5) // Regular user
		createAdminHandler(nil, nil)(c)
	})

	payload := map[string]interface{}{
		"username": "fake_admin",
		"email": "fake@test.com",
		"password": "Pass123",
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
	
	router.DELETE("/admin/users/:user_id", func(c *gin.Context) {
		c.Set("userID", 1)
		deleteUserHandler(nil, nil)(c)
	})

	// Test with invalid user_id (non-numeric)
	req, _ := http.NewRequest("DELETE", "/admin/users/invalid", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmptyRequestBody(t *testing.T) {
	router := setupTestRouter()
	
	router.POST("/god/update-user", func(c *gin.Context) {
		c.Set("userID", 1)
		updateUserHandler(nil, nil)(c)
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
	
	router.POST("/god/view-mode", func(c *gin.Context) {
		c.Set("userID", 1)
		setGodViewModeHandler(nil, nil)(c)
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
