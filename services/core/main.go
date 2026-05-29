package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"core/internal/db"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("core_data_service_starting")

	cfg := db.LoadConfigFromEnv()
	neo, err := db.New(cfg)
	if err != nil {
		slog.Error("failed_to_connect_neo4j", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	defer neo.Close(ctx)

	slog.Info("core_ready", slog.String("neo4j_uri", cfg.URI))

	// TODO: start internal gRPC server for graph data operations
	select {}
}
