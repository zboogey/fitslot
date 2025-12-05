package gym

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestCreateGym(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(dbx)

	mock.ExpectQuery(`INSERT INTO gyms.*`).
		WithArgs("Gym A", "City X").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "location", "created_at"}).
			AddRow(1, "Gym A", "City X", time.Now()))

	gym, err := repo.CreateGym("Gym A", "City X")
	assert.NoError(t, err)
	assert.Equal(t, 1, gym.ID)
	assert.Equal(t, "Gym A", gym.Name)
	assert.Equal(t, "City X", gym.Location)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllGyms(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(dbx)

	mock.ExpectQuery(`SELECT id, name, location, created_at FROM gyms.*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "location", "created_at"}).
			AddRow(1, "Gym A", "City X", time.Now()).
			AddRow(2, "Gym B", "City Y", time.Now()))

	gyms, err := repo.GetAllGyms()
	assert.NoError(t, err)
	assert.Len(t, gyms, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetGymByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(dbx)

	mock.ExpectQuery(`SELECT id, name, location, created_at FROM gyms WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "location", "created_at"}).
			AddRow(1, "Gym A", "City X", time.Now()))

	gym, err := repo.GetGymByID(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, gym.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTimeSlot(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(dbx)

	start := time.Now()
	end := start.Add(time.Hour)

	mock.ExpectQuery(`INSERT INTO time_slots.*`).
		WithArgs(1, start, end, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "gym_id", "start_time", "end_time", "capacity", "created_at"}).
			AddRow(1, 1, start, end, 10, time.Now()))

	slot, err := repo.CreateTimeSlot(1, start, end, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, slot.ID)
	assert.Equal(t, 10, slot.Capacity)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTimeSlotsWithAvailability(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	dbx := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(dbx)

	start := time.Now().Add(time.Hour)
	end := start.Add(time.Hour)

	// Сначала мок для GetTimeSlotsByGym
	mock.ExpectQuery(`SELECT id, gym_id, start_time, end_time, capacity, created_at FROM time_slots.*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "gym_id", "start_time", "end_time", "capacity", "created_at"}).
			AddRow(1, 1, start, end, 10, time.Now()))

	// Мок для подсчёта бронирований
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM bookings.*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	slots, err := repo.GetTimeSlotsWithAvailability(1, true)
	assert.NoError(t, err)
	assert.Len(t, slots, 1)
	assert.Equal(t, 7, slots[0].Available)
	assert.Equal(t, false, slots[0].IsFull)
	assert.NoError(t, mock.ExpectationsWereMet())
}
