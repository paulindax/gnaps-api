-- Migration: Optimize finance_accounts table indexes for better query performance
-- Created: 2025-12-02
-- Purpose: Add indexes to improve LIKE query performance and filtering
-- Database: MySQL

-- Add index on is_deleted column (frequently used in WHERE clause)
CREATE INDEX idx_finance_accounts_is_deleted
ON finance_accounts(is_deleted);

-- Add index on code column (used in searches and uniqueness checks)
CREATE INDEX idx_finance_accounts_code
ON finance_accounts(code);

-- Add composite index on is_deleted and created_at for list queries with ordering
CREATE INDEX idx_finance_accounts_deleted_created
ON finance_accounts(is_deleted, created_at);

-- Add composite index for common filtered queries
CREATE INDEX idx_finance_accounts_code_deleted
ON finance_accounts(code, is_deleted);

-- Add index on account_type for filtering
CREATE INDEX idx_finance_accounts_account_type
ON finance_accounts(account_type);

-- Add composite index for account_type with is_deleted
CREATE INDEX idx_finance_accounts_type_deleted
ON finance_accounts(account_type, is_deleted);

-- Add index on is_income for filtering
CREATE INDEX idx_finance_accounts_is_income
ON finance_accounts(is_income);

-- Add composite index for is_income with is_deleted
CREATE INDEX idx_finance_accounts_income_deleted
ON finance_accounts(is_income, is_deleted);

-- For text search optimization using MySQL FULLTEXT indexes
-- Uncomment if you want FULLTEXT search capabilities (requires MyISAM or InnoDB with MySQL 5.6+)
-- ALTER TABLE finance_accounts ADD FULLTEXT INDEX idx_finance_accounts_name_fulltext (name);
-- ALTER TABLE finance_accounts ADD FULLTEXT INDEX idx_finance_accounts_code_fulltext (code);
-- ALTER TABLE finance_accounts ADD FULLTEXT INDEX idx_finance_accounts_description_fulltext (description);
-- Or create a composite FULLTEXT index for searching across all text fields:
-- ALTER TABLE finance_accounts ADD FULLTEXT INDEX idx_finance_accounts_text_search (name, code, description);
