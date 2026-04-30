package referees

import "time"

// RefereeListItem represents a referee in the management list
type RefereeListItem struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	FirstName    *string    `json:"first_name"`
	LastName     *string    `json:"last_name"`
	DateOfBirth  *time.Time `json:"date_of_birth"`
	Certified    bool       `json:"certified"`
	CertExpiry   *time.Time `json:"cert_expiry"`
	CertStatus   string     `json:"cert_status"` // valid, expiring_soon, expired, none
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	Grade        *string    `json:"grade"`
	CreatedAt    time.Time  `json:"created_at"`
}

// UpdateRequest represents the update payload from assignor
type UpdateRequest struct {
	Status *string `json:"status"` // active, inactive
	Grade  *string `json:"grade"`  // Junior, Mid, Senior, or null
}

// UpdateResult represents the result of an update operation
type UpdateResult struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	FirstName    *string    `json:"first_name"`
	LastName     *string    `json:"last_name"`
	DateOfBirth  *time.Time `json:"date_of_birth"`
	Certified    bool       `json:"certified"`
	CertExpiry   *time.Time `json:"cert_expiry"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	Grade        *string    `json:"grade"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

// RefereeData represents basic referee information from the repository
type RefereeData struct {
	ID     int64
	Email  string
	Name   string
	Role   string
	Status string
}
