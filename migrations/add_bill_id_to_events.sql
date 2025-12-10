-- Migration: Add bill_id to events table
-- This allows events to be associated with a specific bill for registration validation

-- Add bill_id column to events table
ALTER TABLE events
ADD COLUMN bill_id BIGINT NULL;

-- Add foreign key constraint
ALTER TABLE events
ADD CONSTRAINT fk_events_bill_id
FOREIGN KEY (bill_id) REFERENCES bills(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- Add index for bill_id lookups
CREATE INDEX idx_events_bill_id ON events(bill_id);
