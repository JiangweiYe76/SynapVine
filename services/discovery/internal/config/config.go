package config

import "os"

// Config holds the discovery server configuration settings.
type Config struct {
	Port       string // HTTP server port
	ConsoleURL string // URL of the console service (required, for LLM provider config)
	CoreURL    string // URL of the core service (required, for papers and review queue)
}

// Load reads configuration from environment variables with fallback defaults.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	consoleURL := os.Getenv("CONSOLE_URL")
	if consoleURL == "" {
		consoleURL = "http://localhost:8002"
	}

	coreURL := os.Getenv("CORE_URL")
	if coreURL == "" {
		coreURL = "http://localhost:8001"
	}

	return &Config{
		Port:       port,
		ConsoleURL: consoleURL,
		CoreURL:    coreURL,
	}
}
