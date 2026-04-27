package acknowledgment

import (
	"context"
	"time"

	"github.com/msheeley/referee-scheduler/shared/errors"
)

// Service handles acknowledgment business logic
type Service struct {
	repo RepositoryInterface
}

// NewService creates a new acknowledgment service
func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// AcknowledgeAssignment allows a referee to acknowledge their assignment to a match
func (s *Service) AcknowledgeAssignment(ctx context.Context, matchID int64, refereeID int64) (*AcknowledgeResponse, error) {
	// Verify the referee is actually assigned to this match
	roleType, err := s.repo.GetRefereeAssignmentRole(ctx, matchID, refereeID)
	if err != nil {
		return nil, errors.NewInternal("Failed to verify assignment", err)
	}
	if roleType == nil {
		return nil, errors.NewNotFound("Assignment")
	}

	// Mark as acknowledged
	now := time.Now()
	err = s.repo.AcknowledgeAssignment(ctx, matchID, refereeID, now)
	if err != nil {
		return nil, errors.NewInternal("Failed to acknowledge assignment", err)
	}

	return &AcknowledgeResponse{
		Success:        true,
		AcknowledgedAt: now,
	}, nil
}
