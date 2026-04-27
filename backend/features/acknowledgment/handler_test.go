package acknowledgment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	appErrors "github.com/msheeley/referee-scheduler/shared/errors"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// mockService implements ServiceInterface for testing
type mockService struct {
	AcknowledgeAssignmentFunc func(ctx context.Context, matchID int64, refereeID int64) (*AcknowledgeResponse, error)
}

func (m *mockService) AcknowledgeAssignment(ctx context.Context, matchID int64, refereeID int64) (*AcknowledgeResponse, error) {
	if m.AcknowledgeAssignmentFunc != nil {
		return m.AcknowledgeAssignmentFunc(ctx, matchID, refereeID)
	}
	return nil, errors.New("AcknowledgeAssignment not implemented")
}

func TestAcknowledgeAssignmentHandler_Success(t *testing.T) {
	now := time.Now()
	service := &mockService{
		AcknowledgeAssignmentFunc: func(ctx context.Context, matchID int64, refereeID int64) (*AcknowledgeResponse, error) {
			if matchID == 1 && refereeID == 100 {
				return &AcknowledgeResponse{
					Success:        true,
					AcknowledgedAt: now,
				}, nil
			}
			return nil, errors.New("unexpected call")
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("POST", "/api/referee/matches/1/acknowledge", nil)
	req = mux.SetURLVars(req, map[string]string{"match_id": "1"})

	// Add user to context
	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:    100,
		Email: "referee@example.com",
		Name:  "Test Referee",
		Role:  "referee",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.AcknowledgeAssignment(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rr.Code)
	}

	var response AcknowledgeResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.AcknowledgedAt.IsZero() {
		t.Error("Expected acknowledged_at to be set")
	}
}

func TestAcknowledgeAssignmentHandler_AssignorCanAcknowledge(t *testing.T) {
	now := time.Now()
	service := &mockService{
		AcknowledgeAssignmentFunc: func(ctx context.Context, matchID int64, refereeID int64) (*AcknowledgeResponse, error) {
			return &AcknowledgeResponse{
				Success:        true,
				AcknowledgedAt: now,
			}, nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("POST", "/api/referee/matches/1/acknowledge", nil)
	req = mux.SetURLVars(req, map[string]string{"match_id": "1"})

	// Add assignor to context
	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:    200,
		Email: "assignor@example.com",
		Name:  "Test Assignor",
		Role:  "assignor",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.AcknowledgeAssignment(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rr.Code)
	}
}

func TestAcknowledgeAssignmentHandler_UserNotInContext(t *testing.T) {
	service := &mockService{}
	handler := NewHandler(service)

	req := httptest.NewRequest("POST", "/api/referee/matches/1/acknowledge", nil)
	req = mux.SetURLVars(req, map[string]string{"match_id": "1"})

	rr := httptest.NewRecorder()
	handler.AcknowledgeAssignment(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got: %d", rr.Code)
	}
}

func TestAcknowledgeAssignmentHandler_ForbiddenRole(t *testing.T) {
	testCases := []struct {
		name string
		role string
	}{
		{"Admin role", "admin"},
		{"Unknown role", "unknown"},
		{"Empty role", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &mockService{}
			handler := NewHandler(service)

			req := httptest.NewRequest("POST", "/api/referee/matches/1/acknowledge", nil)
			req = mux.SetURLVars(req, map[string]string{"match_id": "1"})

			ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
				ID:    100,
				Email: "user@example.com",
				Name:  "Test User",
				Role:  tc.role,
			})
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.AcknowledgeAssignment(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Errorf("Expected status 403, got: %d", rr.Code)
			}
		})
	}
}

func TestAcknowledgeAssignmentHandler_InvalidMatchID(t *testing.T) {
	testCases := []struct {
		name    string
		matchID string
	}{
		{"Non-numeric ID", "abc"},
		{"Float ID", "1.5"},
		{"Empty ID", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &mockService{}
			handler := NewHandler(service)

			req := httptest.NewRequest("POST", "/api/referee/matches/"+tc.matchID+"/acknowledge", nil)
			req = mux.SetURLVars(req, map[string]string{"match_id": tc.matchID})

			ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
				ID:    100,
				Email: "referee@example.com",
				Name:  "Test Referee",
				Role:  "referee",
			})
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.AcknowledgeAssignment(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got: %d", rr.Code)
			}
		})
	}
}

func TestAcknowledgeAssignmentHandler_NotAssigned(t *testing.T) {
	service := &mockService{
		AcknowledgeAssignmentFunc: func(ctx context.Context, matchID int64, refereeID int64) (*AcknowledgeResponse, error) {
			return nil, appErrors.NewNotFound("Assignment")
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("POST", "/api/referee/matches/1/acknowledge", nil)
	req = mux.SetURLVars(req, map[string]string{"match_id": "1"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:    100,
		Email: "referee@example.com",
		Name:  "Test Referee",
		Role:  "referee",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.AcknowledgeAssignment(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got: %d", rr.Code)
	}
}

func TestAcknowledgeAssignmentHandler_InternalError(t *testing.T) {
	service := &mockService{
		AcknowledgeAssignmentFunc: func(ctx context.Context, matchID int64, refereeID int64) (*AcknowledgeResponse, error) {
			return nil, appErrors.NewInternal("Failed to acknowledge assignment", errors.New("database error"))
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("POST", "/api/referee/matches/1/acknowledge", nil)
	req = mux.SetURLVars(req, map[string]string{"match_id": "1"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:    100,
		Email: "referee@example.com",
		Name:  "Test Referee",
		Role:  "referee",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.AcknowledgeAssignment(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got: %d", rr.Code)
	}
}

func TestAcknowledgeAssignmentHandler_MultipleAcknowledgments(t *testing.T) {
	callCount := 0
	service := &mockService{
		AcknowledgeAssignmentFunc: func(ctx context.Context, matchID int64, refereeID int64) (*AcknowledgeResponse, error) {
			callCount++
			return &AcknowledgeResponse{
				Success:        true,
				AcknowledgedAt: time.Now(),
			}, nil
		},
	}

	handler := NewHandler(service)

	// First acknowledgment
	req1 := httptest.NewRequest("POST", "/api/referee/matches/1/acknowledge", nil)
	req1 = mux.SetURLVars(req1, map[string]string{"match_id": "1"})
	ctx1 := middleware.SetUserInContext(req1.Context(), &middleware.User{
		ID:   100,
		Role: "referee",
	})
	req1 = req1.WithContext(ctx1)

	rr1 := httptest.NewRecorder()
	handler.AcknowledgeAssignment(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("First acknowledgment failed with status: %d", rr1.Code)
	}

	// Second acknowledgment (should still succeed - idempotent)
	req2 := httptest.NewRequest("POST", "/api/referee/matches/1/acknowledge", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"match_id": "1"})
	ctx2 := middleware.SetUserInContext(req2.Context(), &middleware.User{
		ID:   100,
		Role: "referee",
	})
	req2 = req2.WithContext(ctx2)

	rr2 := httptest.NewRecorder()
	handler.AcknowledgeAssignment(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Second acknowledgment failed with status: %d", rr2.Code)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 service calls, got: %d", callCount)
	}
}
