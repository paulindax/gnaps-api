-- ============================================================================
-- GNAPS DATABASE INDEX ROLLBACK SCRIPT
-- ============================================================================
-- This script removes all indexes created by optimize_database_indexes.sql
-- Use this if you need to rollback the index migration
--
-- Created: 2025-12-01
-- Database: MySQL 8.0+
-- ============================================================================

-- IMPORTANT: Back up your database before running this script!

START TRANSACTION;

-- ============================================================================
-- 1. USERS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_users_email ON users;
DROP INDEX IF EXISTS idx_users_username ON users;
DROP INDEX IF EXISTS idx_users_role ON users;
DROP INDEX IF EXISTS idx_users_is_deleted ON users;
DROP INDEX IF EXISTS idx_users_email_is_deleted ON users;
DROP INDEX IF EXISTS idx_users_mobile_no ON users;
DROP INDEX IF EXISTS idx_users_reset_password_token ON users;

-- ============================================================================
-- 2. REGIONS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_regions_name ON regions;
DROP INDEX IF EXISTS idx_regions_code ON regions;
DROP INDEX IF EXISTS idx_regions_is_deleted ON regions;
DROP INDEX IF EXISTS idx_regions_is_deleted_name ON regions;
DROP INDEX IF EXISTS idx_regions_created_at ON regions;

-- ============================================================================
-- 3. ZONES TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_zones_region_id ON zones;
DROP INDEX IF EXISTS idx_zones_name ON zones;
DROP INDEX IF EXISTS idx_zones_code ON zones;
DROP INDEX IF EXISTS idx_zones_is_deleted ON zones;
DROP INDEX IF EXISTS idx_zones_region_id_is_deleted ON zones;
DROP INDEX IF EXISTS idx_zones_region_id_name ON zones;
DROP INDEX IF EXISTS idx_zones_created_at ON zones;

-- ============================================================================
-- 4. SCHOOL_GROUPS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_school_groups_zone_id ON school_groups;
DROP INDEX IF EXISTS idx_school_groups_name ON school_groups;
DROP INDEX IF EXISTS idx_school_groups_is_deleted ON school_groups;
DROP INDEX IF EXISTS idx_school_groups_zone_id_is_deleted ON school_groups;
DROP INDEX IF EXISTS idx_school_groups_created_at ON school_groups;

-- ============================================================================
-- 5. SCHOOLS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_schools_zone_id ON schools;
DROP INDEX IF EXISTS idx_schools_user_id ON schools;
DROP INDEX IF EXISTS idx_schools_name ON schools;
DROP INDEX IF EXISTS idx_schools_member_no ON schools;
DROP INDEX IF EXISTS idx_schools_email ON schools;
DROP INDEX IF EXISTS idx_schools_is_deleted ON schools;
DROP INDEX IF EXISTS idx_schools_zone_id_is_deleted ON schools;
DROP INDEX IF EXISTS idx_schools_is_deleted_name ON schools;
DROP INDEX IF EXISTS idx_schools_member_no_is_deleted ON schools;
DROP INDEX IF EXISTS idx_schools_created_at ON schools;
DROP INDEX IF EXISTS idx_schools_joining_date ON schools;

-- ============================================================================
-- 6. POSITIONS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_positions_name ON positions;
DROP INDEX IF EXISTS idx_positions_created_at ON positions;

-- ============================================================================
-- 7. EXECUTIVES TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_executives_position_id ON executives;
DROP INDEX IF EXISTS idx_executives_user_id ON executives;
DROP INDEX IF EXISTS idx_executives_executive_no ON executives;
DROP INDEX IF EXISTS idx_executives_email ON executives;
DROP INDEX IF EXISTS idx_executives_mobile_no ON executives;
DROP INDEX IF EXISTS idx_executives_is_deleted ON executives;
DROP INDEX IF EXISTS idx_executives_last_first_name ON executives;
DROP INDEX IF EXISTS idx_executives_position_id_is_deleted ON executives;
DROP INDEX IF EXISTS idx_executives_created_at ON executives;

-- ============================================================================
-- 8. CONTACT_PERSONS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_contact_persons_school_id ON contact_persons;
DROP INDEX IF EXISTS idx_contact_persons_email ON contact_persons;
DROP INDEX IF EXISTS idx_contact_persons_mobile_no ON contact_persons;
DROP INDEX IF EXISTS idx_contact_persons_is_deleted ON contact_persons;
DROP INDEX IF EXISTS idx_contact_persons_school_id_is_deleted ON contact_persons;
DROP INDEX IF EXISTS idx_contact_persons_created_at ON contact_persons;

