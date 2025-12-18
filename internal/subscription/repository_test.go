package subscription

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func setupSubscriptionMock(t *testing.T) (Repository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	closer := func() { sqlxDB.Close() }
	return repo, mock, closer
}

func TestCreateSubscription(t *testing.T) {
	repo, mock, close := setupSubscriptionMock(t)
	defer close()

	ctx := context.Background()
	userID := 1
	gymID := 10
	subType := TypeUnlimitedPro
	priceCents := int64(25000)
	var visitsLimit *int

	now := time.Now()
	validUntil := now.AddDate(0, 1, 0)

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO subscriptions (user_id, gym_id, type, status, visits_limit, visits_used, period, price_cents, currency, valid_from, valid_until)
		VALUES ($1, $2, $3, 'active', $4, 0, 'monthly', $5, 'KZT', $6, $7)
		RETURNING id, user_id, gym_id, type, status, visits_limit, visits_used, period, price_cents, currency, valid_from, valid_until, created_at, updated_at
	`)).
		WithArgs(userID, gymID, subType, visitsLimit, priceCents, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "gym_id", "type", "status", "visits_limit", "visits_used",
			"period", "price_cents", "currency", "valid_from", "valid_until", "created_at", "updated_at",
		}).AddRow(
			1, userID, gymID, string(subType), "active", nil, 0,
			"monthly", priceCents, "KZT", now, validUntil, now, now,
		))

	sub, err := repo.CreateSubscription(ctx, userID, &gymID, subType, priceCents, visitsLimit)
	require.NoError(t, err)
	require.NotNil(t, sub)
	require.Equal(t, userID, sub.UserID)
	require.Equal(t, gymID, *sub.GymID)
	require.Equal(t, subType, sub.Type)
}

func TestGetActiveForUserAndGym(t *testing.T) {
	repo, mock, close := setupSubscriptionMock(t)
	defer close()

	ctx := context.Background()
	userID := 1
	gymID := 10

	now := time.Now()
	validUntil := now.AddDate(0, 1, 0)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT *
		FROM subscriptions
		WHERE user_id = $1
		  AND status = 'active'
		  AND valid_from <= NOW()
		  AND valid_until >= NOW()
		  AND (gym_id IS NULL OR gym_id = $2)
		ORDER BY
		  gym_id NULLS LAST,
		  price_cents DESC
		LIMIT 1
	`)).
		WithArgs(userID, gymID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "gym_id", "type", "status", "visits_limit", "visits_used",
			"period", "price_cents", "currency", "valid_from", "valid_until", "created_at", "updated_at",
		}).AddRow(
			1, userID, gymID, "unlimited_pro", "active", nil, 0,
			"monthly", 25000, "KZT", now, validUntil, now, now,
		))

	sub, err := repo.GetActiveForUserAndGym(ctx, userID, gymID)
	require.NoError(t, err)
	require.NotNil(t, sub)
	require.Equal(t, userID, sub.UserID)
}

func TestGetActiveForUserAndGym_NotFound(t *testing.T) {
	repo, mock, close := setupSubscriptionMock(t)
	defer close()

	ctx := context.Background()
	userID := 1
	gymID := 10

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT *
		FROM subscriptions
		WHERE user_id = $1
		  AND status = 'active'
		  AND valid_from <= NOW()
		  AND valid_until >= NOW()
		  AND (gym_id IS NULL OR gym_id = $2)
		ORDER BY
		  gym_id NULLS LAST,
		  price_cents DESC
		LIMIT 1
	`)).
		WithArgs(userID, gymID).
		WillReturnError(sql.ErrNoRows)

	sub, err := repo.GetActiveForUserAndGym(ctx, userID, gymID)
	// sqlx.GetContext returns sql.ErrNoRows when no rows found
	require.Error(t, err)
	// Even though sub may be populated with zero values, the error should be present
	require.ErrorIs(t, err, sql.ErrNoRows)
	// The subscription should be empty/zero-valued when error occurs
	require.Equal(t, 0, sub.ID)
}

func TestIncrementVisits(t *testing.T) {
	repo, mock, close := setupSubscriptionMock(t)
	defer close()

	ctx := context.Background()
	subID := 1

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE subscriptions
		SET visits_used = visits_used + 1,
		    updated_at = NOW()
		WHERE id = $1
	`)).
		WithArgs(subID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.IncrementVisits(ctx, subID)
	require.NoError(t, err)
}

func TestListActiveByUser(t *testing.T) {
	repo, mock, close := setupSubscriptionMock(t)
	defer close()

	ctx := context.Background()
	userID := 1

	now := time.Now()
	validUntil := now.AddDate(0, 1, 0)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT *
		FROM subscriptions
		WHERE user_id = $1
		  AND status = 'active'
		  AND valid_from <= NOW()
		  AND valid_until >= NOW()
		ORDER BY created_at DESC
	`)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "gym_id", "type", "status", "visits_limit", "visits_used",
			"period", "price_cents", "currency", "valid_from", "valid_until", "created_at", "updated_at",
		}).
			AddRow(1, userID, 10, "unlimited_pro", "active", nil, 0, "monthly", 25000, "KZT", now, validUntil, now, now).
			AddRow(2, userID, nil, "multi_gym_flex", "active", intPtr(20), 5, "monthly", 18000, "KZT", now, validUntil, now, now))

	subs, err := repo.ListActiveByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, subs, 2)
	require.Equal(t, userID, subs[0].UserID)
}

func intPtr(i int) *int {
	return &i
}

