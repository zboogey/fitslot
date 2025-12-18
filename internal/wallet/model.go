package wallet

import "time"

// Wallet — кошелёк пользователя.
type Wallet struct {
	ID           int       `db:"id" json:"id"`
	UserID       int       `db:"user_id" json:"user_id"`
	BalanceCents int64     `db:"balance_cents" json:"balance_cents"`
	Currency     string    `db:"currency" json:"currency"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type Transaction struct {
	ID           int       `db:"id" json:"id"`
	WalletID     int       `db:"wallet_id" json:"wallet_id"`
	AmountCents  int64     `db:"amount_cents" json:"amount_cents"`
	Type         string    `db:"type" json:"type"` // topup, booking_payment, subscription_payment, refund и т.п.
	BalanceAfter int64     `db:"balance_after" json:"balance_after"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
