package assignments

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// RepositoryInterface defines the interface for assignment data access
type RepositoryInterface interface {
	// Match queries
	MatchExists(ctx context.Context, matchID int64) (bool, error)
	GetMatchTimeWindow(ctx context.Context, matchID int64) (*MatchTimeWindow, error)

	// Role slot queries
	GetRoleSlot(ctx context.Context, matchID int64, roleType string) (*RoleSlot, error)
	UpdateRoleAssignment(ctx context.Context, roleID int64, refereeID *int64) error

	// Referee queries
	RefereeExists(ctx context.Context, refereeID int64) (bool, error)
	GetRefereeExistingRoleOnMatch(ctx context.Context, matchID int64, refereeID int64, excludeRoleType string) (*string, error)

	// Conflict detection
	FindConflictingAssignments(ctx context.Context, refereeID int64, matchID int64, startTime time.Time, endTime time.Time) ([]ConflictMatch, error)

	// Assignment history
	LogAssignment(ctx context.Context, history *AssignmentHistory) error

	// Referee history
	GetRefereeMatchHistory(ctx context.Context, refereeID int64) ([]RefereeHistoryMatch, error)

	// Assignment change tracking (Story 5.6)
	MarkAssignmentAsViewed(ctx context.Context, matchID int64, refereeID int64) error
	ResetViewedStatusForMatch(ctx context.Context, matchID int64) error
}

// Repository handles assignment data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new assignment repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// MatchExists checks if a match exists, is active, and not archived
func (r *Repository) MatchExists(ctx context.Context, matchID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM matches WHERE id = $1 AND status = 'active' AND archived = FALSE)",
		matchID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check match existence: %w", err)
	}
	return exists, nil
}

// GetMatchTimeWindow retrieves the start and end time of a match
func (r *Repository) GetMatchTimeWindow(ctx context.Context, matchID int64) (*MatchTimeWindow, error) {
	var start, end time.Time
	err := r.db.QueryRowContext(
		ctx,
		`SELECT match_date + start_time::interval, match_date + end_time::interval
		 FROM matches WHERE id = $1`,
		matchID,
	).Scan(&start, &end)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get match time window: %w", err)
	}

	return &MatchTimeWindow{
		MatchID: matchID,
		Start:   start,
		End:     end,
	}, nil
}

// GetRoleSlot retrieves a role slot for a match
func (r *Repository) GetRoleSlot(ctx context.Context, matchID int64, roleType string) (*RoleSlot, error) {
	slot := &RoleSlot{}
	var acknowledgedAt sql.NullTime

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, match_id, role_type, assigned_referee_id, acknowledged, acknowledged_at, updated_at, viewed_by_referee, created_at
		 FROM match_roles
		 WHERE match_id = $1 AND role_type = $2`,
		matchID, roleType,
	).Scan(
		&slot.ID,
		&slot.MatchID,
		&slot.RoleType,
		&slot.AssignedRefereeID,
		&slot.Acknowledged,
		&acknowledgedAt,
		&slot.UpdatedAt,
		&slot.ViewedByReferee,
		&slot.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role slot: %w", err)
	}

	if acknowledgedAt.Valid {
		slot.AcknowledgedAt = &acknowledgedAt.Time
	}

	return slot, nil
}

// UpdateRoleAssignment updates the referee assignment for a role slot
// Also clears acknowledgment when assigning/reassigning/removing
func (r *Repository) UpdateRoleAssignment(ctx context.Context, roleID int64, refereeID *int64) error {
	var nullableRefereeID sql.NullInt64
	if refereeID != nil {
		nullableRefereeID = sql.NullInt64{Int64: *refereeID, Valid: true}
	}

	_, err := r.db.ExecContext(
		ctx,
		`UPDATE match_roles
		 SET assigned_referee_id = $1,
		     acknowledged = false,
		     acknowledged_at = NULL,
		     updated_at = NOW(),
		     viewed_by_referee = false
		 WHERE id = $2`,
		nullableRefereeID, roleID,
	)

	if err != nil {
		return fmt.Errorf("failed to update role assignment: %w", err)
	}
	return nil
}

// RefereeExists checks if a referee exists and is active
func (r *Repository) RefereeExists(ctx context.Context, refereeID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(
			SELECT 1 FROM users
			WHERE id = $1
			  AND (role = 'referee' OR role = 'assignor')
			  AND status = 'active'
		)`,
		refereeID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check referee existence: %w", err)
	}
	return exists, nil
}

// GetRefereeExistingRoleOnMatch checks if a referee is already assigned to another role on the same match
// Returns the role type if found, nil otherwise
func (r *Repository) GetRefereeExistingRoleOnMatch(ctx context.Context, matchID int64, refereeID int64, excludeRoleType string) (*string, error) {
	var roleType sql.NullString
	err := r.db.QueryRowContext(
		ctx,
		`SELECT role_type
		 FROM match_roles
		 WHERE match_id = $1
		   AND assigned_referee_id = $2
		   AND role_type != $3`,
		matchID, refereeID, excludeRoleType,
	).Scan(&roleType)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check existing role: %w", err)
	}

	if !roleType.Valid {
		return nil, nil
	}

	result := roleType.String
	return &result, nil
}

