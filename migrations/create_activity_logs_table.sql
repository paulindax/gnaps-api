-- Migration: Create activity_logs table for user activity tracking
-- Created: 2025-12-08
-- Database: MySQL

-- Create activity_logs table
CREATE TABLE IF NOT EXISTS `activity_logs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `created_at` DATETIME(3) NULL DEFAULT NULL,
    `updated_at` DATETIME(3) NULL DEFAULT NULL,
    `deleted_at` DATETIME(3) NULL DEFAULT NULL,

    -- User information
    `user_id` BIGINT NOT NULL,
    `username` VARCHAR(255) NULL DEFAULT NULL,
    `role` VARCHAR(100) NULL DEFAULT NULL,

    -- Activity details
    `type` VARCHAR(50) NOT NULL,
    `title` VARCHAR(500) NOT NULL,
    `description` TEXT NULL,

    -- API call details
    `method` VARCHAR(10) NULL DEFAULT NULL,
    `endpoint` VARCHAR(500) NULL DEFAULT NULL,
    `status_code` INT NULL DEFAULT NULL,

    -- Navigation/Context
    `url` VARCHAR(500) NULL DEFAULT NULL,

    -- Resource reference
    `resource_type` VARCHAR(100) NULL DEFAULT NULL,
    `resource_id` BIGINT UNSIGNED NULL DEFAULT NULL,

    -- Client information
    `ip_address` VARCHAR(45) NULL DEFAULT NULL,
    `user_agent` TEXT NULL,

    PRIMARY KEY (`id`),
    INDEX `idx_activity_logs_user_id` (`user_id`),
    INDEX `idx_activity_logs_type` (`type`),
    INDEX `idx_activity_logs_created_at` (`created_at`),
    INDEX `idx_activity_logs_deleted_at` (`deleted_at`),
    INDEX `idx_activity_logs_resource_type` (`resource_type`),
    INDEX `idx_activity_logs_user_created` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
