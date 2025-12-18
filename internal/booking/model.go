package booking

import (
	"time"

	"fitslot/internal/subscription"
)

type Booking struct {
	ID         int       `db:"id" json:"id"`
	UserID     int       `db:"user_id" json:"user_id"`
	TimeSlotID int       `db:"time_slot_id" json:"time_slot_id"`
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type BookingWithDetails struct {
	Booking
	TimeSlotStart time.Time `db:"time_slot_start" json:"time_slot_start"`
	TimeSlotEnd   time.Time `db:"time_slot_end" json:"time_slot_end"`
	GymName       string    `db:"gym_name" json:"gym_name"`
	GymLocation   string    `db:"gym_location" json:"gym_location"`
	UserName      string    `db:"user_name" json:"user_name"`
	UserEmail     string    `db:"user_email" json:"user_email"`
}

type BookSlotResponse struct {
	Booking      *Booking                    `json:"booking"`
	PaidWith     string                      `json:"paid_with" example:"wallet"`
	AmountCents  *int64                      `json:"amount_cents,omitempty" example:"1000"`
	Subscription *subscription.Subscription  `json:"subscription,omitempty"`
}

type CancelBookingResponse struct {
	Message string `json:"message" example:"Booking cancelled successfully"`
}
