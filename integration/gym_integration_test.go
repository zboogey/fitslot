package booking_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"fitslot/internal/gym"
)

func cleanGymTables(t *testing.T, db *sqlx.DB) {
	tables := []string{"time_slots", "gyms"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		require.NoError(t, err, "Failed to clean table "+table)
	}
}

func TestCreateGym_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	cleanGymTables(t, db)

	// Create router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := gym.NewHandler(db)

	// Register route
	router.POST("/gyms", handler.CreateGym)

	// Test creating a gym
	reqBody := map[string]string{
		"name":     "Test Gym",
		"location": "Test Location",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/gyms", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var response gym.Gym
	json.Unmarshal(w.Body.Bytes(), &response)
	require.Equal(t, "Test Gym", response.Name)
	require.Equal(t, "Test Location", response.Location)
}

func TestListGyms_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	cleanGymTables(t, db)

	// Create router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := gym.NewHandler(db)

	// Register route
	router.GET("/gyms", handler.ListGyms)

	// Create some test gyms
	repo := gym.NewRepository(db)
	_, _ = repo.CreateGym("Gym 1", "Location 1")
	_, _ = repo.CreateGym("Gym 2", "Location 2")

	// Test listing gyms
	req, _ := http.NewRequest("GET", "/gyms", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response []gym.Gym
	json.Unmarshal(w.Body.Bytes(), &response)
	require.Len(t, response, 2)
}

func TestCreateTimeSlot_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	cleanGymTables(t, db)

	// Create router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := gym.NewHandler(db)

	// Register route
	router.POST("/gyms/:gymID/time-slots", handler.CreateTimeSlot)

	// Create a gym
	repo := gym.NewRepository(db)
	g, _ := repo.CreateGym("Test Gym", "Test Location")

	// Test creating a time slot
	reqBody := map[string]interface{}{
		"start_time":     "2025-02-01T10:00:00Z",
		"end_time":       "2025-02-01T11:00:00Z",
		"capacity":       20,
		"price_cents":    1000,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/gyms/%d/time-slots", g.ID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestGetTimeSlots_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	cleanGymTables(t, db)

	// Create router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := gym.NewHandler(db)

	// Register route
	router.GET("/gyms/:gymID/time-slots", handler.ListTimeSlots)

	// Create a gym and time slots
	repo := gym.NewRepository(db)
	g, _ := repo.CreateGym("Test Gym", "Test Location")

	// Test getting time slots
	req, _ := http.NewRequest("GET", fmt.Sprintf("/gyms/%d/time-slots", g.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
