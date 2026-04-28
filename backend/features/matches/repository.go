package matches

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// RepositoryInterface defines the interface for match data access
type RepositoryInterface interface {
	// Match CRUD operations
	Create(ctx context.Context, match *Match) (*Match, error)
	FindByID(ctx context.Context, id int64) (*Match, error)
	FindByReferenceID(ctx context.Context, referenceID string) (*Match, error) // Story 6.2
	List(ctx context.Context) ([]Match, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) (*Match, error)

	// Role operations
	CreateRole(ctx context.Context, matchID int64, roleType string) error
	GetRoles(ctx context.Context, matchID int64) ([]MatchRole, error)
	DeleteRoles(ctx context.Context, matchID int64, roleTypes []string) error
	RoleExists(ctx context.Context, matchID int64, roleType string) (bool, error)
	GetCurrentRoles(ctx context.Context, matchID int64) ([]string, error)

	// Match queries
	MatchExists(ctx context.Context, matchID int64) (bool, error)
	ListActive(ctx context.Context) ([]Match, error)
	ListArchived(ctx context.Context) ([]Match, error)

	// Archival operations
	Archive(ctx context.Context, matchID int64, archivedBy int64) error
	Unarchive(ctx context.Context, matchID int64) error

	// History logging
	LogEdit(ctx context.Context, matchID int64, actorID int64, changeDescription string) error
}

// Repository handles match data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new match repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new match
func (r *Repository) Create(ctx context.Context, match *Match) (*Match, error) {
	query := `
		INSERT INTO matches (event_name, team_name, age_group, match_date, start_time, end_time,
		                     location, description, reference_id, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		match.EventName,
		match.TeamName,
		match.AgeGroup,
		match.MatchDate,
		match.StartTime,
		match.EndTime,
		match.Location,
		match.Description,
		match.ReferenceID,
		match.Status,
		match.CreatedBy,
	).Scan(&match.ID, &match.CreatedAt, &match.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create match: %w", err)
	}

	return match, nil
}

// FindByID retrieves a match by ID
func (r *Repository) FindByID(ctx context.Context, id int64) (*Match, error) {
	query := `
		SELECT id, event_name, team_name, age_group, match_date, start_time, end_time,
		       location, description, reference_id, status, archived, archived_at, archived_by,
		       created_by, created_at, updated_at
		FROM matches
		WHERE id = $1 AND status != 'deleted'
	`

	match := &Match{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&match.ID,
		&match.EventName,
		&match.TeamName,
		&match.AgeGroup,
		&match.MatchDate,
		&match.StartTime,
		&match.EndTime,
		&match.Location,
		&match.Description,
		&match.ReferenceID,
		&match.Status,
		&match.Archived,
		&match.ArchivedAt,
		&match.ArchivedBy,
		&match.CreatedBy,
		&match.CreatedAt,
		&match.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find match: %w", err)
	}

	return match, nil
}

// FindByReferenceID retrieves a match by reference_id (Story 6.2)
func (r *Repository) FindByReferenceID(ctx context.Context, referenceID string) (*Match, error) {
	query := `
		SELECT id, event_name, team_name, age_group, match_date, start_time, end_time,
		       location, description, reference_id, status, archived, archived_at, archived_by,
		       created_by, created_at, updated_at
		FROM matches
		WHERE reference_id = $1 AND status != 'deleted'
	`

	match := &Match{}
	err := r.db.QueryRowContext(ctx, query, referenceID).Scan(
		&match.ID,
		&match.EventName,
		&match.TeamName,
		&match.AgeGroup,
		&match.MatchDate,
		&match.StartTime,
		&match.EndTime,
		&match.Location,
		&match.Description,
		&match.ReferenceID,
		&match.Status,
		&match.Archived,
		&match.ArchivedAt,
		&match.ArchivedBy,
		&match.CreatedBy,
		&match.CreatedAt,
		&match.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find match by reference_id: %w", err)
	}

	return match, nil
}

// List retrieves all matches
func (r *Repository) List(ctx context.Context) ([]Match, error) {
	query := `
		SELECT id, event_name, team_name, age_group, match_date, start_time, end_time,
		       location, description, reference_id, status, archived, archived_at, archived_by,
		       created_by, created_at, updated_at
		FROM matches
		WHERE status != 'deleted'
		ORDER BY match_date ASC, start_time ASC, id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list matches: %w", err)
	}
	defer rows.Close()

	matches := []Match{}
	for rows.Next() {
		var m Match
		err := rows.Scan(
			&m.ID,
			&m.EventName,
			&m.TeamName,
			&m.AgeGroup,
			&m.MatchDate,
			&m.StartTime,
			&m.EndTime,
			&m.Location,
			&m.Description,
			&m.ReferenceID,
			&m.Status,
			&m.Archived,
			&m.ArchivedAt,
			&m.ArchivedBy,
			&m.CreatedBy,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			continue
		}
		matches = append(matches, m)
	}

	return matches, nil
}

