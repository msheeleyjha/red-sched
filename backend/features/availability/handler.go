package availability

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/msheeley/referee-scheduler/shared/errors"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// Handler handles HTTP requests for availability operations
type Handler struct {
	service ServiceInterface
}

// NewHandler creates a new availability handler
func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

// ToggleMatchAvailability toggles a referee's availability for a match
func (h *Handler) ToggleMatchAvailability(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	// Parse match ID from URL
	vars := mux.Vars(r)
	matchID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid match ID"))
		return
	}

	// Parse request body
	var req ToggleMatchAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid request body"))
		return
	}

	// Call service
	result, err := h.service.ToggleMatchAvailability(r.Context(), matchID, user.ID, &req)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetDayUnavailability returns all days marked as unavailable for the current referee
func (h *Handler) GetDayUnavailability(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	// Call service
	days, err := h.service.GetDayUnavailability(r.Context(), user.ID)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(days)
}

// ToggleDayUnavailability toggles a referee's unavailability for a day
func (h *Handler) ToggleDayUnavailability(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	// Parse date from URL
	vars := mux.Vars(r)
	date := vars["date"]

	// Parse request body
	var req ToggleDayUnavailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid request body"))
		return
	}

	// Call service
	result, err := h.service.ToggleDayUnavailability(r.Context(), user.ID, date, &req)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
