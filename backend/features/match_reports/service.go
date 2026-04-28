package match_reports

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// Service handles business logic for match reports
type Service struct {
	repo *Repository
	db   *sql.DB
}

// NewService creates a new match reports service
func NewService(repo *Repository, db *sql.DB) *Service {
	return &Service{
		repo: repo,
		db:   db,
	}
}

// CreateReport creates a new match report
// Authorization: user must be assigned as CENTER referee for this match OR have manage_matches permission
func (s *Service) CreateReport(ctx context.Context, matchID, userID int64, req *CreateMatchReportRequest) (*MatchReport, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check authorization: user must be center referee or assignor
	authorized, err := s.isAuthorizedForReport(ctx, matchID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}
	if !authorized {
		return nil, ErrUnauthorized
	}

	// Create the report
	report, err := s.repo.Create(ctx, matchID, userID, req)
	if err != nil {
		return nil, err
	}

	// Trigger match archival if final score is provided
	if req.FinalScoreHome != nil && req.FinalScoreAway != nil {
		if err := s.archiveMatch(ctx, matchID, userID); err != nil {
			log.Printf("Warning: Failed to archive match %d after report submission: %v", matchID, err)
			// Don't fail the report creation if archival fails
		}
	}

	return report, nil
}

// UpdateReport updates an existing match report
// Authorization: user must be assigned as CENTER referee for this match OR have manage_matches permission
func (s *Service) UpdateReport(ctx context.Context, matchID, userID int64, req *UpdateMatchReportRequest) (*MatchReport, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check authorization: user must be center referee or assignor
	authorized, err := s.isAuthorizedForReport(ctx, matchID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}
	if !authorized {
		return nil, ErrUnauthorized
	}

	// Update the report
	report, err := s.repo.Update(ctx, matchID, req)
	if err != nil {
		return nil, err
	}

	// Trigger match archival if final score is now provided
	if req.FinalScoreHome != nil && req.FinalScoreAway != nil {
		if err := s.archiveMatch(ctx, matchID, userID); err != nil {
			log.Printf("Warning: Failed to archive match %d after report update: %v", matchID, err)
			// Don't fail the report update if archival fails
		}
	}

	return report, nil
}

// GetReportByMatchID retrieves a match report by match ID
func (s *Service) GetReportByMatchID(ctx context.Context, matchID int64) (*MatchReport, error) {
	return s.repo.GetByMatchID(ctx, matchID)
}

// GetReportsBySubmitter retrieves all match reports submitted by a user
func (s *Service) GetReportsBySubmitter(ctx context.Context, userID int64) ([]MatchReport, error) {
	return s.repo.GetBySubmitter(ctx, userID)
}

// isAuthorizedForReport checks if user is authorized to submit/edit a match report
// Returns true if user is:
// 1. Assigned as CENTER referee for this match, OR
// 2. Has manage_matches permission (assignor/admin)
func (s *Service) isAuthorizedForReport(ctx context.Context, matchID, userID int64) (bool, error) {
	// Check if user has manage_matches permission (assignor/admin)
	var hasManagePermission bool
	permQuery := `
		SELECT EXISTS (
			SELECT 1 FROM user_roles ur
			JOIN role_permissions rp ON ur.role_id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.id
			WHERE ur.user_id = $1 AND p.name = 'can_manage_matches'
		)
	`
	err := s.db.QueryRowContext(ctx, permQuery, userID).Scan(&hasManagePermission)
	if err != nil {
		return false, err
	}
	if hasManagePermission {
		return true, nil
	}

	// Check if user is assigned as CENTER referee for this match
	var isCenterReferee bool
	assignmentQuery := `
		SELECT EXISTS (
			SELECT 1 FROM match_roles
			WHERE match_id = $1
			  AND assigned_referee_id = $2
			  AND role_type = 'center'
		)
	`
	err = s.db.QueryRowContext(ctx, assignmentQuery, matchID, userID).Scan(&isCenterReferee)
	if err != nil {
		return false, err
	}

	return isCenterReferee, nil
}

// archiveMatch archives a match after a report is submitted with final score
// This completes Story 4.2: Automatic Archival Logic
func (s *Service) archiveMatch(ctx context.Context, matchID, userID int64) error {
	// Check if match is already archived
	var isArchived bool
	query := `SELECT archived FROM matches WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, matchID).Scan(&isArchived)
	if err != nil {
		return fmt.Errorf("failed to check match archived status: %w", err)
	}

	if isArchived {
		// Already archived, nothing to do
		return nil
	}

	// Archive the match
	archiveQuery := `
		UPDATE matches
		SET archived = TRUE,
		    archived_at = CURRENT_TIMESTAMP,
		    archived_by = $1
		WHERE id = $2
	`
	_, err = s.db.ExecContext(ctx, archiveQuery, userID, matchID)
	if err != nil {
		return fmt.Errorf("failed to archive match: %w", err)
	}

	log.Printf("Match %d automatically archived by user %d after report submission", matchID, userID)
	return nil
}
