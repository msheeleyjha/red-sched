package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewBadRequest(t *testing.T) {
	err := NewBadRequest("invalid input")

	if err.Message != "invalid input" {
		t.Errorf("Expected message 'invalid input', got '%s'", err.Message)
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, err.StatusCode)
	}
}

func TestNewUnauthorized(t *testing.T) {
	err := NewUnauthorized("authentication required")

	if err.Message != "authentication required" {
		t.Errorf("Expected message 'authentication required', got '%s'", err.Message)
	}
	if err.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, err.StatusCode)
	}
}

func TestNewForbidden(t *testing.T) {
	err := NewForbidden("access denied")

	if err.Message != "access denied" {
		t.Errorf("Expected message 'access denied', got '%s'", err.Message)
	}
	if err.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, err.StatusCode)
	}
}

func TestNewNotFound(t *testing.T) {
	err := NewNotFound("User")

	if err.Message != "User not found" {
		t.Errorf("Expected message 'User not found', got '%s'", err.Message)
	}
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, err.StatusCode)
	}
}

func TestNewConflict(t *testing.T) {
	err := NewConflict("resource already exists")

	if err.Message != "resource already exists" {
		t.Errorf("Expected message 'resource already exists', got '%s'", err.Message)
	}
	if err.StatusCode != http.StatusConflict {
		t.Errorf("Expected status code %d, got %d", http.StatusConflict, err.StatusCode)
	}
}

func TestNewInternal(t *testing.T) {
	underlyingErr := errors.New("database connection failed")
	err := NewInternal("internal error", underlyingErr)

	if err.Message != "internal error" {
		t.Errorf("Expected message 'internal error', got '%s'", err.Message)
	}
	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, err.StatusCode)
	}
	if err.Err != underlyingErr {
		t.Error("Expected underlying error to be stored")
	}
}

func TestAppError_Error(t *testing.T) {
	t.Run("error without underlying error", func(t *testing.T) {
		err := &AppError{
			Message:    "test error",
			StatusCode: http.StatusBadRequest,
		}

		expected := "test error"
		if err.Error() != expected {
			t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("error with underlying error", func(t *testing.T) {
		underlyingErr := errors.New("database error")
		err := &AppError{
			Message:    "test error",
			StatusCode: http.StatusInternalServerError,
			Err:        underlyingErr,
		}

		expected := "test error: database error"
		if err.Error() != expected {
			t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
		}
	})
}

func TestAppError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	err := &AppError{
		Message:    "test error",
		StatusCode: http.StatusInternalServerError,
		Err:        underlyingErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != underlyingErr {
		t.Error("Expected Unwrap to return underlying error")
	}
}

func TestWriteError(t *testing.T) {
	t.Run("writes AppError with correct status code and JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := NewBadRequest("invalid input")

		WriteError(w, err)

		// Check status code
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		// Check content type
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Check response body
		var response map[string]string
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["error"] != "invalid input" {
			t.Errorf("Expected error message 'invalid input', got '%s'", response["error"])
		}
	})

	t.Run("writes 500 for non-AppError", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := errors.New("standard error")

		WriteError(w, err)

		// Check status code
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		// Check content type
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Check response body
		var response map[string]string
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["error"] != "Internal server error" {
			t.Errorf("Expected error message 'Internal server error', got '%s'", response["error"])
		}
	})

	t.Run("writes different status codes for different AppErrors", func(t *testing.T) {
		tests := []struct {
			name       string
			err        *AppError
			statusCode int
		}{
			{"BadRequest", NewBadRequest("bad request"), http.StatusBadRequest},
			{"Unauthorized", NewUnauthorized("unauthorized"), http.StatusUnauthorized},
			{"Forbidden", NewForbidden("forbidden"), http.StatusForbidden},
			{"NotFound", NewNotFound("resource"), http.StatusNotFound},
			{"Conflict", NewConflict("conflict"), http.StatusConflict},
			{"Internal", NewInternal("internal", nil), http.StatusInternalServerError},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				WriteError(w, tt.err)

				if w.Code != tt.statusCode {
					t.Errorf("Expected status code %d, got %d", tt.statusCode, w.Code)
				}
			})
		}
	})
}
