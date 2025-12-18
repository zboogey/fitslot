package user

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrUserNotFound = errors.New("user not found")

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, name, email, passwordHash, role string) (*User, error) {
	query := `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, password_hash, role, created_at
	`

	var user User
	err := r.db.GetContext(ctx, &user, query, name, email, passwordHash, role)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE email = $1
	`

	var user User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *repository) FindByID(ctx context.Context, id int) (*User, error) {
	query := `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *repository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, email)
	if err != nil {
		return false, err
	}

	return exists, nil
}

