package availability

import (
	"database/sql"
	"time"
)

// DayUnavailability represents a day when a referee is unavailable
type DayUnavailability struct {
	ID              int64   `json:"id"`
	RefereeID       int64   `json:"referee_id"`
	UnavailableDate string  `json:"unavailable_date"` // YYYY-MM-DD format
	Reason          *string `json:"reason,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

// DayUnavailabilityData represents day unavailability from the database
type DayUnavailabilityData struct {
	ID              int64
	RefereeID       int64
	UnavailableDate time.Time
	Reason          *string
	CreatedAt       time.Time
}

// ToggleMatchAvailabilityRequest represents the request to toggle match availability
type ToggleMatchAvailabilityRequest struct {
	Available *bool `json:"available"` // Pointer to support tri-state: true=available, false=unavailable, null=no preference
}

// ToggleMatchAvailabilityResponse represents the response after toggling match availability
type ToggleMatchAvailabilityResponse struct {
	Success   bool  `json:"success"`
	Available *bool `json:"available"`
}

// ToggleDayUnavailabilityRequest represents the request to toggle day unavailability
type ToggleDayUnavailabilityRequest struct {
	Unavailable bool    `json:"unavailable"`
	Reason      *string `json:"reason,omitempty"`
}

// ToggleDayUnavailabilityResponse represents the response after toggling day unavailability
type ToggleDayUnavailabilityResponse struct {
	Success     bool   `json:"success"`
	Unavailable bool   `json:"unavailable"`
	Date        string `json:"date"`
}

// PaginatedRefereeMatchesResponse contains paginated referee match results
type PaginatedRefereeMatchesResponse struct {
	Matches    []MatchForReferee `json:"matches"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
}

// ConflictingMatch represents another assignment that conflicts with this one
type ConflictingMatch struct {
	MatchID   int64  `json:"match_id"`
	EventName string `json:"event_name"`
	TeamName  string `json:"team_name"`
	StartTime string `json:"start_time"`
	RoleType  string `json:"position"`
}

// MatchForReferee represents a match with eligibility and availability for a specific referee
type MatchForReferee struct {
	ID                 int64              `json:"id"`
	EventName          string             `json:"event_name"`
	TeamName           string             `json:"team_name"`
	AgeGroup           string             `json:"age_group"`
	MatchDate          string             `json:"match_date"`
	StartTime          string             `json:"start_time"`
	EndTime            string             `json:"end_time"`
	Location           string             `json:"location"`
	Description        *string            `json:"description"`
	Status             string             `json:"status"`
	EligibleRoles      []string           `json:"eligible_roles"`
	IsAvailable        bool               `json:"is_available"`
	IsUnavailable      bool               `json:"is_unavailable"`
	IsAssigned         bool               `json:"is_assigned"`
	AssignedRole       *string            `json:"assigned_role"`
	Acknowledged       bool               `json:"acknowledged"`
	AcknowledgedAt     *string            `json:"acknowledged_at"`
	HasConflict        bool               `json:"has_conflict"`
	ConflictingMatches []ConflictingMatch  `json:"conflicting_matches,omitempty"`
}

// RefereeProfile holds referee details needed for eligibility checking
type RefereeProfile struct {
	ID         int64
	DOB        time.Time
	Certified  bool
	CertExpiry sql.NullTime
}
