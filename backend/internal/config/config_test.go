package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Unset any env vars that might interfere.
	for _, key := range []string{"PORT", "PAPERLESS_BASE_URL", "PAPERLESS_TOKEN", "VISION_API_BASE_URL", "VISION_API_KEY", "DATABASE_URL"} {
		os.Unsetenv(key)
	}

	cfg := Load()

	if cfg.Port != 8080 {
		t.Errorf("expected default Port 8080, got %d", cfg.Port)
	}
	if cfg.PaperlessBaseURL != "http://localhost:8000" {
		t.Errorf("expected default PaperlessBaseURL http://localhost:8000, got %q", cfg.PaperlessBaseURL)
	}
	if cfg.PaperlessToken != "" {
		t.Errorf("expected default PaperlessToken empty, got %q", cfg.PaperlessToken)
	}
	if cfg.VisionAPIBaseURL != "https://ai.havek.es/api" {
		t.Errorf("expected default VisionAPIBaseURL https://ai.havek.es/api, got %q", cfg.VisionAPIBaseURL)
	}
	if cfg.VisionAPIKey != "" {
		t.Errorf("expected default VisionAPIKey empty, got %q", cfg.VisionAPIKey)
	}
	if cfg.DatabaseURL != "postgres://price-tracker:price-tracker@localhost:5433/price-tracker?sslmode=disable" {
		t.Errorf("expected default DatabaseURL, got %q", cfg.DatabaseURL)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	// Set env vars.
	os.Setenv("PORT", "9090")
	os.Setenv("PAPERLESS_BASE_URL", "https://paperless.example.com")
	os.Setenv("PAPERLESS_TOKEN", "secret-token")
	os.Setenv("VISION_API_BASE_URL", "https://vision.example.com")
	os.Setenv("VISION_API_KEY", "vision-key")
	os.Setenv("DATABASE_URL", "postgres://user:pass@host:5432/db?sslmode=require")

	defer func() {
		for _, key := range []string{"PORT", "PAPERLESS_BASE_URL", "PAPERLESS_TOKEN", "VISION_API_BASE_URL", "VISION_API_KEY", "DATABASE_URL"} {
			os.Unsetenv(key)
		}
	}()

	cfg := Load()

	if cfg.Port != 9090 {
		t.Errorf("expected Port 9090, got %d", cfg.Port)
	}
	if cfg.PaperlessBaseURL != "https://paperless.example.com" {
		t.Errorf("expected PaperlessBaseURL https://paperless.example.com, got %q", cfg.PaperlessBaseURL)
	}
	if cfg.PaperlessToken != "secret-token" {
		t.Errorf("expected PaperlessToken secret-token, got %q", cfg.PaperlessToken)
	}
	if cfg.VisionAPIBaseURL != "https://vision.example.com" {
		t.Errorf("expected VisionAPIBaseURL https://vision.example.com, got %q", cfg.VisionAPIBaseURL)
	}
	if cfg.VisionAPIKey != "vision-key" {
		t.Errorf("expected VisionAPIKey vision-key, got %q", cfg.VisionAPIKey)
	}
	if cfg.DatabaseURL != "postgres://user:pass@host:5432/db?sslmode=require" {
		t.Errorf("expected DatabaseURL override, got %q", cfg.DatabaseURL)
	}
}

func TestLoadPortNonNumeric(t *testing.T) {
	os.Setenv("PORT", "not-a-number")
	defer os.Unsetenv("PORT")

	cfg := Load()

	if cfg.Port != 8080 {
		t.Errorf("expected Port fallback 8080 for non-numeric value, got %d", cfg.Port)
	}
}
