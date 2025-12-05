package booking

import "time"

type Booking struct {
	ID         int       `db:"id" json:"id"`
	UserID     int       `db:"user_id" json:"user_id"`
	TimeSlotID int       `db:"time_slot_id" json:"time_slot_id"`
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type BookingWithDetails struct {
	Booking
	TimeSlotStart time.Time `json:"time_slot_start"`
	TimeSlotEnd   time.Time `json:"time_slot_end"`
	GymName       string    `json:"gym_name"`
	GymLocation   string    `json:"gym_location"`
	UserName      string    `json:"user_name"`
	UserEmail     string    `json:"user_email"`
}

