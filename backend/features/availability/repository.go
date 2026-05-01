package availability

import (
	"context"
	"database/sql"
	"fmt"
)

// RepositoryInterface defines the interface for availability data access
type RepositoryInterface interface {
	// ToggleMatchAvailability sets or clears a referee's availability for a match
	// If available is nil, the record is deleted (no preference)
	// If available is non-nil, the record is inserted/updated
	ToggleMatchAvailability(ctx context.Context, matchID int64, refereeID int64, available *bool) error

	// MatchExistsAndActive checks if a match exists and is active/upcoming
	MatchExistsAndActive(ctx context.Context, matchID int64) (bool, error)

	// GetDayUnavailability returns all days marked as unavailable for a referee
	GetDayUnavailability(ctx context.Context, refereeID int64) ([]DayUnavailabilityData, error)

	// ToggleDayUnavailability sets or clears a referee's unavailability for a day
	// If unavailable is true, marks the day as unavailable
	// If unavailable is false, removes the day unavailability
	ToggleDayUnavailability(ctx context.Context, refereeID int64, date string, unavailable bool, reason *string) error

	// ClearMatchAvailabilityForDay removes all match availability records for a specific day
	ClearMatchAvailabilityForDay(ctx context.Context, refereeID int64, date string) error
}

// Repository handles availability data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new availability repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ToggleMatchAvailability sets or clears a referee's availability for a match
func (r *Repository) ToggleMatchAvailability(ctx context.Context, matchID int64, refereeID int64, available *bool) error {
	if available == nil {
		// Clear preference: delete the availability record
		_, err := r.db.ExecContext(
			ctx,
			`DELETE FROM availability
			 WHERE match_id = $1 AND referee_id = $2`,
			matchID, refereeID,
		)
		if err != nil {
			return fmt.Errorf("failed to clear match availability: %w", err)
		}
	} else {
		// Insert or update availability record
		_, err := r.db.ExecContext(
			ctx,
			`INSERT INTO availability (match_id, referee_id, available, created_at)
			 VALUES ($1, $2, $3, NOW())
			 ON CONFLICT (match_id, referee_id)
			 DO UPDATE SET available = $3, created_at = NOW()`,
			matchID, refereeID, *available,
		)
		if err != nil {
			return fmt.Errorf("failed to set match availability: %w", err)
		}
	}
	return nil
}

// MatchExistsAndActive checks if a match exists and is active/upcoming
func (r *Repository) MatchExistsAndActive(ctx context.Context, matchID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(
			SELECT 1 FROM matches
			WHERE id = $1 AND status = 'active' AND match_date >= CURRENT_DATE
		)`,
		matchID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check match existence: %w", err)
	}

	return exists, nil
}

// GetDayUnavailability returns all days marked as unavailable for a referee
func (r *Repository) GetDayUnavailability(ctx context.Context, refereeID int64) ([]DayUnavailabilityData, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, referee_id, unavailable_date, reason, created_at
		 FROM day_unavailability
		 WHERE referee_id = $1
		 ORDER BY unavailable_date`,
		refereeID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get day unavailability: %w", err)
	}
	defer rows.Close()

	var unavailableDays []DayUnavailabilityData

	for rows.Next() {
		var day DayUnavailabilityData
		var reason sql.NullString

		err := rows.Scan(&day.ID, &day.RefereeID, &day.UnavailableDate, &reason, &day.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan day unavailability: %w", err)
		}

		if reason.Valid {
			day.Reason = &reason.String
		}

		unavailableDays = append(unavailableDays, day)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating day unavailability: %w", err)
	}

	return unavailableDays, nil
}

// ToggleDayUnavailability sets or clears a referee's unavailability for a day
func (r *Repository) ToggleDayUnavailability(ctx context.Context, refereeID int64, date string, unavailable bool, reason *string) error {
	if unavailable {
		// Mark day as unavailable
		_, err := r.db.ExecContext(
			ctx,
			`INSERT INTO day_unavailability (referee_id, unavailable_date, reason, created_at)
			 VALUES ($1, $2, $3, NOW())
			 ON CONFLICT (referee_id, unavailable_date)
			 DO UPDATE SET reason = $3`,
			refereeID, date, reason,
		)
		if err != nil {
			return fmt.Errorf("failed to mark day as unavailable: %w", err)
		}
	} else {
		// Remove day unavailability
		_, err := r.db.ExecContext(
			ctx,
			`DELETE FROM day_unavailability
			 WHERE referee_id = $1 AND unavailable_date = $2`,
			refereeID, date,
		)
		if err != nil {
			return fmt.Errorf("failed to remove day unavailability: %w", err)
		}
	}

	return nil
}

// ClearMatchAvailabilityForDay removes all match availability records for a specific day
func (r *Repository) ClearMatchAvailabilityForDay(ctx context.Context, refereeID int64, date string) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM availability
		 WHERE referee_id = $1
		   AND match_id IN (
			 SELECT id FROM matches WHERE match_date = $2
		   )`,
		refereeID, date,
	)
	if err != nil {
		return fmt.Errorf("failed to clear match availability for day: %w", err)
	}

	return nil
}
