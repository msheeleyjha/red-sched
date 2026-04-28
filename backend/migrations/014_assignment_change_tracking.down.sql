-- Rollback: Remove change tracking from match_roles
-- Epic 5: Match Reporting by Referees
-- Story 5.6: Assignment Change Indicator

DROP INDEX IF EXISTS idx_match_roles_viewed;
ALTER TABLE match_roles DROP COLUMN IF EXISTS viewed_by_referee;
ALTER TABLE match_roles DROP COLUMN IF EXISTS updated_at;
