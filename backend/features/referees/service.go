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
	"active":   true,
	"inactive": true,
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
	referee, err := s.repo.FindByID(ctx, refereeID)
	if err != nil {
		return nil, errors.NewInternal("Failed to find referee", err)
	}
	if referee == nil {
		return nil, errors.NewNotFound("Referee")
	}

	if req.Status == nil && req.Grade == nil {
		return nil, errors.NewBadRequest("No updates provided")
	}

	// Don't allow modifying assignors through referee management
	if referee.Role == "assignor" && currentUserID != referee.ID {
		return nil, errors.NewForbidden("Cannot modify assignor accounts through referee management")
	}

	if currentUserID == refereeID && req.Status != nil {
		if *req.Status == "inactive" {
			return nil, errors.NewForbidden("Cannot deactivate your own account")
		}
	}

	if req.Status != nil && *req.Status == "inactive" {
		hasUpcoming, err := s.repo.HasUpcomingAssignments(ctx, refereeID)
		if err != nil {
			return nil, errors.NewInternal("Failed to check for upcoming assignments", err)
		}
		if hasUpcoming {
			return nil, errors.NewBadRequest("Cannot deactivate user with upcoming match assignments")
		}
	}

	updates := make(map[string]interface{})

	if req.Status != nil {
		if !ValidStatuses[*req.Status] {
			return nil, errors.NewBadRequest("Invalid status. Must be: active or inactive")
		}
		updates["status"] = *req.Status

		if *req.Status == "active" && (referee.Role == "pending_referee" || referee.Status == "inactive") {
			updates["role"] = "referee"
		}
	}

	if req.Grade != nil {
		if *req.Grade == "" {
			updates["grade"] = nil
		} else {
			if !ValidGrades[*req.Grade] {
				return nil, errors.NewBadRequest("Invalid grade. Must be: Junior, Mid, or Senior")
			}
			updates["grade"] = *req.Grade
		}
	}

	if len(updates) == 0 {
		return nil, errors.NewBadRequest("No valid updates provided")
	}

	result, err := s.repo.Update(ctx, refereeID, updates)
	if err != nil {
		return nil, errors.NewInternal("Failed to update referee", err)
	}

	// Manage RBAC roles based on status changes
	if req.Status != nil {
		if *req.Status == "active" {
			if err := s.repo.AssignRBACRole(ctx, refereeID, "Referee"); err != nil {
				return nil, errors.NewInternal("Failed to assign referee role", err)
			}
		} else if *req.Status == "inactive" {
			if err := s.repo.RemoveRBACRole(ctx, refereeID, "Referee"); err != nil {
				return nil, errors.NewInternal("Failed to remove referee role", err)
			}
		}
	}

	return result, nil
}
