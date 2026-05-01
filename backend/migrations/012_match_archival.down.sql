-- Remove indexes
DROP INDEX IF EXISTS idx_matches_active_date;
DROP INDEX IF EXISTS idx_matches_archived;

-- Remove archival columns from matches table
ALTER TABLE matches
    DROP COLUMN IF EXISTS archived_by,
    DROP COLUMN IF EXISTS archived_at,
    DROP COLUMN IF EXISTS archived;
