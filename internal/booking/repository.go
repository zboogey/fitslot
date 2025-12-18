package booking

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrBookingNotFoundOrAlreadyCancelled = errors.New("booking not found or already cancelled")

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateBooking(ctx context.Context, userID, timeSlotID int) (*Booking, error) {
	query := `
		INSERT INTO bookings (user_id, time_slot_id, status)
		VALUES ($1, $2, 'booked')
		RETURNING id, user_id, time_slot_id, status, created_at
	`

	var booking Booking
	err := r.db.GetContext(ctx, &booking, query, userID, timeSlotID)
	if err != nil {
		return nil, err
	}

	return &booking, nil
}

func (r *repository) GetBookingByID(ctx context.Context, id int) (*Booking, error) {
	query := `
		SELECT id, user_id, time_slot_id, status, created_at
		FROM bookings
		WHERE id = $1
	`

	var booking Booking
	err := r.db.GetContext(ctx, &booking, query, id)
	if err != nil {
		return nil, err
	}

	return &booking, nil
}

func (r *repository) CancelBooking(ctx context.Context, id int) error {
	query := `
		UPDATE bookings
		SET status = 'cancelled'
		WHERE id = $1 AND status = 'booked'
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrBookingNotFoundOrAlreadyCancelled
	}

	return nil
}

func (r *repository) CountActiveBookingsForSlot(ctx context.Context, timeSlotID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM bookings
		WHERE time_slot_id = $1 AND status = 'booked'
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, timeSlotID)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *repository) UserHasBookingForSlot(ctx context.Context, userID, timeSlotID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM bookings
			WHERE user_id = $1 AND time_slot_id = $2 AND status = 'booked'
		)
	`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, userID, timeSlotID)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *repository) GetUserBookings(ctx context.Context, userID int) ([]Booking, error) {
	query := `
		SELECT id, user_id, time_slot_id, status, created_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var bookings []Booking
	err := r.db.SelectContext(ctx, &bookings, query, userID)
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *repository) GetBookingsByTimeSlot(ctx context.Context, timeSlotID int) ([]BookingWithDetails, error) {
	query := `
		SELECT 
			b.id,
			b.user_id,
			b.time_slot_id,
			b.status,
			b.created_at,
			ts.start_time AS time_slot_start,
			ts.end_time AS time_slot_end,
			g.name AS gym_name,
			g.location AS gym_location,
			u.name AS user_name,
			u.email AS user_email
		FROM bookings b
		JOIN time_slots ts ON b.time_slot_id = ts.id
		JOIN gyms g ON ts.gym_id = g.id
		JOIN users u ON b.user_id = u.id
		WHERE b.time_slot_id = $1
		ORDER BY b.created_at DESC
	`

	var bookings []BookingWithDetails
	err := r.db.SelectContext(ctx, &bookings, query, timeSlotID)
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *repository) GetBookingsByGym(ctx context.Context, gymID int) ([]BookingWithDetails, error) {
	query := `
		SELECT 
			b.id,
			b.user_id,
			b.time_slot_id,
			b.status,
			b.created_at,
			ts.start_time AS time_slot_start,
			ts.end_time AS time_slot_end,
			g.name AS gym_name,
			g.location AS gym_location,
			u.name AS user_name,
			u.email AS user_email
		FROM bookings b
		JOIN time_slots ts ON b.time_slot_id = ts.id
		JOIN gyms g ON ts.gym_id = g.id
		JOIN users u ON b.user_id = u.id
		WHERE g.id = $1
		ORDER BY ts.start_time DESC, b.created_at DESC
	`

	var bookings []BookingWithDetails
	err := r.db.SelectContext(ctx, &bookings, query, gymID)
	if err != nil {
		return nil, err
	}

	return bookings, nil
}
