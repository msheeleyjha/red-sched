package availability

import "time"

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
