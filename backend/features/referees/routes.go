package referees

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all referee routes
func (h *Handler) RegisterRoutes(
	r *mux.Router,
	authMiddleware func(http.HandlerFunc) http.HandlerFunc,
	requirePermission func(string, http.HandlerFunc) http.HandlerFunc,
) {
	// Referee management routes (assignors only)
	r.HandleFunc("/api/referees", requirePermission("can_assign_referees", h.ListReferees)).Methods("GET")
	r.HandleFunc("/api/referees/{id}", requirePermission("can_assign_referees", h.UpdateReferee)).Methods("PUT")
}
