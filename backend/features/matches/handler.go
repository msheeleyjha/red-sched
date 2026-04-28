package matches

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/msheeley/referee-scheduler/shared/errors"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// Handler handles HTTP requests for match operations
type Handler struct {
	service ServiceInterface
}

// NewHandler creates a new match handler
func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

// ParseCSV parses uploaded CSV and returns preview with errors
func (h *Handler) ParseCSV(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (10MB limit)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Failed to parse form"))
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("No file uploaded"))
		return
	}
	defer file.Close()

	// Parse CSV using service
	response, err := h.service.ParseCSV(r.Context(), file, header.Filename)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ImportMatches confirms and imports matches to database
func (h *Handler) ImportMatches(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	var req ImportConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid request body"))
		return
	}

	result, err := h.service.ImportMatches(r.Context(), &req, user.ID)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ListMatches returns paginated matches for assignor schedule view
func (h *Handler) ListMatches(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page := 1
	if v := q.Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 25
	if v := q.Get("per_page"); v != "" {
		if pp, err := strconv.Atoi(v); err == nil && pp > 0 {
			perPage = pp
		}
	}

	params := &MatchListParams{
		Page:             page,
		PerPage:          perPage,
		DateFrom:         q.Get("date_from"),
		DateTo:           q.Get("date_to"),
		AgeGroup:         q.Get("age_group"),
		AssignmentStatus: q.Get("assignment_status"),
		ShowCancelled:    q.Get("show_cancelled") == "true",
	}

	result, err := h.service.ListMatches(r.Context(), params)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// UpdateMatch updates a match
func (h *Handler) UpdateMatch(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	vars := mux.Vars(r)
	matchID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid match ID"))
		return
	}

	var req MatchUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid request body"))
		return
	}

	updated, err := h.service.UpdateMatch(r.Context(), matchID, &req, user.ID)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// AddRoleSlot allows assignor to manually add AR slots to matches (e.g., for U10)
func (h *Handler) AddRoleSlot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID, err := strconv.ParseInt(vars["match_id"], 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid match ID"))
		return
	}

	roleType := vars["role_type"]

	err = h.service.AddRoleSlot(r.Context(), matchID, roleType)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"role_type": roleType,
	})
}

// ListActiveMatches returns all non-archived matches
func (h *Handler) ListActiveMatches(w http.ResponseWriter, r *http.Request) {
	matches, err := h.service.ListActiveMatches(r.Context())
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

// ListArchivedMatches returns all archived matches (history view)
func (h *Handler) ListArchivedMatches(w http.ResponseWriter, r *http.Request) {
	matches, err := h.service.ListArchivedMatches(r.Context())
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

// ArchiveMatch archives a match (marks as completed)
func (h *Handler) ArchiveMatch(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	vars := mux.Vars(r)
	matchID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid match ID"))
		return
	}

	err = h.service.ArchiveMatch(r.Context(), matchID, user.ID)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Match archived successfully",
	})
}

// UnarchiveMatch unarchives a match (for administrative purposes)
func (h *Handler) UnarchiveMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid match ID"))
		return
	}

	err = h.service.UnarchiveMatch(r.Context(), matchID)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Match unarchived successfully",
	})
}
