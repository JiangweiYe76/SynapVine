package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Config holds Neo4j connection settings.
//
// Database selects which Neo4j database (a.k.a. "catalog") the connection
// uses. Leave empty to use the server default ("neo4j"). The test suite
// uses a dedicated database so it cannot pollute the dev graph.
type Config struct {
	URI      string
	Username string
	Password string
	Database string
}

// LoadConfigFromEnv reads Neo4j settings from environment variables.
func LoadConfigFromEnv() Config {
	return Config{
		URI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Username: getEnv("NEO4J_USER", "neo4j"),
		Password: getEnv("NEO4J_PASSWORD", "synapvine123"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Neo4j wraps the official Neo4j Go driver.
type Neo4j struct {
	driver   neo4j.DriverWithContext
	database string
}

// New creates a Neo4j connection manager. If cfg.Database is non-empty,
// all sessions opened by the returned client target that database;
// otherwise the server default ("neo4j") is used.
func New(cfg Config) (*Neo4j, error) {
	driver, err := neo4j.NewDriverWithContext(cfg.URI, neo4j.BasicAuth(cfg.Username, cfg.Password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create neo4j driver: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("neo4j connectivity check failed: %w", err)
	}

	slog.Info("neo4j_connected", slog.String("uri", cfg.URI), slog.String("database", cfg.Database))
	return &Neo4j{driver: driver, database: cfg.Database}, nil
}

// Close shuts down the driver.
func (n *Neo4j) Close(ctx context.Context) error {
	return n.driver.Close(ctx)
}

// sessionConfig returns the base session configuration, pre-populated
// with the database selected at construction time. Callers override
// AccessMode as needed.
func (n *Neo4j) sessionConfig() neo4j.SessionConfig {
	return neo4j.SessionConfig{
		DatabaseName: n.database,
	}
}

// Execute runs a write query without returning records.
func (n *Neo4j) Execute(ctx context.Context, cypher string, params map[string]any) error {
	session := n.driver.NewSession(ctx, n.writeSessionConfig())
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, cypher, params)
		return nil, err
	})
	return err
}

// Query runs a read query and returns records.
func (n *Neo4j) Query(ctx context.Context, cypher string, params map[string]any) ([]*neo4j.Record, error) {
	session := n.driver.NewSession(ctx, n.readSessionConfig())
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		return res.Collect(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]*neo4j.Record), nil
}

// QueryWrite runs a write query and returns records.
func (n *Neo4j) QueryWrite(ctx context.Context, cypher string, params map[string]any) ([]*neo4j.Record, error) {
	session := n.driver.NewSession(ctx, n.writeSessionConfig())
	defer session.Close(ctx)

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		return res.Collect(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]*neo4j.Record), nil
}

func (n *Neo4j) readSessionConfig() neo4j.SessionConfig {
	cfg := n.sessionConfig()
	cfg.AccessMode = neo4j.AccessModeRead
	return cfg
}

func (n *Neo4j) writeSessionConfig() neo4j.SessionConfig {
	cfg := n.sessionConfig()
	cfg.AccessMode = neo4j.AccessModeWrite
	return cfg
}

// Migrate runs all .cypher files in the given directory in lexical order.
func (n *Neo4j) Migrate(ctx context.Context, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cypher") {
			continue
		}

		path := filepath.Join(migrationsDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		slog.Info("running_migration", slog.String("file", entry.Name()))
		if err := n.runCypherStatements(ctx, string(content)); err != nil {
			return fmt.Errorf("migration %s failed: %w", entry.Name(), err)
		}
		slog.Info("migration_completed", slog.String("file", entry.Name()))
	}

	return nil
}

func (n *Neo4j) runCypherStatements(ctx context.Context, script string) error {
	statements := splitStatements(script)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := n.Execute(ctx, stmt, nil); err != nil {
			return err
		}
	}
	return nil
}

func splitStatements(script string) []string {
	var statements []string
	var current strings.Builder
	lines := strings.Split(script, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "//") {
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}

	if current.Len() > 0 {
		statements = append(statements, current.String())
	}
	return statements
}
