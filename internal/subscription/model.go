package subscription

import "time"

type SubscriptionType string
type SubscriptionStatus string

const (
	TypeSingleGymLite SubscriptionType = "single_gym_lite"
	TypeMultiGymFlex  SubscriptionType = "multi_gym_flex"
	TypeUnlimitedPro  SubscriptionType = "unlimited_pro"

	StatusActive   SubscriptionStatus = "active"
	StatusExpired  SubscriptionStatus = "expired"
	StatusCanceled SubscriptionStatus = "cancelled"
)

type Subscription struct {
	ID          int                `db:"id" json:"id"`
	UserID      int                `db:"user_id" json:"user_id"`
	GymID       *int               `db:"gym_id" json:"gym_id,omitempty"`
	Type        SubscriptionType   `db:"type" json:"type"`
	Status      SubscriptionStatus `db:"status" json:"status"`
	VisitsLimit *int               `db:"visits_limit" json:"visits_limit,omitempty"`
	VisitsUsed  int                `db:"visits_used" json:"visits_used"`
	Period      string             `db:"period" json:"period"`
	PriceCents  int64              `db:"price_cents" json:"price_cents"`
	Currency    string             `db:"currency" json:"currency"`
	ValidFrom   time.Time          `db:"valid_from" json:"valid_from"`
	ValidUntil  time.Time          `db:"valid_until" json:"valid_until"`
	CreatedAt   time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `db:"updated_at" json:"updated_at"`
}
