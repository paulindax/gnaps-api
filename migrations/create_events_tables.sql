-- Migration: Create events and event_registrations tables
-- Created: 2025-11-29
-- Database: MySQL/MariaDB

-- Create events table
CREATE TABLE IF NOT EXISTS events (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NULL,
    organization_id BIGINT,
    created_by BIGINT NOT NULL,
    location VARCHAR(255),
    venue VARCHAR(255),
    is_paid BOOLEAN DEFAULT FALSE,
    price DECIMAL(10,2),
    max_attendees INT,
    registration_deadline TIMESTAMP NULL,
    status VARCHAR(50) DEFAULT 'draft',
    image_url TEXT,
    is_deleted BOOLEAN DEFAULT FALSE,

    INDEX idx_events_start_date (start_date),
    INDEX idx_events_status (status),
    INDEX idx_events_is_deleted (is_deleted),
    INDEX idx_events_created_by (created_by),
    INDEX idx_events_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create event_registrations table
CREATE TABLE IF NOT EXISTS event_registrations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    event_id BIGINT UNSIGNED NOT NULL,
    school_id BIGINT NOT NULL,
    registered_by BIGINT NOT NULL,
    payment_status VARCHAR(50) DEFAULT 'pending',
    payment_reference VARCHAR(255),
    registration_date TIMESTAMP NULL,
    number_of_attendees INT DEFAULT 1,
    is_deleted BOOLEAN DEFAULT FALSE,

    INDEX idx_event_registrations_event_id (event_id),
    INDEX idx_event_registrations_school_id (school_id),
    INDEX idx_event_registrations_registered_by (registered_by),
    INDEX idx_event_registrations_payment_status (payment_status),
    INDEX idx_event_registrations_is_deleted (is_deleted),
    INDEX idx_event_registrations_deleted_at (deleted_at),

    UNIQUE KEY unique_event_school (event_id, school_id),

    CONSTRAINT fk_event_registrations_event FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    CONSTRAINT fk_event_registrations_school FOREIGN KEY (school_id) REFERENCES schools(id) ON DELETE CASCADE,
    CONSTRAINT fk_event_registrations_user FOREIGN KEY (registered_by) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
