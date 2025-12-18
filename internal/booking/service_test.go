package booking

import (
	"context"
	"errors"
	"testing"
	"time"

	"fitslot/internal/email"
	"fitslot/internal/gym"
	"fitslot/internal/subscription"
	"fitslot/internal/user"
	"fitslot/internal/wallet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockBookingRepo struct{ mock.Mock }
type MockGymRepo struct{ mock.Mock }
type MockSubscriptionRepo struct{ mock.Mock }
type MockWalletRepo struct{ mock.Mock }
type MockUserRepo struct{ mock.Mock }

func (m *MockBookingRepo) CreateBooking(ctx context.Context, userID, timeSlotID int) (*Booking, error) {
	args := m.Called(ctx, userID, timeSlotID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Booking), args.Error(1)
}

func (m *MockBookingRepo) GetBookingByID(ctx context.Context, id int) (*Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Booking), args.Error(1)
}

func (m *MockBookingRepo) CancelBooking(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockBookingRepo) CountActiveBookingsForSlot(ctx context.Context, timeSlotID int) (int, error) {
	args := m.Called(ctx, timeSlotID)
	return args.Int(0), args.Error(1)
}

func (m *MockBookingRepo) UserHasBookingForSlot(ctx context.Context, userID, timeSlotID int) (bool, error) {
	args := m.Called(ctx, userID, timeSlotID)
	return args.Bool(0), args.Error(1)
}

func (m *MockBookingRepo) GetUserBookings(ctx context.Context, userID int) ([]Booking, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Booking), args.Error(1)
}

func (m *MockBookingRepo) GetBookingsByTimeSlot(ctx context.Context, timeSlotID int) ([]BookingWithDetails, error) {
	args := m.Called(ctx, timeSlotID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]BookingWithDetails), args.Error(1)
}

func (m *MockBookingRepo) GetBookingsByGym(ctx context.Context, gymID int) ([]BookingWithDetails, error) {
	args := m.Called(ctx, gymID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]BookingWithDetails), args.Error(1)
}

func (m *MockGymRepo) CreateGym(ctx context.Context, name, location string) (*gym.Gym, error) {
	args := m.Called(ctx, name, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gym.Gym), args.Error(1)
}

func (m *MockGymRepo) GetAllGyms(ctx context.Context) ([]gym.Gym, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]gym.Gym), args.Error(1)
}

func (m *MockGymRepo) GetGymByID(ctx context.Context, id int) (*gym.Gym, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gym.Gym), args.Error(1)
}

func (m *MockGymRepo) CreateTimeSlot(ctx context.Context, gymID int, startTime, endTime time.Time, capacity int) (*gym.TimeSlot, error) {
	args := m.Called(ctx, gymID, startTime, endTime, capacity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gym.TimeSlot), args.Error(1)
}

func (m *MockGymRepo) GetTimeSlotsByGym(ctx context.Context, gymID int, onlyFuture bool) ([]gym.TimeSlot, error) {
	args := m.Called(ctx, gymID, onlyFuture)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]gym.TimeSlot), args.Error(1)
}

func (m *MockGymRepo) GetTimeSlotByID(ctx context.Context, id int) (*gym.TimeSlot, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gym.TimeSlot), args.Error(1)
}

func (m *MockGymRepo) GetTimeSlotsWithAvailability(ctx context.Context, gymID int, onlyFuture bool) ([]gym.TimeSlotWithAvailability, error) {
	args := m.Called(ctx, gymID, onlyFuture)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]gym.TimeSlotWithAvailability), args.Error(1)
}

func (m *MockSubscriptionRepo) CreateSubscription(ctx context.Context, userID int, gymID *int, stype subscription.SubscriptionType, priceCents int64, visitsLimit *int) (*subscription.Subscription, error) {
	args := m.Called(ctx, userID, gymID, stype, priceCents, visitsLimit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepo) GetActiveForUserAndGym(ctx context.Context, userID int, gymID int) (*subscription.Subscription, error) {
	args := m.Called(ctx, userID, gymID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepo) IncrementVisits(ctx context.Context, subID int) error {
	return m.Called(ctx, subID).Error(0)
}

func (m *MockSubscriptionRepo) ListActiveByUser(ctx context.Context, userID int) ([]*subscription.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*subscription.Subscription), args.Error(1)
}

func (m *MockWalletRepo) GetOrCreateWallet(ctx context.Context, userID int) (*wallet.Wallet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet.Wallet), args.Error(1)
}

func (m *MockWalletRepo) AddTransaction(ctx context.Context, userID int, amountCents int64, txType string) error {
	return m.Called(ctx, userID, amountCents, txType).Error(0)
}

func (m *MockWalletRepo) TopUp(ctx context.Context, userID int, amountCents int64) error {
	return m.Called(ctx, userID, amountCents).Error(0)
}

func (m *MockWalletRepo) GetTransactions(ctx context.Context, userID int, limit, offset int) ([]wallet.Transaction, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]wallet.Transaction), args.Error(1)
}

