package users

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/msheeley/referee-scheduler/shared/errors"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// mockService is a mock implementation of Service for testing
type mockService struct {
	findOrCreateFunc  func(ctx context.Context, googleID, email, name string) (*User, error)
	getByIDFunc       func(ctx context.Context, id int64) (*User, error)
	updateProfileFunc func(ctx context.Context, userID int64, req ProfileUpdateRequest) (*User, error)
	getProfileFunc    func(ctx context.Context, userID int64) (*User, error)
}

func (m *mockService) FindOrCreate(ctx context.Context, googleID, email, name string) (*User, error) {
	if m.findOrCreateFunc != nil {
		return m.findOrCreateFunc(ctx, googleID, email, name)
	}
	return nil, nil
}

func (m *mockService) GetByID(ctx context.Context, id int64) (*User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockService) UpdateProfile(ctx context.Context, userID int64, req ProfileUpdateRequest) (*User, error) {
	if m.updateProfileFunc != nil {
		return m.updateProfileFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *mockService) GetProfile(ctx context.Context, userID int64) (*User, error) {
	if m.getProfileFunc != nil {
		return m.getProfileFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockService) GetByGoogleID(ctx context.Context, googleID string) (*User, error) {
	return nil, nil
}

func TestHandler_GetMe(t *testing.T) {
	t.Run("returns current user info", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("GET", "/api/auth/me", nil)
		w := httptest.NewRecorder()

		// Add user to context (simulating middleware)
		user := &middleware.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
			Role:  "referee",
		}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.GetMe(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["email"] != "test@example.com" {
			t.Errorf("Expected email test@example.com, got %v", response["email"])
		}
		if response["name"] != "Test User" {
			t.Errorf("Expected name Test User, got %v", response["name"])
		}
	})

	t.Run("returns error when user not in context", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("GET", "/api/auth/me", nil)
		w := httptest.NewRecorder()

		// No user in context
		handler.GetMe(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})
}

func TestHandler_GetProfile(t *testing.T) {
	t.Run("returns full user profile", func(t *testing.T) {
		fullUser := &User{
			ID:        1,
			Email:     "test@example.com",
			Name:      "Test User",
			FirstName: stringPtr("Test"),
			LastName:  stringPtr("User"),
			Certified: true,
		}

		mockSvc := &mockService{
			getProfileFunc: func(ctx context.Context, userID int64) (*User, error) {
				if userID == 1 {
					return fullUser, nil
				}
				return nil, nil
			},
		}

		handler := NewHandler(mockSvc)

		req := httptest.NewRequest("GET", "/api/profile", nil)
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 1, Email: "test@example.com", Name: "Test User", Role: "referee"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.GetProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response User
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.ID != 1 {
			t.Errorf("Expected user ID 1, got %d", response.ID)
		}
		if !response.Certified {
			t.Error("Expected user to be certified")
		}
	})

	t.Run("returns error when user not in context", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("GET", "/api/profile", nil)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})
}

func TestHandler_UpdateProfile(t *testing.T) {
	t.Run("successfully updates profile", func(t *testing.T) {
		certExpiry := "2027-12-31"
		reqBody := ProfileUpdateRequest{
			FirstName:  "John",
			LastName:   "Doe",
			Certified:  true,
			CertExpiry: &certExpiry,
		}

		updatedUser := &User{
			ID:        1,
			FirstName: &reqBody.FirstName,
			LastName:  &reqBody.LastName,
			Certified: true,
		}

		mockSvc := &mockService{
			updateProfileFunc: func(ctx context.Context, userID int64, req ProfileUpdateRequest) (*User, error) {
				if userID == 1 {
					return updatedUser, nil
				}
				return nil, nil
			},
		}

		handler := NewHandler(mockSvc)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/api/profile", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 1, Email: "test@example.com", Name: "Test User", Role: "referee"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.UpdateProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response User
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.ID != 1 {
			t.Errorf("Expected user ID 1, got %d", response.ID)
		}
		if *response.FirstName != "John" {
			t.Errorf("Expected first name John, got %s", *response.FirstName)
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("PUT", "/api/profile", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 1, Email: "test@example.com", Name: "Test User", Role: "referee"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.UpdateProfile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns error when user not in context", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		reqBody := ProfileUpdateRequest{
			FirstName: "John",
			LastName:  "Doe",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/api/profile", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("returns validation error from service", func(t *testing.T) {
		mockSvc := &mockService{
			updateProfileFunc: func(ctx context.Context, userID int64, req ProfileUpdateRequest) (*User, error) {
				// Simulate validation error from service
				return nil, errors.NewBadRequest("Certification expiry date is required when certified")
			},
		}

		handler := NewHandler(mockSvc)

		reqBody := ProfileUpdateRequest{
			FirstName: "John",
			LastName:  "Doe",
			Certified: true,
			// Missing CertExpiry
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/api/profile", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 1, Email: "test@example.com", Name: "Test User", Role: "referee"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.UpdateProfile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}
