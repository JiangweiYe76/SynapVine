package main

import (
	"context"
	"log/slog"
	"os"

	"core/internal/config"
	"core/internal/db"
	"core/internal/handler"
	"core/internal/repository"
	"core/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()
	slog.Info("configuration_loaded",
		slog.String("port", cfg.Port),
		slog.String("allowed_origin", cfg.AllowedOrigin),
		slog.String("neo4j_uri", cfg.Neo4jURI),
	)

	neoCfg := db.Config{
		URI:      cfg.Neo4jURI,
		Username: cfg.Neo4jUser,
		Password: cfg.Neo4jPassword,
	}
	neo, err := db.New(neoCfg)
	if err != nil {
		slog.Error("failed_to_connect_neo4j", slog.Any("error", err))
		os.Exit(1)
	}
	defer neo.Close(context.Background())

	nodeRepo := repository.NewNodeRepository(neo)
	commRepo := repository.NewCommunityRepository(neo)

	nodeSvc := service.NewNodeService(nodeRepo)
	commSvc := service.NewCommunityService(commRepo)
	detectorSvc := service.NewCommunityDetectorService(nodeRepo, commRepo)

	nodeHandler := handler.NewNodeHandler(nodeSvc)
	commHandler := handler.NewCommunityHandler(commSvc, detectorSvc)
	healthHandler := handler.NewHealthHandler()

	app := fiber.New(fiber.Config{
		AppName: "AI-Graph Core Server",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigin,
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Content-Type, Authorization",
		AllowCredentials: true,
	}))

	app.Get("/health", healthHandler.Check)

	app.Get("/api/graph/data", nodeHandler.GraphData)

	app.Get("/api/nodes", nodeHandler.List)
	app.Get("/api/nodes/:id", nodeHandler.Get)
	app.Post("/api/nodes", nodeHandler.Create)
	app.Put("/api/nodes/:id", nodeHandler.Update)
	app.Delete("/api/nodes/:id", nodeHandler.Delete)

	app.Get("/api/communities", commHandler.List)
	app.Get("/api/communities/:id", commHandler.Get)
	app.Post("/api/communities", commHandler.Create)
	app.Put("/api/communities/:id", commHandler.Update)
	app.Delete("/api/communities/:id", commHandler.Delete)
	app.Post("/api/communities/detect", commHandler.Detect)

	slog.Info("core_server_starting", slog.String("port", cfg.Port))

	if err := app.Listen(":" + cfg.Port); err != nil {
		slog.Error("server_failed", slog.Any("error", err))
		os.Exit(1)
	}
}
