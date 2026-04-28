package match_reports

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Repository handles database operations for match reports
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new match reports repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new match report
func (r *Repository) Create(ctx context.Context, matchID, submittedBy int64, req *CreateMatchReportRequest) (*MatchReport, error) {
	report := &MatchReport{
		MatchID:        matchID,
		SubmittedBy:    submittedBy,
		FinalScoreHome: req.FinalScoreHome,
		FinalScoreAway: req.FinalScoreAway,
		RedCards:       req.RedCards,
		YellowCards:    req.YellowCards,
		Injuries:       req.Injuries,
		OtherNotes:     req.OtherNotes,
	}

	query := `
		INSERT INTO match_reports (
			match_id, submitted_by, final_score_home, final_score_away,
			red_cards, yellow_cards, injuries, other_notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, submitted_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		report.MatchID,
		report.SubmittedBy,
		report.FinalScoreHome,
		report.FinalScoreAway,
		report.RedCards,
		report.YellowCards,
		report.Injuries,
		report.OtherNotes,
	).Scan(&report.ID, &report.SubmittedAt, &report.UpdatedAt)

	if err != nil {
		// Check for unique constraint violation (report already exists)
		if err.Error() == "pq: duplicate key value violates unique constraint \"match_reports_match_id_key\"" {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}

	return report, nil
}

// Update updates an existing match report
func (r *Repository) Update(ctx context.Context, matchID int64, req *UpdateMatchReportRequest) (*MatchReport, error) {
	query := `
		UPDATE match_reports
		SET final_score_home = $1,
		    final_score_away = $2,
		    red_cards = $3,
		    yellow_cards = $4,
		    injuries = $5,
		    other_notes = $6,
		    updated_at = CURRENT_TIMESTAMP
		WHERE match_id = $7
		RETURNING id, match_id, submitted_by, final_score_home, final_score_away,
		          red_cards, yellow_cards, injuries, other_notes, submitted_at, updated_at
	`

	report := &MatchReport{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		req.FinalScoreHome,
		req.FinalScoreAway,
		req.RedCards,
		req.YellowCards,
		req.Injuries,
		req.OtherNotes,
		matchID,
	).Scan(
		&report.ID,
		&report.MatchID,
		&report.SubmittedBy,
		&report.FinalScoreHome,
		&report.FinalScoreAway,
		&report.RedCards,
		&report.YellowCards,
		&report.Injuries,
		&report.OtherNotes,
		&report.SubmittedAt,
		&report.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return report, nil
}

// GetByMatchID retrieves a match report by match ID
func (r *Repository) GetByMatchID(ctx context.Context, matchID int64) (*MatchReport, error) {
	query := `
		SELECT id, match_id, submitted_by, final_score_home, final_score_away,
		       red_cards, yellow_cards, injuries, other_notes, submitted_at, updated_at
		FROM match_reports
		WHERE match_id = $1
	`

	report := &MatchReport{}
	err := r.db.QueryRowContext(ctx, query, matchID).Scan(
		&report.ID,
		&report.MatchID,
		&report.SubmittedBy,
		&report.FinalScoreHome,
		&report.FinalScoreAway,
		&report.RedCards,
		&report.YellowCards,
		&report.Injuries,
		&report.OtherNotes,
		&report.SubmittedAt,
		&report.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return report, nil
}

// GetByID retrieves a match report by ID
func (r *Repository) GetByID(ctx context.Context, id int64) (*MatchReport, error) {
	query := `
		SELECT id, match_id, submitted_by, final_score_home, final_score_away,
		       red_cards, yellow_cards, injuries, other_notes, submitted_at, updated_at
		FROM match_reports
		WHERE id = $1
	`

	report := &MatchReport{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID,
		&report.MatchID,
		&report.SubmittedBy,
		&report.FinalScoreHome,
		&report.FinalScoreAway,
		&report.RedCards,
		&report.YellowCards,
		&report.Injuries,
		&report.OtherNotes,
		&report.SubmittedAt,
		&report.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return report, nil
}

// GetBySubmitter retrieves all match reports submitted by a specific user
func (r *Repository) GetBySubmitter(ctx context.Context, userID int64) ([]MatchReport, error) {
	query := `
		SELECT id, match_id, submitted_by, final_score_home, final_score_away,
		       red_cards, yellow_cards, injuries, other_notes, submitted_at, updated_at
		FROM match_reports
		WHERE submitted_by = $1
		ORDER BY submitted_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []MatchReport
	for rows.Next() {
		var report MatchReport
		err := rows.Scan(
			&report.ID,
			&report.MatchID,
			&report.SubmittedBy,
			&report.FinalScoreHome,
			&report.FinalScoreAway,
			&report.RedCards,
			&report.YellowCards,
			&report.Injuries,
			&report.OtherNotes,
			&report.SubmittedAt,
			&report.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, rows.Err()
}

// Delete deletes a match report (for admin cleanup only, not exposed via API)
func (r *Repository) Delete(ctx context.Context, matchID int64) error {
	query := `DELETE FROM match_reports WHERE match_id = $1`
	result, err := r.db.ExecContext(ctx, query, matchID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// GetOldReportValues retrieves the old values of a report for audit logging
func (r *Repository) GetOldReportValues(ctx context.Context, matchID int64) (map[string]interface{}, error) {
	report, err := r.GetByMatchID(ctx, matchID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil // No old values if report doesn't exist
		}
		return nil, err
	}

	oldValues := map[string]interface{}{
		"id":               report.ID,
		"match_id":         report.MatchID,
		"submitted_by":     report.SubmittedBy,
		"final_score_home": report.FinalScoreHome,
		"final_score_away": report.FinalScoreAway,
		"red_cards":        report.RedCards,
		"yellow_cards":     report.YellowCards,
		"injuries":         report.Injuries,
		"other_notes":      report.OtherNotes,
		"submitted_at":     report.SubmittedAt.Format(time.RFC3339),
		"updated_at":       report.UpdatedAt.Format(time.RFC3339),
	}

	return oldValues, nil
}
