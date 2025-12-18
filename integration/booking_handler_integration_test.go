package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fitslot/internal/auth"
	"fitslot/internal/booking"
	"fitslot/internal/db"
	"fitslot/internal/email"
	"fitslot/internal/gym"
	"fitslot/internal/subscription"
	"fitslot/internal/user"
	"fitslot/internal/wallet"
)

var (
	testDB           *sqlx.DB
	testEmailService *email.Service
)

func setupTestDB(t *testing.T) *sqlx.DB {
	// Use environment variables for test database connection
	// Default values match docker-compose.test.yml
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("TEST_DB_PORT")
	if dbPort == "" {
		dbPort = "5434"
	}

	dbUser := os.Getenv("TEST_DB_USER")
	if dbUser == "" {
		dbUser = "testuser"
	}

	dbPassword := os.Getenv("TEST_DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "password"
	}

	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = "fitslot_test"
	}

	// Allow full DSN override
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPassword, dbHost, dbPort, dbName)
	}

	database, err := db.Connect(dsn)
	if err != nil {
		t.Skipf("Skipping integration tests: cannot connect to test database: %v", err)
	}

	// Run migrations on test database
	if err := db.RunMigrations(database, "../migrations"); err != nil {
		t.Logf("Warning: Failed to run migrations: %v", err)
	}

	return database
}

func cleanDatabase(t *testing.T, db *sqlx.DB) {
	ctx := context.Background()
	tables := []string{
		"bookings",
		"wallet_transactions",
		"subscriptions",
		"time_slots",
		"gyms",
		"users",
		"wallets",
	}

	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table))
		require.NoError(t, err, "Failed to clean table "+table)
	}
}

func createTestUser(t *testing.T, db *sqlx.DB, email, name string) int {
	hashedPassword, _ := auth.HashPassword("password123")

	ctx := context.Background()
	var userID int
	err := db.QueryRowContext(ctx, `
		INSERT INTO users (email, name, password_hash, role)
		VALUES ($1, $2, $3, 'user')
		RETURNING id
	`, email, name, hashedPassword).Scan(&userID)

	require.NoError(t, err)
	return userID
}

func createTestGym(t *testing.T, db *sqlx.DB, name string) int {
	ctx := context.Background()
	var gymID int
	err := db.QueryRowContext(ctx, `
		INSERT INTO gyms (name, location)
		VALUES ($1, 'Test Location')
		RETURNING id
	`, name).Scan(&gymID)

	require.NoError(t, err)
	return gymID
}

