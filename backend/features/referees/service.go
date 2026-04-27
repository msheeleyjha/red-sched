package referees

import (
	"context"

	"github.com/msheeley/referee-scheduler/shared/errors"
)

// Service handles referee business logic
type Service struct {
	repo RepositoryInterface
}

// NewService creates a new referee service
func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// ValidStatuses are the allowed status values
var ValidStatuses = map[string]bool{
	"pending":  true,
	"active":   true,
	"inactive": true,
	"removed":  true,
}

// ValidRoles are the allowed role values
var ValidRoles = map[string]bool{
	"referee":  true,
	"assignor": true,
}

// ValidGrades are the allowed grade values
var ValidGrades = map[string]bool{
	"Junior": true,
	"Mid":    true,
	"Senior": true,
}

// List returns all referees for assignor management
func (s *Service) List(ctx context.Context) ([]RefereeListItem, error) {
	referees, err := s.repo.List(ctx)
	if err != nil {
		return nil, errors.NewInternal("Failed to list referees", err)
	}

	return referees, nil
}

// Update updates a referee with validation
func (s *Service) Update(ctx context.Context, refereeID int64, currentUserID int64, req *UpdateRequest) (*UpdateResult, error) {
	// Get current referee
	referee, err := s.repo.FindByID(ctx, refereeID)
	if err != nil {
		return nil, errors.NewInternal("Failed to find referee", err)
	}
	if referee == nil {
		return nil, errors.NewNotFound("Referee")
	}

	// Validate that we have at least one update
	if req.Status == nil && req.Role == nil && req.Grade == nil {
		return nil, errors.NewBadRequest("No updates provided")
	}

	// Don't allow assignors to modify other assignors (except to demote them)
	if referee.Role == "assignor" && currentUserID != referee.ID {
		// Allow changing role from assignor to referee, but nothing else
		if req.Role == nil || *req.Role == "assignor" {
			return nil, errors.NewForbidden("Cannot modify other assignor accounts")
		}
	}

	// Prevent self-deactivation
	if currentUserID == refereeID && req.Status != nil {
		if *req.Status == "inactive" || *req.Status == "removed" {
			return nil, errors.NewForbidden("Cannot deactivate your own account")
		}
	}

	// Check for upcoming assignments before allowing deactivation
	if req.Status != nil && (*req.Status == "inactive" || *req.Status == "removed") {
		hasUpcoming, err := s.repo.HasUpcomingAssignments(ctx, refereeID)
		if err != nil {
			return nil, errors.NewInternal("Failed to check for upcoming assignments", err)
		}
		if hasUpcoming {
			return nil, errors.NewBadRequest("Cannot deactivate user with upcoming match assignments")
		}
	}

	// Build updates map
	updates := make(map[string]interface{})

	// Validate and add status
	if req.Status != nil {
		if !ValidStatuses[*req.Status] {
			return nil, errors.NewBadRequest("Invalid status. Must be: pending, active, inactive, or removed")
		}
		updates["status"] = *req.Status

		// When activating a pending_referee, promote to referee role (if no explicit role change)
		if *req.Status == "active" && referee.Role == "pending_referee" && req.Role == nil {
			updates["role"] = "referee"
		}
	}

	// Validate and add role
	if req.Role != nil {
		if !ValidRoles[*req.Role] {
			return nil, errors.NewBadRequest("Invalid role. Must be: referee or assignor")
		}

		// Only apply role change if different from current
		if referee.Role != *req.Role {
			updates["role"] = *req.Role

			// When promoting to assignor, ensure status is active
			if *req.Role == "assignor" && referee.Status != "active" && req.Status == nil {
				updates["status"] = "active"
			}
		}
	}

	// Validate and add grade
	if req.Grade != nil {
		if *req.Grade == "" {
			// Allow setting grade to NULL
			updates["grade"] = nil
		} else {
			if !ValidGrades[*req.Grade] {
				return nil, errors.NewBadRequest("Invalid grade. Must be: Junior, Mid, or Senior")
			}
			updates["grade"] = *req.Grade
		}
	}

	// Ensure we have updates after processing
	if len(updates) == 0 {
		return nil, errors.NewBadRequest("No valid updates provided")
	}

	// Execute update
	result, err := s.repo.Update(ctx, refereeID, updates)
	if err != nil {
		return nil, errors.NewInternal("Failed to update referee", err)
	}

	return result, nil
}
