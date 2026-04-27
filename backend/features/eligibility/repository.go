package eligibility

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// RepositoryInterface defines the interface for eligibility data access
type RepositoryInterface interface {
	// GetMatchData returns match details needed for eligibility checking
	GetMatchData(ctx context.Context, matchID int64) (*MatchData, error)

	// GetActiveReferees returns all active referees with their availability for a specific match
	GetActiveReferees(ctx context.Context, matchID int64) ([]RefereeData, error)
}

// Repository handles eligibility data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new eligibility repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetMatchData returns match details needed for eligibility checking
func (r *Repository) GetMatchData(ctx context.Context, matchID int64) (*MatchData, error) {
	var match MatchData
	var matchDate time.Time

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, age_group, match_date
		 FROM matches
		 WHERE id = $1 AND status = 'active'`,
		matchID,
	).Scan(&match.ID, &match.AgeGroup, &matchDate)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get match data: %w", err)
	}

	match.MatchDate = matchDate.Format("2006-01-02")

	return &match, nil
}

// GetActiveReferees returns all active referees with their availability for a specific match
func (r *Repository) GetActiveReferees(ctx context.Context, matchID int64) ([]RefereeData, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			u.id, u.first_name, u.last_name, u.email, u.grade,
			u.date_of_birth, u.certified, u.cert_expiry,
			COALESCE(a.available, false) as is_available
		 FROM users u
		 LEFT JOIN availability a ON a.referee_id = u.id AND a.match_id = $1
		 WHERE (u.role = 'referee' OR u.role = 'assignor')
		   AND u.status = 'active'
		   AND u.first_name IS NOT NULL
		   AND u.last_name IS NOT NULL
		   AND u.date_of_birth IS NOT NULL
		 ORDER BY
			CASE WHEN a.available = true THEN 0 ELSE 1 END,
			u.last_name, u.first_name`,
		matchID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get active referees: %w", err)
	}
	defer rows.Close()

	var referees []RefereeData

	for rows.Next() {
		var ref RefereeData
		var dob, certExpiry sql.NullTime
		var grade sql.NullString

		err := rows.Scan(
			&ref.ID, &ref.FirstName, &ref.LastName, &ref.Email, &grade,
			&dob, &ref.Certified, &certExpiry, &ref.IsAvailable,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan referee: %w", err)
		}

		// Convert nullable fields
		if grade.Valid {
			ref.Grade = &grade.String
		}

		if dob.Valid {
			dobStr := dob.Time.Format("2006-01-02")
			ref.DateOfBirth = &dobStr
		}

		if certExpiry.Valid {
			certExpiryStr := certExpiry.Time.Format("2006-01-02")
			ref.CertExpiry = &certExpiryStr
		}

		referees = append(referees, ref)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating referees: %w", err)
	}

	return referees, nil
}
