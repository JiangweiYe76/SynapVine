package config

import (
	"os"
)

// Config holds the application configuration settings
type Config struct {
	Port          string // HTTP server port
	AllowedOrigin string // Allowed CORS origin
	CoreURL       string // URL of the core service
	ServiceToken  string // Token presented to core via X-Service-Token
}

// Load reads configuration from environment variables with fallback defaults
// Returns a Config struct with all settings
func Load() *Config {
	// Server port (default: 8000)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Allowed CORS origin (default: http://localhost:5173)
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	// Core service URL (default: http://localhost:8001)
	coreURL := os.Getenv("CORE_URL")
	if coreURL == "" {
		coreURL = "http://localhost:8001"
	}

	return &Config{
		Port:          port,
		AllowedOrigin: allowedOrigin,
		CoreURL:       coreURL,
		ServiceToken:  os.Getenv("SERVICE_TOKEN"),
	}
}
