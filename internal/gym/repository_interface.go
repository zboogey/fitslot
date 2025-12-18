package gym

import (
	"context"
	"time"
)

type Repository interface {
	CreateGym(ctx context.Context, name, location string) (*Gym, error)
	GetAllGyms(ctx context.Context) ([]Gym, error)
	GetGymByID(ctx context.Context, id int) (*Gym, error)
	CreateTimeSlot(ctx context.Context, gymID int, startTime, endTime time.Time, capacity int) (*TimeSlot, error)
	GetTimeSlotsByGym(ctx context.Context, gymID int, onlyFuture bool) ([]TimeSlot, error)
	GetTimeSlotByID(ctx context.Context, id int) (*TimeSlot, error)
	GetTimeSlotsWithAvailability(ctx context.Context, gymID int, onlyFuture bool) ([]TimeSlotWithAvailability, error)
}
