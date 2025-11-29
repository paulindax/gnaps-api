-- Migration to add zone_id and description columns to school_groups table
-- Run this SQL on your database before using the Groups feature

-- Add zone_id column (foreign key to zones table)
ALTER TABLE school_groups
ADD COLUMN IF NOT EXISTS zone_id BIGINT;

-- Add description column (optional text field)
ALTER TABLE school_groups
ADD COLUMN IF NOT EXISTS description TEXT;

-- Add foreign key constraint to zones table
ALTER TABLE school_groups
ADD CONSTRAINT fk_school_groups_zone
FOREIGN KEY (zone_id) REFERENCES zones(id)
ON DELETE RESTRICT;

-- Add index on zone_id for better query performance
CREATE INDEX IF NOT EXISTS idx_school_groups_zone_id ON school_groups(zone_id);

-- Add composite index for name and zone_id (for uniqueness checks)
CREATE INDEX IF NOT EXISTS idx_school_groups_name_zone ON school_groups(name, zone_id);

-- Update existing records to set is_deleted to false if NULL
UPDATE school_groups SET is_deleted = false WHERE is_deleted IS NULL;
