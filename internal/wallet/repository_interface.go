package wallet

import "context"

type Repository interface {
	GetOrCreateWallet(ctx context.Context, userID int) (*Wallet, error)
	AddTransaction(ctx context.Context, userID int, amountCents int64, txType string) error
	TopUp(ctx context.Context, userID int, amountCents int64) error
	GetTransactions(ctx context.Context, userID int, limit, offset int) ([]Transaction, error)
}
