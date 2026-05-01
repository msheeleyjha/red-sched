package utils

import (
	"net/http"
	"testing"
)

func TestGetIPAddress(t *testing.T) {
	tests := []struct {
		name             string
		remoteAddr       string
		xForwardedFor    string
		xRealIP          string
		expectedIP       string
	}{
		{
			name:       "uses RemoteAddr when no headers present",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "uses RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			expectedIP: "192.168.1.1",
		},
		{
			name:          "prefers X-Forwarded-For over RemoteAddr",
			remoteAddr:    "192.168.1.1:12345",
			xForwardedFor: "203.0.113.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:          "takes first IP from X-Forwarded-For list",
			remoteAddr:    "192.168.1.1:12345",
			xForwardedFor: "203.0.113.1, 198.51.100.1, 192.0.2.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:          "trims whitespace from X-Forwarded-For",
			remoteAddr:    "192.168.1.1:12345",
			xForwardedFor: "  203.0.113.1  , 198.51.100.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:        "uses X-Real-IP when no X-Forwarded-For",
			remoteAddr:  "192.168.1.1:12345",
			xRealIP:     "203.0.113.1",
			expectedIP:  "203.0.113.1",
		},
		{
			name:          "prefers X-Forwarded-For over X-Real-IP",
			remoteAddr:    "192.168.1.1:12345",
			xForwardedFor: "203.0.113.1",
			xRealIP:       "198.51.100.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:       "handles IPv6 RemoteAddr",
			remoteAddr: "[2001:db8::1]:12345",
			expectedIP: "[2001:db8::1]",
		},
		{
			name:          "handles IPv6 in X-Forwarded-For",
			remoteAddr:    "192.168.1.1:12345",
			xForwardedFor: "2001:db8::1",
			expectedIP:    "2001:db8::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request
			req, err := http.NewRequest("GET", "http://example.com", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Set RemoteAddr
			req.RemoteAddr = tt.remoteAddr

			// Set headers if provided
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			// Get IP address
			ip := GetIPAddress(req)

			// Check result
			if ip != tt.expectedIP {
				t.Errorf("Expected IP '%s', got '%s'", tt.expectedIP, ip)
			}
		})
	}
}

func TestGetIPAddress_EmptyXForwardedFor(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "")
	req.Header.Set("X-Real-IP", "203.0.113.1")

	ip := GetIPAddress(req)

	// Should fall back to X-Real-IP when X-Forwarded-For is empty
	if ip != "203.0.113.1" {
		t.Errorf("Expected IP '203.0.113.1' (from X-Real-IP), got '%s'", ip)
	}
}

func TestGetIPAddress_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name       string
		scenario   string
		setup      func(*http.Request)
		expectedIP string
	}{
		{
			name:     "direct connection (no proxy)",
			scenario: "User connects directly to server",
			setup: func(r *http.Request) {
				r.RemoteAddr = "203.0.113.1:54321"
			},
			expectedIP: "203.0.113.1",
		},
		{
			name:     "nginx reverse proxy",
			scenario: "User behind nginx proxy",
			setup: func(r *http.Request) {
				r.RemoteAddr = "127.0.0.1:12345"
				r.Header.Set("X-Real-IP", "203.0.113.1")
			},
			expectedIP: "203.0.113.1",
		},
		{
			name:     "AWS ALB (Application Load Balancer)",
			scenario: "User behind AWS ALB",
			setup: func(r *http.Request) {
				r.RemoteAddr = "10.0.1.1:12345"
				r.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.1.1")
			},
			expectedIP: "203.0.113.1",
		},
		{
			name:     "cloudflare proxy",
			scenario: "User behind Cloudflare",
			setup: func(r *http.Request) {
				r.RemoteAddr = "104.16.0.0:12345"
				r.Header.Set("X-Forwarded-For", "203.0.113.1")
			},
			expectedIP: "203.0.113.1",
		},
		{
			name:     "multiple proxies",
			scenario: "User behind multiple proxies",
			setup: func(r *http.Request) {
				r.RemoteAddr = "127.0.0.1:12345"
				r.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1, 192.0.2.1")
			},
			expectedIP: "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			tt.setup(req)

			ip := GetIPAddress(req)

			if ip != tt.expectedIP {
				t.Errorf("%s: Expected IP '%s', got '%s'", tt.scenario, tt.expectedIP, ip)
			}
		})
	}
}
