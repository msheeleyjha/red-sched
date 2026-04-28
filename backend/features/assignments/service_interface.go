package assignments

import "context"

// ServiceInterface defines the interface for assignment business logic
type ServiceInterface interface {
	AssignReferee(ctx context.Context, matchID int64, roleType string, req *AssignmentRequest, actorID int64) (*AssignmentResponse, error)
	CheckConflicts(ctx context.Context, matchID int64, refereeID int64) (*ConflictCheckResponse, error)
	GetRefereeHistory(ctx context.Context, refereeID int64) ([]RefereeHistoryMatch, error)
	MarkMatchAsViewed(ctx context.Context, matchID int64, refereeID int64) error // Story 5.6
}
