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
	Archived    bool       `json:"archived"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
	ArchivedBy  *int64     `json:"archived_by,omitempty"`
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
	EventName    string  `json:"event_name"`
	TeamName     string  `json:"team_name"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time"`
	Description  string  `json:"description"`
	Location     string  `json:"location"`
	ReferenceID  string  `json:"reference_id"`
	AgeGroup     *string `json:"age_group"`
	Error        *string `json:"error"`
	FilterReason *string `json:"filter_reason,omitempty"` // Story 6.4: Why row was filtered
	RowNumber    int     `json:"row_number"`
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
	Filters     *ImportFilters      `json:"filters,omitempty"` // Story 6.4: Optional filters
}

// ImportFilters contains filtering options for CSV import (Story 6.4)
type ImportFilters struct {
	FilterPractices bool     `json:"filter_practices"` // Skip matches with "Practice" in team name
	FilterAway      bool     `json:"filter_away"`      // Skip away matches
	HomeLocations   []string `json:"home_locations"`   // List of home venue names/patterns
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
	Imported int      `json:"imported"` // Deprecated: use Created + Updated
	Created  int      `json:"created"`  // Story 6.2: New matches created
	Updated  int      `json:"updated"`  // Story 6.2: Existing matches updated
	Skipped  int      `json:"skipped"`  // Rows skipped due to errors
	Filtered int      `json:"filtered"` // Story 6.4: Rows filtered (practices/away)
	Excluded int      `json:"excluded"` // Story 6.5: Rows excluded (reference_id in exclusion list)
	Errors   []string `json:"errors"`

	// Story 6.6: Detailed breakdown for import summary report
	CreatedMatches  []ImportedMatchSummary `json:"created_matches,omitempty"`
	UpdatedMatches  []ImportedMatchSummary `json:"updated_matches,omitempty"`
	SkippedRows     []SkippedRowSummary    `json:"skipped_rows,omitempty"`
	FilteredRows    []FilteredRowSummary   `json:"filtered_rows,omitempty"`
	ExcludedRows    []ExcludedRowSummary   `json:"excluded_rows,omitempty"`
}

// ImportedMatchSummary contains details about a created or updated match (Story 6.6)
type ImportedMatchSummary struct {
	ReferenceID string `json:"reference_id"`
	TeamName    string `json:"team_name"`
	MatchDate   string `json:"match_date"`
	StartTime   string `json:"start_time"`
	Location    string `json:"location"`
	Action      string `json:"action"` // "created" or "updated"
}

// SkippedRowSummary contains details about a skipped row (Story 6.6)
type SkippedRowSummary struct {
	RowNumber   int    `json:"row_number"`
	ReferenceID string `json:"reference_id"`
	TeamName    string `json:"team_name"`
	Error       string `json:"error"`
}

// FilteredRowSummary contains details about a filtered row (Story 6.6)
type FilteredRowSummary struct {
	RowNumber   int    `json:"row_number"`
	ReferenceID string `json:"reference_id"`
	TeamName    string `json:"team_name"`
	MatchDate   string `json:"match_date"`
	Reason      string `json:"reason"` // "Practice match" or "Away match"
}

// ExcludedRowSummary contains details about an excluded row (Story 6.6)
type ExcludedRowSummary struct {
	RowNumber   int    `json:"row_number"`
	ReferenceID string `json:"reference_id"`
	TeamName    string `json:"team_name"`
	MatchDate   string `json:"match_date"`
}

// ExcludedReferenceID represents a permanently excluded match reference ID (Story 6.5)
type ExcludedReferenceID struct {
	ID          int64      `json:"id"`
	ReferenceID string     `json:"reference_id"`
	Reason      *string    `json:"reason"`
	ExcludedBy  *int64     `json:"excluded_by"`
	ExcludedAt  time.Time  `json:"excluded_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ExcludeReferenceIDRequest represents the request to exclude a reference ID (Story 6.5)
type ExcludeReferenceIDRequest struct {
	ReferenceID string  `json:"reference_id"`
	Reason      *string `json:"reason"`
}
