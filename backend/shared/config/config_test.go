package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original env vars to restore later
	originalEnv := map[string]string{
		"DATABASE_URL":         os.Getenv("DATABASE_URL"),
		"SESSION_SECRET":       os.Getenv("SESSION_SECRET"),
		"GOOGLE_CLIENT_ID":     os.Getenv("GOOGLE_CLIENT_ID"),
		"GOOGLE_CLIENT_SECRET": os.Getenv("GOOGLE_CLIENT_SECRET"),
		"GOOGLE_REDIRECT_URL":  os.Getenv("GOOGLE_REDIRECT_URL"),
		"FRONTEND_URL":         os.Getenv("FRONTEND_URL"),
		"PORT":                 os.Getenv("PORT"),
		"ENV":                  os.Getenv("ENV"),
		"AUDIT_RETENTION_DAYS": os.Getenv("AUDIT_RETENTION_DAYS"),
	}

	// Restore env vars after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("loads valid configuration", func(t *testing.T) {
		// Set required env vars
		os.Setenv("DATABASE_URL", "postgres://localhost/test")
		os.Setenv("SESSION_SECRET", "test-secret")
		os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
		os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
		os.Setenv("GOOGLE_REDIRECT_URL", "http://localhost:8080/callback")
		os.Setenv("FRONTEND_URL", "http://localhost:3000")
		os.Setenv("PORT", "8080")
		os.Setenv("ENV", "development")
		os.Setenv("AUDIT_RETENTION_DAYS", "365")

		cfg := Load()

		if cfg.DatabaseURL == "" {
			t.Error("Expected DatabaseURL to be set")
		}
		if cfg.SessionSecret != "test-secret" {
			t.Errorf("Expected SessionSecret to be 'test-secret', got %s", cfg.SessionSecret)
		}
		if cfg.GoogleClientID != "test-client-id" {
			t.Errorf("Expected GoogleClientID to be 'test-client-id', got %s", cfg.GoogleClientID)
		}
		if cfg.GoogleClientSecret != "test-client-secret" {
			t.Errorf("Expected GoogleClientSecret to be 'test-client-secret', got %s", cfg.GoogleClientSecret)
		}
		if cfg.GoogleRedirectURL != "http://localhost:8080/callback" {
			t.Errorf("Expected GoogleRedirectURL to be 'http://localhost:8080/callback', got %s", cfg.GoogleRedirectURL)
		}
		if cfg.FrontendURL != "http://localhost:3000" {
			t.Errorf("Expected FrontendURL to be 'http://localhost:3000', got %s", cfg.FrontendURL)
		}
		if cfg.Port != "8080" {
			t.Errorf("Expected Port to be '8080', got %s", cfg.Port)
		}
		if cfg.Env != "development" {
			t.Errorf("Expected Env to be 'development', got %s", cfg.Env)
		}
		if cfg.AuditRetentionDays != 365 {
			t.Errorf("Expected AuditRetentionDays to be 365, got %d", cfg.AuditRetentionDays)
		}
	})

	t.Run("uses default values for optional fields", func(t *testing.T) {
		// Set only required env vars
		os.Setenv("DATABASE_URL", "postgres://localhost/test")
		os.Setenv("SESSION_SECRET", "test-secret")
		os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
		os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
		os.Setenv("GOOGLE_REDIRECT_URL", "http://localhost:8080/callback")
		os.Unsetenv("FRONTEND_URL")
		os.Unsetenv("PORT")
		os.Unsetenv("ENV")
		os.Unsetenv("AUDIT_RETENTION_DAYS")

		cfg := Load()

		if cfg.FrontendURL != "http://localhost:3000" {
			t.Errorf("Expected default FrontendURL to be 'http://localhost:3000', got %s", cfg.FrontendURL)
		}
		if cfg.Port != "8080" {
			t.Errorf("Expected default Port to be '8080', got %s", cfg.Port)
		}
		if cfg.Env != "development" {
			t.Errorf("Expected default Env to be 'development', got %s", cfg.Env)
		}
		if cfg.AuditRetentionDays != 730 {
			t.Errorf("Expected default AuditRetentionDays to be 730, got %d", cfg.AuditRetentionDays)
		}
	})

	t.Run("adds timezone parameter to database URL", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			{
				name:     "adds timezone to URL without query params",
				input:    "postgres://localhost/test",
				expected: "postgres://localhost/test?timezone=America/New_York",
			},
			{
				name:     "adds timezone to URL with existing query params",
				input:    "postgres://localhost/test?sslmode=disable",
				expected: "postgres://localhost/test?sslmode=disable&timezone=America/New_York",
			},
			{
				name:     "does not add timezone if already present",
				input:    "postgres://localhost/test?timezone=UTC",
				expected: "postgres://localhost/test?timezone=UTC",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				os.Setenv("DATABASE_URL", tt.input)
				os.Setenv("SESSION_SECRET", "test-secret")
				os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
				os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
				os.Setenv("GOOGLE_REDIRECT_URL", "http://localhost:8080/callback")

				cfg := Load()

				if cfg.DatabaseURL != tt.expected {
					t.Errorf("Expected DatabaseURL to be '%s', got '%s'", tt.expected, cfg.DatabaseURL)
				}
			})
		}
	})

	t.Run("handles invalid audit retention days", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://localhost/test")
		os.Setenv("SESSION_SECRET", "test-secret")
		os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
		os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
		os.Setenv("GOOGLE_REDIRECT_URL", "http://localhost:8080/callback")
		os.Setenv("AUDIT_RETENTION_DAYS", "invalid")

		cfg := Load()

		// Should fall back to default
		if cfg.AuditRetentionDays != 730 {
			t.Errorf("Expected AuditRetentionDays to fallback to 730, got %d", cfg.AuditRetentionDays)
		}
	})

	t.Run("handles negative audit retention days", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://localhost/test")
		os.Setenv("SESSION_SECRET", "test-secret")
		os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
		os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
		os.Setenv("GOOGLE_REDIRECT_URL", "http://localhost:8080/callback")
		os.Setenv("AUDIT_RETENTION_DAYS", "-10")

		cfg := Load()

		// Should fall back to default (negative is not > 0)
		if cfg.AuditRetentionDays != 730 {
			t.Errorf("Expected AuditRetentionDays to fallback to 730, got %d", cfg.AuditRetentionDays)
		}
	})
}

func TestIsProduction(t *testing.T) {
	tests := []struct {
		env      string
		expected bool
	}{
		{"production", true},
		{"development", false},
		{"test", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &Config{Env: tt.env}
			result := cfg.IsProduction()
			if result != tt.expected {
				t.Errorf("IsProduction() = %v, want %v for env=%s", result, tt.expected, tt.env)
			}
		})
	}
}
