package users

import (
	"encoding/json"
	"net/http"

	"github.com/msheeley/referee-scheduler/shared/errors"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// Handler handles HTTP requests for user operations
type Handler struct {
	service *Service
}

// NewHandler creates a new user handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMe returns the current authenticated user's information
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	// Return user info (from context, already fetched by middleware)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	})
}

// GetProfile returns the current user's full profile
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	// Get fresh data from database (includes all profile fields)
	profile, err := h.service.GetProfile(r.Context(), user.ID)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// UpdateProfile updates the current user's profile
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	// Parse request
	var req ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid request body"))
		return
	}

	// Update profile (service handles validation)
	updatedUser, err := h.service.UpdateProfile(r.Context(), user.ID, req)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}
