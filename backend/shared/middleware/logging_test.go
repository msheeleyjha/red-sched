package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	t.Run("calls next handler", func(t *testing.T) {
		called := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		middleware := LoggingMiddleware(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if !called {
			t.Error("Expected handler to be called")
		}
	})

	t.Run("captures status code", func(t *testing.T) {
		tests := []struct {
			name       string
			statusCode int
		}{
			{"200 OK", http.StatusOK},
			{"201 Created", http.StatusCreated},
			{"400 Bad Request", http.StatusBadRequest},
			{"401 Unauthorized", http.StatusUnauthorized},
			{"404 Not Found", http.StatusNotFound},
			{"500 Internal Server Error", http.StatusInternalServerError},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.statusCode)
				})

				middleware := LoggingMiddleware(handler)

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				if w.Code != tt.statusCode {
					t.Errorf("Expected status code %d, got %d", tt.statusCode, w.Code)
				}
			})
		}
	})

	t.Run("defaults to 200 OK when WriteHeader not called", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't call WriteHeader
			w.Write([]byte("OK"))
		})

		middleware := LoggingMiddleware(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected default status code 200, got %d", w.Code)
		}
	})

	t.Run("works with different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				called := false
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					called = true
					if r.Method != method {
						t.Errorf("Expected method %s, got %s", method, r.Method)
					}
				})

				middleware := LoggingMiddleware(handler)

				req := httptest.NewRequest(method, "/test", nil)
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				if !called {
					t.Errorf("Handler not called for method %s", method)
				}
			})
		}
	})
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	t.Run("captures status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		rw.WriteHeader(http.StatusCreated)

		if rw.statusCode != http.StatusCreated {
			t.Errorf("Expected statusCode to be %d, got %d", http.StatusCreated, rw.statusCode)
		}
	})

	t.Run("calls underlying WriteHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		rw.WriteHeader(http.StatusNotFound)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected underlying status code to be %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}
