-- Add match viewing and management permissions
INSERT INTO permissions (name, display_name, description, resource, action) VALUES
    ('can_view_matches', 'View Matches', 'View match schedules and details', 'matches', 'view'),
    ('can_manage_matches', 'Manage Matches', 'Import, edit, archive, and manage match schedules', 'matches', 'manage')
ON CONFLICT (name) DO NOTHING;

-- Assignor gets both view and manage
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'Assignor'
AND p.name IN ('can_view_matches', 'can_manage_matches')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Referee gets view only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'Referee'
AND p.name = 'can_view_matches'
ON CONFLICT (role_id, permission_id) DO NOTHING;
