package assignments

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/msheeley/referee-scheduler/shared/errors"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// mockService is a mock implementation of ServiceInterface for testing
type mockService struct {
	assignRefereeFunc  func(ctx context.Context, matchID int64, roleType string, req *AssignmentRequest, actorID int64) (*AssignmentResponse, error)
	checkConflictsFunc func(ctx context.Context, matchID int64, refereeID int64) (*ConflictCheckResponse, error)
}

func (m *mockService) AssignReferee(ctx context.Context, matchID int64, roleType string, req *AssignmentRequest, actorID int64) (*AssignmentResponse, error) {
	if m.assignRefereeFunc != nil {
		return m.assignRefereeFunc(ctx, matchID, roleType, req, actorID)
	}
	return nil, nil
}

func (m *mockService) CheckConflicts(ctx context.Context, matchID int64, refereeID int64) (*ConflictCheckResponse, error) {
	if m.checkConflictsFunc != nil {
		return m.checkConflictsFunc(ctx, matchID, refereeID)
	}
	return nil, nil
}

func (m *mockService) GetRefereeHistory(ctx context.Context, refereeID int64) ([]RefereeHistoryMatch, error) {
	return []RefereeHistoryMatch{}, nil
}

func (m *mockService) MarkMatchAsViewed(ctx context.Context, matchID int64, refereeID int64) error {
	return nil
}

