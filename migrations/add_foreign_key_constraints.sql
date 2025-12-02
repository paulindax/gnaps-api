-- ============================================================================
-- GNAPS DATABASE FOREIGN KEY CONSTRAINTS
-- ============================================================================
-- This migration adds foreign key constraints to enforce referential integrity
-- Run this AFTER optimize_database_indexes.sql
--
-- Created: 2025-12-01
-- Database: MySQL 8.0+
-- ============================================================================

-- Start transaction for atomic execution
START TRANSACTION;

-- ============================================================================
-- IMPORTANT: Prerequisites
-- ============================================================================
-- Before running this migration:
-- 1. Ensure all referenced tables exist
-- 2. Ensure all foreign key columns have indexes (done in optimize_database_indexes.sql)
-- 3. Clean up any orphaned records that violate referential integrity
-- 4. Back up your database

-- ============================================================================
-- ZONES TABLE FOREIGN KEYS
-- ============================================================================

-- Zone must belong to a valid region
ALTER TABLE zones
ADD CONSTRAINT fk_zones_region_id
FOREIGN KEY (region_id) REFERENCES regions(id)
ON DELETE RESTRICT
ON UPDATE CASCADE;

-- ============================================================================
-- SCHOOL_GROUPS TABLE FOREIGN KEYS
-- ============================================================================

-- School group must belong to a valid zone
ALTER TABLE school_groups
ADD CONSTRAINT fk_school_groups_zone_id
FOREIGN KEY (zone_id) REFERENCES zones(id)
ON DELETE RESTRICT
ON UPDATE CASCADE;

-- ============================================================================
-- SCHOOLS TABLE FOREIGN KEYS
-- ============================================================================

-- School must belong to a valid zone
ALTER TABLE schools
ADD CONSTRAINT fk_schools_zone_id
FOREIGN KEY (zone_id) REFERENCES zones(id)
ON DELETE RESTRICT
ON UPDATE CASCADE;

-- School user association
ALTER TABLE schools
ADD CONSTRAINT fk_schools_user_id
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- ============================================================================
-- EXECUTIVES TABLE FOREIGN KEYS
-- ============================================================================

-- Executive must have a valid position
ALTER TABLE executives
ADD CONSTRAINT fk_executives_position_id
FOREIGN KEY (position_id) REFERENCES positions(id)
ON DELETE RESTRICT
ON UPDATE CASCADE;

-- Executive user association
ALTER TABLE executives
ADD CONSTRAINT fk_executives_user_id
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- ============================================================================
-- CONTACT_PERSONS TABLE FOREIGN KEYS
-- ============================================================================

