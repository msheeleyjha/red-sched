package referees

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/msheeley/referee-scheduler/shared/errors"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// Handler handles HTTP requests for referee operations
type Handler struct {
	service ServiceInterface
}

// NewHandler creates a new referee handler
func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

// ListReferees returns all referees for assignor management
func (h *Handler) ListReferees(w http.ResponseWriter, r *http.Request) {
	referees, err := h.service.List(r.Context())
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(referees)
}

// UpdateReferee allows assignor to update referee status, role, and grade
func (h *Handler) UpdateReferee(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	// Parse referee ID from URL
	vars := mux.Vars(r)
	refereeID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid referee ID"))
		return
	}

	// Parse request body
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid request body"))
		return
	}

	// Call service
	result, err := h.service.Update(r.Context(), refereeID, user.ID, &req)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
