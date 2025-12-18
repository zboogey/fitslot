package subscription

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func setupSubMock(t *testing.T) (*Repository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	closer := func() { sqlxDB.Close() }
	return repo, mock, closer
}

func TestCreateSubscription(t *testing.T) {
	repo, mock, close := setupSubMock(t)
	defer close()

	ctx := context.Background()
	visitsLimit := 10

	// Don't match exact time values since time.Now() differs between calls
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO subscriptions (user_id, gym_id, type, status, visits_limit, visits_used, period, price_cents, currency, valid_from, valid_until) VALUES ($1, $2, $3, 'active', $4, 0, 'monthly', $5, 'KZT', $6, $7) RETURNING id, user_id, gym_id, type, status, visits_limit, visits_used, period, price_cents, currency, valid_from, valid_until, created_at, updated_at")).
		WithArgs(1, 2, TypeSingleGymLite, &visitsLimit, int64(5000), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "gym_id", "type", "status", "visits_limit", "visits_used", "period", "price_cents", "currency", "valid_from", "valid_until", "created_at", "updated_at"}).
			AddRow(1, 1, 2, TypeSingleGymLite, StatusActive, &visitsLimit, 0, "monthly", int64(5000), "KZT", time.Now(), time.Now().AddDate(0,1,0), time.Now(), time.Now()))

	sub, err := repo.CreateSubscription(ctx, 1, intPtr(2), TypeSingleGymLite, 5000, &visitsLimit)
	require.NoError(t, err)
	require.Equal(t, 1, sub.ID)
	require.Equal(t, StatusActive, sub.Status)
}

func TestIncrementVisits(t *testing.T) {
	repo, mock, close := setupSubMock(t)
	defer close()

	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE subscriptions SET visits_used = visits_used + 1,\n\t\t    updated_at = NOW()\n\t\tWHERE id = $1")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.IncrementVisits(ctx, 1)
	require.NoError(t, err)
}

func TestListActiveByUser(t *testing.T) {
	repo, mock, close := setupSubMock(t)
	defer close()

	ctx := context.Background()
	now := time.Now()
	visitsLimit := 10

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM subscriptions WHERE user_id = $1 AND status = 'active' AND valid_from <= NOW() AND valid_until >= NOW() ORDER BY created_at DESC")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "gym_id", "type", "status", "visits_limit", "visits_used", "period", "price_cents", "currency", "valid_from", "valid_until", "created_at", "updated_at"}).
			AddRow(1, 1, 2, TypeSingleGymLite, StatusActive, &visitsLimit, 0, "monthly", int64(5000), "KZT", now, now.AddDate(0,1,0), now, now))

	subs, err := repo.ListActiveByUser(ctx, 1)
	require.NoError(t, err)
	require.Len(t, subs, 1)
}

func intPtr(i int) *int {
	return &i
}
