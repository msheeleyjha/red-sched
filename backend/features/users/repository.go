package users

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository handles user data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new user repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindByGoogleID retrieves a user by their Google ID
func (r *Repository) FindByGoogleID(ctx context.Context, googleID string) (*User, error) {
	query := `
		SELECT id, google_id, email, name, role, status, first_name, last_name,
		       date_of_birth, certified, cert_expiry, grade, created_at, updated_at
		FROM users
		WHERE google_id = $1 AND status != 'removed'
	`

	user := &User{}
	err := r.db.QueryRowContext(ctx, query, googleID).Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.Status,
		&user.FirstName,
		&user.LastName,
		&user.DateOfBirth,
		&user.Certified,
		&user.CertExpiry,
		&user.Grade,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user by google id: %w", err)
	}

	return user, nil
}

// FindByID retrieves a user by their ID
func (r *Repository) FindByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, google_id, email, name, role, status, first_name, last_name,
		       date_of_birth, certified, cert_expiry, grade, created_at, updated_at
		FROM users
		WHERE id = $1 AND status IN ('active', 'pending')
	`

	user := &User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.Status,
		&user.FirstName,
		&user.LastName,
		&user.DateOfBirth,
		&user.Certified,
		&user.CertExpiry,
		&user.Grade,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}

	return user, nil
}

// Create creates a new user
func (r *Repository) Create(ctx context.Context, googleID, email, name string) (*User, error) {
	query := `
		INSERT INTO users (google_id, email, name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, google_id, email, name, role, status, created_at, updated_at
	`

	user := &User{}
	err := r.db.QueryRowContext(ctx, query, googleID, email, name, "pending_referee", "pending").Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// UpdateProfile updates a user's profile information
func (r *Repository) UpdateProfile(ctx context.Context, userID int64, data ProfileUpdateData) (*User, error) {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, date_of_birth = $3,
		    certified = $4, cert_expiry = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING id, google_id, email, name, role, status, first_name, last_name,
		          date_of_birth, certified, cert_expiry, grade, created_at, updated_at
	`

	user := &User{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		data.FirstName,
		data.LastName,
		data.DOB,
		data.Certified,
		data.CertExpiry,
		userID,
	).Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.Status,
		&user.FirstName,
		&user.LastName,
		&user.DateOfBirth,
		&user.Certified,
		&user.CertExpiry,
		&user.Grade,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return user, nil
}
