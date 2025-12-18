package subscription

import "context"

type Repository interface {
	CreateSubscription(ctx context.Context, userID int, gymID *int, stype SubscriptionType, priceCents int64, visitsLimit *int) (*Subscription, error)
	GetActiveForUserAndGym(ctx context.Context, userID int, gymID int) (*Subscription, error)
	IncrementVisits(ctx context.Context, subID int) error
	ListActiveByUser(ctx context.Context, userID int) ([]*Subscription, error)
}
