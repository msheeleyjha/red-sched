package audit

import (
	"encoding/json"
	"time"
)

// LogResponse represents an audit log entry in API responses
type LogResponse struct {
	ID         int64           `json:"id"`
	UserID     *int64          `json:"user_id"`
	UserName   *string         `json:"user_name"`
	UserEmail  *string         `json:"user_email"`
	ActionType string          `json:"action_type"`
	EntityType string          `json:"entity_type"`
	EntityID   int64           `json:"entity_id"`
	OldValues  json.RawMessage `json:"old_values,omitempty"`
	NewValues  json.RawMessage `json:"new_values,omitempty"`
	IPAddress  *string         `json:"ip_address,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}
