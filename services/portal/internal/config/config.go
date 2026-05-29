package config

import (
	"os"
)

// Config holds the application configuration settings
type Config struct {
	Port          string // HTTP server port
	DataPath      string // Path to the graph data JSON file
	AllowedOrigin string // Allowed CORS origin
}

// Load reads configuration from environment variables with fallback defaults
// Returns a Config struct with all settings
func Load() *Config {
	// Server port (default: 8000)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Path to graph data file (default: ../data/graph.json)
	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "../data/graph.json"
	}

	// Allowed CORS origin (default: http://localhost:5173)
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	return &Config{
		Port:          port,
		DataPath:      dataPath,
		AllowedOrigin: allowedOrigin,
	}
}
