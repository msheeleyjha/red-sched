package referees

import "context"

// ServiceInterface defines the interface for referee business logic
type ServiceInterface interface {
	// List returns all referees for assignor management
	List(ctx context.Context) ([]RefereeListItem, error)

	// Update updates a referee with validation
	Update(ctx context.Context, refereeID int64, currentUserID int64, req *UpdateRequest) (*UpdateResult, error)
}
