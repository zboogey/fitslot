package booking_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"fitslot/internal/auth"
	"fitslot/internal/subscription"
)

func cleanSubscriptionTables(t *testing.T, db *sqlx.DB) {
	tables := []string{"subscriptions", "gyms", "users"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		require.NoError(t, err, "Failed to clean table "+table)
	}
}

func createSubscriptionTestUser(t *testing.T, db *sqlx.DB, email, name string) int {
	hashedPassword, _ := auth.HashPassword("password123")
	
	var userID int
	err := db.QueryRow(`
		INSERT INTO users (email, name, password_hash, role)
		VALUES ($1, $2, $3, 'user')
		RETURNING id
	`, email, name, hashedPassword).Scan(&userID)
	
	require.NoError(t, err)
	return userID
}

func createSubscriptionTestGym(t *testing.T, db *sqlx.DB, name string) int {
	var gymID int
	err := db.QueryRow(`
		INSERT INTO gyms (name, location)
		VALUES ($1, 'Test Location')
		RETURNING id
	`, name).Scan(&gymID)
	
	require.NoError(t, err)
	return gymID
}

func TestCreateSubscription_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()
	
	cleanSubscriptionTables(t, db)

	repo := subscription.NewRepository(db)
	ctx := context.Background()

	userID := createSubscriptionTestUser(t, db, "subuser@test.com", "Sub User")
	gymID := createSubscriptionTestGym(t, db, "Test Gym")

	visitsLimit := 20
	sub, err := repo.CreateSubscription(ctx, userID, &gymID, subscription.TypeSingleGymLite, 5000, &visitsLimit)
	require.NoError(t, err)
	require.Equal(t, userID, sub.UserID)
	require.Equal(t, &gymID, sub.GymID)
	require.Equal(t, subscription.StatusActive, sub.Status)
	require.Equal(t, &visitsLimit, sub.VisitsLimit)
}

func TestGetActiveSubscription_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()
	
	cleanSubscriptionTables(t, db)

	repo := subscription.NewRepository(db)
	ctx := context.Background()

	userID := createSubscriptionTestUser(t, db, "activesub@test.com", "Active Sub User")
	gymID := createSubscriptionTestGym(t, db, "Active Gym")

	visitsLimit := 10
	sub, err := repo.CreateSubscription(ctx, userID, &gymID, subscription.TypeSingleGymLite, 3000, &visitsLimit)
	require.NoError(t, err)

	// Get active subscription
	activeSub, err := repo.GetActiveForUserAndGym(ctx, userID, gymID)
	require.NoError(t, err)
	require.Equal(t, sub.ID, activeSub.ID)
	require.Equal(t, subscription.StatusActive, activeSub.Status)
}

func TestIncrementVisits_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()
	
	cleanSubscriptionTables(t, db)

	repo := subscription.NewRepository(db)
	ctx := context.Background()

	userID := createSubscriptionTestUser(t, db, "visits@test.com", "Visits User")
	gymID := createSubscriptionTestGym(t, db, "Visits Gym")

	visitsLimit := 5
	sub, err := repo.CreateSubscription(ctx, userID, &gymID, subscription.TypeSingleGymLite, 2000, &visitsLimit)
	require.NoError(t, err)
	require.Equal(t, 0, sub.VisitsUsed)

	// Increment visits
	err = repo.IncrementVisits(ctx, sub.ID)
	require.NoError(t, err)

	// Check visits incremented
	activeSub, err := repo.GetActiveForUserAndGym(ctx, userID, gymID)
	require.NoError(t, err)
	require.Equal(t, 1, activeSub.VisitsUsed)
}

func TestListActiveSubscriptions_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()
	
	cleanSubscriptionTables(t, db)

	repo := subscription.NewRepository(db)
	ctx := context.Background()

	userID := createSubscriptionTestUser(t, db, "listuser@test.com", "List User")
	gymID1 := createSubscriptionTestGym(t, db, "Gym 1")
	gymID2 := createSubscriptionTestGym(t, db, "Gym 2")

	// Create two subscriptions
	_, err := repo.CreateSubscription(ctx, userID, &gymID1, subscription.TypeSingleGymLite, 3000, nil)
	require.NoError(t, err)
	
	_, err = repo.CreateSubscription(ctx, userID, &gymID2, subscription.TypeUnlimitedPro, 5000, nil)
	require.NoError(t, err)

	// List active subscriptions
	subs, err := repo.ListActiveByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, subs, 2)
}
