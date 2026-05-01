package availability

import (
	"context"
	"time"

	"github.com/msheeley/referee-scheduler/shared/errors"
)

// Service handles availability business logic
type Service struct {
	repo RepositoryInterface
}

// NewService creates a new availability service
func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// ToggleMatchAvailability toggles a referee's availability for a match
func (s *Service) ToggleMatchAvailability(ctx context.Context, matchID int64, refereeID int64, req *ToggleMatchAvailabilityRequest) (*ToggleMatchAvailabilityResponse, error) {
	// Verify match exists and is active/upcoming
	exists, err := s.repo.MatchExistsAndActive(ctx, matchID)
	if err != nil {
		return nil, errors.NewInternal("Failed to verify match", err)
	}
	if !exists {
		return nil, errors.NewNotFound("Match not found or not available for marking")
	}

	// Toggle availability
	err = s.repo.ToggleMatchAvailability(ctx, matchID, refereeID, req.Available)
	if err != nil {
		return nil, errors.NewInternal("Failed to toggle match availability", err)
	}

	return &ToggleMatchAvailabilityResponse{
		Success:   true,
		Available: req.Available,
	}, nil
}

// GetDayUnavailability returns all days marked as unavailable for a referee
func (s *Service) GetDayUnavailability(ctx context.Context, refereeID int64) ([]DayUnavailability, error) {
	data, err := s.repo.GetDayUnavailability(ctx, refereeID)
	if err != nil {
		return nil, errors.NewInternal("Failed to get day unavailability", err)
	}

	// Convert database models to API models
	result := make([]DayUnavailability, len(data))
	for i, d := range data {
		result[i] = DayUnavailability{
			ID:              d.ID,
			RefereeID:       d.RefereeID,
			UnavailableDate: d.UnavailableDate.Format("2006-01-02"),
			Reason:          d.Reason,
			CreatedAt:       d.CreatedAt.Format(time.RFC3339),
		}
	}

	return result, nil
}

// ToggleDayUnavailability toggles a referee's unavailability for a day
func (s *Service) ToggleDayUnavailability(ctx context.Context, refereeID int64, date string, req *ToggleDayUnavailabilityRequest) (*ToggleDayUnavailabilityResponse, error) {
	// Validate date format
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, errors.NewBadRequest("Invalid date format (expected YYYY-MM-DD)")
	}

	// Toggle day unavailability
	err = s.repo.ToggleDayUnavailability(ctx, refereeID, date, req.Unavailable, req.Reason)
	if err != nil {
		return nil, errors.NewInternal("Failed to toggle day unavailability", err)
	}

	// If marking as unavailable, clear individual match availability for this day
	if req.Unavailable {
		err = s.repo.ClearMatchAvailabilityForDay(ctx, refereeID, date)
		if err != nil {
			return nil, errors.NewInternal("Failed to clear match availability for day", err)
		}
	}

	return &ToggleDayUnavailabilityResponse{
		Success:     true,
		Unavailable: req.Unavailable,
		Date:        date,
	}, nil
}
