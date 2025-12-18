package user

import "context"

type Repository interface {
	Create(ctx context.Context, name, email, passwordHash, role string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id int) (*User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}
