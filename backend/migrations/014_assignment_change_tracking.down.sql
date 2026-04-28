-- Rollback: Remove change tracking from match_roles
-- Epic 5: Match Reporting by Referees
-- Story 5.6: Assignment Change Indicator

-- Note: We do NOT drop updated_at as it was part of the original table creation
-- We only drop the viewed_by_referee column and its index

DROP INDEX IF EXISTS idx_match_roles_viewed;
ALTER TABLE match_roles DROP COLUMN IF EXISTS viewed_by_referee;
