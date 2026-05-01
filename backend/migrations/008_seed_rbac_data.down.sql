-- Epic 1 Story 1.2: Rollback Seed RBAC Data

-- Delete role-permission assignments
DELETE FROM role_permissions;

-- Delete permissions
DELETE FROM permissions WHERE name IN (
    'can_manage_users',
    'can_import_matches',
    'can_assign_referees',
    'can_request_assignments',
    'can_edit_own_match_reports',
    'can_view_audit_logs',
    'can_assign_roles'
);

-- Delete roles
DELETE FROM roles WHERE name IN (
    'Super Admin',
    'Assignor',
    'Referee'
);
