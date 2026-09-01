package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"core/internal/db"
	"core/internal/security"

	_ "github.com/go-sql-driver/mysql"
)

// testMySQLDatabase is the dedicated MySQL schema used by the test
// suite. It is created on first use and is fully isolated from the
// "synapvine_console" schema that `make dev` uses.
const testMySQLDatabase = "core_test"

// MySQLDSN returns the MySQL DSN used by integration tests from
// environment variables, falling back to localhost docker defaults.
// The returned DSN points at the dedicated test schema.
func MySQLDSN() string {
	user := getEnv("TEST_MYSQL_USER", "synapvine")
	pass := getEnv("TEST_MYSQL_PASSWORD", "synapvine123")
	host := getEnv("TEST_MYSQL_HOST", "localhost:3306")
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, host, testMySQLDatabase)
}

// adminMySQLDSN returns a root-level DSN with no schema selected, which
// is required to create the test schema and grant the test user access
// to it. The application user (synapvine) only has privileges on the
// dev schema, so it cannot create databases.
func adminMySQLDSN() string {
	user := getEnv("TEST_MYSQL_ADMIN_USER", "root")
	pass := getEnv("TEST_MYSQL_ADMIN_PASSWORD", "synapvine_root_pw")
	host := getEnv("TEST_MYSQL_HOST", "localhost:3306")
	return fmt.Sprintf("%s:%s@tcp(%s)/?parseTime=true", user, pass, host)
}

// appUserName returns the unprivileged application user that tests
// connect as, so the harness can GRANT it access to the test schema.
func appUserName() string {
	return getEnv("TEST_MYSQL_USER", "synapvine")
}

// TestCipher is a KeyCipher derived from a fixed test-only key. The key
// value is deterministic so tests can be rerun against data written by a
// previous run.
var TestCipher = mustNewTestCipher()

func mustNewTestCipher() *security.KeyCipher {
	// base64 of 32 bytes 0x01.. irrelevant; use a fixed non-zero key.
	c, err := security.NewKeyCipherFromBase64("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	if err != nil {
		panic(fmt.Sprintf("build test cipher: %v", err))
	}
	return c
}

// NewTestMySQL provisions the dedicated test schema (creating it if
// needed), runs core migrations against it, and returns a connected
// *sql.DB with all provider tables truncated. If MySQL is unavailable
// the test is skipped rather than failed. The connection is closed when
// the test finishes.
func NewTestMySQL(t *testing.T) *sql.DB {
	t.Helper()

	// Connect without a schema to create the test database.
	admin, err := db.OpenMySQL(adminMySQLDSN())
	if err != nil {
		t.Skipf("MySQL not available: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := admin.ExecContext(ctx,
		"CREATE DATABASE IF NOT EXISTS `"+testMySQLDatabase+"` DEFAULT CHARACTER SET utf8mb4"); err != nil {
		admin.Close()
		t.Skipf("cannot create test database: %v", err)
	}
	// GRANT does not accept placeholders for account names; the user
	// name comes from an env var or a hardcoded dev default, never from
	// test input.
	if _, err := admin.ExecContext(ctx,
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%%'", testMySQLDatabase, appUserName())); err != nil {
		admin.Close()
		t.Skipf("cannot grant test database privileges: %v", err)
	}
	admin.Close()

	conn, err := db.OpenMySQL(MySQLDSN())
	if err != nil {
		t.Skipf("MySQL not available: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if err := db.MigrateMySQL(context.Background(), conn); err != nil {
		conn.Close()
		t.Fatalf("migrate test database: %v", err)
	}

	cleanupMySQL(t, conn)
	return conn
}

// cleanupMySQL truncates the feature tables so each test starts empty.
// Foreign key checks are disabled during truncation because review_queue
// references papers.
func cleanupMySQL(t *testing.T, conn *sql.DB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		t.Fatalf("disable FK checks: %v", err)
	}
	for _, table := range []string{"llm_providers", "embedding_providers", "review_queue", "papers"} {
		if _, err := conn.ExecContext(ctx, "TRUNCATE TABLE `"+table+"`"); err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}
	if _, err := conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		t.Fatalf("enable FK checks: %v", err)
	}
}
