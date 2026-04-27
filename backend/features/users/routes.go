package users

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all user-related routes
func (h *Handler) RegisterRoutes(r *mux.Router, authMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	// Authenticated user endpoints
	r.HandleFunc("/api/auth/me", authMiddleware(h.GetMe)).Methods("GET")
	r.HandleFunc("/api/profile", authMiddleware(h.GetProfile)).Methods("GET")
	r.HandleFunc("/api/profile", authMiddleware(h.UpdateProfile)).Methods("PUT")
}
