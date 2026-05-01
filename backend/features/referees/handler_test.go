package referees

import (
	"bytes"
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
	ListFunc   func(ctx context.Context) ([]RefereeListItem, error)
	UpdateFunc func(ctx context.Context, refereeID int64, currentUserID int64, req *UpdateRequest) (*UpdateResult, error)
}

func (m *mockService) List(ctx context.Context) ([]RefereeListItem, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, errors.New("List not implemented")
}

func (m *mockService) Update(ctx context.Context, refereeID int64, currentUserID int64, req *UpdateRequest) (*UpdateResult, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, refereeID, currentUserID, req)
	}
	return nil, errors.New("Update not implemented")
}

func TestListReferees_Success(t *testing.T) {
	now := time.Now()
	certExpiry := now.AddDate(0, 2, 0)

	mockReferees := []RefereeListItem{
		{
			ID:         1,
			Email:      "ref1@example.com",
			Name:       "Referee One",
			Role:       "referee",
			Status:     "active",
			CertStatus: "valid",
			CertExpiry: &certExpiry,
			CreatedAt:  now,
		},
		{
			ID:         2,
			Email:      "ref2@example.com",
			Name:       "Referee Two",
			Role:       "pending_referee",
			Status:     "pending",
			CertStatus: "none",
			CreatedAt:  now,
		},
	}

	service := &mockService{
		ListFunc: func(ctx context.Context) ([]RefereeListItem, error) {
			return mockReferees, nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/api/referees", nil)
	rr := httptest.NewRecorder()

	handler.ListReferees(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rr.Code)
	}

	var referees []RefereeListItem
	if err := json.NewDecoder(rr.Body).Decode(&referees); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(referees) != 2 {
		t.Errorf("Expected 2 referees, got: %d", len(referees))
	}

	if referees[0].Email != "ref1@example.com" {
		t.Errorf("Expected first referee email ref1@example.com, got: %s", referees[0].Email)
	}
}

func TestListReferees_ServiceError(t *testing.T) {
	service := &mockService{
		ListFunc: func(ctx context.Context) ([]RefereeListItem, error) {
			return nil, appErrors.NewInternal("Failed to list referees", errors.New("database error"))
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/api/referees", nil)
	rr := httptest.NewRecorder()

	handler.ListReferees(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got: %d", rr.Code)
	}
}

func TestUpdateReferee_Success(t *testing.T) {
	grade := "Senior"
	service := &mockService{
		UpdateFunc: func(ctx context.Context, refereeID int64, currentUserID int64, req *UpdateRequest) (*UpdateResult, error) {
			if refereeID != 1 {
				t.Errorf("Expected refereeID 1, got: %d", refereeID)
			}
			if currentUserID != 100 {
				t.Errorf("Expected currentUserID 100, got: %d", currentUserID)
			}
			if req.Grade == nil || *req.Grade != "Senior" {
				t.Errorf("Expected grade Senior, got: %v", req.Grade)
			}

			return &UpdateResult{
				ID:     1,
				Email:  "ref@example.com",
				Name:   "Test Referee",
				Role:   "referee",
				Status: "active",
				Grade:  &grade,
			}, nil
		},
	}

	handler := NewHandler(service)

	reqBody := UpdateRequest{
		Grade: &grade,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/referees/1", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:    100,
		Email: "assignor@example.com",
		Name:  "Test Assignor",
		Role:  "assignor",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateReferee(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rr.Code)
	}

	var result UpdateResult
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Grade == nil || *result.Grade != "Senior" {
		t.Errorf("Expected grade Senior, got: %v", result.Grade)
	}
}

func TestUpdateReferee_UserNotInContext(t *testing.T) {
	service := &mockService{}
	handler := NewHandler(service)

	reqBody := UpdateRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/referees/1", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handler.UpdateReferee(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got: %d", rr.Code)
	}
}

func TestUpdateReferee_InvalidRefereeID(t *testing.T) {
	testCases := []struct {
		name string
		id   string
	}{
		{"Non-numeric ID", "abc"},
		{"Float ID", "1.5"},
		{"Empty ID", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &mockService{}
			handler := NewHandler(service)

			reqBody := UpdateRequest{}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("PUT", "/api/referees/"+tc.id, bytes.NewReader(body))
			req = mux.SetURLVars(req, map[string]string{"id": tc.id})

			ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
				ID:   100,
				Role: "assignor",
			})
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.UpdateReferee(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got: %d", rr.Code)
			}
		})
	}
}

