package config

import (
	"os"
)

// Config holds the console server configuration settings
type Config struct {
	Port          string // HTTP server port
	AllowedOrigin string // Allowed CORS origin
	JWTSecret     string // Secret key for JWT signing
}

// Load reads configuration from environment variables with fallback defaults
// Returns a Config struct with all settings
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5174"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "console-dev-secret-key-change-in-production"
	}

	return &Config{
		Port:          port,
		AllowedOrigin: allowedOrigin,
		JWTSecret:     jwtSecret,
	}
}
