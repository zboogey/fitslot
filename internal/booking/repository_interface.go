package booking

import "context"

type Repository interface {
	CreateBooking(ctx context.Context, userID, timeSlotID int) (*Booking, error)
	GetBookingByID(ctx context.Context, id int) (*Booking, error)
	CancelBooking(ctx context.Context, id int) error
	CountActiveBookingsForSlot(ctx context.Context, timeSlotID int) (int, error)
	UserHasBookingForSlot(ctx context.Context, userID, timeSlotID int) (bool, error)
	GetUserBookings(ctx context.Context, userID int) ([]Booking, error)
	GetBookingsByTimeSlot(ctx context.Context, timeSlotID int) ([]BookingWithDetails, error)
	GetBookingsByGym(ctx context.Context, gymID int) ([]BookingWithDetails, error)
}