func TestUpdateReferee_InvalidRequestBody(t *testing.T) {
	service := &mockService{}
	handler := NewHandler(service)

	req := httptest.NewRequest("PUT", "/api/referees/1", bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:   100,
		Role: "assignor",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateReferee(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got: %d", rr.Code)
	}
}

func TestUpdateReferee_RefereeNotFound(t *testing.T) {
	service := &mockService{
		UpdateFunc: func(ctx context.Context, refereeID int64, currentUserID int64, req *UpdateRequest) (*UpdateResult, error) {
			return nil, appErrors.NewNotFound("Referee")
		},
	}

	handler := NewHandler(service)

	status := "active"
	reqBody := UpdateRequest{
		Status: &status,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/referees/999", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "999"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:   100,
		Role: "assignor",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateReferee(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got: %d", rr.Code)
	}
}

func TestUpdateReferee_CannotModifyOtherAssignor(t *testing.T) {
	service := &mockService{
		UpdateFunc: func(ctx context.Context, refereeID int64, currentUserID int64, req *UpdateRequest) (*UpdateResult, error) {
			return nil, appErrors.NewForbidden("Cannot modify other assignor accounts")
		},
	}

	handler := NewHandler(service)

	status := "inactive"
	reqBody := UpdateRequest{
		Status: &status,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/referees/2", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "2"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:   1,
		Role: "assignor",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateReferee(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got: %d", rr.Code)
	}
}

func TestUpdateReferee_CannotDeactivateSelf(t *testing.T) {
	service := &mockService{
		UpdateFunc: func(ctx context.Context, refereeID int64, currentUserID int64, req *UpdateRequest) (*UpdateResult, error) {
			return nil, appErrors.NewForbidden("Cannot deactivate your own account")
		},
	}

	handler := NewHandler(service)

	status := "inactive"
	reqBody := UpdateRequest{
		Status: &status,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/referees/1", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:   1,
		Role: "assignor",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateReferee(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got: %d", rr.Code)
	}
}

func TestUpdateReferee_CannotDeactivateWithUpcomingAssignments(t *testing.T) {
	service := &mockService{
		UpdateFunc: func(ctx context.Context, refereeID int64, currentUserID int64, req *UpdateRequest) (*UpdateResult, error) {
			return nil, appErrors.NewBadRequest("Cannot deactivate user with upcoming match assignments")
		},
	}

	handler := NewHandler(service)

	status := "inactive"
	reqBody := UpdateRequest{
		Status: &status,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/referees/2", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "2"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:   1,
		Role: "assignor",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateReferee(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got: %d", rr.Code)
	}
}

func TestUpdateReferee_InvalidStatus(t *testing.T) {
	service := &mockService{
		UpdateFunc: func(ctx context.Context, refereeID int64, currentUserID int64, req *UpdateRequest) (*UpdateResult, error) {
			return nil, appErrors.NewBadRequest("Invalid status. Must be: pending, active, inactive, or removed")
		},
	}

	handler := NewHandler(service)

	status := "invalid_status"
	reqBody := UpdateRequest{
		Status: &status,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/referees/1", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:   100,
		Role: "assignor",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateReferee(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got: %d", rr.Code)
	}
}

func TestUpdateReferee_InvalidGrade(t *testing.T) {
	service := &mockService{
		UpdateFunc: func(ctx context.Context, refereeID int64, currentUserID int64, req *UpdateRequest) (*UpdateResult, error) {
			return nil, appErrors.NewBadRequest("Invalid grade. Must be: Junior, Mid, or Senior")
		},
	}

	handler := NewHandler(service)

	grade := "Expert"
	reqBody := UpdateRequest{
		Grade: &grade,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/referees/1", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:   100,
		Role: "assignor",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateReferee(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got: %d", rr.Code)
	}
}
