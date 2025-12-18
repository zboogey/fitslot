package subscription

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateSubscription(
	ctx context.Context,
	userID int,
	gymID *int,
	stype SubscriptionType,
	priceCents int64,
	visitsLimit *int,
) (*Subscription, error) {
	now := time.Now()
	validUntil := now.AddDate(0, 1, 0)

	sub := &Subscription{}
	err := r.db.QueryRowxContext(ctx, `
		INSERT INTO subscriptions (user_id, gym_id, type, status, visits_limit, visits_used, period, price_cents, currency, valid_from, valid_until)
		VALUES ($1, $2, $3, 'active', $4, 0, 'monthly', $5, 'KZT', $6, $7)
		RETURNING id, user_id, gym_id, type, status, visits_limit, visits_used, period, price_cents, currency, valid_from, valid_until, created_at, updated_at
	`, userID, gymID, stype, visitsLimit, priceCents, now, validUntil).StructScan(sub)

	return sub, err
}

func (r *Repository) GetActiveForUserAndGym(ctx context.Context, userID int, gymID int) (*Subscription, error) {
	sub := &Subscription{}
	err := r.db.GetContext(ctx, sub, `
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
	`, userID, gymID)

	return sub, err
}

func (r *Repository) IncrementVisits(ctx context.Context, subID int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET visits_used = visits_used + 1,
		    updated_at = NOW()
		WHERE id = $1
	`, subID)
	return err
}

func (r *Repository) ListActiveByUser(ctx context.Context, userID int) ([]*Subscription, error) {
	subs := []*Subscription{}
	err := r.db.SelectContext(ctx, &subs, `
		SELECT *
		FROM subscriptions
		WHERE user_id = $1
		  AND status = 'active'
		  AND valid_from <= NOW()
		  AND valid_until >= NOW()
		ORDER BY created_at DESC
	`, userID)
	return subs, err
}
