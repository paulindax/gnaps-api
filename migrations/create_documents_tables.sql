-- Create documents table
CREATE TABLE IF NOT EXISTS documents (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    status VARCHAR(50) DEFAULT 'draft',

    -- JSON field to store the document structure (fields, positions, sizes, etc.)
    template_data JSON NOT NULL,

    -- Permissions and targeting
    created_by BIGINT NOT NULL,
    is_required BOOLEAN DEFAULT FALSE,
    region_ids JSON,
    zone_ids JSON,
    group_ids JSON,
    school_ids JSON,

    -- Metadata
    version INT DEFAULT 1,
    is_deleted BOOLEAN DEFAULT FALSE,

    INDEX idx_documents_status (status),
    INDEX idx_documents_category (category),
    INDEX idx_documents_created_by (created_by),
    INDEX idx_documents_deleted (is_deleted)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create document_submissions table
CREATE TABLE IF NOT EXISTS document_submissions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    document_id BIGINT UNSIGNED NOT NULL,
    school_id BIGINT NOT NULL,
    submitted_by BIGINT NOT NULL,

    -- JSON field to store the filled form data
    form_data JSON NOT NULL,

    -- Status tracking
    status VARCHAR(50) DEFAULT 'draft',
    submitted_at TIMESTAMP NULL,
    reviewed_at TIMESTAMP NULL,
    reviewed_by BIGINT,
    review_notes TEXT,

    -- Metadata
    is_deleted BOOLEAN DEFAULT FALSE,

    INDEX idx_submissions_document (document_id),
    INDEX idx_submissions_school (school_id),
    INDEX idx_submissions_status (status),
    INDEX idx_submissions_submitted_by (submitted_by),
    INDEX idx_submissions_deleted (is_deleted),

    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create document_attachments table (for supporting files)
CREATE TABLE IF NOT EXISTS document_attachments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    submission_id BIGINT UNSIGNED NOT NULL,
    field_name VARCHAR(100) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(100),
    file_size BIGINT,

    INDEX idx_attachments_submission (submission_id),

    FOREIGN KEY (submission_id) REFERENCES document_submissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