// Update updates a match with dynamic fields
func (r *Repository) Update(ctx context.Context, id int64, updates map[string]interface{}) (*Match, error) {
	if len(updates) == 0 {
		return r.FindByID(ctx, id)
	}

	// Build UPDATE query dynamically
	setClauses := []string{}
	args := []interface{}{}
	argCount := 1

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argCount))
		args = append(args, value)
		argCount++
	}

	// Always update updated_at
	setClauses = append(setClauses, "updated_at = NOW()")

	// Add WHERE clause
	args = append(args, id)
	whereClause := fmt.Sprintf("WHERE id = $%d", argCount)

	query := fmt.Sprintf("UPDATE matches SET %s %s", strings.Join(setClauses, ", "), whereClause)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update match: %w", err)
	}

	return r.FindByID(ctx, id)
}

// CreateRole creates a role slot for a match
func (r *Repository) CreateRole(ctx context.Context, matchID int64, roleType string) error {
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO match_roles (match_id, role_type) VALUES ($1, $2)",
		matchID, roleType,
	)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	return nil
}

// GetRoles retrieves all role slots for a match
func (r *Repository) GetRoles(ctx context.Context, matchID int64) ([]MatchRole, error) {
	query := `
		SELECT mr.id, mr.match_id, mr.role_type, mr.assigned_referee_id,
		       COALESCE(u.first_name || ' ' || u.last_name, u.name) as referee_name,
		       mr.acknowledged, mr.acknowledged_at,
		       mr.created_at, mr.updated_at
		FROM match_roles mr
		LEFT JOIN users u ON mr.assigned_referee_id = u.id
		WHERE mr.match_id = $1
		ORDER BY mr.role_type
	`

	rows, err := r.db.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	defer rows.Close()

	roles := []MatchRole{}
	for rows.Next() {
		var role MatchRole
		var refereeName *string
		var acknowledgedAt sql.NullTime

		err := rows.Scan(
			&role.ID,
			&role.MatchID,
			&role.RoleType,
			&role.AssignedRefereeID,
			&refereeName,
			&role.Acknowledged,
			&acknowledgedAt,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			continue
		}

		role.AssignedRefereeName = refereeName
		if acknowledgedAt.Valid {
			ackTime := acknowledgedAt.Time.Format(time.RFC3339)
			role.AcknowledgedAt = &ackTime
		}

		// Check if acknowledgment is overdue (assigned >24h ago and not acknowledged)
		if role.AssignedRefereeID != nil && !role.Acknowledged {
			hoursSinceAssignment := time.Since(role.UpdatedAt).Hours()
			if hoursSinceAssignment > 24 {
				role.AckOverdue = true
			}
		}

		roles = append(roles, role)
	}

	return roles, nil
}

// DeleteRoles deletes role slots for a match
func (r *Repository) DeleteRoles(ctx context.Context, matchID int64, roleTypes []string) error {
	if len(roleTypes) == 0 {
		return nil
	}

	// Build placeholders for IN clause
	placeholders := []string{}
	args := []interface{}{matchID}
	for i, roleType := range roleTypes {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
		args = append(args, roleType)
	}

	query := fmt.Sprintf(
		"DELETE FROM match_roles WHERE match_id = $1 AND role_type IN (%s)",
		strings.Join(placeholders, ", "),
	)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete roles: %w", err)
	}
	return nil
}

// RoleExists checks if a role slot exists for a match
func (r *Repository) RoleExists(ctx context.Context, matchID int64, roleType string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM match_roles WHERE match_id = $1 AND role_type = $2)",
		matchID, roleType,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check role existence: %w", err)
	}
	return exists, nil
}

