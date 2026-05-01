-- Rollback: Drop match_reports table
-- Epic 5: Match Reporting by Referees
-- Story 5.1: Match Report Database Schema

DROP INDEX IF EXISTS idx_match_reports_submitted_at;
DROP INDEX IF EXISTS idx_match_reports_submitted_by;
DROP INDEX IF EXISTS idx_match_reports_match_id;
DROP TABLE IF EXISTS match_reports;
