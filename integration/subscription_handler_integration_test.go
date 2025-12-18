package booking_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"fitslot/internal/auth"
	"fitslot/internal/subscription"
	"fitslot/internal/wallet"
)

func cleanSubscriptionHandlerTables(t *testing.T, db *sqlx.DB) {
	tables := []string{"subscriptions", "wallets", "wallet_transactions", "users", "gyms"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		require.NoError(t, err, "Failed to clean table "+table)
	}
}

func createSubHandlerTestUser(t *testing.T, db *sqlx.DB, email, name string) (int, string) {
	hashedPassword, _ := auth.HashPassword("password123")

	var userID int
	err := db.QueryRow(`
		INSERT INTO users (email, name, password_hash, role)
		VALUES ($1, $2, $3, 'user')
		RETURNING id
	`, email, name, hashedPassword).Scan(&userID)

	require.NoError(t, err)

	// Generate token for authenticated requests
	token, _ := auth.GenerateAccessToken(userID, email, "user", "test-secret")
	return userID, token
}

func createSubHandlerTestGym(t *testing.T, db *sqlx.DB, name string) int {
	var gymID int
	err := db.QueryRow(`
		INSERT INTO gyms (name, location)
		VALUES ($1, 'Test Location')
		RETURNING id
	`, name).Scan(&gymID)

	require.NoError(t, err)
	return gymID
}

func TestGetPlans_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	// Create router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := subscription.NewHandler(db)

	// Register route
	router.GET("/subscriptions/plans", handler.ListPlans)

	// Test getting plans
	req, _ := http.NewRequest("GET", "/subscriptions/plans", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var plans []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &plans)
	require.Greater(t, len(plans), 0)
}

func TestCreateSubscriptionHandler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	cleanSubscriptionHandlerTables(t, db)

	// Create test user and gym
	userID, token := createSubHandlerTestUser(t, db, "subhandler@test.com", "Sub Handler User")
	gymID := createSubHandlerTestGym(t, db, "Test Gym")

	// Add wallet balance
	walletRepo := wallet.NewRepository(db)
	ctx := context.Background()
	_ = walletRepo.AddTransaction(ctx, userID, 10000, "topup")

	// Create router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := subscription.NewHandler(db)

	// Add auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("email", "subhandler@test.com")
		c.Next()
	})

	// Register route
	router.POST("/subscriptions", handler.Create)

	// Test creating subscription
	reqBody := map[string]interface{}{
		"type":     "single_gym_lite",
		"gym_id":   gymID,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/subscriptions", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestGetActiveSubscriptions_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	cleanSubscriptionHandlerTables(t, db)

	// Create test user, gym, and subscription
	userID, _ := createSubHandlerTestUser(t, db, "activesubhandler@test.com", "Active Sub Handler")
	gymID := createSubHandlerTestGym(t, db, "Active Gym")

	subRepo := subscription.NewRepository(db)
	ctx := context.Background()
	visitsLimit := 8
	_, _ = subRepo.CreateSubscription(ctx, userID, &gymID, subscription.TypeSingleGymLite, 5000, &visitsLimit)

	// Create router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := subscription.NewHandler(db)

	// Add auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})

	// Register route
	router.GET("/subscriptions/active", handler.ListMy)

	// Test getting active subscriptions
	req, _ := http.NewRequest("GET", "/subscriptions/active", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var subs []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &subs)
	require.Greater(t, len(subs), 0)
}

func TestListPlansHandler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	// Create router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := subscription.NewHandler(db)

	// Register route
	router.GET("/subscriptions/plans", handler.ListPlans)

	// Test getting plans
	req, _ := http.NewRequest("GET", "/subscriptions/plans", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var plans []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &plans)
	require.Greater(t, len(plans), 0)
}
