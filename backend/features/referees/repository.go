package referees

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// RepositoryInterface defines the interface for referee data access
type RepositoryInterface interface {
	// List returns all referees for assignor management
	List(ctx context.Context) ([]RefereeListItem, error)

	// FindByID returns a referee by ID
	FindByID(ctx context.Context, id int64) (*RefereeData, error)

	// Update updates a referee with the provided fields
	Update(ctx context.Context, id int64, updates map[string]interface{}) (*UpdateResult, error)

	// HasUpcomingAssignments checks if a referee has upcoming match assignments
	HasUpcomingAssignments(ctx context.Context, refereeID int64) (bool, error)
}

// Repository handles referee data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new referee repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List returns all referees for assignor management
func (r *Repository) List(ctx context.Context) ([]RefereeListItem, error) {
	query := `
		SELECT id, email, name, first_name, last_name, date_of_birth,
		       certified, cert_expiry, role, status, grade, created_at
		FROM users
		WHERE role IN ('pending_referee', 'referee', 'assignor') AND status != 'removed'
		ORDER BY
		  CASE
		    WHEN role = 'assignor' THEN 0
		    WHEN status = 'pending' THEN 1
		    WHEN status = 'active' THEN 2
		    WHEN status = 'inactive' THEN 3
		  END,
		  created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list referees: %w", err)
	}
	defer rows.Close()

	referees := []RefereeListItem{}
	now := time.Now()

	for rows.Next() {
		var ref RefereeListItem
		err := rows.Scan(
			&ref.ID,
			&ref.Email,
			&ref.Name,
			&ref.FirstName,
			&ref.LastName,
			&ref.DateOfBirth,
			&ref.Certified,
			&ref.CertExpiry,
			&ref.Role,
			&ref.Status,
			&ref.Grade,
			&ref.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan referee: %w", err)
		}

		// Determine certification status
		ref.CertStatus = DetermineCertStatus(ref.Certified, ref.CertExpiry, now)

		referees = append(referees, ref)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating referees: %w", err)
	}

	return referees, nil
}

// FindByID returns a referee by ID
func (r *Repository) FindByID(ctx context.Context, id int64) (*RefereeData, error) {
	var ref RefereeData
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, email, name, role, status FROM users WHERE id = $1`,
		id,
	).Scan(&ref.ID, &ref.Email, &ref.Name, &ref.Role, &ref.Status)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find referee: %w", err)
	}

	return &ref, nil
}

// Update updates a referee with the provided fields
func (r *Repository) Update(ctx context.Context, id int64, updates map[string]interface{}) (*UpdateResult, error) {
	if len(updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}

	setClauses := []string{}
	args := []interface{}{}
	argCount := 1

	for field, value := range updates {
		if value == nil {
			// Handle NULL values
			setClauses = append(setClauses, fmt.Sprintf("%s = NULL", field))
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argCount))
			args = append(args, value)
			argCount++
		}
	}

	// Always update updated_at
	setClauses = append(setClauses, "updated_at = NOW()")

	// Add WHERE clause
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE users SET %s WHERE id = $%d
		 RETURNING id, email, name, first_name, last_name, date_of_birth,
		           certified, cert_expiry, role, status, grade, created_at, updated_at`,
		strings.Join(setClauses, ", "),
		argCount,
	)

	var result UpdateResult
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&result.ID,
		&result.Email,
		&result.Name,
		&result.FirstName,
		&result.LastName,
		&result.DateOfBirth,
		&result.Certified,
		&result.CertExpiry,
		&result.Role,
		&result.Status,
		&result.Grade,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("referee not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update referee: %w", err)
	}

	return &result, nil
}

// HasUpcomingAssignments checks if a referee has upcoming match assignments
func (r *Repository) HasUpcomingAssignments(ctx context.Context, refereeID int64) (bool, error) {
	var hasAssignments bool
	err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(
			SELECT 1 FROM assignments mr
			JOIN matches m ON mr.match_id = m.id
			WHERE mr.referee_id = $1
			  AND m.match_date >= CURRENT_DATE
			  AND m.status = 'active'
		)`,
		refereeID,
	).Scan(&hasAssignments)

	if err != nil {
		return false, fmt.Errorf("failed to check for upcoming assignments: %w", err)
	}

	return hasAssignments, nil
}

// DetermineCertStatus determines the certification status based on certification and expiry
func DetermineCertStatus(certified bool, certExpiry *time.Time, now time.Time) string {
	if !certified {
		return "none"
	}
	if certExpiry == nil {
		return "none"
	}
	if certExpiry.Before(now) {
		return "expired"
	}
	// Check if expiring within 30 days
	thirtyDaysFromNow := now.AddDate(0, 0, 30)
	if certExpiry.Before(thirtyDaysFromNow) {
		return "expiring_soon"
	}
	return "valid"
}
