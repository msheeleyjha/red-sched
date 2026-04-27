package matches

import "time"

// Match represents a scheduled match
type Match struct {
	ID          int64      `json:"id"`
	EventName   string     `json:"event_name"`
	TeamName    string     `json:"team_name"`
	AgeGroup    *string    `json:"age_group"`
	MatchDate   time.Time  `json:"match_date"`
	StartTime   string     `json:"start_time"`
	EndTime     string     `json:"end_time"`
	Location    string     `json:"location"`
	Description *string    `json:"description"`
	ReferenceID *string    `json:"reference_id"`
	Status      string     `json:"status"`
	CreatedBy   int64      `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// MatchRole represents a role slot for a match
type MatchRole struct {
	ID                  int64     `json:"id"`
	MatchID             int64     `json:"match_id"`
	RoleType            string    `json:"role_type"` // center, assistant_1, assistant_2
	AssignedRefereeID   *int64    `json:"assigned_referee_id"`
	AssignedRefereeName *string   `json:"assigned_referee_name,omitempty"`
	Acknowledged        bool      `json:"acknowledged"`
	AcknowledgedAt      *string   `json:"acknowledged_at,omitempty"`
	AckOverdue          bool      `json:"ack_overdue"` // True if assigned >24h ago and not acknowledged
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// MatchWithRoles includes match data and role assignments
type MatchWithRoles struct {
	Match
	Roles            []MatchRole `json:"roles"`
	AssignmentStatus string      `json:"assignment_status"` // unassigned, partial, full
	HasOverdueAck    bool        `json:"has_overdue_ack"`   // true if any assignment is overdue
}

// CSVRow represents a parsed row from Stack Team App CSV
type CSVRow struct {
	EventName   string  `json:"event_name"`
	TeamName    string  `json:"team_name"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	ReferenceID string  `json:"reference_id"`
	AgeGroup    *string `json:"age_group"`
	Error       *string `json:"error"`
	RowNumber   int     `json:"row_number"`
}

// ImportPreviewResponse contains parsed rows and any duplicates found
type ImportPreviewResponse struct {
	Rows       []CSVRow              `json:"rows"`
	Duplicates []DuplicateMatchGroup `json:"duplicates"`
}

// DuplicateMatchGroup represents a set of duplicate matches
type DuplicateMatchGroup struct {
	Signal   string   `json:"signal"` // reference_id or datetime_location
	Matches  []CSVRow `json:"matches"`
	Existing *Match   `json:"existing,omitempty"` // If duplicate with existing DB record
}

// ImportConfirmRequest contains the user's resolution of duplicates
type ImportConfirmRequest struct {
	Rows        []CSVRow            `json:"rows"`
	Resolutions map[string][]CSVRow `json:"resolutions"` // Key: duplicate group ID, Value: rows to import
}

// MatchUpdateRequest represents the update payload for a match
type MatchUpdateRequest struct {
	EventName   *string `json:"event_name"`
	TeamName    *string `json:"team_name"`
	AgeGroup    *string `json:"age_group"`
	MatchDate   *string `json:"match_date"`
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
	Location    *string `json:"location"`
	Description *string `json:"description"`
	Status      *string `json:"status"` // active, cancelled
}

// ImportResult contains the result of a match import operation
type ImportResult struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors"`
}
