package assignments

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all assignment routes
func (h *Handler) RegisterRoutes(r *mux.Router, authMiddleware func(http.HandlerFunc) http.HandlerFunc, requirePermission func(string, http.HandlerFunc) http.HandlerFunc) {
	// Assign or remove referee - requires manage_assignments permission
	r.HandleFunc("/api/matches/{match_id}/roles/{role_type}/assign", authMiddleware(requirePermission("manage_assignments", h.AssignReferee))).Methods("POST")

	// Check for conflicting assignments - requires manage_assignments permission
	r.HandleFunc("/api/matches/{match_id}/conflicts", authMiddleware(requirePermission("manage_assignments", h.CheckConflicts))).Methods("GET")

	// Get referee's match history (all assignments, active and archived)
	r.HandleFunc("/api/referee/my-history", authMiddleware(h.GetRefereeHistory)).Methods("GET")
}
