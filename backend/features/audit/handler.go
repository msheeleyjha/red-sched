package audit

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Handler handles HTTP requests for audit log operations
type Handler struct {
	db               *sql.DB
	retentionService *RetentionService
}

// NewHandler creates a new audit handler
func NewHandler(db *sql.DB, retentionService *RetentionService) *Handler {
	return &Handler{
		db:               db,
		retentionService: retentionService,
	}
}

// GetAuditLogs returns paginated audit logs with filters
func (h *Handler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

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

	userIDFilter := queryParams.Get("user_id")
	entityTypeFilter := queryParams.Get("entity_type")
	actionTypeFilter := queryParams.Get("action_type")
	startDateFilter := queryParams.Get("start_date")
	endDateFilter := queryParams.Get("end_date")

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

	query += " ORDER BY a.created_at DESC"

	query += " LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, pageSize, offset)

	var totalCount int
	err := h.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		log.Printf("Error counting audit logs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rows, err := h.db.Query(query, args...)
	if err != nil {
		log.Printf("Error querying audit logs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	auditLogs := []LogResponse{}
	for rows.Next() {
		var logEntry LogResponse
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

		if oldValues.Valid {
			logEntry.OldValues = json.RawMessage(oldValues.String)
		}
		if newValues.Valid {
			logEntry.NewValues = json.RawMessage(newValues.String)
		}

		auditLogs = append(auditLogs, logEntry)
	}

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

// ExportAuditLogs exports audit logs as CSV or JSON
func (h *Handler) ExportAuditLogs(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	format := queryParams.Get("format")
	if format != "csv" && format != "json" {
		format = "csv"
	}

	userIDFilter := queryParams.Get("user_id")
	entityTypeFilter := queryParams.Get("entity_type")
	actionTypeFilter := queryParams.Get("action_type")
	startDateFilter := queryParams.Get("start_date")
	endDateFilter := queryParams.Get("end_date")

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

	var totalCount int
	err := h.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		log.Printf("Error counting audit logs for export: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if totalCount > 10000 {
		w.Header().Set("X-Export-Warning", "Results limited to 10,000 records")
	}

	query += " ORDER BY a.created_at DESC"
	query += " LIMIT 10000"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		log.Printf("Error querying audit logs for export: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	auditLogs := []LogResponse{}
	for rows.Next() {
		var logEntry LogResponse
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

		if oldValues.Valid {
			logEntry.OldValues = json.RawMessage(oldValues.String)
		}
		if newValues.Valid {
			logEntry.NewValues = json.RawMessage(newValues.String)
		}

		auditLogs = append(auditLogs, logEntry)
	}

	if format == "json" {
		h.exportAsJSON(w, auditLogs)
	} else {
		h.exportAsCSV(w, auditLogs)
	}
}

func (h *Handler) exportAsJSON(w http.ResponseWriter, logs []LogResponse) {
	filename := fmt.Sprintf("audit_logs_%s.json", time.Now().Format("2006-01-02_15-04-05"))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	json.NewEncoder(w).Encode(logs)
}

func (h *Handler) exportAsCSV(w http.ResponseWriter, logs []LogResponse) {
	filename := fmt.Sprintf("audit_logs_%s.csv", time.Now().Format("2006-01-02_15-04-05"))

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	header := "ID,User ID,User Name,User Email,Action Type,Entity Type,Entity ID,Old Values,New Values,IP Address,Created At\n"
	w.Write([]byte(header))

	for _, log := range logs {
		oldValuesStr := ""
		if log.OldValues != nil {
			oldValuesStr = string(log.OldValues)
		}

		newValuesStr := ""
		if log.NewValues != nil {
			newValuesStr = string(log.NewValues)
		}

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

func escapeCSV(field string) string {
	if field == "" {
		return ""
	}

	needsQuotes := false
	for _, c := range field {
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			needsQuotes = true
			break
		}
	}

	if needsQuotes {
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

// PurgeAuditLogs manually triggers audit log purge
func (h *Handler) PurgeAuditLogs(w http.ResponseWriter, r *http.Request) {
	if h.retentionService == nil {
		http.Error(w, "Retention service not initialized", http.StatusInternalServerError)
		return
	}

	log.Println("Manual audit log purge triggered by admin")

	result, err := h.retentionService.PurgeOldLogs()
	if err != nil {
		log.Printf("Error during manual purge: %v", err)
		http.Error(w, "Failed to purge audit logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
