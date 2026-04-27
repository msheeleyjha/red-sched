package eligibility

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	appErrors "github.com/msheeley/referee-scheduler/shared/errors"
)

// mockService implements ServiceInterface for testing
type mockService struct {
	GetEligibleRefereesFunc func(ctx context.Context, matchID int64, roleType string) ([]EligibleReferee, error)
}

func (m *mockService) GetEligibleReferees(ctx context.Context, matchID int64, roleType string) ([]EligibleReferee, error) {
	if m.GetEligibleRefereesFunc != nil {
		return m.GetEligibleRefereesFunc(ctx, matchID, roleType)
	}
	return nil, errors.New("GetEligibleReferees not implemented")
}

func TestGetEligibleRefereesHandler_Success(t *testing.T) {
	dob := "1990-01-01"
	certExpiry := "2028-12-31"
	grade := "Senior"
	age := 37

	mockReferees := []EligibleReferee{
		{
			ID:               100,
			FirstName:        "John",
			LastName:         "Doe",
			Email:            "john@example.com",
			Grade:            &grade,
			DateOfBirth:      &dob,
			Certified:        true,
			CertExpiry:       &certExpiry,
			AgeAtMatch:       &age,
			IsEligible:       true,
			IneligibleReason: nil,
			IsAvailable:      true,
		},
	}

	service := &mockService{
		GetEligibleRefereesFunc: func(ctx context.Context, matchID int64, roleType string) ([]EligibleReferee, error) {
			if matchID != 1 {
				t.Errorf("Expected matchID=1, got: %d", matchID)
			}
			if roleType != "center" {
				t.Errorf("Expected roleType=center, got: %s", roleType)
			}
			return mockReferees, nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/api/matches/1/eligible-referees", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handler.GetEligibleReferees(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rr.Code)
	}

	var response []EligibleReferee
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("Expected 1 referee, got: %d", len(response))
	}

	if response[0].ID != 100 {
		t.Errorf("Expected ID=100, got: %d", response[0].ID)
	}

	if !response[0].IsEligible {
		t.Error("Expected referee to be eligible")
	}
}

func TestGetEligibleRefereesHandler_WithRoleQuery(t *testing.T) {
	service := &mockService{
		GetEligibleRefereesFunc: func(ctx context.Context, matchID int64, roleType string) ([]EligibleReferee, error) {
			if roleType != "assistant_1" {
				t.Errorf("Expected roleType=assistant_1, got: %s", roleType)
			}
			return []EligibleReferee{}, nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/api/matches/1/eligible-referees?role=assistant_1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handler.GetEligibleReferees(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rr.Code)
	}
}

func TestGetEligibleRefereesHandler_DefaultsToCenter(t *testing.T) {
	service := &mockService{
		GetEligibleRefereesFunc: func(ctx context.Context, matchID int64, roleType string) ([]EligibleReferee, error) {
			if roleType != "center" {
				t.Errorf("Expected default roleType=center, got: %s", roleType)
			}
			return []EligibleReferee{}, nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/api/matches/1/eligible-referees", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handler.GetEligibleReferees(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rr.Code)
	}
}

func TestGetEligibleRefereesHandler_InvalidMatchID(t *testing.T) {
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

			req := httptest.NewRequest("GET", "/api/matches/"+tc.id+"/eligible-referees", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.id})

			rr := httptest.NewRecorder()
			handler.GetEligibleReferees(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got: %d", rr.Code)
			}
		})
	}
}

func TestGetEligibleRefereesHandler_InvalidRoleType(t *testing.T) {
	service := &mockService{
		GetEligibleRefereesFunc: func(ctx context.Context, matchID int64, roleType string) ([]EligibleReferee, error) {
			return nil, appErrors.NewBadRequest("Invalid role type. Must be: center, assistant_1, or assistant_2")
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/api/matches/1/eligible-referees?role=invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handler.GetEligibleReferees(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got: %d", rr.Code)
	}
}

func TestGetEligibleRefereesHandler_MatchNotFound(t *testing.T) {
	service := &mockService{
		GetEligibleRefereesFunc: func(ctx context.Context, matchID int64, roleType string) ([]EligibleReferee, error) {
			return nil, appErrors.NewNotFound("Match")
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/api/matches/999/eligible-referees", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})

	rr := httptest.NewRecorder()
	handler.GetEligibleReferees(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got: %d", rr.Code)
	}
}

func TestGetEligibleRefereesHandler_InternalError(t *testing.T) {
	service := &mockService{
		GetEligibleRefereesFunc: func(ctx context.Context, matchID int64, roleType string) ([]EligibleReferee, error) {
			return nil, appErrors.NewInternal("Failed to get match data", errors.New("database error"))
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/api/matches/1/eligible-referees", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handler.GetEligibleReferees(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got: %d", rr.Code)
	}
}

func TestGetEligibleRefereesHandler_EmptyResult(t *testing.T) {
	service := &mockService{
		GetEligibleRefereesFunc: func(ctx context.Context, matchID int64, roleType string) ([]EligibleReferee, error) {
			return []EligibleReferee{}, nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/api/matches/1/eligible-referees", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handler.GetEligibleReferees(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rr.Code)
	}

	var response []EligibleReferee
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 0 {
		t.Errorf("Expected 0 referees, got: %d", len(response))
	}
}

func TestGetEligibleRefereesHandler_MultipleReferees(t *testing.T) {
	dob1 := "1990-01-01"
	dob2 := "1995-06-15"
	age1 := 37
	age2 := 31

	mockReferees := []EligibleReferee{
		{
			ID:          100,
			FirstName:   "John",
			LastName:    "Doe",
			Email:       "john@example.com",
			DateOfBirth: &dob1,
			AgeAtMatch:  &age1,
			IsEligible:  true,
			IsAvailable: true,
		},
		{
			ID:          101,
			FirstName:   "Jane",
			LastName:    "Smith",
			Email:       "jane@example.com",
			DateOfBirth: &dob2,
			AgeAtMatch:  &age2,
			IsEligible:  false,
			IsAvailable: false,
		},
	}

	service := &mockService{
		GetEligibleRefereesFunc: func(ctx context.Context, matchID int64, roleType string) ([]EligibleReferee, error) {
			return mockReferees, nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/api/matches/1/eligible-referees", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handler.GetEligibleReferees(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rr.Code)
	}

	var response []EligibleReferee
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 referees, got: %d", len(response))
	}

	if response[0].IsEligible != true {
		t.Error("Expected first referee to be eligible")
	}

	if response[1].IsEligible != false {
		t.Error("Expected second referee to be not eligible")
	}
}
