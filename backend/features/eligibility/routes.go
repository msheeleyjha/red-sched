package eligibility

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all eligibility routes
func (h *Handler) RegisterRoutes(
	r *mux.Router,
	authMiddleware func(http.HandlerFunc) http.HandlerFunc,
	requirePermission func(string, http.HandlerFunc) http.HandlerFunc,
) {
	// Get eligible referees for a match (assignors only)
	r.HandleFunc("/api/matches/{id}/eligible-referees", requirePermission("can_assign_referees", h.GetEligibleReferees)).Methods("GET")
}
