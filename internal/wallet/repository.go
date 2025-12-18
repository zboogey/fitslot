package wallet

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetOrCreateWallet(ctx context.Context, userID int) (*Wallet, error) {
	w := &Wallet{}
	err := r.db.GetContext(ctx, w, `SELECT * FROM wallets WHERE user_id = $1`, userID)
	if err == nil {
		return w, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	err = r.db.QueryRowxContext(ctx,
		`INSERT INTO wallets (user_id)
		 VALUES ($1)
		 RETURNING id, user_id, balance_cents, currency, created_at, updated_at`,
		userID,
	).StructScan(w)

	if err != nil {
		return nil, err
	}

	return w, nil
}

func (r *Repository) AddTransaction(ctx context.Context, userID int, amountCents int64, txType string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var w Wallet
	err = tx.QueryRowxContext(ctx,
		`SELECT id, user_id, balance_cents, currency, created_at, updated_at
		 FROM wallets
		 WHERE user_id = $1
		 FOR UPDATE`,
		userID,
	).StructScan(&w)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = tx.QueryRowxContext(ctx,
				`INSERT INTO wallets (user_id)
				 VALUES ($1)
				 RETURNING id, user_id, balance_cents, currency, created_at, updated_at`,
				userID,
			).StructScan(&w)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	newBalance := w.BalanceCents + amountCents
	if newBalance < 0 {
		return ErrInsufficientBalance
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE wallets
		 SET balance_cents = $1, updated_at = NOW()
		 WHERE id = $2`,
		newBalance, w.ID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO wallet_transactions (wallet_id, amount_cents, type, balance_after)
		 VALUES ($1, $2, $3, $4)`,
		w.ID, amountCents, txType, newBalance,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) TopUp(ctx context.Context, userID int, amountCents int64) error {
	if amountCents <= 0 {
		return errors.New("top up amount must be positive")
	}
	return r.AddTransaction(ctx, userID, amountCents, "topup")
}

func (r *Repository) GetTransactions(ctx context.Context, userID int, limit, offset int) ([]Transaction, error) {
	if limit <= 0 {
		limit = 50
	}

	var walletID int
	err := r.db.GetContext(ctx, &walletID, `SELECT id FROM wallets WHERE user_id = $1`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Transaction{}, nil
		}
		return nil, err
	}

	var txs []Transaction
	err = r.db.SelectContext(ctx, &txs, `
		SELECT id, wallet_id, amount_cents, type, balance_after, created_at
		FROM wallet_transactions
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, walletID, limit, offset)
	if err != nil {
		return nil, err
	}

	return txs, nil
}
