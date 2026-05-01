package acknowledgment

import "context"

// ServiceInterface defines the interface for acknowledgment business logic
type ServiceInterface interface {
	AcknowledgeAssignment(ctx context.Context, matchID int64, refereeID int64) (*AcknowledgeResponse, error)
}
