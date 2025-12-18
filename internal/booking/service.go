package booking

import (
	"context"
	"errors"
	"time"

	"fitslot/internal/email"
	"fitslot/internal/gym"
	"fitslot/internal/subscription"
	"fitslot/internal/user"
	"fitslot/internal/wallet"
)

var (
	ErrBookingNotFound   = errors.New("booking not found")
	ErrInsufficientFunds = errors.New("insufficient wallet balance")
)

type Service interface {
	BookSlot(ctx context.Context, userID, slotID int) (*Booking, string, interface{}, error)
	CancelBooking(ctx context.Context, userID, bookingID int) error
	GetUserBookings(ctx context.Context, userID int) ([]Booking, error)
	GetBookingsByTimeSlot(ctx context.Context, slotID int) ([]BookingWithDetails, error)
	GetBookingsByGym(ctx context.Context, gymID int) ([]BookingWithDetails, error)
}

type service struct {
	bookingRepo      Repository
	gymRepo          gym.Repository
	subscriptionRepo subscription.Repository
	walletRepo       wallet.Repository
	userRepo         user.Repository
	emailService     *email.Service
}

func NewService(
	bookingRepo Repository,
	gymRepo gym.Repository,
	subscriptionRepo subscription.Repository,
	walletRepo wallet.Repository,
	userRepo user.Repository,
	emailService *email.Service,
) Service {
	return &service{
		bookingRepo:      bookingRepo,
		gymRepo:          gymRepo,
		subscriptionRepo: subscriptionRepo,
		walletRepo:       walletRepo,
		userRepo:         userRepo,
		emailService:     emailService,
	}
}

func (s *service) BookSlot(ctx context.Context, userID, slotID int) (*Booking, string, interface{}, error) {
	slot, err := s.gymRepo.GetTimeSlotByID(ctx, slotID)
	if err != nil {
		return nil, "", nil, errors.New("time slot not found")
	}

	if slot.StartTime.Before(time.Now()) {
		return nil, "", nil, errors.New("cannot book a slot in the past")
	}

	bookedCount, err := s.bookingRepo.CountActiveBookingsForSlot(ctx, slotID)
	if err != nil {
		return nil, "", nil, err
	}

	if bookedCount >= slot.Capacity {
		return nil, "", nil, errors.New("time slot is full")
	}

	hasBooking, err := s.bookingRepo.UserHasBookingForSlot(ctx, userID, slotID)
	if err != nil {
		return nil, "", nil, err
	}

	if hasBooking {
		return nil, "", nil, errors.New("user already has a booking for this slot")
	}

	// Check for active subscription
	useSubscription := false
	var activeSub *subscription.Subscription

	sub, err := s.subscriptionRepo.GetActiveForUserAndGym(ctx, userID, slot.GymID)
	if err == nil && sub.Status == subscription.StatusActive {
		if sub.VisitsLimit == nil {
			useSubscription = true
			activeSub = sub
		} else if sub.VisitsUsed < *sub.VisitsLimit {
			useSubscription = true
			activeSub = sub
		}
	}

	// Create booking
	booking, err := s.bookingRepo.CreateBooking(ctx, userID, slotID)
	if err != nil {
		return nil, "", nil, err
	}

	// Handle payment
	if useSubscription && activeSub != nil {
		if err := s.subscriptionRepo.IncrementVisits(ctx, activeSub.ID); err != nil {
			// Booking already created, return warning
			return booking, "subscription", activeSub, err
		}

		// Send confirmation email
		user, _ := s.userRepo.FindByID(ctx, userID)
		if user != nil {
			s.emailService.SendBookingConfirmation(
				ctx,
				user.Email,
				user.Name,
				"Gym Slot",
				slot.StartTime.Format("Jan 2, 2006 at 3:04 PM"),
				slot.StartTime,
			)
		}

		return booking, "subscription", activeSub, nil
	}

	// Pay with wallet
	const priceCents int64 = 1000
	if err := s.walletRepo.AddTransaction(ctx, userID, -priceCents, "booking_payment"); err != nil {
		if err.Error() == "insufficient balance" {
			return nil, "", nil, ErrInsufficientFunds
		}
		return nil, "", nil, err
	}

	// Send confirmation email
	user, _ := s.userRepo.FindByID(ctx, userID)
	if user != nil {
		s.emailService.SendBookingConfirmation(
			ctx,
			user.Email,
			user.Name,
			"Gym Slot",
			slot.StartTime.Format("Jan 2, 2006 at 3:04 PM"),
			slot.StartTime,
		)
	}

	return booking, "wallet", map[string]interface{}{"amount_cents": priceCents}, nil
}

func (s *service) CancelBooking(ctx context.Context, userID, bookingID int) error {
	booking, err := s.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	if booking.UserID != userID {
		return errors.New("unauthorized: can only cancel own bookings")
	}

	err = s.bookingRepo.CancelBooking(ctx, bookingID)
	if err != nil {
		if err == ErrBookingNotFoundOrAlreadyCancelled {
			return ErrBookingNotFound
		}
		return err
	}

	return nil
}

func (s *service) GetUserBookings(ctx context.Context, userID int) ([]Booking, error) {
	return s.bookingRepo.GetUserBookings(ctx, userID)
}

func (s *service) GetBookingsByTimeSlot(ctx context.Context, slotID int) ([]BookingWithDetails, error) {
	return s.bookingRepo.GetBookingsByTimeSlot(ctx, slotID)
}

func (s *service) GetBookingsByGym(ctx context.Context, gymID int) ([]BookingWithDetails, error) {
	return s.bookingRepo.GetBookingsByGym(ctx, gymID)
}
