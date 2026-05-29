package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"core/internal/db"
)

func main() {
	migrationsDir := flag.String("dir", "internal/db/migrations", "Path to migrations directory")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := db.LoadConfigFromEnv()
	neo, err := db.New(cfg)
	if err != nil {
		slog.Error("failed_to_connect_neo4j", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	defer neo.Close(ctx)

	if err := neo.Migrate(ctx, *migrationsDir); err != nil {
		slog.Error("migration_failed", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("all_migrations_completed")
}
