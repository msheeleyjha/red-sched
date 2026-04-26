-- Epic 1 Story 1.7: Rollback Assignment Position Field
-- Revert assignments table back to match_roles

-- Remove CHECK constraints
ALTER TABLE assignments DROP CONSTRAINT IF EXISTS chk_assignment_position;
ALTER TABLE assignment_history DROP CONSTRAINT IF EXISTS chk_assignment_history_position;

-- Rename columns back
ALTER TABLE assignments RENAME COLUMN position TO role_type;
ALTER TABLE assignments RENAME COLUMN referee_id TO assigned_referee_id;
ALTER TABLE assignment_history RENAME COLUMN position TO role_type;

-- Rename table back
ALTER TABLE assignments RENAME TO match_roles;
