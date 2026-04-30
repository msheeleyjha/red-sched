-- Remove match permission assignments
DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE name IN ('can_view_matches', 'can_manage_matches'));

-- Remove match permissions
DELETE FROM permissions WHERE name IN ('can_view_matches', 'can_manage_matches');
