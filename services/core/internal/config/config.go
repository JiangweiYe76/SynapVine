package config

import (
	"os"
	"strings"
)

// Config holds application configuration.
type Config struct {
	Port                  string
	AllowedOrigin         string
	Neo4jURI              string
	Neo4jUser             string
	Neo4jPassword         string
	MySQLDSN              string // MySQL DSN for papers and review queue (optional)
	ServiceTokens         map[string]string
	ProviderEncryptionKey string // base64-encoded 32-byte AES-256 key for provider API keys (required when MySQL is enabled)
}

// Load reads configuration from environment variables.
func Load() Config {
	return Config{
		Port:                  getEnv("PORT", "8001"),
		AllowedOrigin:         getEnv("ALLOWED_ORIGIN", "http://localhost:5174"),
		Neo4jURI:              getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUser:             getEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword:         getEnv("NEO4J_PASSWORD", "synapvine123"),
		MySQLDSN:              os.Getenv("MYSQL_DSN"),
		ServiceTokens:         ParseServiceTokens(os.Getenv("SERVICE_TOKENS")),
		ProviderEncryptionKey: os.Getenv("PROVIDER_ENCRYPTION_KEY"),
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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
