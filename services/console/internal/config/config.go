package config

import (
	"os"
	"strconv"
)

// Config holds the console server configuration settings
type Config struct {
	Port          string // HTTP server port
	AllowedOrigin string // Allowed CORS origin
	JWTSecret     string // Secret key for JWT signing
	CoreURL       string // URL of the core service (required)
	DiscoveryURL  string // URL of the discovery service (optional, for auto-triggering paper analysis)
	MySQLDSN      string // MySQL DSN for the console auth database (required)
	CookieSecure  bool   // Whether the refresh-token cookie gets the Secure attribute
	ServiceToken  string // Token presented to core and discovery via X-Service-Token
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

	discoveryURL := os.Getenv("DISCOVERY_URL")
	// Optional: defaults to empty string (auto-trigger disabled) when not set.

	mysqlDSN := os.Getenv("MYSQL_DSN")
	// Intentionally no default: MySQL is mandatory.

	// COOKIE_SECURE controls the Secure attribute on the refresh-token
	// cookie. Defaults to true (production-safe). Dev runs over plain
	// HTTP on localhost, where the browser rejects Secure cookies, so
	// the dev script sets COOKIE_SECURE=false.
	cookieSecure := true
	if raw := os.Getenv("COOKIE_SECURE"); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			cookieSecure = parsed
		}
	}

	return &Config{
		Port:          port,
		AllowedOrigin: allowedOrigin,
		JWTSecret:     jwtSecret,
		CoreURL:       coreURL,
		DiscoveryURL:  discoveryURL,
		MySQLDSN:      mysqlDSN,
		CookieSecure:  cookieSecure,
		ServiceToken:  os.Getenv("SERVICE_TOKEN"),
	}
}
