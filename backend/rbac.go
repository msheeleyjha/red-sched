package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

// Permission represents a system permission
type Permission struct {
	ID          int64
	Name        string
	DisplayName string
	Description string
	Resource    string
	Action      string
}

// Role represents a system role
type Role struct {
	ID          int64
	Name        string
	Description string
}

// UserPermissions caches a user's permissions for the current request
type UserPermissions struct {
	UserID       int64
	Roles        []Role
	Permissions  []Permission
	IsSuperAdmin bool
}

const userPermissionsKey contextKey = "userPermissions"

// getUserPermissions retrieves all permissions for a user from the database
// Implements "most permissive wins" - user has permission if ANY of their roles includes it
func getUserPermissions(userID int64) (*UserPermissions, error) {
	up := &UserPermissions{
		UserID: userID,
		Roles:  []Role{},
		Permissions: []Permission{},
	}

	// Get user's roles
	roleQuery := `
		SELECT r.id, r.name, r.description
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`
	rows, err := db.Query(roleQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		up.Roles = append(up.Roles, role)

		// Check if user is Super Admin
		if role.Name == "Super Admin" {
			up.IsSuperAdmin = true
		}
	}

	// Get all unique permissions from all user's roles (union of permissions)
	permQuery := `
		SELECT DISTINCT p.id, p.name, p.display_name, p.description, p.resource, p.action
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
	`
	rows, err = db.Query(permQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var perm Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.DisplayName, &perm.Description, &perm.Resource, &perm.Action); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		up.Permissions = append(up.Permissions, perm)
	}

	return up, nil
}

// hasPermission checks if user has a specific permission
func (up *UserPermissions) hasPermission(permissionName string) bool {
	// Super Admin auto-passes all permission checks
	if up.IsSuperAdmin {
		return true
	}

	// Check if user has the specific permission
	for _, perm := range up.Permissions {
		if perm.Name == permissionName {
			return true
		}
	}

	return false
}

// requirePermission is middleware that enforces permission-based authorization
// Usage: router.HandleFunc("/api/admin/users", requirePermission("can_manage_users", handlerFunc))
func requirePermission(permissionName string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from session
		session, err := sessionStore.Get(r, "auth-session")
		if err != nil {
			log.Printf("Session error: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := session.Values["user_id"].(int64)
		if !ok || userID == 0 {
			http.Error(w, "Unauthorized - not authenticated", http.StatusUnauthorized)
			return
		}

		// Get user permissions
		userPerms, err := getUserPermissions(userID)
		if err != nil {
			log.Printf("Failed to get user permissions: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Check if user has required permission
		if !userPerms.hasPermission(permissionName) {
			log.Printf("User %d denied access - missing permission: %s", userID, permissionName)
			http.Error(w, "Forbidden - insufficient permissions", http.StatusForbidden)
			return
		}

		// Store user permissions in context for handler to use if needed
		ctx := context.WithValue(r.Context(), userPermissionsKey, userPerms)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// requireAuth is middleware that only checks if user is authenticated (no permission check)
// Used for endpoints like profile edit that all authenticated users can access
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionStore.Get(r, "auth-session")
		if err != nil {
			log.Printf("Session error: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := session.Values["user_id"].(int64)
		if !ok || userID == 0 {
			http.Error(w, "Unauthorized - not authenticated", http.StatusUnauthorized)
			return
		}

		// For authenticated users with no roles, they can only edit their own profile
		// This is handled by individual route logic
		next.ServeHTTP(w, r)
	}
}

// getUserPermissionsFromContext retrieves user permissions from request context
func getUserPermissionsFromContext(ctx context.Context) (*UserPermissions, bool) {
	userPerms, ok := ctx.Value(userPermissionsKey).(*UserPermissions)
	return userPerms, ok
}

// getCurrentUserID gets the current user ID from session
func getCurrentUserID(r *http.Request) (int64, error) {
	session, err := sessionStore.Get(r, "auth-session")
	if err != nil {
		return 0, fmt.Errorf("session error: %w", err)
	}

	userID, ok := session.Values["user_id"].(int64)
	if !ok || userID == 0 {
		return 0, fmt.Errorf("user not authenticated")
	}

	return userID, nil
}
