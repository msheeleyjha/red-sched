package acknowledgment

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/msheeley/referee-scheduler/shared/errors"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// Handler handles HTTP requests for acknowledgment operations
type Handler struct {
	service ServiceInterface
}

// NewHandler creates a new acknowledgment handler
func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

// AcknowledgeAssignment allows a referee to acknowledge their assignment
func (h *Handler) AcknowledgeAssignment(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	// Only referees and assignors can acknowledge assignments
	if user.Role != "referee" && user.Role != "assignor" {
		errors.WriteError(w, errors.NewForbidden("Only referees can acknowledge assignments"))
		return
	}

	vars := mux.Vars(r)
	matchID, err := strconv.ParseInt(vars["match_id"], 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid match ID"))
		return
	}

	result, err := h.service.AcknowledgeAssignment(r.Context(), matchID, user.ID)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
