package availability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	appErrors "github.com/msheeley/referee-scheduler/shared/errors"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// mockService implements ServiceInterface for testing
type mockService struct {
	ToggleMatchAvailabilityFunc  func(ctx context.Context, matchID int64, refereeID int64, req *ToggleMatchAvailabilityRequest) (*ToggleMatchAvailabilityResponse, error)
	GetDayUnavailabilityFunc     func(ctx context.Context, refereeID int64) ([]DayUnavailability, error)
	ToggleDayUnavailabilityFunc  func(ctx context.Context, refereeID int64, date string, req *ToggleDayUnavailabilityRequest) (*ToggleDayUnavailabilityResponse, error)
}

func (m *mockService) ToggleMatchAvailability(ctx context.Context, matchID int64, refereeID int64, req *ToggleMatchAvailabilityRequest) (*ToggleMatchAvailabilityResponse, error) {
	if m.ToggleMatchAvailabilityFunc != nil {
		return m.ToggleMatchAvailabilityFunc(ctx, matchID, refereeID, req)
	}
	return nil, errors.New("ToggleMatchAvailability not implemented")
}

func (m *mockService) GetDayUnavailability(ctx context.Context, refereeID int64) ([]DayUnavailability, error) {
	if m.GetDayUnavailabilityFunc != nil {
		return m.GetDayUnavailabilityFunc(ctx, refereeID)
	}
	return nil, errors.New("GetDayUnavailability not implemented")
}

func (m *mockService) ToggleDayUnavailability(ctx context.Context, refereeID int64, date string, req *ToggleDayUnavailabilityRequest) (*ToggleDayUnavailabilityResponse, error) {
	if m.ToggleDayUnavailabilityFunc != nil {
		return m.ToggleDayUnavailabilityFunc(ctx, refereeID, date, req)
	}
	return nil, errors.New("ToggleDayUnavailability not implemented")
}

func TestToggleMatchAvailabilityHandler_Success(t *testing.T) {
	available := true
	service := &mockService{
		ToggleMatchAvailabilityFunc: func(ctx context.Context, matchID int64, refereeID int64, req *ToggleMatchAvailabilityRequest) (*ToggleMatchAvailabilityResponse, error) {
			if matchID != 1 || refereeID != 100 {
				t.Errorf("Unexpected parameters: matchID=%d, refereeID=%d", matchID, refereeID)
			}
			return &ToggleMatchAvailabilityResponse{
				Success:   true,
				Available: &available,
			}, nil
		},
	}

	handler := NewHandler(service)

	reqBody := ToggleMatchAvailabilityRequest{
		Available: &available,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/referee/matches/1/availability", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:    100,
		Email: "referee@example.com",
		Name:  "Test Referee",
		Role:  "referee",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ToggleMatchAvailability(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rr.Code)
	}

	var response ToggleMatchAvailabilityResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.Available == nil || *response.Available != true {
		t.Errorf("Expected available=true, got: %v", response.Available)
	}
}

func TestToggleMatchAvailabilityHandler_UserNotInContext(t *testing.T) {
	service := &mockService{}
	handler := NewHandler(service)

	reqBody := ToggleMatchAvailabilityRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/referee/matches/1/availability", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handler.ToggleMatchAvailability(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got: %d", rr.Code)
	}
}

func TestToggleMatchAvailabilityHandler_InvalidMatchID(t *testing.T) {
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

			reqBody := ToggleMatchAvailabilityRequest{}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/api/referee/matches/"+tc.id+"/availability", bytes.NewReader(body))
			req = mux.SetURLVars(req, map[string]string{"id": tc.id})

			ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
				ID:   100,
				Role: "referee",
			})
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.ToggleMatchAvailability(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got: %d", rr.Code)
			}
		})
	}
}

