package users

import "time"

// User represents a user in the system
type User struct {
	ID          int64      `json:"id"`
	GoogleID    string     `json:"google_id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Role        string     `json:"role"`        // Legacy field, kept for backward compatibility
	Status      string     `json:"status"`      // active, pending, removed
	FirstName   *string    `json:"first_name,omitempty"`
	LastName    *string    `json:"last_name,omitempty"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Certified   bool       `json:"certified"`
	CertExpiry  *time.Time `json:"cert_expiry,omitempty"`
	Grade       *string    `json:"grade,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ProfileUpdateRequest represents the profile update payload
type ProfileUpdateRequest struct {
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	DateOfBirth *string `json:"date_of_birth"` // YYYY-MM-DD format
	Certified   bool    `json:"certified"`
	CertExpiry  *string `json:"cert_expiry"` // YYYY-MM-DD format
}

// ProfileUpdateData contains parsed and validated profile update data
type ProfileUpdateData struct {
	FirstName  string
	LastName   string
	DOB        *time.Time
	Certified  bool
	CertExpiry *time.Time
}
