package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"console/internal/auth"

	_ "github.com/go-sql-driver/mysql"
)

// verify_seed is a manual end-to-end check: it reads the admin user
// from MySQL and confirms that auth.CheckPassword accepts "admin123"
// against the stored hash. Run after `make dev` (or after manually
// running cmd/seed) to confirm the seeded credentials work.
//
// Usage:
//   MYSQL_DSN="..." go run ./cmd/verify_seed
func main() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "MYSQL_DSN is required")
		os.Exit(1)
	}

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		os.Exit(1)
	}
	defer conn.Close()

	var hash string
	err = conn.QueryRowContext(context.Background(),
		`SELECT password FROM users WHERE username = 'admin'`,
	).Scan(&hash)
	if err != nil {
		fmt.Fprintln(os.Stderr, "query:", err)
		os.Exit(1)
	}

	if !auth.CheckPassword("admin123", hash) {
		fmt.Fprintln(os.Stderr, "FAIL: admin123 does not match stored hash")
		os.Exit(1)
	}
	fmt.Println("OK: admin/admin123 verifies against the stored hash")
}
