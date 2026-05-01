package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Common error constructors

// NewBadRequest creates a 400 Bad Request error
func NewBadRequest(message string) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewUnauthorized creates a 401 Unauthorized error
func NewUnauthorized(message string) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewForbidden creates a 403 Forbidden error
func NewForbidden(message string) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewNotFound creates a 404 Not Found error
func NewNotFound(resource string) *AppError {
	return &AppError{
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

// NewConflict creates a 409 Conflict error
func NewConflict(message string) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// NewInternal creates a 500 Internal Server Error with wrapped error
func NewInternal(message string, err error) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// WriteError writes an error response to HTTP writer
func WriteError(w http.ResponseWriter, err error) {
	// Check if it's an AppError
	if appErr, ok := err.(*AppError); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.StatusCode)
		json.NewEncoder(w).Encode(map[string]string{
			"error": appErr.Message,
		})
		return
	}

	// Default to 500 Internal Server Error
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "Internal server error",
	})
}
