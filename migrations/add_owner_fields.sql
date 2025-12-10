-- Migration: Add owner_type and owner_id fields to multiple tables
-- Date: 2025-12-10
-- Description: Adds owner_type (string) and owner_id (bigint) for multi-tenancy/ownership tracking
-- Excluded tables: contact_persons, payment_gateways, positions, regions, roles, role_users,
--                  school_billing_particulars, school_bills, schools, users, zones

-- ============================================
-- Activity Logs
-- ============================================
ALTER TABLE activity_logs
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_activity_logs_owner ON activity_logs(owner_type, owner_id);

-- ============================================
-- Bill Assignments
-- ============================================
ALTER TABLE bill_assignments
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_bill_assignments_owner ON bill_assignments(owner_type, owner_id);

-- ============================================
-- Bill Items
-- ============================================
ALTER TABLE bill_items
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_bill_items_owner ON bill_items(owner_type, owner_id);

-- ============================================
-- Bill Particulars
-- ============================================
ALTER TABLE bill_particulars
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_bill_particulars_owner ON bill_particulars(owner_type, owner_id);

-- ============================================
-- Bills
-- ============================================
ALTER TABLE bills
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_bills_owner ON bills(owner_type, owner_id);

-- ============================================
-- Document Attachments
-- ============================================
ALTER TABLE document_attachments
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_document_attachments_owner ON document_attachments(owner_type, owner_id);

-- ============================================
-- Document Submissions
-- ============================================
ALTER TABLE document_submissions
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_document_submissions_owner ON document_submissions(owner_type, owner_id);

-- ============================================
-- Documents
-- ============================================
ALTER TABLE documents
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_documents_owner ON documents(owner_type, owner_id);

-- ============================================
-- Events
-- ============================================
ALTER TABLE events
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_events_owner ON events(owner_type, owner_id);

-- ============================================
-- Event Registrations
-- ============================================
ALTER TABLE event_registrations
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_event_registrations_owner ON event_registrations(owner_type, owner_id);

-- ============================================
-- Finance Accounts
-- ============================================
ALTER TABLE finance_accounts
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_finance_accounts_owner ON finance_accounts(owner_type, owner_id);

-- ============================================
-- Finance Expenses
-- ============================================
ALTER TABLE finance_expenses
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_finance_expenses_owner ON finance_expenses(owner_type, owner_id);

-- ============================================
-- Finance Transactions
-- ============================================
ALTER TABLE finance_transactions
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_finance_transactions_owner ON finance_transactions(owner_type, owner_id);

-- ============================================
-- Message Logs
-- ============================================
ALTER TABLE message_logs
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_message_logs_owner ON message_logs(owner_type, owner_id);

-- ============================================
-- Message Packages
-- ============================================
ALTER TABLE message_packages
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_message_packages_owner ON message_packages(owner_type, owner_id);

-- ============================================
-- Momo Payments
-- ============================================
ALTER TABLE momo_payments
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_momo_payments_owner ON momo_payments(owner_type, owner_id);

-- ============================================
-- News
-- ============================================
ALTER TABLE news
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_news_owner ON news(owner_type, owner_id);

-- ============================================
-- News Comments
-- ============================================
ALTER TABLE news_comments
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_news_comments_owner ON news_comments(owner_type, owner_id);

-- ============================================
-- Particular Payments
-- ============================================
ALTER TABLE particular_payments
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_particular_payments_owner ON particular_payments(owner_type, owner_id);

-- ============================================
-- School Groups
-- ============================================
ALTER TABLE school_groups
ADD COLUMN owner_type VARCHAR(50) NULL DEFAULT NULL COMMENT 'Owner entity type',
ADD COLUMN owner_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT 'Owner entity ID';

CREATE INDEX idx_school_groups_owner ON school_groups(owner_type, owner_id);
