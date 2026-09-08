package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the discovery server configuration settings.
type Config struct {
	Port          string            // HTTP server port
	CoreURL       string            // URL of the core service (required, for papers, review queue, and LLM provider config)
	ServiceToken  string            // Token presented to core via X-Service-Token
	ServiceTokens map[string]string // Tokens accepted from callers on /api/analyze
	AllowedOrigin string            // Allowed CORS origin (the console frontend)

	ArxivEnabled    bool          // whether the ArXiv ingestion scheduler runs
	ArxivCategories []string      // ArXiv categories to watch
	ArxivInterval   time.Duration // ingestion cycle interval
	ArxivMaxPerRun  int           // max papers fetched per cycle
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

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5174"
	}

	// ArXiv ingestion defaults: disabled unless explicitly enabled, one
	// cycle per hour, watching the main AI-related categories.
	arxivEnabled := os.Getenv("ARXIV_ENABLED") == "true"
	categories := []string{"cs.AI", "cs.CL", "cs.LG"}
	if raw := os.Getenv("ARXIV_CATEGORIES"); raw != "" {
		categories = nil
		for _, c := range strings.Split(raw, ",") {
			if c = strings.TrimSpace(c); c != "" {
				categories = append(categories, c)
			}
		}
	}
	interval := time.Hour
	if raw := os.Getenv("ARXIV_INTERVAL"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			interval = d
		}
	}
	maxPerRun := 5
	if raw := os.Getenv("ARXIV_MAX_PER_RUN"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			maxPerRun = n
		}
	}

	return &Config{
		Port:          port,
		CoreURL:       coreURL,
		ServiceToken:  os.Getenv("SERVICE_TOKEN"),
		ServiceTokens: ParseServiceTokens(os.Getenv("SERVICE_TOKENS")),
		AllowedOrigin: allowedOrigin,

		ArxivEnabled:    arxivEnabled,
		ArxivCategories: categories,
		ArxivInterval:   interval,
		ArxivMaxPerRun:  maxPerRun,
	}
}

// ParseServiceTokens parses a SERVICE_TOKENS value in
// "name=token,name=token" format into a map. Entries with empty names
// or empty tokens are skipped so a trailing comma or typo cannot
// register a blank token that would match empty headers.
func ParseServiceTokens(raw string) map[string]string {
	tokens := make(map[string]string)
	for _, entry := range strings.Split(raw, ",") {
		name, token, found := strings.Cut(strings.TrimSpace(entry), "=")
		if !found {
			continue
		}
		name = strings.TrimSpace(name)
		token = strings.TrimSpace(token)
		if name == "" || token == "" {
			continue
		}
		tokens[name] = token
	}
	return tokens
}
