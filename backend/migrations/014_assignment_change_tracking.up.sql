-- Migration: Add change tracking to match_roles (assignments)
-- Epic 5: Match Reporting by Referees
-- Story 5.6: Assignment Change Indicator

-- Note: updated_at already exists from migration 002_matches_schema.up.sql
-- We only need to add the viewed_by_referee column

-- Add viewed_by_referee boolean to track if referee has seen the update
ALTER TABLE match_roles
ADD COLUMN viewed_by_referee BOOLEAN NOT NULL DEFAULT FALSE;

-- Create index for efficient querying of unviewed assignments
CREATE INDEX idx_match_roles_viewed ON match_roles(assigned_referee_id, viewed_by_referee) WHERE viewed_by_referee = FALSE;

-- Comments for documentation
COMMENT ON COLUMN match_roles.updated_at IS 'Timestamp when assignment was last modified (for change tracking)';
COMMENT ON COLUMN match_roles.viewed_by_referee IS 'Whether referee has viewed the match since last update';
