-- Migration: Add role-based fields to executives table
-- Date: 2025-12-03
-- Description: Adds role, region_id, zone_id, status, and bio fields for executive role management

-- Add new columns to executives table (MySQL syntax)
-- Note: Run each statement separately if a column already exists

ALTER TABLE executives
ADD COLUMN role VARCHAR(50) NULL DEFAULT NULL COMMENT 'Executive admin role: national_admin, region_admin, zone_admin',
ADD COLUMN region_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Assigned region for region_admin and zone_admin',
ADD COLUMN zone_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Assigned zone for zone_admin',
ADD COLUMN status VARCHAR(20) NULL DEFAULT 'active' COMMENT 'Executive status: active, inactive',
ADD COLUMN bio TEXT NULL DEFAULT NULL COMMENT 'Executive bio/description';

-- Add indexes for better query performance
CREATE INDEX idx_executives_role ON executives(role);
CREATE INDEX idx_executives_region_id ON executives(region_id);
CREATE INDEX idx_executives_zone_id ON executives(zone_id);
CREATE INDEX idx_executives_status ON executives(status);

-- Add foreign key constraints (optional - uncomment if you want referential integrity)
-- ALTER TABLE executives
-- ADD CONSTRAINT fk_executives_region FOREIGN KEY (region_id) REFERENCES regions(id) ON DELETE SET NULL,
-- ADD CONSTRAINT fk_executives_zone FOREIGN KEY (zone_id) REFERENCES zones(id) ON DELETE SET NULL;
