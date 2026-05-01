package rbac

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
