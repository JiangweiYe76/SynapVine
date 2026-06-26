package config

import "os"

// Config holds the discovery server configuration settings.
type Config struct {
	Port    string // HTTP server port
	CoreURL string // URL of the core service (required, for papers, review queue, and LLM provider config)
}

// Load reads configuration from environment variables with fallback defaults.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	coreURL := os.Getenv("CORE_URL")
	if coreURL == "" {
		coreURL = "http://localhost:8001"
	}

	return &Config{
		Port:    port,
		CoreURL: coreURL,
	}
}
