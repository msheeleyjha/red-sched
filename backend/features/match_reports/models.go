package match_reports

import "time"

// MatchReport represents a referee's report after completing a match
type MatchReport struct {
	ID             int64      `json:"id"`
	MatchID        int64      `json:"match_id"`
	SubmittedBy    int64      `json:"submitted_by"`
	FinalScoreHome *int       `json:"final_score_home"`
	FinalScoreAway *int       `json:"final_score_away"`
	RedCards       int        `json:"red_cards"`
	YellowCards    int        `json:"yellow_cards"`
	Injuries       *string    `json:"injuries,omitempty"`
	OtherNotes     *string    `json:"other_notes,omitempty"`
	SubmittedAt    time.Time  `json:"submitted_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateMatchReportRequest is the payload for creating a match report
type CreateMatchReportRequest struct {
	FinalScoreHome *int    `json:"final_score_home"`
	FinalScoreAway *int    `json:"final_score_away"`
	RedCards       int     `json:"red_cards"`
	YellowCards    int     `json:"yellow_cards"`
	Injuries       *string `json:"injuries,omitempty"`
	OtherNotes     *string `json:"other_notes,omitempty"`
}

// UpdateMatchReportRequest is the payload for updating a match report
type UpdateMatchReportRequest struct {
	FinalScoreHome *int    `json:"final_score_home"`
	FinalScoreAway *int    `json:"final_score_away"`
	RedCards       int     `json:"red_cards"`
	YellowCards    int     `json:"yellow_cards"`
	Injuries       *string `json:"injuries,omitempty"`
	OtherNotes     *string `json:"other_notes,omitempty"`
}

// Validate validates a create match report request
func (r *CreateMatchReportRequest) Validate() error {
	// Scores can be nil (optional) but if provided must be non-negative
	// This is handled by DB constraint, but we can add validation here too
	if r.FinalScoreHome != nil && *r.FinalScoreHome < 0 {
		return ErrInvalidScore
	}
	if r.FinalScoreAway != nil && *r.FinalScoreAway < 0 {
		return ErrInvalidScore
	}
	if r.RedCards < 0 {
		return ErrInvalidCards
	}
	if r.YellowCards < 0 {
		return ErrInvalidCards
	}
	return nil
}

// Validate validates an update match report request
func (r *UpdateMatchReportRequest) Validate() error {
	if r.FinalScoreHome != nil && *r.FinalScoreHome < 0 {
		return ErrInvalidScore
	}
	if r.FinalScoreAway != nil && *r.FinalScoreAway < 0 {
		return ErrInvalidScore
	}
	if r.RedCards < 0 {
		return ErrInvalidCards
	}
	if r.YellowCards < 0 {
		return ErrInvalidCards
	}
	return nil
}
