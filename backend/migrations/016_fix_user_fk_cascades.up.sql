-- Fix foreign key constraints on users table to allow hard deletion
-- matches.created_by: add ON DELETE SET NULL
ALTER TABLE matches DROP CONSTRAINT IF EXISTS matches_created_by_fkey;
ALTER TABLE matches ADD CONSTRAINT matches_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

-- assignment_history.actor_id: make nullable and add ON DELETE SET NULL
ALTER TABLE assignment_history ALTER COLUMN actor_id DROP NOT NULL;
ALTER TABLE assignment_history DROP CONSTRAINT IF EXISTS assignment_history_actor_id_fkey;
ALTER TABLE assignment_history ADD CONSTRAINT assignment_history_actor_id_fkey
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL;

-- match_reports.submitted_by: make nullable and add ON DELETE SET NULL
ALTER TABLE match_reports ALTER COLUMN submitted_by DROP NOT NULL;
ALTER TABLE match_reports DROP CONSTRAINT IF EXISTS match_reports_submitted_by_fkey;
ALTER TABLE match_reports ADD CONSTRAINT match_reports_submitted_by_fkey
    FOREIGN KEY (submitted_by) REFERENCES users(id) ON DELETE SET NULL;
