-- Rollback Migration: Remove finance_accounts table indexes
-- Created: 2025-12-02
-- Purpose: Rollback optimize_finance_accounts_indexes.sql migration
-- Database: MySQL

-- Drop composite indexes first (in reverse order)
DROP INDEX idx_finance_accounts_income_deleted ON finance_accounts;
DROP INDEX idx_finance_accounts_is_income ON finance_accounts;
DROP INDEX idx_finance_accounts_type_deleted ON finance_accounts;
DROP INDEX idx_finance_accounts_account_type ON finance_accounts;
DROP INDEX idx_finance_accounts_code_deleted ON finance_accounts;
DROP INDEX idx_finance_accounts_deleted_created ON finance_accounts;
DROP INDEX idx_finance_accounts_code ON finance_accounts;
DROP INDEX idx_finance_accounts_is_deleted ON finance_accounts;

-- If you enabled FULLTEXT indexes, uncomment to drop them
-- DROP INDEX idx_finance_accounts_text_search ON finance_accounts;
-- DROP INDEX idx_finance_accounts_description_fulltext ON finance_accounts;
-- DROP INDEX idx_finance_accounts_code_fulltext ON finance_accounts;
-- DROP INDEX idx_finance_accounts_name_fulltext ON finance_accounts;
