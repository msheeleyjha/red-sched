package assignments

import (
	"context"

	"github.com/msheeley/referee-scheduler/shared/errors"
)

// Service handles assignment business logic
type Service struct {
	repo RepositoryInterface
}

// NewService creates a new assignment service
func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// ValidRoleTypes defines the valid role types for assignments
var ValidRoleTypes = map[string]bool{
	"center":      true,
	"assistant_1": true,
	"assistant_2": true,
}

// RoleTypeDisplayNames maps role types to human-readable names
var RoleTypeDisplayNames = map[string]string{
	"center":      "Center Referee",
	"assistant_1": "Assistant Referee 1",
	"assistant_2": "Assistant Referee 2",
}

// AssignReferee assigns or removes a referee from a match role
func (s *Service) AssignReferee(ctx context.Context, matchID int64, roleType string, req *AssignmentRequest, actorID int64) (*AssignmentResponse, error) {
	// Validate role type
	if !ValidRoleTypes[roleType] {
		return nil, errors.NewBadRequest("Invalid role type")
	}

	// Verify match exists and is active
	matchExists, err := s.repo.MatchExists(ctx, matchID)
	if err != nil {
		return nil, errors.NewInternal("Failed to verify match", err)
	}
	if !matchExists {
		return nil, errors.NewNotFound("Match")
	}

	// Get role slot
	roleSlot, err := s.repo.GetRoleSlot(ctx, matchID, roleType)
	if err != nil {
		return nil, errors.NewInternal("Failed to get role slot", err)
	}
	if roleSlot == nil {
		return nil, errors.NewNotFound("Role slot")
	}

	// Store current referee ID for history
	var currentRefereeID *int64
	if roleSlot.AssignedRefereeID != nil {
		currentRefereeID = roleSlot.AssignedRefereeID
	}

	// If assigning (not removing)
	if req.RefereeID != nil {
		// Verify referee exists and is active
		refereeExists, err := s.repo.RefereeExists(ctx, *req.RefereeID)
		if err != nil {
			return nil, errors.NewInternal("Failed to verify referee", err)
		}
		if !refereeExists {
			return nil, errors.NewBadRequest("Referee not found or not active")
		}

		// Check if referee is already assigned to another role on this match
		existingRole, err := s.repo.GetRefereeExistingRoleOnMatch(ctx, matchID, *req.RefereeID, roleType)
		if err != nil {
			return nil, errors.NewInternal("Failed to check existing role", err)
		}
		if existingRole != nil {
			displayName := RoleTypeDisplayNames[*existingRole]
			if displayName == "" {
				displayName = *existingRole
			}
			return nil, errors.NewBadRequest("Referee is already assigned as " + displayName + " for this match")
		}

		// TODO: Check eligibility (optional for v1, can assign anyone)
		// TODO: Check for double-booking conflicts (could be done here or in frontend)
	}

	// Update assignment
	err = s.repo.UpdateRoleAssignment(ctx, roleSlot.ID, req.RefereeID)
	if err != nil {
		return nil, errors.NewInternal("Failed to update assignment", err)
	}

	// Determine action type
	action := "unassigned"
	if req.RefereeID != nil {
		if currentRefereeID != nil {
			action = "reassigned"
		} else {
			action = "assigned"
		}
	}

	// Log assignment history
	history := &AssignmentHistory{
		MatchID:      matchID,
		RoleType:     roleType,
		OldRefereeID: currentRefereeID,
		NewRefereeID: req.RefereeID,
		Action:       action,
		ActorID:      actorID,
	}

	err = s.repo.LogAssignment(ctx, history)
	if err != nil {
		// Log error but don't fail the request
		// In production, use proper logger
	}

	return &AssignmentResponse{
		Success: true,
		Action:  action,
	}, nil
}

// CheckConflicts checks if a referee has conflicting assignments for a given match
func (s *Service) CheckConflicts(ctx context.Context, matchID int64, refereeID int64) (*ConflictCheckResponse, error) {
	// Get match time window
	timeWindow, err := s.repo.GetMatchTimeWindow(ctx, matchID)
	if err != nil {
		return nil, errors.NewInternal("Failed to get match time window", err)
	}
	if timeWindow == nil {
		return nil, errors.NewNotFound("Match")
	}

	// Find overlapping assignments
	conflicts, err := s.repo.FindConflictingAssignments(ctx, refereeID, matchID, timeWindow.Start, timeWindow.End)
	if err != nil {
		return nil, errors.NewInternal("Failed to find conflicts", err)
	}

	return &ConflictCheckResponse{
		HasConflict: len(conflicts) > 0,
		Conflicts:   conflicts,
	}, nil
}