-- Contact person must belong to a valid school
ALTER TABLE contact_persons
ADD CONSTRAINT fk_contact_persons_school_id
FOREIGN KEY (school_id) REFERENCES schools(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- ============================================================================
-- NEWS TABLE FOREIGN KEYS
-- ============================================================================

-- News executive association
ALTER TABLE news
ADD CONSTRAINT fk_news_executive_id
FOREIGN KEY (executive_id) REFERENCES executives(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- News author association
ALTER TABLE news
ADD CONSTRAINT fk_news_author_id
FOREIGN KEY (author_id) REFERENCES users(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- ============================================================================
-- NEWS_COMMENTS TABLE FOREIGN KEYS
-- ============================================================================

-- Comment must belong to a valid news article
ALTER TABLE news_comments
ADD CONSTRAINT fk_news_comments_news_id
FOREIGN KEY (news_id) REFERENCES news(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- Comment author
ALTER TABLE news_comments
ADD CONSTRAINT fk_news_comments_user_id
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- ============================================================================
-- EVENTS TABLE FOREIGN KEYS
-- ============================================================================

-- Event creator
ALTER TABLE events
ADD CONSTRAINT fk_events_created_by
FOREIGN KEY (created_by) REFERENCES users(id)
ON DELETE RESTRICT
ON UPDATE CASCADE;

-- Event organization association
ALTER TABLE events
ADD CONSTRAINT fk_events_organization_id
FOREIGN KEY (organization_id) REFERENCES regions(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- ============================================================================
-- EVENT_REGISTRATIONS TABLE FOREIGN KEYS
-- ============================================================================

-- Registration must belong to a valid event
ALTER TABLE event_registrations
ADD CONSTRAINT fk_event_registrations_event_id
FOREIGN KEY (event_id) REFERENCES events(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- Registration user association
ALTER TABLE event_registrations
ADD CONSTRAINT fk_event_registrations_user_id
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- Registration school association
ALTER TABLE event_registrations
ADD CONSTRAINT fk_event_registrations_school_id
FOREIGN KEY (school_id) REFERENCES schools(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- ============================================================================
-- DOCUMENTS TABLE FOREIGN KEYS
-- ============================================================================

-- Document creator
ALTER TABLE documents
ADD CONSTRAINT fk_documents_created_by
FOREIGN KEY (created_by) REFERENCES users(id)
ON DELETE RESTRICT
ON UPDATE CASCADE;

-- ============================================================================
-- DOCUMENT_SUBMISSIONS TABLE FOREIGN KEYS
-- ============================================================================

-- Submission must belong to a valid document
ALTER TABLE document_submissions
ADD CONSTRAINT fk_document_submissions_document_id
FOREIGN KEY (document_id) REFERENCES documents(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- Submission must belong to a valid school
ALTER TABLE document_submissions
ADD CONSTRAINT fk_document_submissions_school_id
FOREIGN KEY (school_id) REFERENCES schools(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- Submission submitter
ALTER TABLE document_submissions
ADD CONSTRAINT fk_document_submissions_submitted_by
FOREIGN KEY (submitted_by) REFERENCES users(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- Submission reviewer
ALTER TABLE document_submissions
ADD CONSTRAINT fk_document_submissions_reviewed_by
FOREIGN KEY (reviewed_by) REFERENCES users(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- ============================================================================
-- FINANCE_ACCOUNTS TABLE FOREIGN KEYS
-- ============================================================================

-- Finance account approver
ALTER TABLE finance_accounts
ADD CONSTRAINT fk_finance_accounts_approver_id
FOREIGN KEY (approver_id) REFERENCES users(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- ============================================================================
-- BILL_ITEMS TABLE FOREIGN KEYS
-- ============================================================================

-- Bill item must belong to a valid bill
ALTER TABLE bill_items
ADD CONSTRAINT fk_bill_items_bill_id
FOREIGN KEY (bill_id) REFERENCES bills(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- Bill item must reference a valid bill particular
ALTER TABLE bill_items
ADD CONSTRAINT fk_bill_items_bill_particular_id
FOREIGN KEY (bill_particular_id) REFERENCES bill_particulars(id)
ON DELETE RESTRICT
ON UPDATE CASCADE;

-- Bill item finance account association
ALTER TABLE bill_items
ADD CONSTRAINT fk_bill_items_finance_account_id
FOREIGN KEY (finance_account_id) REFERENCES finance_accounts(id)
ON DELETE SET NULL
ON UPDATE CASCADE;

-- ============================================================================
-- BILL_ASSIGNMENTS TABLE FOREIGN KEYS
-- ============================================================================

-- Bill assignment must belong to a valid bill item
ALTER TABLE bill_assignments
ADD CONSTRAINT fk_bill_assignments_bill_item_id
FOREIGN KEY (bill_item_id) REFERENCES bill_items(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- Note: entity_id references different tables based on entity_type
-- Cannot create FK constraint for polymorphic relationship
-- Application-level validation required

-- ============================================================================
-- SCHOOL_BILLINGS TABLE FOREIGN KEYS (if exists)
-- ============================================================================

-- School billing must belong to a valid school
ALTER TABLE school_billings
ADD CONSTRAINT fk_school_billings_school_id
FOREIGN KEY (school_id) REFERENCES schools(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- School billing must reference a valid bill
ALTER TABLE school_billings
ADD CONSTRAINT fk_school_billings_bill_id
FOREIGN KEY (bill_id) REFERENCES bills(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- ============================================================================
-- FINANCE_TRANSACTIONS TABLE FOREIGN KEYS (if exists)
-- ============================================================================

-- Transaction must belong to a valid finance account
ALTER TABLE finance_transactions
ADD CONSTRAINT fk_finance_transactions_finance_account_id
FOREIGN KEY (finance_account_id) REFERENCES finance_accounts(id)
ON DELETE RESTRICT
ON UPDATE CASCADE;

-- Transaction school association
ALTER TABLE finance_transactions
ADD CONSTRAINT fk_finance_transactions_school_id
FOREIGN KEY (school_id) REFERENCES schools(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- ============================================================================
-- COMMIT TRANSACTION
-- ============================================================================

COMMIT;

-- ============================================================================
-- NOTES ON FOREIGN KEY CONSTRAINTS
-- ============================================================================
/*

1. **ON DELETE Strategies:**
   - RESTRICT: Prevents deletion if referenced records exist (default for core entities)
   - CASCADE: Automatically deletes child records (used for dependent entities)
   - SET NULL: Sets foreign key to NULL (used for optional associations)

2. **ON UPDATE CASCADE:**
   - All foreign keys use CASCADE to automatically update references
   - Ensures data consistency when primary keys change (rare but important)

3. **Performance Considerations:**
   - Foreign keys add overhead to INSERT/UPDATE/DELETE operations
   - Benefits: Data integrity, query optimizer hints, documentation
   - Ensure indexes exist on all foreign key columns (already done)

4. **Orphaned Records Cleanup:**
   Before running this migration, clean up orphaned records:

   -- Find orphaned zones (zones without valid region)
   SELECT z.* FROM zones z
   LEFT JOIN regions r ON z.region_id = r.id
   WHERE z.region_id IS NOT NULL AND r.id IS NULL;

   -- Find orphaned schools (schools without valid zone)
   SELECT s.* FROM schools s
   LEFT JOIN zones z ON s.zone_id = z.id
   WHERE s.zone_id IS NOT NULL AND z.id IS NULL;

   -- Similar queries for other tables...

5. **Application-Level Validation:**
   Some relationships require application-level validation:
   - bill_assignments.entity_id (polymorphic relationship)
   - JSON array fields (region_ids, zone_ids, school_ids, etc.)

6. **Rollback Procedure:**
   If needed, constraints can be dropped:
   ALTER TABLE zones DROP FOREIGN KEY fk_zones_region_id;
   (Repeat for all constraints)

*/

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================
/*

-- List all foreign key constraints
SELECT
    TABLE_NAME,
    CONSTRAINT_NAME,
    REFERENCED_TABLE_NAME,
    DELETE_RULE,
    UPDATE_RULE
FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA = 'gnaps'
ORDER BY TABLE_NAME;

-- Check for orphaned records before adding constraints
SELECT 'zones with invalid region_id' AS issue, COUNT(*) AS count
FROM zones z
LEFT JOIN regions r ON z.region_id = r.id
WHERE z.region_id IS NOT NULL AND r.id IS NULL

UNION ALL

SELECT 'schools with invalid zone_id', COUNT(*)
FROM schools s
LEFT JOIN zones z ON s.zone_id = z.id
WHERE s.zone_id IS NOT NULL AND z.id IS NULL

UNION ALL

SELECT 'bill_items with invalid bill_id', COUNT(*)
FROM bill_items bi
LEFT JOIN bills b ON bi.bill_id = b.id
WHERE bi.bill_id IS NOT NULL AND b.id IS NULL;

*/
