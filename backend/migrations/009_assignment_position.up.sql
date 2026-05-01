-- Epic 1 Story 1.7: Assignment Position Field
-- Rename match_roles to assignments and role_type to position for clarity
-- This aligns with V2 PRD terminology

-- Rename table
ALTER TABLE match_roles RENAME TO assignments;

-- Rename column
ALTER TABLE assignments RENAME COLUMN role_type TO position;

-- Rename column for clarity
ALTER TABLE assignments RENAME COLUMN assigned_referee_id TO referee_id;

-- Add CHECK constraint to ensure valid position values
-- Position can be configurable per match type in the future, but start with standard positions
ALTER TABLE assignments ADD CONSTRAINT chk_assignment_position
    CHECK (position IN ('center', 'assistant_1', 'assistant_2'));

-- Update assignment_history table to match new column name
ALTER TABLE assignment_history RENAME COLUMN role_type TO position;

-- Add CHECK constraint to assignment_history as well
ALTER TABLE assignment_history ADD CONSTRAINT chk_assignment_history_position
    CHECK (position IN ('center', 'assistant_1', 'assistant_2'));

-- Update comments
COMMENT ON TABLE assignments IS 'Referee assignments to matches with position (center, assistant_1, assistant_2)';
COMMENT ON COLUMN assignments.position IS 'Referee position for this match: center (can submit reports), assistant_1, or assistant_2';
COMMENT ON COLUMN assignments.referee_id IS 'Referee assigned to this position (NULL if unassigned)';

-- Note: Indexes are automatically renamed by PostgreSQL when renaming tables
-- Old: idx_match_roles_match_id -> New: idx_assignments_match_id
-- Old: idx_match_roles_referee_id -> New: idx_assignments_referee_id
