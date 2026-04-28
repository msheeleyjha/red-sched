package matches

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all match routes
func (h *Handler) RegisterRoutes(r *mux.Router, authMiddleware func(http.HandlerFunc) http.HandlerFunc, requirePermission func(string, http.HandlerFunc) http.HandlerFunc) {
	// Parse CSV (preview) - requires manage_matches permission
	r.HandleFunc("/api/matches/import/parse", authMiddleware(requirePermission("manage_matches", h.ParseCSV))).Methods("POST")

	// Import matches (confirm) - requires manage_matches permission
	r.HandleFunc("/api/matches/import/confirm", authMiddleware(requirePermission("manage_matches", h.ImportMatches))).Methods("POST")

	// List all matches - requires manage_matches permission
	r.HandleFunc("/api/matches", authMiddleware(requirePermission("manage_matches", h.ListMatches))).Methods("GET")

	// List active matches (non-archived) - requires manage_matches permission
	r.HandleFunc("/api/matches/active", authMiddleware(requirePermission("manage_matches", h.ListActiveMatches))).Methods("GET")

	// List archived matches (history) - authenticated users can view
	r.HandleFunc("/api/matches/archived", authMiddleware(h.ListArchivedMatches)).Methods("GET")

	// Archive match - requires manage_matches permission
	r.HandleFunc("/api/matches/{id}/archive", authMiddleware(requirePermission("manage_matches", h.ArchiveMatch))).Methods("POST")

	// Unarchive match - requires manage_matches permission
	r.HandleFunc("/api/matches/{id}/unarchive", authMiddleware(requirePermission("manage_matches", h.UnarchiveMatch))).Methods("POST")

	// Update match - requires manage_matches permission
	r.HandleFunc("/api/matches/{id}", authMiddleware(requirePermission("manage_matches", h.UpdateMatch))).Methods("PUT")

	// Add role slot to match - requires manage_matches permission
	r.HandleFunc("/api/matches/{match_id}/roles/{role_type}", authMiddleware(requirePermission("manage_matches", h.AddRoleSlot))).Methods("POST")
}