-- ============================================================================
-- 9. NEWS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_news_executive_id ON news;
DROP INDEX IF EXISTS idx_news_author_id ON news;
DROP INDEX IF EXISTS idx_news_title ON news;
DROP INDEX IF EXISTS idx_news_category ON news;
DROP INDEX IF EXISTS idx_news_status ON news;
DROP INDEX IF EXISTS idx_news_featured ON news;
DROP INDEX IF EXISTS idx_news_is_deleted ON news;
DROP INDEX IF EXISTS idx_news_status_featured_is_deleted ON news;
DROP INDEX IF EXISTS idx_news_status_is_deleted_created_at ON news;
DROP INDEX IF EXISTS idx_news_created_at ON news;

-- ============================================================================
-- 10. EVENTS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_events_organization_id ON events;
DROP INDEX IF EXISTS idx_events_created_by ON events;
DROP INDEX IF EXISTS idx_events_title ON events;
DROP INDEX IF EXISTS idx_events_status ON events;
DROP INDEX IF EXISTS idx_events_registration_code ON events;
DROP INDEX IF EXISTS idx_events_is_deleted ON events;
DROP INDEX IF EXISTS idx_events_start_end_date ON events;
DROP INDEX IF EXISTS idx_events_is_deleted_start_date ON events;
DROP INDEX IF EXISTS idx_events_status_is_deleted ON events;
DROP INDEX IF EXISTS idx_events_registration_deadline ON events;
DROP INDEX IF EXISTS idx_events_created_at ON events;

-- ============================================================================
-- 11. EVENT_REGISTRATIONS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_event_registrations_event_id ON event_registrations;
DROP INDEX IF EXISTS idx_event_registrations_user_id ON event_registrations;
DROP INDEX IF EXISTS idx_event_registrations_school_id ON event_registrations;
DROP INDEX IF EXISTS idx_event_registrations_email ON event_registrations;
DROP INDEX IF EXISTS idx_event_registrations_status ON event_registrations;
DROP INDEX IF EXISTS idx_event_registrations_event_id_status ON event_registrations;
DROP INDEX IF EXISTS idx_event_registrations_user_id_event_id ON event_registrations;
DROP INDEX IF EXISTS idx_event_registrations_created_at ON event_registrations;

-- ============================================================================
-- 12. NEWS_COMMENTS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_news_comments_news_id ON news_comments;
DROP INDEX IF EXISTS idx_news_comments_user_id ON news_comments;
DROP INDEX IF EXISTS idx_news_comments_is_deleted ON news_comments;
DROP INDEX IF EXISTS idx_news_comments_news_id_is_deleted ON news_comments;
DROP INDEX IF EXISTS idx_news_comments_created_at ON news_comments;

-- ============================================================================
-- 13. DOCUMENTS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_documents_created_by ON documents;
DROP INDEX IF EXISTS idx_documents_title ON documents;
DROP INDEX IF EXISTS idx_documents_category ON documents;
DROP INDEX IF EXISTS idx_documents_status ON documents;
DROP INDEX IF EXISTS idx_documents_is_required ON documents;
DROP INDEX IF EXISTS idx_documents_is_deleted ON documents;
DROP INDEX IF EXISTS idx_documents_status_is_required_is_deleted ON documents;
DROP INDEX IF EXISTS idx_documents_version ON documents;
DROP INDEX IF EXISTS idx_documents_created_at ON documents;

-- ============================================================================
-- 14. DOCUMENT_SUBMISSIONS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_document_submissions_document_id ON document_submissions;
DROP INDEX IF EXISTS idx_document_submissions_school_id ON document_submissions;
DROP INDEX IF EXISTS idx_document_submissions_submitted_by ON document_submissions;
DROP INDEX IF EXISTS idx_document_submissions_reviewed_by ON document_submissions;
DROP INDEX IF EXISTS idx_document_submissions_status ON document_submissions;
DROP INDEX IF EXISTS idx_document_submissions_school_id_document_id ON document_submissions;
DROP INDEX IF EXISTS idx_document_submissions_status_reviewed_by ON document_submissions;
DROP INDEX IF EXISTS idx_document_submissions_submitted_at ON document_submissions;
DROP INDEX IF EXISTS idx_document_submissions_created_at ON document_submissions;

-- ============================================================================
-- 15. FINANCE_ACCOUNTS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_finance_accounts_name ON finance_accounts;
DROP INDEX IF EXISTS idx_finance_accounts_code ON finance_accounts;
DROP INDEX IF EXISTS idx_finance_accounts_account_type ON finance_accounts;
DROP INDEX IF EXISTS idx_finance_accounts_is_income ON finance_accounts;
DROP INDEX IF EXISTS idx_finance_accounts_approver_id ON finance_accounts;
DROP INDEX IF EXISTS idx_finance_accounts_is_deleted ON finance_accounts;
DROP INDEX IF EXISTS idx_finance_accounts_account_type_is_deleted ON finance_accounts;
DROP INDEX IF EXISTS idx_finance_accounts_created_at ON finance_accounts;