func TestToggleMatchAvailabilityHandler_InvalidRequestBody(t *testing.T) {
	service := &mockService{}
	handler := NewHandler(service)

	req := httptest.NewRequest("POST", "/api/referee/matches/1/availability", bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:   100,
		Role: "referee",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ToggleMatchAvailability(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got: %d", rr.Code)
	}
}

func TestToggleMatchAvailabilityHandler_MatchNotFound(t *testing.T) {
	available := true
	service := &mockService{
		ToggleMatchAvailabilityFunc: func(ctx context.Context, matchID int64, refereeID int64, req *ToggleMatchAvailabilityRequest) (*ToggleMatchAvailabilityResponse, error) {
			return nil, appErrors.NewNotFound("Match not found or not available for marking")
		},
	}

	handler := NewHandler(service)

	reqBody := ToggleMatchAvailabilityRequest{
		Available: &available,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/referee/matches/999/availability", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "999"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:   100,
		Role: "referee",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ToggleMatchAvailability(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got: %d", rr.Code)
	}
}

func TestGetDayUnavailabilityHandler_Success(t *testing.T) {
	reason := "Vacation"
	mockDays := []DayUnavailability{
		{
			ID:              1,
			RefereeID:       100,
			UnavailableDate: "2027-12-31",
			Reason:          &reason,
			CreatedAt:       "2026-04-27T00:00:00Z",
		},
		{
			ID:              2,
			RefereeID:       100,
			UnavailableDate: "2028-01-01",
			Reason:          nil,
			CreatedAt:       "2026-04-27T00:00:00Z",
		},
	}

	service := &mockService{
		GetDayUnavailabilityFunc: func(ctx context.Context, refereeID int64) ([]DayUnavailability, error) {
			if refereeID != 100 {
				t.Errorf("Expected refereeID=100, got: %d", refereeID)
			}
			return mockDays, nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/api/referee/day-unavailability", nil)

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:    100,
		Email: "referee@example.com",
		Name:  "Test Referee",
		Role:  "referee",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetDayUnavailability(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rr.Code)
	}

	var response []DayUnavailability
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 days, got: %d", len(response))
	}

	if response[0].UnavailableDate != "2027-12-31" {
		t.Errorf("Expected date=2027-12-31, got: %s", response[0].UnavailableDate)
	}

	if response[0].Reason == nil || *response[0].Reason != "Vacation" {
		t.Errorf("Expected reason=Vacation, got: %v", response[0].Reason)
	}
}

func TestGetDayUnavailabilityHandler_UserNotInContext(t *testing.T) {
	service := &mockService{}
	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/api/referee/day-unavailability", nil)

	rr := httptest.NewRecorder()
	handler.GetDayUnavailability(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got: %d", rr.Code)
	}
}

func TestToggleDayUnavailabilityHandler_Success(t *testing.T) {
	reason := "Out of town"
	service := &mockService{
		ToggleDayUnavailabilityFunc: func(ctx context.Context, refereeID int64, date string, req *ToggleDayUnavailabilityRequest) (*ToggleDayUnavailabilityResponse, error) {
			if refereeID != 100 {
				t.Errorf("Expected refereeID=100, got: %d", refereeID)
			}
			if date != "2027-12-31" {
				t.Errorf("Expected date=2027-12-31, got: %s", date)
			}
			return &ToggleDayUnavailabilityResponse{
				Success:     true,
				Unavailable: true,
				Date:        date,
			}, nil
		},
	}

	handler := NewHandler(service)

	reqBody := ToggleDayUnavailabilityRequest{
		Unavailable: true,
		Reason:      &reason,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/referee/day-unavailability/2027-12-31", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"date": "2027-12-31"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:    100,
		Email: "referee@example.com",
		Name:  "Test Referee",
		Role:  "referee",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ToggleDayUnavailability(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rr.Code)
	}

	var response ToggleDayUnavailabilityResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if !response.Unavailable {
		t.Error("Expected unavailable to be true")
	}

	if response.Date != "2027-12-31" {
		t.Errorf("Expected date=2027-12-31, got: %s", response.Date)
	}
}

func TestToggleDayUnavailabilityHandler_InvalidDate(t *testing.T) {
	service := &mockService{
		ToggleDayUnavailabilityFunc: func(ctx context.Context, refereeID int64, date string, req *ToggleDayUnavailabilityRequest) (*ToggleDayUnavailabilityResponse, error) {
			return nil, appErrors.NewBadRequest("Invalid date format (expected YYYY-MM-DD)")
		},
	}

	handler := NewHandler(service)

	reqBody := ToggleDayUnavailabilityRequest{
		Unavailable: true,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/referee/day-unavailability/invalid-date", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"date": "invalid-date"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:   100,
		Role: "referee",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ToggleDayUnavailability(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got: %d", rr.Code)
	}
}

func TestToggleDayUnavailabilityHandler_UserNotInContext(t *testing.T) {
	service := &mockService{}
	handler := NewHandler(service)

	reqBody := ToggleDayUnavailabilityRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/referee/day-unavailability/2027-12-31", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"date": "2027-12-31"})

	rr := httptest.NewRecorder()
	handler.ToggleDayUnavailability(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got: %d", rr.Code)
	}
}

func TestToggleDayUnavailabilityHandler_InvalidRequestBody(t *testing.T) {
	service := &mockService{}
	handler := NewHandler(service)

	req := httptest.NewRequest("POST", "/api/referee/day-unavailability/2027-12-31", bytes.NewReader([]byte("invalid json")))
	req = mux.SetURLVars(req, map[string]string{"date": "2027-12-31"})

	ctx := middleware.SetUserInContext(req.Context(), &middleware.User{
		ID:   100,
		Role: "referee",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ToggleDayUnavailability(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got: %d", rr.Code)
	}
}
