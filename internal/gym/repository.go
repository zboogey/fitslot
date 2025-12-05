package gym

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateGym(name, location string) (*Gym, error) {
	query := `
		INSERT INTO gyms (name, location)
		VALUES ($1, $2)
		RETURNING id, name, location, created_at
	`

	var gym Gym
	err := r.db.Get(&gym, query, name, location)
	if err != nil {
		return nil, err
	}

	return &gym, nil
}

func (r *Repository) GetAllGyms() ([]Gym, error) {
	query := `
		SELECT id, name, location, created_at
		FROM gyms
		ORDER BY created_at DESC
	`

	var gyms []Gym
	err := r.db.Select(&gyms, query)
	if err != nil {
		return nil, err
	}

	return gyms, nil
}

func (r *Repository) GetGymByID(id int) (*Gym, error) {
	query := `
		SELECT id, name, location, created_at
		FROM gyms
		WHERE id = $1
	`

	var gym Gym
	err := r.db.Get(&gym, query, id)
	if err != nil {
		return nil, err
	}

	return &gym, nil
}

func (r *Repository) CreateTimeSlot(gymID int, startTime, endTime time.Time, capacity int) (*TimeSlot, error) {
	query := `
		INSERT INTO time_slots (gym_id, start_time, end_time, capacity)
		VALUES ($1, $2, $3, $4)
		RETURNING id, gym_id, start_time, end_time, capacity, created_at
	`

	var slot TimeSlot
	err := r.db.Get(&slot, query, gymID, startTime, endTime, capacity)
	if err != nil {
		return nil, err
	}

	return &slot, nil
}

func (r *Repository) GetTimeSlotsByGym(gymID int, onlyFuture bool) ([]TimeSlot, error) {
	query := `
		SELECT id, gym_id, start_time, end_time, capacity, created_at
		FROM time_slots
		WHERE gym_id = $1
	`
	args := []interface{}{gymID}

	if onlyFuture {
		query += " AND start_time > NOW()"
	}

	query += " ORDER BY start_time ASC"

	var slots []TimeSlot
	err := r.db.Select(&slots, query, args...)
	if err != nil {
		return nil, err
	}

	return slots, nil
}

func (r *Repository) GetTimeSlotByID(id int) (*TimeSlot, error) {
	query := `
		SELECT id, gym_id, start_time, end_time, capacity, created_at
		FROM time_slots
		WHERE id = $1
	`

	var slot TimeSlot
	err := r.db.Get(&slot, query, id)
	if err != nil {
		return nil, err
	}

	return &slot, nil
}

func (r *Repository) GetTimeSlotsWithAvailability(gymID int, onlyFuture bool) ([]TimeSlotWithAvailability, error) {

	slots, err := r.GetTimeSlotsByGym(gymID, onlyFuture)
	if err != nil {
		return nil, err
	}

	result := make([]TimeSlotWithAvailability, 0, len(slots))
	for _, slot := range slots {
		var bookedCount int
		countQuery := `
			SELECT COUNT(*)
			FROM bookings
			WHERE time_slot_id = $1 AND status = 'booked'
		`
		err := r.db.Get(&bookedCount, countQuery, slot.ID)
		if err != nil {
			return nil, err
		}

		available := slot.Capacity - bookedCount
		isFull := available <= 0

		result = append(result, TimeSlotWithAvailability{
			TimeSlot:    slot,
			BookedCount: bookedCount,
			Available:   available,
			IsFull:      isFull,
		})
	}

	return result, nil
}


