package config

import (
	"os"
)

// Config holds the console server configuration settings
type Config struct {
	Port          string // HTTP server port
	AllowedOrigin string // Allowed CORS origin
	JWTSecret     string // Secret key for JWT signing
	CoreURL       string // URL of the core service (required)
	MySQLDSN      string // MySQL DSN for the console auth database (required)
}

// Load reads configuration from environment variables with fallback defaults.
// Returns a Config struct with all settings.
//
// CORE_URL and MYSQL_DSN have no default: when unset the application must
// fail fast at startup. CORE_URL is required because the core service is
// the authoritative source of node data and statistics. MYSQL_DSN is
// required because the console persists users, refresh tokens, and audit
// events in MySQL; we no longer keep an in-memory user map.
//
// JWT_SECRET also has no default in production-style runs. The dev script
// (scripts/dev-up.sh) sets a known value so developers can log in
// without managing secrets.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5174"
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	coreURL := os.Getenv("CORE_URL")
	// Intentionally no default: core is mandatory.

	mysqlDSN := os.Getenv("MYSQL_DSN")
	// Intentionally no default: MySQL is mandatory.

	return &Config{
		Port:          port,
		AllowedOrigin: allowedOrigin,
		JWTSecret:     jwtSecret,
		CoreURL:       coreURL,
		MySQLDSN:      mysqlDSN,
	}
}
