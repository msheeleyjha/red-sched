package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

// AuditLogResponse represents an audit log entry in API responses
type AuditLogResponse struct {
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

// getAuditLogsHandler returns paginated audit logs with filters
// GET /api/admin/audit-logs
func getAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryParams := r.URL.Query()

	// Pagination
	page := 1
	if pageStr := queryParams.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 100
	if pageSizeStr := queryParams.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 1000 {
			pageSize = ps
		}
	}

	offset := (page - 1) * pageSize

	// Filters
	userIDFilter := queryParams.Get("user_id")
	entityTypeFilter := queryParams.Get("entity_type")
	actionTypeFilter := queryParams.Get("action_type")
	startDateFilter := queryParams.Get("start_date")
	endDateFilter := queryParams.Get("end_date")

	// Build query
	query := `
		SELECT
			a.id, a.user_id, u.name as user_name, u.email as user_email,
			a.action_type, a.entity_type, a.entity_id,
			a.old_values, a.new_values, a.ip_address, a.created_at
		FROM audit_logs a
		LEFT JOIN users u ON a.user_id = u.id
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE 1=1`

	args := []interface{}{}
	countArgs := []interface{}{}
	argCount := 1

	// Apply filters
	if userIDFilter != "" {
		query += " AND a.user_id = $" + strconv.Itoa(argCount)
		countQuery += " AND user_id = $" + strconv.Itoa(argCount)
		args = append(args, userIDFilter)
		countArgs = append(countArgs, userIDFilter)
		argCount++
	}

	if entityTypeFilter != "" {
		query += " AND a.entity_type = $" + strconv.Itoa(argCount)
		countQuery += " AND entity_type = $" + strconv.Itoa(argCount)
		args = append(args, entityTypeFilter)
		countArgs = append(countArgs, entityTypeFilter)
		argCount++
	}

	if actionTypeFilter != "" {
		query += " AND a.action_type = $" + strconv.Itoa(argCount)
		countQuery += " AND action_type = $" + strconv.Itoa(argCount)
		args = append(args, actionTypeFilter)
		countArgs = append(countArgs, actionTypeFilter)
		argCount++
	}

	if startDateFilter != "" {
		query += " AND a.created_at >= $" + strconv.Itoa(argCount)
		countQuery += " AND created_at >= $" + strconv.Itoa(argCount)
		args = append(args, startDateFilter)
		countArgs = append(countArgs, startDateFilter)
		argCount++
	}

	if endDateFilter != "" {
		query += " AND a.created_at <= $" + strconv.Itoa(argCount)
		countQuery += " AND created_at <= $" + strconv.Itoa(argCount)
		args = append(args, endDateFilter)
		countArgs = append(countArgs, endDateFilter)
		argCount++
	}

	// Order by timestamp descending (newest first)
	query += " ORDER BY a.created_at DESC"

	// Add pagination
	query += " LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, pageSize, offset)

	// Get total count
	var totalCount int
	err := db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		log.Printf("Error counting audit logs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Execute query
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error querying audit logs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	auditLogs := []AuditLogResponse{}
	for rows.Next() {
		var logEntry AuditLogResponse
		var oldValues, newValues sql.NullString

		err := rows.Scan(
			&logEntry.ID,
			&logEntry.UserID,
			&logEntry.UserName,
			&logEntry.UserEmail,
			&logEntry.ActionType,
			&logEntry.EntityType,
			&logEntry.EntityID,
			&oldValues,
			&newValues,
			&logEntry.IPAddress,
			&logEntry.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning audit log: %v", err)
			continue
		}

		// Handle JSON fields
		if oldValues.Valid {
			logEntry.OldValues = json.RawMessage(oldValues.String)
		}
		if newValues.Valid {
			logEntry.NewValues = json.RawMessage(newValues.String)
		}

		auditLogs = append(auditLogs, logEntry)
	}

	// Return response
	response := map[string]interface{}{
		"logs":        auditLogs,
		"total_count": totalCount,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (totalCount + pageSize - 1) / pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
