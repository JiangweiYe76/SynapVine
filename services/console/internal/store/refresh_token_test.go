package store_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"console/internal/db"
	"console/internal/model"
	"console/internal/store"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// openTestDB connects to the MySQL instance described by MYSQL_TEST_DSN
// (default: the dev container at localhost:3306). The test is skipped
// when MYSQL_TEST_DSN is unset and no local MySQL is reachable, so the
// suite stays green on CI without infrastructure.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		dsn = "synapvine:synapvine123@tcp(localhost:3306)/synapvine_console?parseTime=true"
	}

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Skipf("MYSQL unavailable, skipping: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		t.Skipf("MYSQL ping failed, skipping: %v", err)
	}
	if err := db.Migrate(ctx, conn); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	return conn
}

func TestRefreshTokenCreate_FitsInColumn(t *testing.T) {
	conn := openTestDB(t)
	defer conn.Close()

	rt := store.NewRefreshTokenStore(conn)
	users := store.NewUserStore(conn)

	ctx := context.Background()

	// Create a throwaway user.
	u := &model.User{
		ID:        uuid.NewString(),
		Username:  "rt-test-" + uuid.NewString()[:8],
		Password:  "x",
		Role:      model.RoleViewer,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := users.Create(ctx, u); err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = conn.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, u.ID)
	})

	// Simulate the issueSession path: a UUID-shaped refresh id
	// (36 chars) inserted into refresh_tokens.id (VARCHAR(36)).
	id := uuid.NewString()
	if len(id) > 36 {
		t.Fatalf("uuid.NewString produced %d chars, expected <= 36", len(id))
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := rt.Create(ctx, id, u.ID, expiresAt); err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	// Lookup round-trips.
	gotUser, gotExp, err := rt.Lookup(ctx, id)
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if gotUser != u.ID {
		t.Errorf("expected user_id=%s, got %s", u.ID, gotUser)
	}
	if gotExp.IsZero() {
		t.Error("expected non-zero expires_at")
	}

	// Delete + lookup-not-found.
	if err := rt.Delete(ctx, id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, _, err := rt.Lookup(ctx, id); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}
