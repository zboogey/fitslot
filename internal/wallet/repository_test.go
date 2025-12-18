package wallet

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

func setupWalletMock(t *testing.T) (Repository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	closer := func() { sqlxDB.Close() }
	return repo, mock, closer
}

func TestGetOrCreateWallet_WhenNotExists(t *testing.T) {
	repo, mock, close := setupWalletMock(t)
	defer close()

	ctx := context.Background()

	// GetContext should return no rows -> insert
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM wallets WHERE user_id = $1")).
		WithArgs(10).
		WillReturnError(sql.ErrNoRows)

	// Insert returning
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO wallets (user_id) VALUES ($1) RETURNING id, user_id, balance_cents, currency, created_at, updated_at")).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance_cents", "currency", "created_at", "updated_at"}).AddRow(5, 10, 1000, "KZT", time.Now(), time.Now()))

	w, err := repo.GetOrCreateWallet(ctx, 10)
	require.NoError(t, err)
	require.Equal(t, 5, w.ID)
}

func TestAddTransaction_Success_UpdateAndInsert(t *testing.T) {
	repo, mock, close := setupWalletMock(t)
	defer close()

	ctx := context.Background()

	// Begin
	mock.ExpectBegin()

	// SELECT FOR UPDATE returns existing wallet
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, balance_cents, currency, created_at, updated_at FROM wallets WHERE user_id = $1 FOR UPDATE")).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance_cents", "currency", "created_at", "updated_at"}).AddRow(7, 20, 2000, "KZT", time.Now(), time.Now()))

	// UPDATE wallets
	mock.ExpectExec(regexp.QuoteMeta("UPDATE wallets SET balance_cents = $1, updated_at = NOW() WHERE id = $2")).
		WithArgs(1500, 7).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// INSERT wallet_transactions
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO wallet_transactions (wallet_id, amount_cents, type, balance_after) VALUES ($1, $2, $3, $4)")).
		WithArgs(7, -500, "booking_payment", 1500).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.AddTransaction(ctx, 20, -500, "booking_payment")
	require.NoError(t, err)
}
