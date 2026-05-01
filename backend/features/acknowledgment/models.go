package acknowledgment

import "time"

// AcknowledgeResponse represents the response after acknowledging an assignment
type AcknowledgeResponse struct {
	Success         bool      `json:"success"`
	AcknowledgedAt  time.Time `json:"acknowledged_at"`
}
