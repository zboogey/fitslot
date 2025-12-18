package gym

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateGym(ctx context.Context, name, location string) (*Gym, error) {
	args := m.Called(ctx, name, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Gym), args.Error(1)
}

func (m *MockRepository) GetAllGyms(ctx context.Context) ([]Gym, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Gym), args.Error(1)
}

func (m *MockRepository) GetGymByID(ctx context.Context, id int) (*Gym, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Gym), args.Error(1)
}

func (m *MockRepository) CreateTimeSlot(ctx context.Context, gymID int, startTime, endTime time.Time, capacity int) (*TimeSlot, error) {
	args := m.Called(ctx, gymID, startTime, endTime, capacity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TimeSlot), args.Error(1)
}

func (m *MockRepository) GetTimeSlotsByGym(ctx context.Context, gymID int, onlyFuture bool) ([]TimeSlot, error) {
	args := m.Called(ctx, gymID, onlyFuture)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]TimeSlot), args.Error(1)
}

func (m *MockRepository) GetTimeSlotByID(ctx context.Context, id int) (*TimeSlot, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TimeSlot), args.Error(1)
}

func (m *MockRepository) GetTimeSlotsWithAvailability(ctx context.Context, gymID int, onlyFuture bool) ([]TimeSlotWithAvailability, error) {
	args := m.Called(ctx, gymID, onlyFuture)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]TimeSlotWithAvailability), args.Error(1)
}

func TestService_CreateGym(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	req := CreateGymRequest{
		Name:     "Test Gym",
		Location: "Test Location",
	}

	mockRepo.On("CreateGym", mock.Anything, "Test Gym", "Test Location").Return(&Gym{
		ID:       1,
		Name:     "Test Gym",
		Location: "Test Location",
	}, nil)

	gym, err := service.CreateGym(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, gym)
	assert.Equal(t, "Test Gym", gym.Name)
	mockRepo.AssertExpectations(t)
}

func TestService_CreateTimeSlot(t *testing.T) {
	tests := []struct {
		name        string
		gymID       int
		req         CreateTimeSlotRequest
		setupMock   func(*MockRepository)
		expectError bool
	}{
		{
			name:  "successful creation",
			gymID: 1,
			req: CreateTimeSlotRequest{
				StartTime: "2024-12-20T10:00:00Z",
				EndTime:   "2024-12-20T11:00:00Z",
				Capacity:  20,
			},
			setupMock: func(m *MockRepository) {
				m.On("GetGymByID", mock.Anything, 1).Return(&Gym{ID: 1}, nil)
				start, _ := time.Parse(time.RFC3339, "2024-12-20T10:00:00Z")
				end, _ := time.Parse(time.RFC3339, "2024-12-20T11:00:00Z")
				m.On("CreateTimeSlot", mock.Anything, 1, start, end, 20).Return(&TimeSlot{
					ID:        1,
					GymID:     1,
					StartTime: start,
					EndTime:   end,
					Capacity:  20,
				}, nil)
			},
			expectError: false,
		},
		{
			name:  "gym not found",
			gymID: 999,
			req: CreateTimeSlotRequest{
				StartTime: "2024-12-20T10:00:00Z",
				EndTime:   "2024-12-20T11:00:00Z",
				Capacity:  20,
			},
			setupMock: func(m *MockRepository) {
				m.On("GetGymByID", mock.Anything, 999).Return(nil, errors.New("not found"))
			},
			expectError: true,
		},
		{
			name:  "invalid time format",
			gymID: 1,
			req: CreateTimeSlotRequest{
				StartTime: "invalid",
				EndTime:   "2024-12-20T11:00:00Z",
				Capacity:  20,
			},
			setupMock: func(m *MockRepository) {
				m.On("GetGymByID", mock.Anything, 1).Return(&Gym{ID: 1}, nil)
			},
			expectError: true,
		},
		{
			name:  "end time before start time",
			gymID: 1,
			req: CreateTimeSlotRequest{
				StartTime: "2024-12-20T11:00:00Z",
				EndTime:   "2024-12-20T10:00:00Z",
				Capacity:  20,
			},
			setupMock: func(m *MockRepository) {
				m.On("GetGymByID", mock.Anything, 1).Return(&Gym{ID: 1}, nil)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			service := NewService(mockRepo)
			slot, err := service.CreateTimeSlot(context.Background(), tt.gymID, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, slot)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, slot)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetTimeSlots(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	mockRepo.On("GetGymByID", mock.Anything, 1).Return(&Gym{ID: 1}, nil)
	mockRepo.On("GetTimeSlotsWithAvailability", mock.Anything, 1, true).Return([]TimeSlotWithAvailability{
		{
			TimeSlot: TimeSlot{
				ID:        1,
				GymID:     1,
				StartTime: time.Now().Add(24 * time.Hour),
				EndTime:   time.Now().Add(25 * time.Hour),
				Capacity:  20,
			},
			BookedCount: 5,
			Available:  15,
			IsFull:     false,
		},
	}, nil)

	slots, err := service.GetTimeSlots(context.Background(), 1, true)

	assert.NoError(t, err)
	assert.Len(t, slots, 1)
	mockRepo.AssertExpectations(t)
}


