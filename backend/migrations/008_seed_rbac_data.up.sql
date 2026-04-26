-- Epic 1 Story 1.2: Seed Initial Roles and Permissions

-- Insert initial roles
INSERT INTO roles (name, description) VALUES
    ('Super Admin', 'System administrator with full access to all features including audit logs and role management'),
    ('Assignor', 'League administrator who can import schedules, assign referees, and manage users'),
    ('Referee', 'Official who can request assignments, submit match reports, and view their match history')
ON CONFLICT (name) DO NOTHING;

-- Insert initial permissions with technical names and display names
INSERT INTO permissions (name, display_name, description, resource, action) VALUES
    -- User Management
    ('can_manage_users', 'Manage Users', 'Approve new users and manage user profiles', 'users', 'manage'),

    -- Match Management
    ('can_import_matches', 'Import Match Schedule', 'Import matches from CSV files', 'matches', 'import'),
    ('can_assign_referees', 'Assign Referees to Matches', 'Assign referees to match positions', 'assignments', 'assign'),

    -- Referee Actions
    ('can_request_assignments', 'Request Match Assignments', 'Mark availability and request to work matches', 'assignments', 'request'),
    ('can_edit_own_match_reports', 'Edit Own Match Reports', 'Submit and edit match reports for assigned matches', 'match_reports', 'edit'),

    -- System Administration
    ('can_view_audit_logs', 'View Audit Logs', 'View system audit logs and user activity', 'audit_logs', 'view'),
    ('can_assign_roles', 'Assign User Roles', 'Assign and revoke roles for users', 'roles', 'assign')
ON CONFLICT (name) DO NOTHING;

-- Assign permissions to roles
-- Super Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'Super Admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assignor gets user management, match import, and assignment permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'Assignor'
AND p.name IN (
    'can_manage_users',
    'can_import_matches',
    'can_assign_referees'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Referee gets assignment request and match report permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'Referee'
AND p.name IN (
    'can_request_assignments',
    'can_edit_own_match_reports'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;
