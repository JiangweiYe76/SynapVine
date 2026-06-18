// Command seed creates the first admin user in the console's MySQL
// database. It runs idempotently: if any user already exists, it is a
// no-op. The username and password are taken from the ADMIN_USERNAME
// and ADMIN_PASSWORD environment variables. Running this command is the
// only supported way to bootstrap the console auth database; the
// in-memory admin that previously lived in NewAuthHandler has been
// removed.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"console/internal/auth"
	"console/internal/db"
	"console/internal/model"
	"console/internal/store"

	"github.com/google/uuid"
)

func main() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "MYSQL_DSN is required")
		os.Exit(1)
	}
	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		fmt.Fprintln(os.Stderr, "ADMIN_PASSWORD is required")
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	conn, err := db.Open(dsn)
	if err != nil {
		slog.Error("seed_mysql_open_failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer conn.Close()

	if err := db.Migrate(context.Background(), conn); err != nil {
		slog.Error("seed_mysql_migrate_failed", slog.Any("error", err))
		os.Exit(1)
	}

	users := store.NewUserStore(conn)

	count, err := users.Count(context.Background())
	if err != nil {
		slog.Error("seed_user_count_failed", slog.Any("error", err))
		os.Exit(1)
	}
	if count > 0 {
		slog.Info("seed_skipped", slog.Int("existing_users", count))
		fmt.Printf("Users table is not empty (%d users). Nothing to do.\n", count)
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		slog.Error("seed_hash_failed", slog.Any("error", err))
		os.Exit(1)
	}

	now := time.Now()
	u := &model.User{
		ID:        uuid.NewString(),
		Username:  username,
		Password:  hash,
		Role:      model.RoleAdmin,
		TokenVer:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := users.Create(context.Background(), u); err != nil {
		slog.Error("seed_user_create_failed", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("seed_admin_created",
		slog.String("user_id", u.ID),
		slog.String("username", username),
	)
	fmt.Printf("Seeded admin user %q (id=%s). Use ADMIN_PASSWORD to sign in.\n", username, u.ID)
}