-- ============================================================================
-- 16. BILL_PARTICULARS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_bill_particulars_name ON bill_particulars;
DROP INDEX IF EXISTS idx_bill_particulars_priority ON bill_particulars;
DROP INDEX IF EXISTS idx_bill_particulars_is_deleted ON bill_particulars;
DROP INDEX IF EXISTS idx_bill_particulars_is_deleted_priority ON bill_particulars;
DROP INDEX IF EXISTS idx_bill_particulars_created_at ON bill_particulars;

-- ============================================================================
-- 17. BILLS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_bills_name ON bills;
DROP INDEX IF EXISTS idx_bills_is_approved ON bills;
DROP INDEX IF EXISTS idx_bills_is_generating ON bills;
DROP INDEX IF EXISTS idx_bills_is_deleted ON bills;
DROP INDEX IF EXISTS idx_bills_is_deleted_is_approved ON bills;
DROP INDEX IF EXISTS idx_bills_created_at ON bills;

-- ============================================================================
-- 18. BILL_ITEMS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_bill_items_bill_id ON bill_items;
DROP INDEX IF EXISTS idx_bill_items_bill_particular_id ON bill_items;
DROP INDEX IF EXISTS idx_bill_items_finance_account_id ON bill_items;
DROP INDEX IF EXISTS idx_bill_items_priority ON bill_items;
DROP INDEX IF EXISTS idx_bill_items_is_approved ON bill_items;
DROP INDEX IF EXISTS idx_bill_items_is_deleted ON bill_items;
DROP INDEX IF EXISTS idx_bill_items_bill_id_is_deleted ON bill_items;
DROP INDEX IF EXISTS idx_bill_items_bill_id_priority ON bill_items;
DROP INDEX IF EXISTS idx_bill_items_bill_particular_id_is_deleted ON bill_items;
DROP INDEX IF EXISTS idx_bill_items_created_at ON bill_items;

-- ============================================================================
-- 19. BILL_ASSIGNMENTS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_bill_assignments_bill_item_id ON bill_assignments;
DROP INDEX IF EXISTS idx_bill_assignments_entity_type_entity_id ON bill_assignments;
DROP INDEX IF EXISTS idx_bill_assignments_bill_item_id_entity_type ON bill_assignments;
DROP INDEX IF EXISTS idx_bill_assignments_entity_id_entity_type ON bill_assignments;
DROP INDEX IF EXISTS idx_bill_assignments_created_at ON bill_assignments;

-- ============================================================================
-- 20. SCHOOL_BILLINGS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_school_billings_school_id ON school_billings;
DROP INDEX IF EXISTS idx_school_billings_bill_id ON school_billings;
DROP INDEX IF EXISTS idx_school_billings_payment_status ON school_billings;
DROP INDEX IF EXISTS idx_school_billings_school_id_bill_id ON school_billings;
DROP INDEX IF EXISTS idx_school_billings_payment_status_school_id ON school_billings;
DROP INDEX IF EXISTS idx_school_billings_due_date ON school_billings;
DROP INDEX IF EXISTS idx_school_billings_created_at ON school_billings;

-- ============================================================================
-- 21. FINANCE_TRANSACTIONS TABLE INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_finance_transactions_finance_account_id ON finance_transactions;
DROP INDEX IF EXISTS idx_finance_transactions_school_id ON finance_transactions;
DROP INDEX IF EXISTS idx_finance_transactions_transaction_type ON finance_transactions;
DROP INDEX IF EXISTS idx_finance_transactions_transaction_date ON finance_transactions;
DROP INDEX IF EXISTS idx_finance_transactions_finance_account_id_transaction_date ON finance_transactions;
DROP INDEX IF EXISTS idx_finance_transactions_school_id_transaction_date ON finance_transactions;
DROP INDEX IF EXISTS idx_finance_transactions_created_at ON finance_transactions;

-- ============================================================================
-- COMMIT TRANSACTION
-- ============================================================================

COMMIT;

SELECT 'All database indexes have been successfully removed.' AS status;

-- ============================================================================
-- POST-ROLLBACK VERIFICATION
-- ============================================================================
/*

-- Verify all custom indexes have been removed
SELECT
    TABLE_NAME,
    INDEX_NAME,
    INDEX_TYPE,
    COLUMN_NAME
FROM INFORMATION_SCHEMA.STATISTICS
WHERE TABLE_SCHEMA = 'gnaps'
    AND INDEX_NAME LIKE 'idx_%'
ORDER BY TABLE_NAME, INDEX_NAME;

-- This should return empty or only show PRIMARY and deleted_at indexes (from GORM)

*/