// FindConflictingAssignments finds all active (non-archived) assignments for a referee that overlap with the given time window
func (r *Repository) FindConflictingAssignments(ctx context.Context, refereeID int64, matchID int64, startTime time.Time, endTime time.Time) ([]ConflictMatch, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT m.id, m.event_name, m.team_name, m.match_date, m.start_time, mr.role_type
		 FROM matches m
		 JOIN match_roles mr ON mr.match_id = m.id
		 WHERE mr.assigned_referee_id = $1
		   AND m.id != $2
		   AND m.status = 'active'
		   AND m.archived = FALSE
		   AND (
			 (m.match_date + m.start_time::interval, m.match_date + m.end_time::interval)
			 OVERLAPS
			 ($3::timestamp, $4::timestamp)
		   )`,
		refereeID, matchID, startTime, endTime,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to find conflicts: %w", err)
	}
	defer rows.Close()

	conflicts := []ConflictMatch{}
	for rows.Next() {
		var c ConflictMatch
		var matchDate time.Time

		err := rows.Scan(
			&c.MatchID,
			&c.EventName,
			&c.TeamName,
			&matchDate,
			&c.StartTime,
			&c.RoleType,
		)
		if err != nil {
			continue
		}

		c.MatchDate = matchDate.Format("2006-01-02")
		conflicts = append(conflicts, c)
	}

	return conflicts, nil
}

// LogAssignment logs an assignment history record
func (r *Repository) LogAssignment(ctx context.Context, history *AssignmentHistory) error {
	var oldRefereeID, newRefereeID sql.NullInt64

	if history.OldRefereeID != nil {
		oldRefereeID = sql.NullInt64{Int64: *history.OldRefereeID, Valid: true}
	}
	if history.NewRefereeID != nil {
		newRefereeID = sql.NullInt64{Int64: *history.NewRefereeID, Valid: true}
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO assignment_history (match_id, role_type, old_referee_id, new_referee_id, action, actor_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		history.MatchID,
		history.RoleType,
		oldRefereeID,
		newRefereeID,
		history.Action,
		history.ActorID,
	)

	if err != nil {
		return fmt.Errorf("failed to log assignment: %w", err)
	}
	return nil
}

// GetRefereeMatchHistory retrieves all matches (active and archived) assigned to a referee
func (r *Repository) GetRefereeMatchHistory(ctx context.Context, refereeID int64) ([]RefereeHistoryMatch, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			m.id, m.event_name, m.team_name, m.age_group, m.match_date,
			m.start_time, m.end_time, m.location, m.status,
			m.archived, m.archived_at,
			mr.role_type, mr.acknowledged, mr.acknowledged_at,
			mr.updated_at, mr.viewed_by_referee
		FROM matches m
		JOIN match_roles mr ON mr.match_id = m.id
		WHERE mr.assigned_referee_id = $1
		  AND m.status != 'deleted'
		ORDER BY m.match_date DESC, m.start_time DESC`,
		refereeID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get referee match history: %w", err)
	}
	defer rows.Close()

	history := []RefereeHistoryMatch{}
	for rows.Next() {
		var h RefereeHistoryMatch
		var ageGroup sql.NullString
		var archivedAt sql.NullTime
		var acknowledgedAt sql.NullTime

		err := rows.Scan(
			&h.MatchID,
			&h.EventName,
			&h.TeamName,
			&ageGroup,
			&h.MatchDate,
			&h.StartTime,
			&h.EndTime,
			&h.Location,
			&h.Status,
			&h.Archived,
			&archivedAt,
			&h.RoleType,
			&h.Acknowledged,
			&acknowledgedAt,
			&h.UpdatedAt,
			&h.ViewedByReferee,
		)
		if err != nil {
			continue
		}

		if ageGroup.Valid {
			h.AgeGroup = &ageGroup.String
		}
		if archivedAt.Valid {
			h.ArchivedAt = &archivedAt.Time
		}
		if acknowledgedAt.Valid {
			h.AcknowledgedAt = &acknowledgedAt.Time
		}

		history = append(history, h)
	}

	return history, nil
}

// MarkAssignmentAsViewed marks a referee's assignment as viewed
// Story 5.6: Called when referee views match detail page
func (r *Repository) MarkAssignmentAsViewed(ctx context.Context, matchID int64, refereeID int64) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE match_roles
		 SET viewed_by_referee = true
		 WHERE match_id = $1 AND assigned_referee_id = $2`,
		matchID, refereeID,
	)

	if err != nil {
		return fmt.Errorf("failed to mark assignment as viewed: %w", err)
	}
	return nil
}

// ResetViewedStatusForMatch resets viewed_by_referee for all assignments on a match
// Story 5.6: Called when match details (time/location) are updated via CSV import
func (r *Repository) ResetViewedStatusForMatch(ctx context.Context, matchID int64) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE match_roles
		 SET viewed_by_referee = false,
		     updated_at = NOW()
		 WHERE match_id = $1 AND assigned_referee_id IS NOT NULL`,
		matchID,
	)

	if err != nil {
		return fmt.Errorf("failed to reset viewed status: %w", err)
	}
	return nil
}