func createTestTimeSlot(t *testing.T, db *sqlx.DB, gymID int, startTime time.Time, capacity int) int {
	ctx := context.Background()
	var slotID int
	err := db.QueryRowContext(ctx, `
		INSERT INTO time_slots (gym_id, start_time, end_time, capacity)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, gymID, startTime, startTime.Add(1*time.Hour), capacity).Scan(&slotID)

	require.NoError(t, err)
	return slotID
}

func addWalletBalance(t *testing.T, db *sqlx.DB, userID int, amountCents int64) {
	ctx := context.Background()
	// Ensure wallet exists
	_, err := db.ExecContext(ctx, `
		INSERT INTO wallets (user_id, balance_cents, currency)
		VALUES ($1, $2, 'KZT')
		ON CONFLICT (user_id) DO UPDATE SET balance_cents = wallets.balance_cents + EXCLUDED.balance_cents
	`, userID, amountCents)
	if err != nil {
		require.NoError(t, err)
		return
	}

	// Insert wallet transaction
	_, err = db.ExecContext(ctx, `
		INSERT INTO wallet_transactions (wallet_id, amount_cents, type, balance_after, created_at)
		SELECT id, $2, 'topup', balance_cents, NOW() FROM wallets WHERE user_id = $1
	`, userID, amountCents)

	require.NoError(t, err)
}

func createTestSubscription(t *testing.T, db *sqlx.DB, userID int, gymID int, visitsLimit *int) int {
	ctx := context.Background()
	var subID int
	err := db.QueryRowContext(ctx, `
		INSERT INTO subscriptions (user_id, gym_id, type, status, visits_limit, visits_used, period, price_cents, currency, valid_from, valid_until)
		VALUES ($1, $2, 'unlimited_pro', 'active', $3, 0, 'monthly', 25000, 'KZT', NOW(), NOW() + INTERVAL '1 month')
		RETURNING id
	`, userID, gymID, visitsLimit).Scan(&subID)

	require.NoError(t, err)
	return subID
}

func generateTestToken(userID int, email, role, secret string) string {
	accessToken, _, _ := auth.GenerateTokens(userID, email, role, secret, secret)
	return accessToken
}

func TestBookingFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanDatabase(t, db)

	// Setup
	emailService := email.New("test@fitslot.com", "FitSlot", "mailhog", "1025", "", "", "localhost:6380")

	// Initialize repositories and services
	bookingRepo := booking.NewRepository(db)
	gymRepo := gym.NewRepository(db)
	subscriptionRepo := subscription.NewRepository(db)
	walletRepo := wallet.NewRepository(db)
	userRepo := user.NewRepository(db)

	bookingService := booking.NewService(
		bookingRepo,
		gymRepo,
		subscriptionRepo,
		walletRepo,
		userRepo,
		emailService,
	)

	handler := booking.NewHandler(bookingService)

	router := gin.New()
	router.POST("/bookings/:slotID", auth.AuthMiddleware("test-secret"), handler.BookSlot)

	t.Run("Successfully book slot with wallet", func(t *testing.T) {
		cleanDatabase(t, db)

		// Create test data
		userID := createTestUser(t, db, "user@example.com", "Test User")
		gymID := createTestGym(t, db, "Test Gym")
		futureTime := time.Now().Add(24 * time.Hour)
		slotID := createTestTimeSlot(t, db, gymID, futureTime, 10)
		addWalletBalance(t, db, userID, 5000) // 50.00 in cents

		// Generate token
		token := generateTestToken(userID, "user@example.com", "user", "test-secret")

		// Make request
		req := httptest.NewRequest("POST", fmt.Sprintf("/bookings/%d", slotID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "wallet", response["paid_with"])
		assert.NotNil(t, response["booking"])
	})

	t.Run("Successfully book slot with subscription", func(t *testing.T) {
		cleanDatabase(t, db)

		userID := createTestUser(t, db, "sub@example.com", "Sub User")
		gymID := createTestGym(t, db, "Test Gym")
		futureTime := time.Now().Add(24 * time.Hour)
		slotID := createTestTimeSlot(t, db, gymID, futureTime, 10)

		// Create unlimited subscription
		createTestSubscription(t, db, userID, gymID, nil)

		token := generateTestToken(userID, "sub@example.com", "user", "test-secret")

		req := httptest.NewRequest("POST", fmt.Sprintf("/bookings/%d", slotID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "subscription", response["paid_with"])
		assert.NotNil(t, response["booking"])
	})

	t.Run("Fail booking slot in the past", func(t *testing.T) {
		cleanDatabase(t, db)

		userID := createTestUser(t, db, "user@example.com", "Test User")
		gymID := createTestGym(t, db, "Test Gym")
		pastTime := time.Now().Add(-24 * time.Hour)
		slotID := createTestTimeSlot(t, db, gymID, pastTime, 10)
		addWalletBalance(t, db, userID, 5000)

		token := generateTestToken(userID, "user@example.com", "user", "test-secret")

		req := httptest.NewRequest("POST", fmt.Sprintf("/bookings/%d", slotID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Cannot book a slot in the past")
	})

	t.Run("Fail booking full slot", func(t *testing.T) {
		cleanDatabase(t, db)

		user1ID := createTestUser(t, db, "user1@example.com", "User 1")
		user2ID := createTestUser(t, db, "user2@example.com", "User 2")
		gymID := createTestGym(t, db, "Test Gym")
		futureTime := time.Now().Add(24 * time.Hour)
		slotID := createTestTimeSlot(t, db, gymID, futureTime, 1) // Capacity 1

		addWalletBalance(t, db, user1ID, 5000)
		addWalletBalance(t, db, user2ID, 5000)

		// First user books
		token1 := generateTestToken(user1ID, "user1@example.com", "user", "test-secret")
		req1 := httptest.NewRequest("POST", fmt.Sprintf("/bookings/%d", slotID), nil)
		req1.Header.Set("Authorization", "Bearer "+token1)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusCreated, w1.Code)

		// Second user tries to book (should fail)
		token2 := generateTestToken(user2ID, "user2@example.com", "user", "test-secret")
		req2 := httptest.NewRequest("POST", fmt.Sprintf("/bookings/%d", slotID), nil)
		req2.Header.Set("Authorization", "Bearer "+token2)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusConflict, w2.Code)
		assert.Contains(t, w2.Body.String(), "Time slot is full")
	})

	t.Run("Fail double booking same slot", func(t *testing.T) {
		cleanDatabase(t, db)

		userID := createTestUser(t, db, "user@example.com", "Test User")
		gymID := createTestGym(t, db, "Test Gym")
		futureTime := time.Now().Add(24 * time.Hour)
		slotID := createTestTimeSlot(t, db, gymID, futureTime, 10)
		addWalletBalance(t, db, userID, 10000)

		token := generateTestToken(userID, "user@example.com", "user", "test-secret")

		// First booking
		req1 := httptest.NewRequest("POST", fmt.Sprintf("/bookings/%d", slotID), nil)
		req1.Header.Set("Authorization", "Bearer "+token)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusCreated, w1.Code)

		// Second booking (should fail)
		req2 := httptest.NewRequest("POST", fmt.Sprintf("/bookings/%d", slotID), nil)
		req2.Header.Set("Authorization", "Bearer "+token)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusConflict, w2.Code)
		assert.Contains(t, w2.Body.String(), "already have a booking")
	})

	t.Run("Fail booking with insufficient wallet balance", func(t *testing.T) {
		cleanDatabase(t, db)

		userID := createTestUser(t, db, "user@example.com", "Test User")
		gymID := createTestGym(t, db, "Test Gym")
		futureTime := time.Now().Add(24 * time.Hour)
		slotID := createTestTimeSlot(t, db, gymID, futureTime, 10)
		addWalletBalance(t, db, userID, 500) // Only 5.00, need 10.00

		token := generateTestToken(userID, "user@example.com", "user", "test-secret")

		req := httptest.NewRequest("POST", fmt.Sprintf("/bookings/%d", slotID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusPaymentRequired, w.Code)
		assert.Contains(t, w.Body.String(), "insufficient wallet balance")
	})

	t.Run("Fail booking non-existent slot", func(t *testing.T) {
		cleanDatabase(t, db)

		userID := createTestUser(t, db, "user@example.com", "Test User")
		addWalletBalance(t, db, userID, 5000)

		token := generateTestToken(userID, "user@example.com", "user", "test-secret")

		req := httptest.NewRequest("POST", "/bookings/99999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Time slot not found")
	})

	t.Run("Fail booking without authentication", func(t *testing.T) {
		cleanDatabase(t, db)

		gymID := createTestGym(t, db, "Test Gym")
		futureTime := time.Now().Add(24 * time.Hour)
		slotID := createTestTimeSlot(t, db, gymID, futureTime, 10)

		req := httptest.NewRequest("POST", fmt.Sprintf("/bookings/%d", slotID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCancelBooking(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanDatabase(t, db)

	emailService := email.New("test@fitslot.com", "FitSlot", "mailhog", "1025", "", "", "localhost:6380")

	// Initialize repositories and services
	bookingRepo := booking.NewRepository(db)
	gymRepo := gym.NewRepository(db)
	subscriptionRepo := subscription.NewRepository(db)
	walletRepo := wallet.NewRepository(db)
	userRepo := user.NewRepository(db)

	bookingService := booking.NewService(
		bookingRepo,
		gymRepo,
		subscriptionRepo,
		walletRepo,
		userRepo,
		emailService,
	)

	handler := booking.NewHandler(bookingService)

	router := gin.New()
	router.POST("/bookings/:slotID", auth.AuthMiddleware("test-secret"), handler.BookSlot)
	router.DELETE("/bookings/:bookingID", auth.AuthMiddleware("test-secret"), handler.CancelBooking)

	t.Run("Successfully cancel booking", func(t *testing.T) {
		cleanDatabase(t, db)

		userID := createTestUser(t, db, "user@example.com", "Test User")
		gymID := createTestGym(t, db, "Test Gym")
		futureTime := time.Now().Add(24 * time.Hour)
		slotID := createTestTimeSlot(t, db, gymID, futureTime, 10)
		addWalletBalance(t, db, userID, 5000)

		token := generateTestToken(userID, "user@example.com", "user", "test-secret")

		// Create booking
		reqBook := httptest.NewRequest("POST", fmt.Sprintf("/bookings/%d", slotID), nil)
		reqBook.Header.Set("Authorization", "Bearer "+token)
		wBook := httptest.NewRecorder()
		router.ServeHTTP(wBook, reqBook)

		var bookingResponse map[string]interface{}
		json.Unmarshal(wBook.Body.Bytes(), &bookingResponse)
		bookingMap := bookingResponse["booking"].(map[string]interface{})
		bookingID := int(bookingMap["id"].(float64))

		// Cancel booking
		reqCancel := httptest.NewRequest("DELETE", fmt.Sprintf("/bookings/%d", bookingID), nil)
		reqCancel.Header.Set("Authorization", "Bearer "+token)
		wCancel := httptest.NewRecorder()
		router.ServeHTTP(wCancel, reqCancel)

		assert.Equal(t, http.StatusOK, wCancel.Code)
		assert.Contains(t, wCancel.Body.String(), "cancelled successfully")
	})

	t.Run("Fail cancelling other user's booking", func(t *testing.T) {
		cleanDatabase(t, db)

		user1ID := createTestUser(t, db, "user1@example.com", "User 1")
		user2ID := createTestUser(t, db, "user2@example.com", "User 2")
		gymID := createTestGym(t, db, "Test Gym")
		futureTime := time.Now().Add(24 * time.Hour)
		slotID := createTestTimeSlot(t, db, gymID, futureTime, 10)
		addWalletBalance(t, db, user1ID, 5000)

		token1 := generateTestToken(user1ID, "user1@example.com", "user", "test-secret")
		token2 := generateTestToken(user2ID, "user2@example.com", "user", "test-secret")

		// User 1 creates booking
		reqBook := httptest.NewRequest("POST", fmt.Sprintf("/bookings/%d", slotID), nil)
		reqBook.Header.Set("Authorization", "Bearer "+token1)
		wBook := httptest.NewRecorder()
		router.ServeHTTP(wBook, reqBook)

		var bookingResponse map[string]interface{}
		json.Unmarshal(wBook.Body.Bytes(), &bookingResponse)
		bookingMap := bookingResponse["booking"].(map[string]interface{})
		bookingID := int(bookingMap["id"].(float64))

		// User 2 tries to cancel
		reqCancel := httptest.NewRequest("DELETE", fmt.Sprintf("/bookings/%d", bookingID), nil)
		reqCancel.Header.Set("Authorization", "Bearer "+token2)
		wCancel := httptest.NewRecorder()
		router.ServeHTTP(wCancel, reqCancel)

		assert.Equal(t, http.StatusForbidden, wCancel.Code)
		assert.Contains(t, wCancel.Body.String(), "can only cancel your own bookings")
	})

	t.Run("Fail cancelling non-existent booking", func(t *testing.T) {
		cleanDatabase(t, db)

		userID := createTestUser(t, db, "user@example.com", "Test User")
		token := generateTestToken(userID, "user@example.com", "user", "test-secret")

		req := httptest.NewRequest("DELETE", "/bookings/99999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestListMyBookings(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanDatabase(t, db)

	emailService := email.New("test@fitslot.com", "FitSlot", "mailhog", "1025", "", "", "localhost:6380")

	// Initialize repositories and services
	bookingRepo := booking.NewRepository(db)
	gymRepo := gym.NewRepository(db)
	subscriptionRepo := subscription.NewRepository(db)
	walletRepo := wallet.NewRepository(db)
	userRepo := user.NewRepository(db)

	bookingService := booking.NewService(
		bookingRepo,
		gymRepo,
		subscriptionRepo,
		walletRepo,
		userRepo,
		emailService,
	)

	handler := booking.NewHandler(bookingService)

	router := gin.New()
	router.GET("/bookings/my", auth.AuthMiddleware("test-secret"), handler.ListMyBookings)

	t.Run("List user bookings", func(t *testing.T) {
		cleanDatabase(t, db)

		userID := createTestUser(t, db, "user@example.com", "Test User")
		token := generateTestToken(userID, "user@example.com", "user", "test-secret")

		req := httptest.NewRequest("GET", "/bookings/my", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var bookings []interface{}
		json.Unmarshal(w.Body.Bytes(), &bookings)

		// Should be empty initially
		assert.Equal(t, 0, len(bookings))
	})
}

func init() {
	// Initialize logger for tests
	// logger.Init() // Commented out to avoid initialization issues in tests
}
