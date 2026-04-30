package matches

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all match routes
func (h *Handler) RegisterRoutes(r *mux.Router, authMiddleware func(http.HandlerFunc) http.HandlerFunc, requirePermission func(string, http.HandlerFunc) http.HandlerFunc) {
	// Import routes - requires can_manage_matches
	r.HandleFunc("/api/matches/import/parse", authMiddleware(requirePermission("can_manage_matches", h.ParseCSV))).Methods("POST")
	r.HandleFunc("/api/matches/import/confirm", authMiddleware(requirePermission("can_manage_matches", h.ImportMatches))).Methods("POST")

	// List routes - requires can_view_matches
	r.HandleFunc("/api/matches", authMiddleware(requirePermission("can_view_matches", h.ListMatches))).Methods("GET")
	r.HandleFunc("/api/matches/active", authMiddleware(requirePermission("can_view_matches", h.ListActiveMatches))).Methods("GET")
	r.HandleFunc("/api/matches/archived", authMiddleware(h.ListArchivedMatches)).Methods("GET")

	// Match management routes - requires can_manage_matches
	r.HandleFunc("/api/matches/{id}/archive", authMiddleware(requirePermission("can_manage_matches", h.ArchiveMatch))).Methods("POST")
	r.HandleFunc("/api/matches/{id}/unarchive", authMiddleware(requirePermission("can_manage_matches", h.UnarchiveMatch))).Methods("POST")
	r.HandleFunc("/api/matches/{id}", authMiddleware(requirePermission("can_manage_matches", h.UpdateMatch))).Methods("PUT")

	// Add role slot - requires can_assign_referees
	r.HandleFunc("/api/matches/{match_id}/roles/{role_type}", authMiddleware(requirePermission("can_assign_referees", h.AddRoleSlot))).Methods("POST")
}
