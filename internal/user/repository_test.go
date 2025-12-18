package user

import (
    "regexp"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/jmoiron/sqlx"
    "github.com/stretchr/testify/require"
)

func setupUserMock(t *testing.T) (*Repository, sqlmock.Sqlmock, func()) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)

    sqlxDB := sqlx.NewDb(db, "sqlmock")
    repo := NewRepository(sqlxDB)

    closer := func() { sqlxDB.Close() }
    return repo, mock, closer
}

func TestCreateAndFindUser(t *testing.T) {
    repo, mock, close := setupUserMock(t)
    defer close()

    now := time.Now()

    // Create
    mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id, name, email, password_hash, role, created_at")).
        WithArgs("Alice", "a@example.com", "hash", "user").
        WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "role", "created_at"}).AddRow(1, "Alice", "a@example.com", "hash", "user", now))

    u, err := repo.Create("Alice", "a@example.com", "hash", "user")
    require.NoError(t, err)
    require.Equal(t, 1, u.ID)

    // FindByEmail
    mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = $1")).
        WithArgs("a@example.com").
        WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "role", "created_at"}).AddRow(1, "Alice", "a@example.com", "hash", "user", now))

    fu, err := repo.FindByEmail("a@example.com")
    require.NoError(t, err)
    require.Equal(t, "Alice", fu.Name)

    // EmailExists true
    mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)")).
        WithArgs("a@example.com").
        WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

    ok, err := repo.EmailExists("a@example.com")
    require.NoError(t, err)
    require.True(t, ok)
}
