package gym

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func setupGymMock(t *testing.T) (*Repository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	closer := func() { sqlxDB.Close() }
	return repo, mock, closer
}

func TestCreateGym(t *testing.T) {
	repo, mock, close := setupGymMock(t)
	defer close()

	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO gyms (name, location) VALUES ($1, $2) RETURNING id, name, location, created_at")).
		WithArgs("Fitness Club", "Downtown").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "location", "created_at"}).AddRow(1, "Fitness Club", "Downtown", now))

	gym, err := repo.CreateGym("Fitness Club", "Downtown")
	require.NoError(t, err)
	require.Equal(t, 1, gym.ID)
	require.Equal(t, "Fitness Club", gym.Name)
}

func TestGetAllGyms(t *testing.T) {
	repo, mock, close := setupGymMock(t)
	defer close()

	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, location, created_at FROM gyms ORDER BY created_at DESC")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "location", "created_at"}).
			AddRow(1, "Gym A", "Loc A", now).
			AddRow(2, "Gym B", "Loc B", now))

	gyms, err := repo.GetAllGyms()
	require.NoError(t, err)
	require.Len(t, gyms, 2)
	require.Equal(t, "Gym A", gyms[0].Name)
}

func TestGetGymByID(t *testing.T) {
	repo, mock, close := setupGymMock(t)
	defer close()

	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, location, created_at FROM gyms WHERE id = $1")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "location", "created_at"}).AddRow(1, "Gym A", "Loc A", now))

	gym, err := repo.GetGymByID(1)
	require.NoError(t, err)
	require.Equal(t, "Gym A", gym.Name)
}

func TestCreateTimeSlot(t *testing.T) {
	repo, mock, close := setupGymMock(t)
	defer close()

	start := time.Now().Add(24 * time.Hour)
	end := start.Add(time.Hour)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO time_slots (gym_id, start_time, end_time, capacity) VALUES ($1, $2, $3, $4) RETURNING id, gym_id, start_time, end_time, capacity, created_at")).
		WithArgs(1, start, end, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "gym_id", "start_time", "end_time", "capacity", "created_at"}).AddRow(5, 1, start, end, 10, now))

	slot, err := repo.CreateTimeSlot(1, start, end, 10)
	require.NoError(t, err)
	require.Equal(t, 5, slot.ID)
	require.Equal(t, 10, slot.Capacity)
}

func TestGetTimeSlotByID(t *testing.T) {
	repo, mock, close := setupGymMock(t)
	defer close()

	start := time.Now()
	end := start.Add(time.Hour)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, gym_id, start_time, end_time, capacity, created_at FROM time_slots WHERE id = $1")).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "gym_id", "start_time", "end_time", "capacity", "created_at"}).AddRow(5, 1, start, end, 10, now))

	slot, err := repo.GetTimeSlotByID(5)
	require.NoError(t, err)
	require.Equal(t, 5, slot.ID)
}

func TestGetTimeSlotsByGym(t *testing.T) {
	repo, mock, close := setupGymMock(t)
	defer close()

	start := time.Now()
	end := start.Add(time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, gym_id, start_time, end_time, capacity, created_at FROM time_slots WHERE gym_id = $1 ORDER BY start_time ASC")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "gym_id", "start_time", "end_time", "capacity", "created_at"}).
			AddRow(5, 1, start, end, 10, time.Now()))

	slots, err := repo.GetTimeSlotsByGym(1, false)
	require.NoError(t, err)
	require.Len(t, slots, 1)
}
