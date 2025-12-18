package gym

import (
	"context"
	"errors"
	"time"
)

var (
	ErrGymNotFound     = errors.New("gym not found")
	ErrTimeSlotInvalid = errors.New("invalid time slot")
)

type Service interface {
	CreateGym(ctx context.Context, req CreateGymRequest) (*Gym, error)
	GetAllGyms(ctx context.Context) ([]Gym, error)
	GetGymByID(ctx context.Context, id int) (*Gym, error)
	CreateTimeSlot(ctx context.Context, gymID int, req CreateTimeSlotRequest) (*TimeSlot, error)
	GetTimeSlots(ctx context.Context, gymID int, onlyFuture bool) ([]TimeSlotWithAvailability, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateGym(ctx context.Context, req CreateGymRequest) (*Gym, error) {
	return s.repo.CreateGym(ctx, req.Name, req.Location)
}

func (s *service) GetAllGyms(ctx context.Context) ([]Gym, error) {
	return s.repo.GetAllGyms(ctx)
}

func (s *service) GetGymByID(ctx context.Context, id int) (*Gym, error) {
	gym, err := s.repo.GetGymByID(ctx, id)
	if err != nil {
		return nil, ErrGymNotFound
	}
	return gym, nil
}

func (s *service) CreateTimeSlot(ctx context.Context, gymID int, req CreateTimeSlotRequest) (*TimeSlot, error) {
	_, err := s.repo.GetGymByID(ctx, gymID)
	if err != nil {
		return nil, ErrGymNotFound
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, ErrTimeSlotInvalid
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, ErrTimeSlotInvalid
	}

	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return nil, ErrTimeSlotInvalid
	}

	if req.Capacity <= 0 {
		return nil, ErrTimeSlotInvalid
	}

	return s.repo.CreateTimeSlot(ctx, gymID, startTime, endTime, req.Capacity)
}

func (s *service) GetTimeSlots(ctx context.Context, gymID int, onlyFuture bool) ([]TimeSlotWithAvailability, error) {
	// Validate gym exists
	_, err := s.repo.GetGymByID(ctx, gymID)
	if err != nil {
		return nil, ErrGymNotFound
	}

	return s.repo.GetTimeSlotsWithAvailability(ctx, gymID, onlyFuture)
}
