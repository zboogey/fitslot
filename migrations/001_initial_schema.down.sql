-- +goose Down

ALTER TABLE bookings DROP CONSTRAINT IF EXISTS unique_user_slot;

DROP INDEX IF EXISTS idx_bookings_status;
DROP INDEX IF EXISTS idx_bookings_time_slot_id;
DROP INDEX IF EXISTS idx_bookings_user_id;
DROP TABLE IF EXISTS bookings;

DROP INDEX IF EXISTS idx_time_slots_start_time;
DROP INDEX IF EXISTS idx_time_slots_gym_id;
DROP TABLE IF EXISTS time_slots;

DROP TABLE IF EXISTS gyms;

DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;