package assignments

import "time"

// AssignmentRequest represents a request to assign or remove a referee
type AssignmentRequest struct {
	RefereeID *int64 `json:"referee_id"` // null to remove assignment
}

// AssignmentResponse represents the result of an assignment operation
type AssignmentResponse struct {
	Success bool   `json:"success"`
	Action  string `json:"action"` // assigned, reassigned, unassigned
}

// AssignmentHistory represents a historical assignment record
type AssignmentHistory struct {
	ID            int64     `json:"id"`
	MatchID       int64     `json:"match_id"`
	RoleType      string    `json:"role_type"`
	OldRefereeID  *int64    `json:"old_referee_id,omitempty"`
	NewRefereeID  *int64    `json:"new_referee_id,omitempty"`
	Action        string    `json:"action"` // assigned, reassigned, unassigned, match_edit
	ActorID       int64     `json:"actor_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// RoleSlot represents a match role slot from the database
type RoleSlot struct {
	ID                  int64     `json:"id"`
	MatchID             int64     `json:"match_id"`
	RoleType            string    `json:"role_type"`
	AssignedRefereeID   *int64    `json:"assigned_referee_id"`
	Acknowledged        bool      `json:"acknowledged"`
	AcknowledgedAt      *time.Time `json:"acknowledged_at,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// ConflictMatch represents a conflicting match assignment
type ConflictMatch struct {
	MatchID   int64  `json:"match_id"`
	EventName string `json:"event_name"`
	TeamName  string `json:"team_name"`
	MatchDate string `json:"match_date"`
	StartTime string `json:"start_time"`
	RoleType  string `json:"role_type"`
}

// ConflictCheckResponse represents the result of a conflict check
type ConflictCheckResponse struct {
	HasConflict bool            `json:"has_conflict"`
	Conflicts   []ConflictMatch `json:"conflicts"`
}

// MatchTimeWindow represents the time boundaries of a match
type MatchTimeWindow struct {
	MatchID int64
	Start   time.Time
	End     time.Time
}

// RefereeHistoryMatch represents a match in a referee's history
type RefereeHistoryMatch struct {
	MatchID         int64      `json:"match_id"`
	EventName       string     `json:"event_name"`
	TeamName        string     `json:"team_name"`
	AgeGroup        *string    `json:"age_group,omitempty"`
	MatchDate       time.Time  `json:"match_date"`
	StartTime       string     `json:"start_time"`
	EndTime         string     `json:"end_time"`
	Location        string     `json:"location"`
	Status          string     `json:"status"`
	Archived        bool       `json:"archived"`
	ArchivedAt      *time.Time `json:"archived_at,omitempty"`
	RoleType        string     `json:"role_type"`
	Acknowledged    bool       `json:"acknowledged"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty"`
}
