package availability

import "context"

// ServiceInterface defines the interface for availability business logic
type ServiceInterface interface {
	// ToggleMatchAvailability toggles a referee's availability for a match
	ToggleMatchAvailability(ctx context.Context, matchID int64, refereeID int64, req *ToggleMatchAvailabilityRequest) (*ToggleMatchAvailabilityResponse, error)

	// GetDayUnavailability returns all days marked as unavailable for a referee
	GetDayUnavailability(ctx context.Context, refereeID int64) ([]DayUnavailability, error)

	// ToggleDayUnavailability toggles a referee's unavailability for a day
	ToggleDayUnavailability(ctx context.Context, refereeID int64, date string, req *ToggleDayUnavailabilityRequest) (*ToggleDayUnavailabilityResponse, error)
}
