package booking_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"fitslot/internal/auth"
	"fitslot/internal/booking"
	"fitslot/internal/email"
	"fitslot/internal/subscription"
	"fitslot/internal/wallet"
)

func cleanBookingHandlerTables(t *testing.T, db *sqlx.DB) {
	tables := []string{
		"bookings",
		"wallet_transactions",
		"subscriptions",
		"time_slots",
		"gyms",
		"wallets",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		require.NoError(t, err, "Failed to clean table "+table)
	}
}

func createBookingHandlerTestUser(t *testing.T, db *sqlx.DB, email, name string) (int, string) {
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

func createBookingHandlerTestGym(t *testing.T, db *sqlx.DB, name string) int {
	var gymID int
	err := db.QueryRow(`
		INSERT INTO gyms (name, location)
		VALUES ($1, 'Test Location')
		RETURNING id
	`, name).Scan(&gymID)

	require.NoError(t, err)
	return gymID
}

func createBookingHandlerTestTimeSlot(t *testing.T, db *sqlx.DB, gymID int) int {
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(time.Hour)

	var slotID int
	err := db.QueryRow(`
		INSERT INTO time_slots (gym_id, start_time, end_time, capacity, price_cents)
		VALUES ($1, $2, $3, 20, 1000)
		RETURNING id
	`, gymID, startTime, endTime).Scan(&slotID)

	require.NoError(t, err)
	return slotID
}

func TestBookSlotHandler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	cleanBookingHandlerTables(t, db)

	// Create test data
	userID, token := createBookingHandlerTestUser(t, db, "booker@test.com", "Booker")
	gymID := createBookingHandlerTestGym(t, db, "Test Gym")
	slotID := createBookingHandlerTestTimeSlot(t, db, gymID)

	// Add wallet balance
	walletRepo := wallet.NewRepository(db)
	ctx := context.Background()
	_ = walletRepo.AddTransaction(ctx, userID, 5000, "topup")

	// Create router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create email service (mock)
	emailService := email.New("", "", "", "", "", "", "")

	// Create booking handler
	handler := booking.NewHandler(db, emailService)

	// Add auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("email", "booker@test.com")
		c.Next()
	})

	// Register routes
	router.POST("/bookings", handler.BookSlot)

	// Test booking a slot with wallet
	reqBody := map[string]interface{}{
		"slot_id": slotID,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/bookings", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestBookSlotWithSubscription_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	cleanBookingHandlerTables(t, db)

	// Create test data
	userID, token := createBookingHandlerTestUser(t, db, "bookersub@test.com", "Booker Sub")
	gymID := createBookingHandlerTestGym(t, db, "Sub Gym")
	slotID := createBookingHandlerTestTimeSlot(t, db, gymID)

	// Create subscription
	subRepo := subscription.NewRepository(db)
	ctx := context.Background()
	visitsLimit := 10
	_, _ = subRepo.CreateSubscription(ctx, userID, &gymID, subscription.TypeSingleGymLite, 5000, &visitsLimit)

	// Create router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	emailService := email.New("", "", "", "", "", "", "")
	handler := booking.NewHandler(db, emailService)

	// Add auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("email", "bookersub@test.com")
		c.Next()
	})

	// Register route
	router.POST("/bookings", handler.BookSlot)

	// Test booking with subscription
	reqBody := map[string]interface{}{
		"slot_id": slotID,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/bookings", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestCancelBookingHandler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	cleanBookingHandlerTables(t, db)

	// Create test data and booking
	userID, token := createBookingHandlerTestUser(t, db, "canceller@test.com", "Canceller")
	gymID := createBookingHandlerTestGym(t, db, "Cancel Gym")
	slotID := createBookingHandlerTestTimeSlot(t, db, gymID)

	// Add wallet and create booking
	walletRepo := wallet.NewRepository(db)
	ctx := context.Background()
	_ = walletRepo.AddTransaction(ctx, userID, 5000, "topup")

	bookingRepo := booking.NewRepository(db)
	b, _ := bookingRepo.CreateBooking(userID, slotID)

	// Create router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	emailService := email.New("", "", "", "", "", "", "")
	handler := booking.NewHandler(db, emailService)

	// Add auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("email", "canceller@test.com")
		c.Next()
	})

	// Register route
	router.POST("/bookings/:bookingID/cancel", handler.CancelBooking)

	// Test cancelling booking
	req, _ := http.NewRequest("POST", fmt.Sprintf("/bookings/%d/cancel", b.ID), bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestListMyBookingsHandler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	cleanBookingHandlerTables(t, db)

	// Create test data
	userID, _ := createBookingHandlerTestUser(t, db, "lister@test.com", "Lister")
	gymID := createBookingHandlerTestGym(t, db, "List Gym")
	slotID := createBookingHandlerTestTimeSlot(t, db, gymID)

	// Add wallet and create booking
	walletRepo := wallet.NewRepository(db)
	ctx := context.Background()
	_ = walletRepo.AddTransaction(ctx, userID, 5000, "topup")

	bookingRepo := booking.NewRepository(db)
	_, _ = bookingRepo.CreateBooking(userID, slotID)

	// Create router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	emailService := email.New("", "", "", "", "", "", "")
	handler := booking.NewHandler(db, emailService)

	// Add auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})

	// Register route
	router.GET("/bookings/my", handler.ListMyBookings)

	// Test listing bookings
	req, _ := http.NewRequest("GET", "/bookings/my", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var bookings []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &bookings)
	require.Greater(t, len(bookings), 0)
}
