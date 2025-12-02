-- Fix foreign key constraint issue by making registered_by nullable
-- This allows public event registrations without a logged-in user

-- First, drop the foreign key constraint
ALTER TABLE event_registrations DROP FOREIGN KEY fk_event_registrations_user;

-- Modify the column to allow NULL values
ALTER TABLE event_registrations MODIFY COLUMN registered_by BIGINT NULL;

-- Re-add the foreign key constraint with NULL support
ALTER TABLE event_registrations
ADD CONSTRAINT fk_event_registrations_user
FOREIGN KEY (registered_by)
REFERENCES users(id)
ON DELETE CASCADE;
