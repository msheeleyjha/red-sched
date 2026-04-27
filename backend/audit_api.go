package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

// exportAuditLogsHandler exports audit logs as CSV or JSON
// GET /api/admin/audit-logs/export
func exportAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	// Get export format (default: csv)
	format := queryParams.Get("format")
	if format != "csv" && format != "json" {
		format = "csv"
	}

	// Apply same filters as viewer
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

	// Apply filters (same as viewer)
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

	// Get total count
	var totalCount int
	err := db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		log.Printf("Error counting audit logs for export: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Warn if over 10,000 records
	if totalCount > 10000 {
		w.Header().Set("X-Export-Warning", "Results limited to 10,000 records")
	}

	// Order by timestamp descending
	query += " ORDER BY a.created_at DESC"

	// Limit to 10,000 records
	query += " LIMIT 10000"

	// Execute query
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error querying audit logs for export: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Collect logs
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
			log.Printf("Error scanning audit log for export: %v", err)
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

	// Export based on format
	if format == "json" {
		exportAsJSON(w, auditLogs)
	} else {
		exportAsCSV(w, auditLogs)
	}
}

// exportAsJSON exports audit logs as JSON file
func exportAsJSON(w http.ResponseWriter, logs []AuditLogResponse) {
	filename := fmt.Sprintf("audit_logs_%s.json", time.Now().Format("2006-01-02_15-04-05"))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	json.NewEncoder(w).Encode(logs)
}

// exportAsCSV exports audit logs as CSV file
func exportAsCSV(w http.ResponseWriter, logs []AuditLogResponse) {
	filename := fmt.Sprintf("audit_logs_%s.csv", time.Now().Format("2006-01-02_15-04-05"))

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Write CSV header
	header := "ID,User ID,User Name,User Email,Action Type,Entity Type,Entity ID,Old Values,New Values,IP Address,Created At\n"
	w.Write([]byte(header))

	// Write rows
	for _, log := range logs {
		// Flatten JSON fields to strings
		oldValuesStr := ""
		if log.OldValues != nil {
			oldValuesStr = string(log.OldValues)
		}

		newValuesStr := ""
		if log.NewValues != nil {
			newValuesStr = string(log.NewValues)
		}

		// Escape CSV fields
		row := fmt.Sprintf("%d,%s,%s,%s,%s,%s,%d,%s,%s,%s,%s\n",
			log.ID,
			escapeCSV(ptrToString(log.UserID)),
			escapeCSV(ptrToString(log.UserName)),
			escapeCSV(ptrToString(log.UserEmail)),
			escapeCSV(log.ActionType),
			escapeCSV(log.EntityType),
			log.EntityID,
			escapeCSV(oldValuesStr),
			escapeCSV(newValuesStr),
			escapeCSV(ptrToString(log.IPAddress)),
			log.CreatedAt.Format(time.RFC3339),
		)

		w.Write([]byte(row))
	}
}

// escapeCSV escapes CSV fields that contain commas, quotes, or newlines
func escapeCSV(field string) string {
	if field == "" {
		return ""
	}

	// If field contains comma, quote, or newline, wrap in quotes and escape quotes
	needsQuotes := false
	for _, c := range field {
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			needsQuotes = true
			break
		}
	}

	if needsQuotes {
		// Escape quotes by doubling them
		escaped := ""
		for _, c := range field {
			if c == '"' {
				escaped += "\"\""
			} else {
				escaped += string(c)
			}
		}
		return "\"" + escaped + "\""
	}

	return field
}

// ptrToString converts various pointer types to string
func ptrToString(ptr interface{}) string {
	if ptr == nil {
		return ""
	}

	switch v := ptr.(type) {
	case *int64:
		if v != nil {
			return strconv.FormatInt(*v, 10)
		}
	case *string:
		if v != nil {
			return *v
		}
	}

	return ""
}

// purgeAuditLogsHandler manually triggers audit log purge (admin only)
// POST /api/admin/audit-logs/purge
func purgeAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	// Trigger purge via the retention service
	if retentionService == nil {
		http.Error(w, "Retention service not initialized", http.StatusInternalServerError)
		return
	}

	log.Println("Manual audit log purge triggered by admin")

	result, err := retentionService.PurgeOldLogs()
	if err != nil {
		log.Printf("Error during manual purge: %v", err)
		http.Error(w, "Failed to purge audit logs", http.StatusInternalServerError)
		return
	}

	// Return purge statistics
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
