package user

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(name, email, passwordHash, role string) (*User, error) {
	query := `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, password_hash, role, created_at
	`

	var user User
	err := r.db.Get(&user, query, name, email, passwordHash, role)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) FindByEmail(email string) (*User, error) {
	query := `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE email = $1
	`

	var user User
	err := r.db.Get(&user, query, email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) FindByID(id int) (*User, error) {
	query := `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) EmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.Get(&exists, query, email)
	if err != nil {
		return false, err
	}

	return exists, nil
}

var ErrUserNotFound = errors.New("user not found")

