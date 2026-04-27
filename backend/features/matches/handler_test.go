package matches

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/msheeley/referee-scheduler/shared/errors"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// mockService is a mock implementation of ServiceInterface for testing
type mockService struct {
	parseCSVFunc            func(file multipart.File, filename string) (*ImportPreviewResponse, error)
	importMatchesFunc       func(ctx context.Context, req *ImportConfirmRequest, currentUserID int64) (*ImportResult, error)
	createRoleSlotsFunc     func(ctx context.Context, matchID int64, ageGroup string) error
	listMatchesFunc         func(ctx context.Context) ([]MatchWithRoles, error)
	getMatchWithRolesFunc   func(ctx context.Context, matchID int64) (*MatchWithRoles, error)
	updateMatchFunc         func(ctx context.Context, matchID int64, req *MatchUpdateRequest, actorID int64) (*MatchWithRoles, error)
	addRoleSlotFunc         func(ctx context.Context, matchID int64, roleType string) error
}

func (m *mockService) ParseCSV(file multipart.File, filename string) (*ImportPreviewResponse, error) {
	if m.parseCSVFunc != nil {
		return m.parseCSVFunc(file, filename)
	}
	return nil, nil
}

func (m *mockService) ImportMatches(ctx context.Context, req *ImportConfirmRequest, currentUserID int64) (*ImportResult, error) {
	if m.importMatchesFunc != nil {
		return m.importMatchesFunc(ctx, req, currentUserID)
	}
	return nil, nil
}

func (m *mockService) CreateRoleSlotsForMatch(ctx context.Context, matchID int64, ageGroup string) error {
	if m.createRoleSlotsFunc != nil {
		return m.createRoleSlotsFunc(ctx, matchID, ageGroup)
	}
	return nil
}

func (m *mockService) ListMatches(ctx context.Context) ([]MatchWithRoles, error) {
	if m.listMatchesFunc != nil {
		return m.listMatchesFunc(ctx)
	}
	return []MatchWithRoles{}, nil
}

func (m *mockService) GetMatchWithRoles(ctx context.Context, matchID int64) (*MatchWithRoles, error) {
	if m.getMatchWithRolesFunc != nil {
		return m.getMatchWithRolesFunc(ctx, matchID)
	}
	return nil, nil
}

func (m *mockService) UpdateMatch(ctx context.Context, matchID int64, req *MatchUpdateRequest, actorID int64) (*MatchWithRoles, error) {
	if m.updateMatchFunc != nil {
		return m.updateMatchFunc(ctx, matchID, req, actorID)
	}
	return nil, nil
}

func (m *mockService) AddRoleSlot(ctx context.Context, matchID int64, roleType string) error {
	if m.addRoleSlotFunc != nil {
		return m.addRoleSlotFunc(ctx, matchID, roleType)
	}
	return nil
}

