package availability

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all availability routes
func (h *Handler) RegisterRoutes(
	r *mux.Router,
	authMiddleware func(http.HandlerFunc) http.HandlerFunc,
) {
	// Match availability routes (authenticated referees)
	r.HandleFunc("/api/referee/matches/{id}/availability", authMiddleware(h.ToggleMatchAvailability)).Methods("POST")

	// Day unavailability routes (authenticated referees)
	r.HandleFunc("/api/referee/day-unavailability", authMiddleware(h.GetDayUnavailability)).Methods("GET")
	r.HandleFunc("/api/referee/day-unavailability/{date}", authMiddleware(h.ToggleDayUnavailability)).Methods("POST")

	// Referee match listings (authenticated referees)
	r.HandleFunc("/api/referee/matches", authMiddleware(h.GetEligibleMatchesForReferee)).Methods("GET")
	r.HandleFunc("/api/referee/assignments", authMiddleware(h.GetRefereeAssignments)).Methods("GET")
}
