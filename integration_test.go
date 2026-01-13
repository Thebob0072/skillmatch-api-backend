package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite provides full integration tests
type IntegrationTestSuite struct {
	suite.Suite
	router    *gin.Engine
	dbPool    *pgxpool.Pool
	ctx       context.Context
	authToken string
	userID    int
}

// SetupSuite runs before all tests
func (suite *IntegrationTestSuite) SetupSuite() {
	// Load test environment
	_ = godotenv.Load(".env.test")

	// Set test database URL
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "postgres://admin:mysecretpassword@localhost:5432/skillmatch_test?sslmode=disable"
	}

	var err error
	suite.ctx = context.Background()

	// Connect to test database
	suite.dbPool, err = pgxpool.New(suite.ctx, testDBURL)
	if err != nil {
		suite.T().Fatalf("Failed to connect to test database: %v", err)
	}

	// Verify connection
	if err = suite.dbPool.Ping(suite.ctx); err != nil {
		suite.T().Fatalf("Failed to ping test database: %v", err)
	}

	// Setup router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.setupRoutes()
}

// TearDownSuite runs after all tests
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.dbPool != nil {
		suite.dbPool.Close()
	}
}

// SetupTest runs before each test
func (suite *IntegrationTestSuite) SetupTest() {
	// Clean up test data before each test
	suite.cleanupTestData()
}

func (suite *IntegrationTestSuite) setupRoutes() {
	// Public routes
	suite.router.POST("/register", createUserHandler(suite.dbPool, suite.ctx))
	suite.router.POST("/login", loginHandler(suite.dbPool, suite.ctx))

	// Protected routes
	protected := suite.router.Group("/")
	protected.Use(authMiddleware())
	{
		protected.GET("/users/me", getMeHandler(suite.dbPool, suite.ctx))
		protected.GET("/profile/me", getMyProfileHandler(suite.dbPool, suite.ctx))
		protected.PUT("/profile/me", updateMyProfileHandler(suite.dbPool, suite.ctx))
	}
}

func (suite *IntegrationTestSuite) cleanupTestData() {
	// Delete test users
	_, _ = suite.dbPool.Exec(suite.ctx, "DELETE FROM users WHERE email LIKE '%@test.automation%'")
	_, _ = suite.dbPool.Exec(suite.ctx, "DELETE FROM users WHERE username LIKE 'test_%'")
}

// Test 1: User Registration Flow
func (suite *IntegrationTestSuite) TestUserRegistrationFlow() {
	t := suite.T()

	// Test successful registration
	payload := map[string]interface{}{
		"username": "test_user_" + fmt.Sprint(time.Now().Unix()),
		"email":    fmt.Sprintf("user%d@test.automation", time.Now().Unix()),
		"password": "TestPass123!",
		"role":     "client",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "token")
	assert.Contains(t, response, "user")

	// Save token for later tests
	suite.authToken = response["token"].(string)
	user := response["user"].(map[string]interface{})
	suite.userID = int(user["user_id"].(float64))
}

// Test 2: Duplicate Email Registration
func (suite *IntegrationTestSuite) TestDuplicateEmailRegistration() {
	t := suite.T()

	email := fmt.Sprintf("duplicate%d@test.automation", time.Now().Unix())

	// First registration
	payload := map[string]interface{}{
		"username": "test_dup1",
		"email":    email,
		"password": "TestPass123!",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Second registration with same email
	payload["username"] = "test_dup2"
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Email already registered")
}

// Test 3: Login Flow
func (suite *IntegrationTestSuite) TestLoginFlow() {
	t := suite.T()

	// Register a user first
	email := fmt.Sprintf("login%d@test.automation", time.Now().Unix())
	password := "TestPass123!"

	regPayload := map[string]interface{}{
		"username": "test_login",
		"email":    email,
		"password": password,
	}

	body, _ := json.Marshal(regPayload)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Test successful login
	loginPayload := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "token")
	assert.Contains(t, response, "user")
}

// Test 4: Invalid Login Credentials
func (suite *IntegrationTestSuite) TestInvalidLoginCredentials() {
	t := suite.T()

	// Test with non-existent user
	payload := map[string]interface{}{
		"email":    "nonexistent@test.automation",
		"password": "WrongPass123!",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test 5: Protected Route Access Without Token
func (suite *IntegrationTestSuite) TestProtectedRouteWithoutToken() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/users/me", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test 6: Protected Route Access With Valid Token
func (suite *IntegrationTestSuite) TestProtectedRouteWithValidToken() {
	t := suite.T()

	// Register user and get token
	suite.TestUserRegistrationFlow()

	req := httptest.NewRequest("GET", "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(suite.userID), response["user_id"])
}

// Test 7: Profile Update Flow
func (suite *IntegrationTestSuite) TestProfileUpdateFlow() {
	t := suite.T()

	// Register user and get token
	suite.TestUserRegistrationFlow()

	// Update profile
	updatePayload := map[string]interface{}{
		"bio":      "Test bio for automation",
		"location": "Bangkok, Thailand",
	}

	body, _ := json.Marshal(updatePayload)
	req := httptest.NewRequest("PUT", "/profile/me", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify update
	req = httptest.NewRequest("GET", "/profile/me", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Test bio for automation", response["bio"])
}

// Test 8: Invalid Input Validation
func (suite *IntegrationTestSuite) TestInvalidInputValidation() {
	t := suite.T()

	testCases := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Missing required fields",
			payload: map[string]interface{}{
				"email": "test@test.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid email format",
			payload: map[string]interface{}{
				"username": "testuser",
				"email":    "invalid-email",
				"password": "Pass123!",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

// Run the test suite
func TestIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(IntegrationTestSuite))
}
