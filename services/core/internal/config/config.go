package config

import (
	"os"
)

// Config holds application configuration.
type Config struct {
	Port          string
	AllowedOrigin string
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string
}

// Load reads configuration from environment variables.
func Load() Config {
	return Config{
		Port:          getEnv("PORT", "8001"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:5174"),
		Neo4jURI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUser:     getEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword: getEnv("NEO4J_PASSWORD", "synapvine123"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
