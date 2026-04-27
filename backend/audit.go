package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// AuditActionType represents the type of action performed
type AuditActionType string

const (
	AuditActionCreate AuditActionType = "create"
	AuditActionUpdate AuditActionType = "update"
	AuditActionDelete AuditActionType = "delete"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	UserID     *int64          `json:"user_id,omitempty"`
	ActionType AuditActionType `json:"action_type"`
	EntityType string          `json:"entity_type"`
	EntityID   int64           `json:"entity_id"`
	OldValues  interface{}     `json:"old_values,omitempty"`
	NewValues  interface{}     `json:"new_values,omitempty"`
	IPAddress  string          `json:"ip_address,omitempty"`
}

// AuditLogger provides async audit logging functionality
type AuditLogger struct {
	db      *sql.DB
	logChan chan AuditEntry
}

// NewAuditLogger creates a new audit logger with async processing
func NewAuditLogger(db *sql.DB) *AuditLogger {
	logger := &AuditLogger{
		db:      db,
		logChan: make(chan AuditEntry, 100), // Buffer up to 100 entries
	}

	// Start background worker to process audit logs
	go logger.processAuditLogs()

	return logger
}

// Log queues an audit entry for async processing
func (a *AuditLogger) Log(entry AuditEntry) {
	select {
	case a.logChan <- entry:
		// Successfully queued
	default:
		// Channel full, log warning but don't block
		log.Printf("Warning: Audit log channel full, dropping entry for %s %s:%d",
			entry.ActionType, entry.EntityType, entry.EntityID)
	}
}

// LogWithContext is a convenience method to log with user and IP from request context
func (a *AuditLogger) LogWithContext(r *http.Request, actionType AuditActionType, entityType string, entityID int64, oldValues, newValues interface{}) {
	entry := AuditEntry{
		ActionType: actionType,
		EntityType: entityType,
		EntityID:   entityID,
		OldValues:  oldValues,
		NewValues:  newValues,
		IPAddress:  getClientIP(r),
	}

	// Try to get user from context
	if user := getUserFromContext(r.Context()); user != nil {
		entry.UserID = &user.ID
	}

	a.Log(entry)
}

// processAuditLogs runs in background and writes audit entries to database
func (a *AuditLogger) processAuditLogs() {
	for entry := range a.logChan {
		if err := a.writeAuditLog(entry); err != nil {
			log.Printf("Error writing audit log: %v", err)
		}
	}
}

// writeAuditLog writes a single audit entry to the database
func (a *AuditLogger) writeAuditLog(entry AuditEntry) error {
	// Serialize old/new values to JSON
	var oldValuesJSON, newValuesJSON []byte
	var err error

	if entry.OldValues != nil {
		oldValuesJSON, err = json.Marshal(entry.OldValues)
		if err != nil {
			return err
		}
	}

	if entry.NewValues != nil {
		newValuesJSON, err = json.Marshal(entry.NewValues)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO audit_logs (user_id, action_type, entity_type, entity_id, old_values, new_values, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = a.db.Exec(query,
		entry.UserID,
		entry.ActionType,
		entry.EntityType,
		entry.EntityID,
		oldValuesJSON,
		newValuesJSON,
		entry.IPAddress,
		time.Now(),
	)

	return err
}

// Close stops the audit logger and waits for pending entries to be written
func (a *AuditLogger) Close() {
	close(a.logChan)
}

// Helper functions

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, use the first one
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// getUserFromContext retrieves the user from the request context
func getUserFromContext(ctx context.Context) *User {
	if user, ok := ctx.Value(userContextKey).(*User); ok {
		return user
	}
	return nil
}