func TestHandler_AssignReferee(t *testing.T) {
	t.Run("successfully assigns referee", func(t *testing.T) {
		mockSvc := &mockService{
			assignRefereeFunc: func(ctx context.Context, matchID int64, roleType string, req *AssignmentRequest, actorID int64) (*AssignmentResponse, error) {
				return &AssignmentResponse{
					Success: true,
					Action:  "assigned",
				}, nil
			},
		}

		handler := NewHandler(mockSvc)

		reqBody := AssignmentRequest{RefereeID: int64Ptr(10)}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/matches/1/roles/center/assign", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"match_id": "1", "role_type": "center"})
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 5, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.AssignReferee(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response AssignmentResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.Success {
			t.Error("Expected success to be true")
		}
		if response.Action != "assigned" {
			t.Errorf("Expected action 'assigned', got '%s'", response.Action)
		}
	})

	t.Run("successfully reassigns referee", func(t *testing.T) {
		mockSvc := &mockService{
			assignRefereeFunc: func(ctx context.Context, matchID int64, roleType string, req *AssignmentRequest, actorID int64) (*AssignmentResponse, error) {
				return &AssignmentResponse{
					Success: true,
					Action:  "reassigned",
				}, nil
			},
		}

		handler := NewHandler(mockSvc)

		reqBody := AssignmentRequest{RefereeID: int64Ptr(10)}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/matches/1/roles/center/assign", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"match_id": "1", "role_type": "center"})
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 5, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.AssignReferee(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response AssignmentResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Action != "reassigned" {
			t.Errorf("Expected action 'reassigned', got '%s'", response.Action)
		}
	})

	t.Run("successfully removes referee", func(t *testing.T) {
		mockSvc := &mockService{
			assignRefereeFunc: func(ctx context.Context, matchID int64, roleType string, req *AssignmentRequest, actorID int64) (*AssignmentResponse, error) {
				if req.RefereeID != nil {
					t.Error("Expected nil referee ID for removal")
				}
				return &AssignmentResponse{
					Success: true,
					Action:  "unassigned",
				}, nil
			},
		}

		handler := NewHandler(mockSvc)

		reqBody := AssignmentRequest{RefereeID: nil}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/matches/1/roles/center/assign", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"match_id": "1", "role_type": "center"})
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 5, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.AssignReferee(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response AssignmentResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Action != "unassigned" {
			t.Errorf("Expected action 'unassigned', got '%s'", response.Action)
		}
	})

	t.Run("returns error when user not in context", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("POST", "/api/matches/1/roles/center/assign", nil)
		req = mux.SetURLVars(req, map[string]string{"match_id": "1", "role_type": "center"})
		w := httptest.NewRecorder()

		handler.AssignReferee(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("returns error for invalid match ID", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("POST", "/api/matches/invalid/roles/center/assign", nil)
		req = mux.SetURLVars(req, map[string]string{"match_id": "invalid", "role_type": "center"})
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 5, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.AssignReferee(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns error for invalid request body", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("POST", "/api/matches/1/roles/center/assign", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"match_id": "1", "role_type": "center"})
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 5, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.AssignReferee(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns error from service", func(t *testing.T) {
		mockSvc := &mockService{
			assignRefereeFunc: func(ctx context.Context, matchID int64, roleType string, req *AssignmentRequest, actorID int64) (*AssignmentResponse, error) {
				return nil, errors.NewNotFound("Match")
			},
		}

		handler := NewHandler(mockSvc)

		reqBody := AssignmentRequest{RefereeID: int64Ptr(10)}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/matches/999/roles/center/assign", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"match_id": "999", "role_type": "center"})
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 5, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.AssignReferee(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

func TestHandler_CheckConflicts(t *testing.T) {
	t.Run("returns no conflicts", func(t *testing.T) {
		mockSvc := &mockService{
			checkConflictsFunc: func(ctx context.Context, matchID int64, refereeID int64) (*ConflictCheckResponse, error) {
				return &ConflictCheckResponse{
					HasConflict: false,
					Conflicts:   []ConflictMatch{},
				}, nil
			},
		}

		handler := NewHandler(mockSvc)

		req := httptest.NewRequest("GET", "/api/matches/1/conflicts?referee_id=10", nil)
		req = mux.SetURLVars(req, map[string]string{"match_id": "1"})
		w := httptest.NewRecorder()

		handler.CheckConflicts(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response ConflictCheckResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.HasConflict {
			t.Error("Expected no conflicts")
		}
	})

	t.Run("returns conflicts when they exist", func(t *testing.T) {
		mockSvc := &mockService{
			checkConflictsFunc: func(ctx context.Context, matchID int64, refereeID int64) (*ConflictCheckResponse, error) {
				return &ConflictCheckResponse{
					HasConflict: true,
					Conflicts: []ConflictMatch{
						{MatchID: 2, EventName: "Other League", TeamName: "Under 12 Boys"},
					},
				}, nil
			},
		}

		handler := NewHandler(mockSvc)

		req := httptest.NewRequest("GET", "/api/matches/1/conflicts?referee_id=10", nil)
		req = mux.SetURLVars(req, map[string]string{"match_id": "1"})
		w := httptest.NewRecorder()

		handler.CheckConflicts(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response ConflictCheckResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.HasConflict {
			t.Error("Expected conflicts to be detected")
		}
		if len(response.Conflicts) != 1 {
			t.Errorf("Expected 1 conflict, got %d", len(response.Conflicts))
		}
	})

	t.Run("returns error for invalid match ID", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("GET", "/api/matches/invalid/conflicts?referee_id=10", nil)
		req = mux.SetURLVars(req, map[string]string{"match_id": "invalid"})
		w := httptest.NewRecorder()

		handler.CheckConflicts(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns error for missing referee_id param", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("GET", "/api/matches/1/conflicts", nil)
		req = mux.SetURLVars(req, map[string]string{"match_id": "1"})
		w := httptest.NewRecorder()

		handler.CheckConflicts(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns error for invalid referee_id param", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("GET", "/api/matches/1/conflicts?referee_id=invalid", nil)
		req = mux.SetURLVars(req, map[string]string{"match_id": "1"})
		w := httptest.NewRecorder()

		handler.CheckConflicts(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns error from service", func(t *testing.T) {
		mockSvc := &mockService{
			checkConflictsFunc: func(ctx context.Context, matchID int64, refereeID int64) (*ConflictCheckResponse, error) {
				return nil, errors.NewNotFound("Match")
			},
		}

		handler := NewHandler(mockSvc)

		req := httptest.NewRequest("GET", "/api/matches/999/conflicts?referee_id=10", nil)
		req = mux.SetURLVars(req, map[string]string{"match_id": "999"})
		w := httptest.NewRecorder()

		handler.CheckConflicts(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}
