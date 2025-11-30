-- Add registration_code to events table
ALTER TABLE events
ADD COLUMN registration_code VARCHAR(50) UNIQUE AFTER image_url;

-- Add payment method and phone fields to event_registrations table
ALTER TABLE event_registrations
ADD COLUMN payment_method VARCHAR(50) AFTER payment_reference,
ADD COLUMN payment_phone VARCHAR(20) AFTER payment_method;

-- Add index on registration_code for faster lookups
CREATE INDEX idx_events_registration_code ON events(registration_code);
