-- Migration: Add change tracking to assignments table
-- Epic 5: Match Reporting by Referees
-- Story 5.6: Assignment Change Indicator

-- Note: Table was renamed from match_roles to assignments in migration 009
-- Note: updated_at already exists from the original table creation in migration 002
-- Note: assigned_referee_id was renamed to referee_id in migration 009
-- We only need to add the viewed_by_referee column

-- Add viewed_by_referee boolean to track if referee has seen the update
ALTER TABLE assignments
ADD COLUMN viewed_by_referee BOOLEAN NOT NULL DEFAULT FALSE;

-- Create index for efficient querying of unviewed assignments
CREATE INDEX idx_assignments_viewed ON assignments(referee_id, viewed_by_referee) WHERE viewed_by_referee = FALSE;

-- Comments for documentation
COMMENT ON COLUMN assignments.updated_at IS 'Timestamp when assignment was last modified (for change tracking)';
COMMENT ON COLUMN assignments.viewed_by_referee IS 'Whether referee has viewed the match since last update';
