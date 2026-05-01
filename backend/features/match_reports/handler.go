package match_reports

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// Handler handles HTTP requests for match reports
type Handler struct {
	service ServiceInterface
	db      *sql.DB
}

// NewHandler creates a new match reports handler
func NewHandler(service ServiceInterface, db *sql.DB) *Handler {
	return &Handler{
		service: service,
		db:      db,
	}
}

// CreateReportHandler handles POST /api/matches/:id/report
func (h *Handler) CreateReportHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized - user not found in context", http.StatusUnauthorized)
		return
	}

	// Get match ID from URL
	vars := mux.Vars(r)
	matchIDStr := vars["id"]
	matchID, err := strconv.ParseInt(matchIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req CreateMatchReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create report
	report, err := h.service.CreateReport(r.Context(), matchID, user.ID, &req)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			http.Error(w, "Match report already exists. Use PUT to update.", http.StatusConflict)
			return
		}
		if errors.Is(err, ErrUnauthorized) {
			http.Error(w, "Unauthorized: Only center referee or assignor can submit reports", http.StatusForbidden)
			return
		}
		if errors.Is(err, ErrInvalidScore) || errors.Is(err, ErrInvalidCards) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Error creating match report: %v", err)
		http.Error(w, "Failed to create match report", http.StatusInternalServerError)
		return
	}

	// Create audit log entry
	h.logReportAction(r.Context(), user.ID, "create", report.MatchID, nil, report)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(report)
}

// UpdateReportHandler handles PUT /api/matches/:id/report
func (h *Handler) UpdateReportHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized - user not found in context", http.StatusUnauthorized)
		return
	}

	// Get match ID from URL
	vars := mux.Vars(r)
	matchIDStr := vars["id"]
	matchID, err := strconv.ParseInt(matchIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	// Get old values for audit log
	oldReport, _ := h.service.GetReportByMatchID(r.Context(), matchID)

	// Parse request body
	var req UpdateMatchReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update report
	report, err := h.service.UpdateReport(r.Context(), matchID, user.ID, &req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "Match report not found. Use POST to create.", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrUnauthorized) {
			http.Error(w, "Unauthorized: Only center referee or assignor can edit reports", http.StatusForbidden)
			return
		}
		if errors.Is(err, ErrInvalidScore) || errors.Is(err, ErrInvalidCards) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Error updating match report: %v", err)
		http.Error(w, "Failed to update match report", http.StatusInternalServerError)
		return
	}

	// Create audit log entry
	h.logReportAction(r.Context(), user.ID, "update", report.MatchID, oldReport, report)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetReportHandler handles GET /api/matches/:id/report
func (h *Handler) GetReportHandler(w http.ResponseWriter, r *http.Request) {
	// Get match ID from URL
	vars := mux.Vars(r)
	matchIDStr := vars["id"]
	matchID, err := strconv.ParseInt(matchIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	// Get report
	report, err := h.service.GetReportByMatchID(r.Context(), matchID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "Match report not found", http.StatusNotFound)
			return
		}
		log.Printf("Error retrieving match report: %v", err)
		http.Error(w, "Failed to retrieve match report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetMyReportsHandler handles GET /api/referee/my-reports
func (h *Handler) GetMyReportsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized - user not found in context", http.StatusUnauthorized)
		return
	}

	// Get reports submitted by this user
	reports, err := h.service.GetReportsBySubmitter(r.Context(), user.ID)
	if err != nil {
		log.Printf("Error retrieving user reports: %v", err)
		http.Error(w, "Failed to retrieve reports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

// logReportAction creates an audit log entry for match report actions
func (h *Handler) logReportAction(ctx context.Context, userID int64, action string, matchID int64, oldReport, newReport *MatchReport) {
	var oldValues, newValues interface{}

	if oldReport != nil {
		oldValues = map[string]interface{}{
			"final_score_home": oldReport.FinalScoreHome,
			"final_score_away": oldReport.FinalScoreAway,
			"red_cards":        oldReport.RedCards,
			"yellow_cards":     oldReport.YellowCards,
			"injuries":         oldReport.Injuries,
			"other_notes":      oldReport.OtherNotes,
		}
	}

	if newReport != nil {
		newValues = map[string]interface{}{
			"final_score_home": newReport.FinalScoreHome,
			"final_score_away": newReport.FinalScoreAway,
			"red_cards":        newReport.RedCards,
			"yellow_cards":     newReport.YellowCards,
			"injuries":         newReport.Injuries,
			"other_notes":      newReport.OtherNotes,
		}
	}

	oldJSON, _ := json.Marshal(oldValues)
	newJSON, _ := json.Marshal(newValues)

	_, err := h.db.ExecContext(
		ctx,
		`INSERT INTO audit_logs (user_id, action_type, entity_type, entity_id, old_values, new_values, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)`,
		userID,
		action,
		"match_report",
		matchID,
		oldJSON,
		newJSON,
	)

	if err != nil {
		log.Printf("Warning: Failed to create audit log for match report %s: %v", action, err)
	}
}
