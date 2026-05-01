-- Epic 1 Story 1.1: RBAC Database Schema
-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on role name for fast lookups
CREATE INDEX idx_roles_name ON roles(name);

-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL, -- Technical name: can_import_matches
    display_name VARCHAR(100) NOT NULL, -- UI-friendly: "Import Match Schedule"
    description TEXT,
    resource VARCHAR(50), -- e.g., "matches", "users", "audit_logs"
    action VARCHAR(50), -- e.g., "import", "assign_roles", "view"
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on permission name for fast lookups
CREATE INDEX idx_permissions_name ON permissions(name);
CREATE INDEX idx_permissions_resource_action ON permissions(resource, action);

-- Create user_roles junction table (multi-role support)
CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by BIGINT REFERENCES users(id) ON DELETE SET NULL, -- System Admin who assigned the role
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id) -- Prevents duplicate assignments
);

-- Create indexes on user_roles for query performance
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- Create role_permissions junction table
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id) -- Prevents duplicate permission assignments
);

-- Create indexes on role_permissions for query performance
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Add comments
COMMENT ON TABLE roles IS 'RBAC roles: Super Admin, Assignor, Referee';
COMMENT ON TABLE permissions IS 'Granular permissions for actions and resources';
COMMENT ON TABLE user_roles IS 'Junction table allowing users to have multiple roles';
COMMENT ON TABLE role_permissions IS 'Junction table defining which permissions each role has';

-- Note: The old users.role field is deprecated in favor of the new user_roles table
-- Migration script will handle assigning users to appropriate roles
