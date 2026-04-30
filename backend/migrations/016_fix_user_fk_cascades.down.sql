-- Revert FK constraint changes
ALTER TABLE matches DROP CONSTRAINT IF EXISTS matches_created_by_fkey;
ALTER TABLE matches ADD CONSTRAINT matches_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id);

ALTER TABLE match_reports DROP CONSTRAINT IF EXISTS match_reports_submitted_by_fkey;
ALTER TABLE match_reports ADD CONSTRAINT match_reports_submitted_by_fkey
    FOREIGN KEY (submitted_by) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE match_reports ALTER COLUMN submitted_by SET NOT NULL;
