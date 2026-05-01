package assignments

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/msheeley/referee-scheduler/shared/errors"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// Handler handles HTTP requests for assignment operations
type Handler struct {
	service ServiceInterface
}

// NewHandler creates a new assignment handler
func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

// AssignReferee assigns or removes a referee from a role slot
func (h *Handler) AssignReferee(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	vars := mux.Vars(r)
	matchID, err := strconv.ParseInt(vars["match_id"], 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid match ID"))
		return
	}

	roleType := vars["role_type"]

	var req AssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid request body"))
		return
	}

	result, err := h.service.AssignReferee(r.Context(), matchID, roleType, &req, user.ID)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// CheckConflicts checks if a referee has conflicting assignments
func (h *Handler) CheckConflicts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID, err := strconv.ParseInt(vars["match_id"], 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid match ID"))
		return
	}

	refereeIDStr := r.URL.Query().Get("referee_id")
	if refereeIDStr == "" {
		errors.WriteError(w, errors.NewBadRequest("referee_id query parameter required"))
		return
	}

	refereeID, err := strconv.ParseInt(refereeIDStr, 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid referee_id"))
		return
	}

	result, err := h.service.CheckConflicts(r.Context(), matchID, refereeID)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetRefereeHistory returns all matches assigned to the current referee
func (h *Handler) GetRefereeHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	history, err := h.service.GetRefereeHistory(r.Context(), user.ID)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// MarkMatchAsViewed marks a referee's assignment as viewed
// Story 5.6: Called when referee views match detail page
func (h *Handler) MarkMatchAsViewed(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	vars := mux.Vars(r)
	matchID, err := strconv.ParseInt(vars["match_id"], 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid match ID"))
		return
	}

	err = h.service.MarkMatchAsViewed(r.Context(), matchID, user.ID)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