func (m *MockUserRepo) Create(ctx context.Context, name, email, passwordHash, role string) (*user.User, error) {
	args := m.Called(ctx, name, email, passwordHash, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepo) FindByID(ctx context.Context, id int) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func TestService_BookSlot(t *testing.T) {
	futureTime := time.Now().Add(24 * time.Hour)
	pastTime := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name        string
		userID      int
		slotID      int
		setupMocks  func(*MockBookingRepo, *MockGymRepo, *MockSubscriptionRepo, *MockWalletRepo, *MockUserRepo)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful booking with wallet",
			userID: 1,
			slotID: 1,
			setupMocks: func(br *MockBookingRepo, gr *MockGymRepo, sr *MockSubscriptionRepo, wr *MockWalletRepo, ur *MockUserRepo) {
				gr.On("GetTimeSlotByID", mock.Anything, 1).Return(&gym.TimeSlot{
					ID:        1,
					GymID:     1,
					StartTime: futureTime,
					EndTime:   futureTime.Add(time.Hour),
					Capacity:  20,
				}, nil)
				br.On("CountActiveBookingsForSlot", mock.Anything, 1).Return(5, nil)
				br.On("UserHasBookingForSlot", mock.Anything, 1, 1).Return(false, nil)
				sr.On("GetActiveForUserAndGym", mock.Anything, 1, 1).Return(nil, errors.New("no subscription"))
				br.On("CreateBooking", mock.Anything, 1, 1).Return(&Booking{
					ID:         1,
					UserID:     1,
					TimeSlotID: 1,
					Status:     "booked",
				}, nil)
				wr.On("AddTransaction", mock.Anything, 1, int64(-1000), "booking_payment").Return(nil)
				ur.On("FindByID", mock.Anything, 1).Return(&user.User{
					ID:    1,
					Email: "test@example.com",
					Name:  "Test User",
				}, nil)
			},
			expectError: false,
		},
		{
			name:   "slot not found",
			userID: 1,
			slotID: 999,
			setupMocks: func(br *MockBookingRepo, gr *MockGymRepo, sr *MockSubscriptionRepo, wr *MockWalletRepo, ur *MockUserRepo) {
				gr.On("GetTimeSlotByID", mock.Anything, 999).Return(nil, errors.New("not found"))
			},
			expectError: true,
			errorMsg:    "time slot not found",
		},
		{
			name:   "slot in past",
			userID: 1,
			slotID: 1,
			setupMocks: func(br *MockBookingRepo, gr *MockGymRepo, sr *MockSubscriptionRepo, wr *MockWalletRepo, ur *MockUserRepo) {
				gr.On("GetTimeSlotByID", mock.Anything, 1).Return(&gym.TimeSlot{
					ID:        1,
					StartTime: pastTime,
					EndTime:   pastTime.Add(time.Hour),
					Capacity:  20,
				}, nil)
			},
			expectError: true,
			errorMsg:    "cannot book a slot in the past",
		},
		{
			name:   "slot full",
			userID: 1,
			slotID: 1,
			setupMocks: func(br *MockBookingRepo, gr *MockGymRepo, sr *MockSubscriptionRepo, wr *MockWalletRepo, ur *MockUserRepo) {
				gr.On("GetTimeSlotByID", mock.Anything, 1).Return(&gym.TimeSlot{
					ID:        1,
					StartTime: futureTime,
					EndTime:   futureTime.Add(time.Hour),
					Capacity:  20,
				}, nil)
				br.On("CountActiveBookingsForSlot", mock.Anything, 1).Return(20, nil)
			},
			expectError: true,
			errorMsg:    "time slot is full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := new(MockBookingRepo)
			gr := new(MockGymRepo)
			sr := new(MockSubscriptionRepo)
			wr := new(MockWalletRepo)
			ur := new(MockUserRepo)

			tt.setupMocks(br, gr, sr, wr, ur)

			emailService := email.New("from@test.com", "Test", "localhost", "1025", "", "", "localhost:6379")
			service := NewService(br, gr, sr, wr, ur, emailService)

			booking, paymentMethod, _, err := service.BookSlot(context.Background(), tt.userID, tt.slotID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, booking)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, booking)
				assert.NotEmpty(t, paymentMethod)
			}
		})
	}
}

func TestService_CancelBooking(t *testing.T) {
	br := new(MockBookingRepo)
	gr := new(MockGymRepo)
	sr := new(MockSubscriptionRepo)
	wr := new(MockWalletRepo)
	ur := new(MockUserRepo)

	br.On("GetBookingByID", mock.Anything, 1).Return(&Booking{
		ID:     1,
		UserID: 1,
		Status: "booked",
	}, nil)
	br.On("CancelBooking", mock.Anything, 1).Return(nil)

	emailService := email.New("from@test.com", "Test", "localhost", "1025", "", "", "localhost:6379")
	service := NewService(br, gr, sr, wr, ur, emailService)

	err := service.CancelBooking(context.Background(), 1, 1)

	assert.NoError(t, err)
	br.AssertExpectations(t)
}
