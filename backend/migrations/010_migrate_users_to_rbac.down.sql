-- Epic 1 Story 1.6: Rollback User Migration to RBAC
-- This removes all user role assignments created by the migration

-- Note: This does NOT restore the old role field values
-- In a rollback scenario, users.role field will remain as-is from before migration

DELETE FROM user_roles;

-- Remove deprecation comment
COMMENT ON COLUMN users.role IS 'User role: pending_referee, referee, or assignor';
