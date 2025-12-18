package booking

import (
    "regexp"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/jmoiron/sqlx"
    "github.com/stretchr/testify/require"
)

func setupMock(t *testing.T) (*Repository, sqlmock.Sqlmock, func()) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)

    sqlxDB := sqlx.NewDb(db, "sqlmock")
    repo := NewRepository(sqlxDB)

    closer := func() {
        sqlxDB.Close()
    }

    return repo, mock, closer
}

func TestCreateAndGetBooking(t *testing.T) {
    repo, mock, close := setupMock(t)
    defer close()

    now := time.Now()

    // Expect INSERT ... RETURNING
    mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO bookings (user_id, time_slot_id, status) VALUES ($1, $2, 'booked') RETURNING id, user_id, time_slot_id, status, created_at")).
        WithArgs(1, 2).
        WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "time_slot_id", "status", "created_at"}).AddRow(10, 1, 2, "booked", now))

    b, err := repo.CreateBooking(1, 2)
    require.NoError(t, err)
    require.Equal(t, 10, b.ID)

    // Expect SELECT by id
    mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, time_slot_id, status, created_at FROM bookings WHERE id = $1")).
        WithArgs(10).
        WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "time_slot_id", "status", "created_at"}).AddRow(10, 1, 2, "booked", now))

    got, err := repo.GetBookingByID(10)
    require.NoError(t, err)
    require.Equal(t, 10, got.ID)
}

func TestCancelBooking(t *testing.T) {
    repo, mock, close := setupMock(t)
    defer close()

    // success case
    mock.ExpectExec(regexp.QuoteMeta("UPDATE bookings SET status = 'cancelled' WHERE id = $1 AND status = 'booked'")).
        WithArgs(5).
        WillReturnResult(sqlmock.NewResult(0, 1))

    err := repo.CancelBooking(5)
    require.NoError(t, err)

    // failure case: zero rows affected
    mock.ExpectExec(regexp.QuoteMeta("UPDATE bookings SET status = 'cancelled' WHERE id = $1 AND status = 'booked'")).
        WithArgs(6).
        WillReturnResult(sqlmock.NewResult(0, 0))

    err = repo.CancelBooking(6)
    require.Error(t, err)
    require.Equal(t, ErrBookingNotFoundOrAlreadyCancelled, err)
}

func TestCountsAndExists(t *testing.T) {
    repo, mock, close := setupMock(t)
    defer close()

    // CountActiveBookingsForSlot
    mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM bookings WHERE time_slot_id = $1 AND status = 'booked'")).
        WithArgs(3).
        WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

    cnt, err := repo.CountActiveBookingsForSlot(3)
    require.NoError(t, err)
    require.Equal(t, 2, cnt)

    // UserHasBookingForSlot true
    mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS( SELECT 1 FROM bookings WHERE user_id = $1 AND time_slot_id = $2 AND status = 'booked' )")).
        WithArgs(1, 3).
        WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

    ok, err := repo.UserHasBookingForSlot(1, 3)
    require.NoError(t, err)
    require.True(t, ok)
}

func TestGetUserBookingsAndByTimeSlot(t *testing.T) {
    repo, mock, close := setupMock(t)
    defer close()

    now := time.Now()

    // GetUserBookings
    rows := sqlmock.NewRows([]string{"id", "user_id", "time_slot_id", "status", "created_at"}).
        AddRow(1, 1, 10, "booked", now).
        AddRow(2, 1, 11, "booked", now.Add(-time.Hour))

    mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, time_slot_id, status, created_at FROM bookings WHERE user_id = $1 ORDER BY created_at DESC")).
        WithArgs(1).
        WillReturnRows(rows)

    list, err := repo.GetUserBookings(1)
    require.NoError(t, err)
    require.Len(t, list, 2)

    // GetBookingsByTimeSlot — return only booking columns (extra fields will be zero-valued)
    rows2 := sqlmock.NewRows([]string{"id", "user_id", "time_slot_id", "status", "created_at"}).
        AddRow(1, 1, 10, "booked", now)

    mock.ExpectQuery(regexp.QuoteMeta("SELECT b.id, b.user_id, b.time_slot_id, b.status, b.created_at, ts.start_time AS time_slot_start, ts.end_time AS time_slot_end, g.name AS gym_name, g.location AS gym_location, u.name AS user_name, u.email AS user_email FROM bookings b JOIN time_slots ts ON b.time_slot_id = ts.id JOIN gyms g ON ts.gym_id = g.id JOIN users u ON b.user_id = u.id WHERE b.time_slot_id = $1 ORDER BY b.created_at DESC")).
        WithArgs(10).
        WillReturnRows(rows2)

    details, err := repo.GetBookingsByTimeSlot(10)
    require.NoError(t, err)
    require.Len(t, details, 1)
    require.Equal(t, 1, details[0].ID)
    require.Equal(t, 10, details[0].TimeSlotID)
}
