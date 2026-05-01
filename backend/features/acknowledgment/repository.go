package acknowledgment

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// RepositoryInterface defines the interface for acknowledgment data access
type RepositoryInterface interface {
	// GetRefereeAssignmentRole checks if a referee is assigned to a match and returns the role type
	GetRefereeAssignmentRole(ctx context.Context, matchID int64, refereeID int64) (*string, error)

	// AcknowledgeAssignment marks an assignment as acknowledged
	AcknowledgeAssignment(ctx context.Context, matchID int64, refereeID int64, acknowledgedAt time.Time) error
}

// Repository handles acknowledgment data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new acknowledgment repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetRefereeAssignmentRole checks if a referee is assigned to a match and returns the role type
func (r *Repository) GetRefereeAssignmentRole(ctx context.Context, matchID int64, refereeID int64) (*string, error) {
	var roleType string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT position
		 FROM assignments
		 WHERE match_id = $1 AND referee_id = $2`,
		matchID, refereeID,
	).Scan(&roleType)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get referee assignment: %w", err)
	}

	return &roleType, nil
}

// AcknowledgeAssignment marks an assignment as acknowledged
func (r *Repository) AcknowledgeAssignment(ctx context.Context, matchID int64, refereeID int64, acknowledgedAt time.Time) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE assignments
		 SET acknowledged = true, acknowledged_at = $1
		 WHERE match_id = $2 AND referee_id = $3`,
		acknowledgedAt, matchID, refereeID,
	)

	if err != nil {
		return fmt.Errorf("failed to acknowledge assignment: %w", err)
	}
	return nil
}
