package booking

import "time"

type BookingStatsByBucket struct {
	Bucket            string `db:"bucket" json:"bucket"`
	BookingsCreated   int    `db:"bookings_created" json:"bookings_created"`
	BookingsCancelled int    `db:"bookings_cancelled" json:"bookings_cancelled"`
}

type BookingStatsByGym struct {
	GymID             int    `db:"gym_id" json:"gym_id"`
	GymName           string `db:"gym_name" json:"gym_name"`
	BookingsCreated   int    `db:"bookings_created" json:"bookings_created"`
	BookingsCancelled int    `db:"bookings_cancelled" json:"bookings_cancelled"`
}

func (r *Repository) GetBookingStatsByDay(from, to time.Time) ([]BookingStatsByBucket, error) {
	query := `
SELECT
  DATE(created_at) AS bucket,
  COUNT(*) FILTER (WHERE status = 'booked')    AS bookings_created,
  COUNT(*) FILTER (WHERE status = 'cancelled') AS bookings_cancelled
FROM bookings
WHERE created_at BETWEEN $1 AND $2
GROUP BY DATE(created_at)
ORDER BY bucket;
`
	var stats []BookingStatsByBucket
	if err := r.db.Select(&stats, query, from, to); err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *Repository) GetBookingStatsByGym(from, to time.Time) ([]BookingStatsByGym, error) {
	query := `
SELECT
  g.id   AS gym_id,
  g.name AS gym_name,
  COUNT(b.*) FILTER (WHERE b.status = 'booked')    AS bookings_created,
  COUNT(b.*) FILTER (WHERE b.status = 'cancelled') AS bookings_cancelled
FROM gyms g
LEFT JOIN time_slots ts ON ts.id = g.id
LEFT JOIN bookings b ON b.time_slot_id = ts.id
WHERE b.created_at BETWEEN $1 AND $2
GROUP BY g.id, g.name
ORDER BY g.id;
`
	var stats []BookingStatsByGym
	if err := r.db.Select(&stats, query, from, to); err != nil {
		return nil, err
	}
	return stats, nil
}