// LogEdit logs a match edit to assignment_history
func (r *Repository) LogEdit(ctx context.Context, matchID int64, actorID int64, changeDescription string) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO assignment_history (match_id, role_type, action, actor_id)
		 VALUES ($1, 'match_edit', $2, $3)`,
		matchID, changeDescription, actorID,
	)
	if err != nil {
		return fmt.Errorf("failed to log edit: %w", err)
	}
	return nil
}

// GetAgeGroup retrieves just the age group for a match
func (r *Repository) GetAgeGroup(ctx context.Context, matchID int64) (*string, error) {
	var ageGroup sql.NullString
	err := r.db.QueryRowContext(
		ctx,
		"SELECT age_group FROM matches WHERE id = $1",
		matchID,
	).Scan(&ageGroup)

	if err != nil {
		return nil, fmt.Errorf("failed to get age group: %w", err)
	}

	if !ageGroup.Valid {
		return nil, nil
	}

	result := ageGroup.String
	return &result, nil
}

// GetCurrentRoles retrieves current role types for a match
func (r *Repository) GetCurrentRoles(ctx context.Context, matchID int64) ([]string, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT role_type FROM match_roles WHERE match_id = $1 ORDER BY role_type",
		matchID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get current roles: %w", err)
	}
	defer rows.Close()

	roleTypes := []string{}
	for rows.Next() {
		var roleType string
		if err := rows.Scan(&roleType); err != nil {
			continue
		}
		roleTypes = append(roleTypes, roleType)
	}

	return roleTypes, nil
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

// GetAgeGroupInt parses age group and returns numeric age
func GetAgeGroupInt(ageGroup *string) (int, error) {
	if ageGroup == nil {
		return 0, fmt.Errorf("age group is nil")
	}

	ageStr := strings.TrimPrefix(*ageGroup, "U")
	age, err := strconv.Atoi(ageStr)
	if err != nil {
		return 0, fmt.Errorf("invalid age group format: %s", *ageGroup)
	}

	return age, nil
}

// ListActive retrieves all non-archived matches
func (r *Repository) ListActive(ctx context.Context) ([]Match, error) {
	query := `
		SELECT id, event_name, team_name, age_group, match_date, start_time, end_time,
		       location, description, reference_id, status, archived, archived_at, archived_by,
		       created_by, created_at, updated_at
		FROM matches
		WHERE status != 'deleted' AND archived = FALSE
		ORDER BY match_date ASC, start_time ASC, id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active matches: %w", err)
	}
	defer rows.Close()

	matches := []Match{}
	for rows.Next() {
		var m Match
		err := rows.Scan(
			&m.ID,
			&m.EventName,
			&m.TeamName,
			&m.AgeGroup,
			&m.MatchDate,
			&m.StartTime,
			&m.EndTime,
			&m.Location,
			&m.Description,
			&m.ReferenceID,
			&m.Status,
			&m.Archived,
			&m.ArchivedAt,
			&m.ArchivedBy,
			&m.CreatedBy,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			continue
		}
		matches = append(matches, m)
	}

	return matches, nil
}

// ListArchived retrieves all archived matches
func (r *Repository) ListArchived(ctx context.Context) ([]Match, error) {
	query := `
		SELECT id, event_name, team_name, age_group, match_date, start_time, end_time,
		       location, description, reference_id, status, archived, archived_at, archived_by,
		       created_by, created_at, updated_at
		FROM matches
		WHERE status != 'deleted' AND archived = TRUE
		ORDER BY archived_at DESC, match_date DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list archived matches: %w", err)
	}
	defer rows.Close()

	matches := []Match{}
	for rows.Next() {
		var m Match
		err := rows.Scan(
			&m.ID,
			&m.EventName,
			&m.TeamName,
			&m.AgeGroup,
			&m.MatchDate,
			&m.StartTime,
			&m.EndTime,
			&m.Location,
			&m.Description,
			&m.ReferenceID,
			&m.Status,
			&m.Archived,
			&m.ArchivedAt,
			&m.ArchivedBy,
			&m.CreatedBy,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			continue
		}
		matches = append(matches, m)
	}

	return matches, nil
}

// Archive marks a match as archived
func (r *Repository) Archive(ctx context.Context, matchID int64, archivedBy int64) error {
	query := `
		UPDATE matches
		SET archived = TRUE,
		    archived_at = NOW(),
		    archived_by = $1,
		    updated_at = NOW()
		WHERE id = $2 AND archived = FALSE
	`

	result, err := r.db.ExecContext(ctx, query, archivedBy, matchID)
	if err != nil {
		return fmt.Errorf("failed to archive match: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("match %d not found or already archived", matchID)
	}

	return nil
}

// Unarchive marks a match as not archived (for administrative purposes)
func (r *Repository) Unarchive(ctx context.Context, matchID int64) error {
	query := `
		UPDATE matches
		SET archived = FALSE,
		    archived_at = NULL,
		    archived_by = NULL,
		    updated_at = NOW()
		WHERE id = $1 AND archived = TRUE
	`

	result, err := r.db.ExecContext(ctx, query, matchID)
	if err != nil {
		return fmt.Errorf("failed to unarchive match: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("match %d not found or not archived", matchID)
	}

	return nil
}
