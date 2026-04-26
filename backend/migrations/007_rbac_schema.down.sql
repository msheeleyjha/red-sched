-- Epic 1 Story 1.1: Rollback RBAC Database Schema
-- Drop tables in reverse order of creation (respecting foreign key constraints)

DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
