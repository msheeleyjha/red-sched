-- Rollback: Remove change tracking from assignments table
-- Epic 5: Match Reporting by Referees
-- Story 5.6: Assignment Change Indicator

-- Note: Table was renamed from match_roles to assignments in migration 009
-- Note: We do NOT drop updated_at as it was part of the original table creation
-- We only drop the viewed_by_referee column and its index

DROP INDEX IF EXISTS idx_assignments_viewed;
ALTER TABLE assignments DROP COLUMN IF EXISTS viewed_by_referee;
