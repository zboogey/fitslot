-- Seed data for development/testing
-- This file should be run manually after migrations

-- Insert admin user (password: admin123)
INSERT INTO users (name, email, password_hash, role) VALUES
('Admin User', 'admin@fitslot.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin')
ON CONFLICT (email) DO NOTHING;

-- Insert test member user (password: member123)
INSERT INTO users (name, email, password_hash, role) VALUES
('Test Member', 'member@fitslot.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'member')
ON CONFLICT (email) DO NOTHING;

-- Insert sample gyms
INSERT INTO gyms (name, location) VALUES
('FitZone Downtown', '123 Main St, Downtown'),
('PowerGym Central', '456 Oak Ave, Central District'),
('Elite Fitness North', '789 Pine Rd, North Side')
ON CONFLICT DO NOTHING;

-- Insert sample time slots for gym 1 (next 7 days)
INSERT INTO time_slots (gym_id, start_time, end_time, capacity)
SELECT 
    1,
    (CURRENT_DATE + INTERVAL '1 day' + (n || ' hours')::interval)::timestamp,
    (CURRENT_DATE + INTERVAL '1 day' + ((n + 1) || ' hours')::interval)::timestamp,
    20
FROM generate_series(8, 20) n
WHERE NOT EXISTS (
    SELECT 1 FROM time_slots 
    WHERE gym_id = 1 
    AND start_time = (CURRENT_DATE + INTERVAL '1 day' + (n || ' hours')::interval)::timestamp
)
LIMIT 13;

-- Insert sample time slots for gym 2 (next 7 days)
INSERT INTO time_slots (gym_id, start_time, end_time, capacity)
SELECT 
    2,
    (CURRENT_DATE + INTERVAL '1 day' + (n || ' hours')::interval)::timestamp,
    (CURRENT_DATE + INTERVAL '1 day' + ((n + 1) || ' hours')::interval)::timestamp,
    15
FROM generate_series(9, 21) n
WHERE NOT EXISTS (
    SELECT 1 FROM time_slots 
    WHERE gym_id = 2 
    AND start_time = (CURRENT_DATE + INTERVAL '1 day' + (n || ' hours')::interval)::timestamp
)
LIMIT 12;

-- Add initial wallet balance for test member
INSERT INTO wallets (user_id, balance_cents, currency)
SELECT id, 50000, 'KZT'
FROM users WHERE email = 'member@fitslot.com'
ON CONFLICT (user_id) DO NOTHING;


