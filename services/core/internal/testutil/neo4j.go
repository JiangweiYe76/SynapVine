package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"core/internal/db"
)

// Neo4jConfig returns test Neo4j configuration from environment variables.
// Falls back to localhost defaults suitable for local docker development.
func Neo4jConfig() db.Config {
	return db.Config{
		URI:      getEnv("TEST_NEO4J_URI", "bolt://localhost:7687"),
		Username: getEnv("TEST_NEO4J_USER", "neo4j"),
		Password: getEnv("TEST_NEO4J_PASSWORD", "synapvine123"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// NewTestNeo4j creates a Neo4j connection for integration tests.
// If Neo4j is unavailable, the test is skipped rather than failed.
// The connection is automatically closed when the test finishes.
func NewTestNeo4j(t *testing.T) *db.Neo4j {
	cfg := Neo4jConfig()
	neo, err := db.New(cfg)
	if err != nil {
		t.Skipf("Neo4j not available at %s: %v", cfg.URI, err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = neo.Close(ctx)
	})
	return neo
}

// CleanupAllData removes all nodes and relationships from the test database.
func CleanupAllData(t *testing.T, neo *db.Neo4j) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := neo.Execute(ctx, "MATCH (n) DETACH DELETE n", nil); err != nil {
		t.Fatalf("failed to cleanup neo4j data: %v", err)
	}
}
