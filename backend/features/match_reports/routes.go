package match_reports

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers match report routes
func (h *Handler) RegisterRoutes(r *mux.Router, requireAuth func(http.HandlerFunc) http.HandlerFunc) {
	// Match report endpoints (authentication required)
	r.HandleFunc("/api/matches/{id}/report", requireAuth(h.CreateReportHandler)).Methods("POST")
	r.HandleFunc("/api/matches/{id}/report", requireAuth(h.UpdateReportHandler)).Methods("PUT")
	r.HandleFunc("/api/matches/{id}/report", requireAuth(h.GetReportHandler)).Methods("GET")

	// Referee endpoints
	r.HandleFunc("/api/referee/my-reports", requireAuth(h.GetMyReportsHandler)).Methods("GET")
}
