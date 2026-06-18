// Package db owns the MySQL connection lifecycle and schema migrations
// for the console service. The console uses MySQL to persist users,
// refresh tokens, and audit events.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Open creates a *sql.DB pool targeting the given MySQL DSN. It pings
// the database to fail fast when the server is unreachable and tunes the
// pool for a small dev workload.
func Open(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql open: %w", err)
	}

	conn.SetMaxOpenConns(20)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("mysql ping: %w", err)
	}

	return conn, nil
}
