package eligibility

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/msheeley/referee-scheduler/shared/errors"
)

// Handler handles HTTP requests for eligibility operations
type Handler struct {
	service ServiceInterface
}

// NewHandler creates a new eligibility handler
func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

// GetEligibleReferees returns all referees with eligibility status for a specific match and role
func (h *Handler) GetEligibleReferees(w http.ResponseWriter, r *http.Request) {
	// Parse match ID from URL
	vars := mux.Vars(r)
	matchID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid match ID"))
		return
	}

	// Get role type from query parameter (default to center)
	roleType := r.URL.Query().Get("role")
	if roleType == "" {
		roleType = "center"
	}

	// Call service
	referees, err := h.service.GetEligibleReferees(r.Context(), matchID, roleType)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(referees)
}
