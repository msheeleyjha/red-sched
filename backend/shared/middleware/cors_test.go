package middleware

import (
	"testing"
)

func TestNewCORSHandler(t *testing.T) {
	frontendURL := "http://localhost:3000"
	corsHandler := NewCORSHandler(frontendURL)

	if corsHandler == nil {
		t.Error("Expected CORS handler to be created, got nil")
	}

	// Note: Testing CORS behavior requires integration tests with actual HTTP requests
	// The rs/cors package is well-tested, so we just verify it initializes
}

func TestNewCORSHandler_DifferentURLs(t *testing.T) {
	tests := []string{
		"http://localhost:3000",
		"https://example.com",
		"https://app.example.com",
	}

	for _, url := range tests {
		t.Run(url, func(t *testing.T) {
			corsHandler := NewCORSHandler(url)
			if corsHandler == nil {
				t.Errorf("Expected CORS handler to be created for URL %s, got nil", url)
			}
		})
	}
}
