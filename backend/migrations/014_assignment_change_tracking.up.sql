-- Migration: Add change tracking to match_roles (assignments)
-- Epic 5: Match Reporting by Referees
-- Story 5.6: Assignment Change Indicator

-- Add updated_at timestamp to track when assignment was last modified
ALTER TABLE match_roles
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Add viewed_by_referee boolean to track if referee has seen the update
ALTER TABLE match_roles
ADD COLUMN viewed_by_referee BOOLEAN NOT NULL DEFAULT FALSE;

-- Create index for efficient querying of unviewed assignments
CREATE INDEX idx_match_roles_viewed ON match_roles(assigned_referee_id, viewed_by_referee) WHERE viewed_by_referee = FALSE;

-- Comments for documentation
COMMENT ON COLUMN match_roles.updated_at IS 'Timestamp when assignment was last modified (for change tracking)';
COMMENT ON COLUMN match_roles.viewed_by_referee IS 'Whether referee has viewed the match since last update';
