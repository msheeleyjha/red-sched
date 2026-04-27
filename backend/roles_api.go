package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// RoleResponse represents a role in API responses
type RoleResponse struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions,omitempty"`
}

// PermissionResponse represents a permission in API responses
type PermissionResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// AssignRoleRequest represents the request body for assigning a role
type AssignRoleRequest struct {
	RoleID int64 `json:"role_id"`
}

// UserRolesResponse represents a user's roles
type UserRolesResponse struct {
	UserID int64          `json:"user_id"`
	Roles  []RoleResponse `json:"roles"`
}

// assignRoleToUser assigns a role to a user
// POST /api/admin/users/:id/roles
func assignRoleToUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	targetUserID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current user ID for audit trail
	currentUserID, err := getCurrentUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify target user exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", targetUserID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Verify role exists
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)", req.RoleID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "Role not found", http.StatusNotFound)
		return
	}

	// Check if user is trying to assign Super Admin role to themselves
	// This is allowed (no special restriction)

	// Check if assignment already exists
	err = db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)",
		targetUserID, req.RoleID,
	).Scan(&exists)
	if err != nil {
		log.Printf("Error checking existing role assignment: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "User already has this role", http.StatusConflict)
		return
	}

	// Assign role
	_, err = db.Exec(
		"INSERT INTO user_roles (user_id, role_id, assigned_by) VALUES ($1, $2, $3)",
		targetUserID, req.RoleID, currentUserID,
	)
	if err != nil {
		log.Printf("Error assigning role: %v", err)
		http.Error(w, "Failed to assign role", http.StatusInternalServerError)
		return
	}

	// Create audit log entry
	auditLogger.LogWithContext(r, AuditActionCreate, "user_role", targetUserID, nil, map[string]interface{}{
		"user_id":     targetUserID,
		"role_id":     req.RoleID,
		"assigned_by": currentUserID,
	})

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Role assigned successfully",
	})
}

// revokeRoleFromUser removes a role from a user
// DELETE /api/admin/users/:id/roles/:roleId
func revokeRoleFromUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	roleIDStr := vars["roleId"]

	targetUserID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	// Get current user ID for safety check
	currentUserID, err := getCurrentUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if role is Super Admin
	var roleName string
	err = db.QueryRow("SELECT name FROM roles WHERE id = $1", roleID).Scan(&roleName)
	if err == sql.ErrNoRows {
		http.Error(w, "Role not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error querying role: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Prevent user from revoking their own Super Admin role (prevent lockout)
	if roleName == "Super Admin" && currentUserID == targetUserID {
		http.Error(w, "Cannot revoke your own Super Admin role", http.StatusForbidden)
		return
	}

	// Revoke role
	result, err := db.Exec(
		"DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2",
		targetUserID, roleID,
	)
	if err != nil {
		log.Printf("Error revoking role: %v", err)
		http.Error(w, "Failed to revoke role", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "User does not have this role", http.StatusNotFound)
		return
	}

	// Create audit log entry
	auditLogger.LogWithContext(r, AuditActionDelete, "user_role", targetUserID, map[string]interface{}{
		"user_id": targetUserID,
		"role_id": roleID,
	}, nil)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Role revoked successfully",
	})
}

// getUserRoles returns all roles for a user
// GET /api/admin/users/:id/roles
func getUserRoles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Verify user exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get user's roles
	rows, err := db.Query(`
		SELECT r.id, r.name, r.description
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`, userID)
	if err != nil {
		log.Printf("Error querying user roles: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	roles := []RoleResponse{}
	for rows.Next() {
		var role RoleResponse
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			log.Printf("Error scanning role: %v", err)
			continue
		}
		roles = append(roles, role)
	}

	response := UserRolesResponse{
		UserID: userID,
		Roles:  roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getAllRoles returns all available roles
// GET /api/admin/roles
func getAllRoles(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT r.id, r.name, r.description
		FROM roles r
		ORDER BY r.name
	`)
	if err != nil {
		log.Printf("Error querying roles: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	roles := []RoleResponse{}
	for rows.Next() {
		var role RoleResponse
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			log.Printf("Error scanning role: %v", err)
			continue
		}

		// Get permissions for this role
		permRows, err := db.Query(`
			SELECT p.name
			FROM permissions p
			INNER JOIN role_permissions rp ON p.id = rp.permission_id
			WHERE rp.role_id = $1
			ORDER BY p.name
		`, role.ID)
		if err != nil {
			log.Printf("Error querying role permissions: %v", err)
			continue
		}

		permissions := []string{}
		for permRows.Next() {
			var permName string
			if err := permRows.Scan(&permName); err != nil {
				log.Printf("Error scanning permission: %v", err)
				continue
			}
			permissions = append(permissions, permName)
		}
		permRows.Close()

		role.Permissions = permissions
		roles = append(roles, role)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

// getAllPermissions returns all available permissions with display names
// GET /api/admin/permissions
func getAllPermissions(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT id, name, display_name, description, resource, action
		FROM permissions
		ORDER BY resource, action
	`)
	if err != nil {
		log.Printf("Error querying permissions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	permissions := []PermissionResponse{}
	for rows.Next() {
		var perm PermissionResponse
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.DisplayName, &perm.Description, &perm.Resource, &perm.Action); err != nil {
			log.Printf("Error scanning permission: %v", err)
			continue
		}
		permissions = append(permissions, perm)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}
