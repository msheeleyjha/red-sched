package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
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

// RBACMiddleware provides role-based access control
type RBACMiddleware struct {
	sessionStore *sessions.CookieStore
	db           *sql.DB
}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware(sessionStore *sessions.CookieStore, db *sql.DB) *RBACMiddleware {
	return &RBACMiddleware{
		sessionStore: sessionStore,
		db:           db,
	}
}

// RequirePermission middleware enforces permission-based authorization
func (rm *RBACMiddleware) RequirePermission(permissionName string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from session
		session, err := rm.sessionStore.Get(r, "auth-session")
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
		userPerms, err := rm.getUserPermissions(userID)
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

// getUserPermissions retrieves all permissions for a user from the database
// Implements "most permissive wins" - user has permission if ANY of their roles includes it
func (rm *RBACMiddleware) getUserPermissions(userID int64) (*UserPermissions, error) {
	up := &UserPermissions{
		UserID:      userID,
		Roles:       []Role{},
		Permissions: []Permission{},
	}

	// Get user's roles
	roleQuery := `
		SELECT r.id, r.name, r.description
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`
	rows, err := rm.db.Query(roleQuery, userID)
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
	rows, err = rm.db.Query(permQuery, userID)
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

// GetUserPermissionsFromContext retrieves user permissions from request context
func GetUserPermissionsFromContext(ctx context.Context) (*UserPermissions, bool) {
	userPerms, ok := ctx.Value(userPermissionsKey).(*UserPermissions)
	return userPerms, ok
}
