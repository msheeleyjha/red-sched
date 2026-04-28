-- Add archival fields to matches table
ALTER TABLE matches
    ADD COLUMN archived BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN archived_at TIMESTAMP,
    ADD COLUMN archived_by INTEGER REFERENCES users(id) ON DELETE SET NULL;

-- Create index on archived field for efficient filtering
CREATE INDEX idx_matches_archived ON matches(archived);

-- Create composite index for common query patterns (active matches by date)
CREATE INDEX idx_matches_active_date ON matches(archived, match_date) WHERE archived = FALSE;

-- Add comments documenting the archival system
COMMENT ON COLUMN matches.archived IS 'Whether this match has been archived (completed and removed from active views)';
COMMENT ON COLUMN matches.archived_at IS 'Timestamp when the match was archived (typically when final score submitted)';
COMMENT ON COLUMN matches.archived_by IS 'User ID of the referee or assignor who archived the match';

-- Set existing matches to archived = false (already default, but explicit for clarity)
-- This is safe because we're using DEFAULT FALSE in the ALTER TABLE above
UPDATE matches SET archived = FALSE WHERE archived IS NULL;
