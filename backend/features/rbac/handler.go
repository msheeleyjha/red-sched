package rbac

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/msheeley/referee-scheduler/shared/audit"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// Handler handles HTTP requests for RBAC admin operations
type Handler struct {
	db          *sql.DB
	auditLogger *audit.Logger
	authMW      *middleware.AuthMiddleware
}

// NewHandler creates a new RBAC handler
func NewHandler(db *sql.DB, auditLogger *audit.Logger, authMW *middleware.AuthMiddleware) *Handler {
	return &Handler{
		db:          db,
		auditLogger: auditLogger,
		authMW:      authMW,
	}
}

// AssignRoleToUser assigns a role to a user
func (h *Handler) AssignRoleToUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	targetUserID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	currentUserID, err := h.authMW.GetCurrentUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", targetUserID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)", req.RoleID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "Role not found", http.StatusNotFound)
		return
	}

	err = h.db.QueryRow(
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

	_, err = h.db.Exec(
		"INSERT INTO user_roles (user_id, role_id, assigned_by) VALUES ($1, $2, $3)",
		targetUserID, req.RoleID, currentUserID,
	)
	if err != nil {
		log.Printf("Error assigning role: %v", err)
		http.Error(w, "Failed to assign role", http.StatusInternalServerError)
		return
	}

	h.auditLogger.LogWithContext(r, audit.ActionCreate, "user_role", targetUserID, nil, map[string]interface{}{
		"user_id":     targetUserID,
		"role_id":     req.RoleID,
		"assigned_by": currentUserID,
	})

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Role assigned successfully",
	})
}

// RevokeRoleFromUser removes a role from a user
func (h *Handler) RevokeRoleFromUser(w http.ResponseWriter, r *http.Request) {
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

	currentUserID, err := h.authMW.GetCurrentUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var roleName string
	err = h.db.QueryRow("SELECT name FROM roles WHERE id = $1", roleID).Scan(&roleName)
	if err == sql.ErrNoRows {
		http.Error(w, "Role not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error querying role: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if roleName == "Super Admin" && currentUserID == targetUserID {
		http.Error(w, "Cannot revoke your own Super Admin role", http.StatusForbidden)
		return
	}

	result, err := h.db.Exec(
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

	h.auditLogger.LogWithContext(r, audit.ActionDelete, "user_role", targetUserID, map[string]interface{}{
		"user_id": targetUserID,
		"role_id": roleID,
	}, nil)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Role revoked successfully",
	})
}

// GetUserRoles returns all roles for a user
func (h *Handler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	rows, err := h.db.Query(`
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

// GetAllRoles returns all available roles
func (h *Handler) GetAllRoles(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
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

		permRows, err := h.db.Query(`
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

// GetAllPermissions returns all available permissions
func (h *Handler) GetAllPermissions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
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
