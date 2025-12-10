-- Rollback Migration: Remove owner_type and owner_id fields from multiple tables
-- Date: 2025-12-10
-- Description: Removes owner_type (string) and owner_id (bigint) columns

-- ============================================
-- Activity Logs
-- ============================================
DROP INDEX idx_activity_logs_owner ON activity_logs;
ALTER TABLE activity_logs DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Bill Assignments
-- ============================================
DROP INDEX idx_bill_assignments_owner ON bill_assignments;
ALTER TABLE bill_assignments DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Bill Items
-- ============================================
DROP INDEX idx_bill_items_owner ON bill_items;
ALTER TABLE bill_items DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Bill Particulars
-- ============================================
DROP INDEX idx_bill_particulars_owner ON bill_particulars;
ALTER TABLE bill_particulars DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Bills
-- ============================================
DROP INDEX idx_bills_owner ON bills;
ALTER TABLE bills DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Document Attachments
-- ============================================
DROP INDEX idx_document_attachments_owner ON document_attachments;
ALTER TABLE document_attachments DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Document Submissions
-- ============================================
DROP INDEX idx_document_submissions_owner ON document_submissions;
ALTER TABLE document_submissions DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Documents
-- ============================================
DROP INDEX idx_documents_owner ON documents;
ALTER TABLE documents DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Events
-- ============================================
DROP INDEX idx_events_owner ON events;
ALTER TABLE events DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Event Registrations
-- ============================================
DROP INDEX idx_event_registrations_owner ON event_registrations;
ALTER TABLE event_registrations DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Executives
-- ============================================
DROP INDEX idx_executives_owner ON executives;
ALTER TABLE executives DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Finance Accounts
-- ============================================
DROP INDEX idx_finance_accounts_owner ON finance_accounts;
ALTER TABLE finance_accounts DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Finance Expenses
-- ============================================
DROP INDEX idx_finance_expenses_owner ON finance_expenses;
ALTER TABLE finance_expenses DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Finance Transactions
-- ============================================
DROP INDEX idx_finance_transactions_owner ON finance_transactions;
ALTER TABLE finance_transactions DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Message Logs
-- ============================================
DROP INDEX idx_message_logs_owner ON message_logs;
ALTER TABLE message_logs DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Message Packages
-- ============================================
DROP INDEX idx_message_packages_owner ON message_packages;
ALTER TABLE message_packages DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Momo Payments
-- ============================================
DROP INDEX idx_momo_payments_owner ON momo_payments;
ALTER TABLE momo_payments DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- News
-- ============================================
DROP INDEX idx_news_owner ON news;
ALTER TABLE news DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- News Comments
-- ============================================
DROP INDEX idx_news_comments_owner ON news_comments;
ALTER TABLE news_comments DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- Particular Payments
-- ============================================
DROP INDEX idx_particular_payments_owner ON particular_payments;
ALTER TABLE particular_payments DROP COLUMN owner_type, DROP COLUMN owner_id;

-- ============================================
-- School Groups
-- ============================================
DROP INDEX idx_school_groups_owner ON school_groups;
ALTER TABLE school_groups DROP COLUMN owner_type, DROP COLUMN owner_id;
