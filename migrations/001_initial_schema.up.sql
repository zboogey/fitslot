-- +goose Up

CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

CREATE TABLE IF NOT EXISTS gyms (
                                    id SERIAL PRIMARY KEY,
                                    name VARCHAR(255) NOT NULL,
    location VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS time_slots (
                                          id SERIAL PRIMARY KEY,
                                          gym_id INTEGER NOT NULL REFERENCES gyms(id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    capacity INTEGER NOT NULL DEFAULT 10,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_time_order CHECK (end_time > start_time),
    CONSTRAINT check_capacity_positive CHECK (capacity > 0)
    );
CREATE INDEX IF NOT EXISTS idx_time_slots_gym_id ON time_slots(gym_id);
CREATE INDEX IF NOT EXISTS idx_time_slots_start_time ON time_slots(start_time);

CREATE TABLE IF NOT EXISTS bookings (
                                        id SERIAL PRIMARY KEY,
                                        user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    time_slot_id INTEGER NOT NULL REFERENCES time_slots(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'booked',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_status_valid CHECK (status IN ('booked', 'cancelled'))
    );
CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_time_slot_id ON bookings(time_slot_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);

ALTER TABLE bookings ADD CONSTRAINT unique_user_slot UNIQUE (user_id, time_slot_id);
