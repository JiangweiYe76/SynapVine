package testutil

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"core/internal/db"
)

// testDatabaseName is the dedicated Neo4j database used by the test
// suite. It is created on first use and is fully isolated from the
// default "neo4j" database that `make dev` uses.
const testDatabaseName = "neo4j-test"

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
//
// Tests run against a dedicated database (default: "neo4j-test") instead
// of the dev database, so they cannot pollute the dev graph. The test
// database is created on first use if it does not yet exist. Cleanup of
// the test database is the caller's responsibility — typically by
// calling CleanupAllData at the start of each test.
func NewTestNeo4j(t *testing.T) *db.Neo4j {
	cfg := Neo4jConfig()
	if err := ensureTestDatabase(cfg); err != nil {
		t.Skipf("cannot provision test database: %v", err)
	}
	cfg.Database = testDatabaseName

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

// ensureTestDatabase makes sure the dedicated test database exists. It
// connects as the system database (which is the only place the
// `CREATE DATABASE` admin command is valid) and runs a no-op
// `CREATE DATABASE IF NOT EXISTS`. An existing database is not an error.
func ensureTestDatabase(cfg db.Config) error {
	adminCfg := cfg
	adminCfg.Database = "system"

	admin, err := db.New(adminCfg)
	if err != nil {
		return err
	}
	defer admin.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := admin.Execute(ctx,
		"CREATE DATABASE `"+testDatabaseName+"` IF NOT EXISTS",
		nil,
	); err != nil {
		// Some Neo4j editions or configs do not allow creating databases
		// (e.g. community edition restriction, locked-down permissions).
		// Treat "already exists" as success; surface everything else.
		if !strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return err
		}
	}
	return nil
}

// CleanupAllData removes all nodes and relationships from the test
// database. Because the test database is dedicated and isolated from
// the dev database, this is safe to call without risk of wiping real
// data.
func CleanupAllData(t *testing.T, neo *db.Neo4j) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := neo.Execute(ctx, "MATCH (n) DETACH DELETE n", nil); err != nil {
		t.Fatalf("failed to cleanup test database: %v", err)
	}
}
