-- Epic 1 Story 1.6: User Migration to RBAC
-- Migrate existing users from old role field to new user_roles table

-- Create a temporary Super Admin user (first user in system becomes Super Admin)
-- In production, this would be configured via environment variable or manual assignment
DO $$
DECLARE
    first_user_id BIGINT;
    super_admin_role_id BIGINT;
    assignor_role_id BIGINT;
    referee_role_id BIGINT;
BEGIN
    -- Get role IDs
    SELECT id INTO super_admin_role_id FROM roles WHERE name = 'Super Admin';
    SELECT id INTO assignor_role_id FROM roles WHERE name = 'Assignor';
    SELECT id INTO referee_role_id FROM roles WHERE name = 'Referee';

    -- Get first user ID (will become Super Admin)
    SELECT id INTO first_user_id FROM users ORDER BY created_at LIMIT 1;

    -- If there's at least one user, make them Super Admin
    IF first_user_id IS NOT NULL THEN
        INSERT INTO user_roles (user_id, role_id, assigned_by)
        VALUES (first_user_id, super_admin_role_id, first_user_id)
        ON CONFLICT (user_id, role_id) DO NOTHING;

        RAISE NOTICE 'Assigned Super Admin role to user ID: %', first_user_id;
    END IF;

    -- Migrate users with old 'assignor' role
    INSERT INTO user_roles (user_id, role_id, assigned_by)
    SELECT id, assignor_role_id, first_user_id
    FROM users
    WHERE role = 'assignor' AND id != first_user_id
    ON CONFLICT (user_id, role_id) DO NOTHING;

    -- Migrate users with old 'referee' role (and 'pending_referee' for completeness)
    INSERT INTO user_roles (user_id, role_id, assigned_by)
    SELECT id, referee_role_id, first_user_id
    FROM users
    WHERE role IN ('referee', 'pending_referee')
    ON CONFLICT (user_id, role_id) DO NOTHING;

    -- Handle users with 'assignor' role - they likely also work as referees
    -- Give assignors the Referee role as well (multi-role)
    INSERT INTO user_roles (user_id, role_id, assigned_by)
    SELECT id, referee_role_id, first_user_id
    FROM users
    WHERE role = 'assignor'
    ON CONFLICT (user_id, role_id) DO NOTHING;

    -- Log migration summary
    RAISE NOTICE 'User migration complete';
    RAISE NOTICE 'Super Admins: %', (SELECT COUNT(*) FROM user_roles WHERE role_id = super_admin_role_id);
    RAISE NOTICE 'Assignors: %', (SELECT COUNT(*) FROM user_roles WHERE role_id = assignor_role_id);
    RAISE NOTICE 'Referees: %', (SELECT COUNT(*) FROM user_roles WHERE role_id = referee_role_id);
    RAISE NOTICE 'Users with no roles: %', (SELECT COUNT(*) FROM users WHERE id NOT IN (SELECT user_id FROM user_roles));
END $$;

-- Add comment to deprecated field
COMMENT ON COLUMN users.role IS 'DEPRECATED: Use user_roles table instead. This field is kept for backward compatibility during V1->V2 transition.';
