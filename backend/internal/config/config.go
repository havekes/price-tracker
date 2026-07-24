package config

import (
	"log/slog"
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port int

	PaperlessBaseURL string
	PaperlessToken   string

	VisionAPIBaseURL string
	VisionAPIKey     string

	DatabaseURL string
}

// Load reads environment variables and returns a Config with defaults applied.
// Currently all vars have reasonable defaults except the Paperless/Vision tokens
// which will be required in later phases.
func Load() Config {
	return Config{
		Port:             getEnvInt("PORT", 8080),
		PaperlessBaseURL: getEnv("PAPERLESS_BASE_URL", "http://localhost:8000"),
		PaperlessToken:   getEnv("PAPERLESS_TOKEN", ""),
		VisionAPIBaseURL: getEnv("VISION_API_BASE_URL", "https://ai.havek.es/api"),
		VisionAPIKey:     getEnv("VISION_API_KEY", ""),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://price-tracker:price-tracker@localhost:5432/price-tracker?sslmode=disable"),
}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
		slog.Warn("non-numeric environment variable, using fallback",
			"var", key,
			"value", v,
			"fallback", fallback,
		)
	}
	return fallback
}
