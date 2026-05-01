package acknowledgment

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all acknowledgment routes
func (h *Handler) RegisterRoutes(r *mux.Router, authMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	// Acknowledge assignment - authenticated referees only
	r.HandleFunc("/api/referee/matches/{match_id}/acknowledge", authMiddleware(h.AcknowledgeAssignment)).Methods("POST")
}
