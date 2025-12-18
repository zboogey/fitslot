package gym

import "time"

type Gym struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Location  string    `db:"location" json:"location"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type TimeSlot struct {
	ID        int       `db:"id" json:"id"`
	GymID     int       `db:"gym_id" json:"gym_id"`
	StartTime time.Time `db:"start_time" json:"start_time"`
	EndTime   time.Time `db:"end_time" json:"end_time"`
	Capacity  int       `db:"capacity" json:"capacity"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type TimeSlotWithAvailability struct {
	TimeSlot
	BookedCount int  `json:"booked_count"`
	Available   int  `json:"available"`
	IsFull      bool `json:"is_full"`
}

type CreateGymRequest struct {
	Name     string `json:"name" binding:"required"`
	Location string `json:"location" binding:"required"`
}

type CreateTimeSlotRequest struct {
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	Capacity  int    `json:"capacity" binding:"required,min=1"`
}