func TestHandler_ParseCSV(t *testing.T) {
	t.Run("successfully parses CSV", func(t *testing.T) {
		ageGroup := "U12"
		mockSvc := &mockService{
			parseCSVFunc: func(file multipart.File, filename string) (*ImportPreviewResponse, error) {
				return &ImportPreviewResponse{
					Rows: []CSVRow{
						{EventName: "Spring League", TeamName: "Under 12 Girls", AgeGroup: &ageGroup},
					},
					Duplicates: []DuplicateMatchGroup{},
				}, nil
			},
		}

		handler := NewHandler(mockSvc)

		// Create multipart form
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.csv")
		part.Write([]byte("event_name,team_name,start_date,start_time,end_time,location\nSpring,Under 12 Girls,2027-05-15,10:00,11:30,Field A"))
		writer.Close()

		req := httptest.NewRequest("POST", "/api/matches/import/parse", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		handler.ParseCSV(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response ImportPreviewResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Rows) != 1 {
			t.Errorf("Expected 1 row, got %d", len(response.Rows))
		}
	})

	t.Run("returns error for invalid form", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("POST", "/api/matches/import/parse", nil)
		w := httptest.NewRecorder()

		handler.ParseCSV(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		// Create multipart form without file
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest("POST", "/api/matches/import/parse", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		handler.ParseCSV(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns error from service", func(t *testing.T) {
		mockSvc := &mockService{
			parseCSVFunc: func(file multipart.File, filename string) (*ImportPreviewResponse, error) {
				return nil, errors.NewBadRequest("Invalid CSV format")
			},
		}

		handler := NewHandler(mockSvc)

		// Create multipart form
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.csv")
		part.Write([]byte("invalid"))
		writer.Close()

		req := httptest.NewRequest("POST", "/api/matches/import/parse", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		handler.ParseCSV(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestHandler_ImportMatches(t *testing.T) {
	t.Run("successfully imports matches", func(t *testing.T) {
		mockSvc := &mockService{
			importMatchesFunc: func(ctx context.Context, req *ImportConfirmRequest, currentUserID int64) (*ImportResult, error) {
				return &ImportResult{
					Imported: 5,
					Skipped:  2,
					Errors:   []string{},
				}, nil
			},
		}

		handler := NewHandler(mockSvc)

		reqBody := ImportConfirmRequest{
			Rows: []CSVRow{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/matches/import/confirm", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 1, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.ImportMatches(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response ImportResult
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Imported != 5 {
			t.Errorf("Expected 5 imported, got %d", response.Imported)
		}
	})

	t.Run("returns error when user not in context", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("POST", "/api/matches/import/confirm", nil)
		w := httptest.NewRecorder()

		handler.ImportMatches(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("returns error for invalid request body", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("POST", "/api/matches/import/confirm", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 1, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.ImportMatches(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns error from service", func(t *testing.T) {
		mockSvc := &mockService{
			importMatchesFunc: func(ctx context.Context, req *ImportConfirmRequest, currentUserID int64) (*ImportResult, error) {
				return nil, errors.NewInternal("Database error", nil)
			},
		}

		handler := NewHandler(mockSvc)

		reqBody := ImportConfirmRequest{Rows: []CSVRow{}}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/matches/import/confirm", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 1, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.ImportMatches(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_ListMatches(t *testing.T) {
	t.Run("returns list of matches", func(t *testing.T) {
		ageGroup := "U12"
		mockSvc := &mockService{
			listMatchesFunc: func(ctx context.Context) ([]MatchWithRoles, error) {
				return []MatchWithRoles{
					{
						Match:            Match{ID: 1, EventName: "Spring League", TeamName: "Under 12 Girls", AgeGroup: &ageGroup},
						Roles:            []MatchRole{},
						AssignmentStatus: "unassigned",
					},
				}, nil
			},
		}

		handler := NewHandler(mockSvc)

		req := httptest.NewRequest("GET", "/api/matches", nil)
		w := httptest.NewRecorder()

		handler.ListMatches(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response []MatchWithRoles
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response) != 1 {
			t.Errorf("Expected 1 match, got %d", len(response))
		}
		if response[0].EventName != "Spring League" {
			t.Errorf("Expected 'Spring League', got '%s'", response[0].EventName)
		}
	})

	t.Run("returns error from service", func(t *testing.T) {
		mockSvc := &mockService{
			listMatchesFunc: func(ctx context.Context) ([]MatchWithRoles, error) {
				return nil, errors.NewInternal("Database error", nil)
			},
		}

		handler := NewHandler(mockSvc)

		req := httptest.NewRequest("GET", "/api/matches", nil)
		w := httptest.NewRecorder()

		handler.ListMatches(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_UpdateMatch(t *testing.T) {
	t.Run("successfully updates match", func(t *testing.T) {
		ageGroup := "U12"
		mockSvc := &mockService{
			updateMatchFunc: func(ctx context.Context, matchID int64, req *MatchUpdateRequest, actorID int64) (*MatchWithRoles, error) {
				return &MatchWithRoles{
					Match: Match{ID: matchID, EventName: "Updated League", TeamName: "Under 12 Girls", AgeGroup: &ageGroup},
				}, nil
			},
		}

		handler := NewHandler(mockSvc)

		reqBody := MatchUpdateRequest{
			EventName: stringPtr("Updated League"),
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/api/matches/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 1, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.UpdateMatch(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response MatchWithRoles
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.EventName != "Updated League" {
			t.Errorf("Expected 'Updated League', got '%s'", response.EventName)
		}
	})

	t.Run("returns error for invalid match ID", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("PUT", "/api/matches/invalid", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 1, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.UpdateMatch(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns error when user not in context", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("PUT", "/api/matches/1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		w := httptest.NewRecorder()

		handler.UpdateMatch(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("returns error for invalid request body", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("PUT", "/api/matches/1", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 1, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.UpdateMatch(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns error from service", func(t *testing.T) {
		mockSvc := &mockService{
			updateMatchFunc: func(ctx context.Context, matchID int64, req *MatchUpdateRequest, actorID int64) (*MatchWithRoles, error) {
				return nil, errors.NewNotFound("Match")
			},
		}

		handler := NewHandler(mockSvc)

		reqBody := MatchUpdateRequest{EventName: stringPtr("Updated")}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/api/matches/999", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"id": "999"})
		w := httptest.NewRecorder()

		// Add user to context
		user := &middleware.User{ID: 1, Email: "test@example.com", Name: "Test", Role: "assignor"}
		ctx := middleware.SetUserInContext(req.Context(), user)
		req = req.WithContext(ctx)

		handler.UpdateMatch(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

func TestHandler_AddRoleSlot(t *testing.T) {
	t.Run("successfully adds role slot", func(t *testing.T) {
		mockSvc := &mockService{
			addRoleSlotFunc: func(ctx context.Context, matchID int64, roleType string) error {
				return nil
			},
		}

		handler := NewHandler(mockSvc)

		req := httptest.NewRequest("POST", "/api/matches/1/roles/assistant_1", nil)
		req = mux.SetURLVars(req, map[string]string{"match_id": "1", "role_type": "assistant_1"})
		w := httptest.NewRecorder()

		handler.AddRoleSlot(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["success"] != true {
			t.Error("Expected success to be true")
		}
		if response["role_type"] != "assistant_1" {
			t.Errorf("Expected role_type 'assistant_1', got '%v'", response["role_type"])
		}
	})

	t.Run("returns error for invalid match ID", func(t *testing.T) {
		handler := NewHandler(&mockService{})

		req := httptest.NewRequest("POST", "/api/matches/invalid/roles/assistant_1", nil)
		req = mux.SetURLVars(req, map[string]string{"match_id": "invalid", "role_type": "assistant_1"})
		w := httptest.NewRecorder()

		handler.AddRoleSlot(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns error from service", func(t *testing.T) {
		mockSvc := &mockService{
			addRoleSlotFunc: func(ctx context.Context, matchID int64, roleType string) error {
				return errors.NewBadRequest("Can only add assistant referee slots")
			},
		}

		handler := NewHandler(mockSvc)

		req := httptest.NewRequest("POST", "/api/matches/1/roles/center", nil)
		req = mux.SetURLVars(req, map[string]string{"match_id": "1", "role_type": "center"})
		w := httptest.NewRecorder()

		handler.AddRoleSlot(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}
