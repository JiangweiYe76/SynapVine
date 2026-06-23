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
		slog.String("mysql_dsn", cfg.MySQLDSN),
	)

	// Neo4j: always required.
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

	// MySQL: optional (for papers and review queue).
	var paperHandler *handler.PaperHandler
	var reviewHandler *handler.ReviewQueueHandler
	if cfg.MySQLDSN != "" {
		mysqlDB, err := db.OpenMySQL(cfg.MySQLDSN)
		if err != nil {
			slog.Error("failed_to_connect_mysql", slog.Any("error", err))
			os.Exit(1)
		}
		defer mysqlDB.Close()

		if err := db.MigrateMySQL(context.Background(), mysqlDB); err != nil {
			slog.Error("mysql_migrate_failed", slog.Any("error", err))
			os.Exit(1)
		}
		slog.Info("mysql_connected_and_migrated")

		paperRepo := repository.NewPaperRepository(mysqlDB)
		reviewRepo := repository.NewReviewQueueRepository(mysqlDB)
		paperHandler = handler.NewPaperHandler(paperRepo)
		reviewHandler = handler.NewReviewQueueHandler(reviewRepo, paperRepo)
	} else {
		slog.Warn("mysql_not_configured", slog.String("hint", "Set MYSQL_DSN to enable papers and review queue"))
	}

	nodeRepo := repository.NewNodeRepository(neo)
	edgeRepo := repository.NewEdgeRepository(neo)
	commRepo := repository.NewCommunityRepository(neo)

	nodeSvc := service.NewNodeService(nodeRepo, commRepo)
	edgeSvc := service.NewEdgeService(edgeRepo)
	commSvc := service.NewCommunityService(commRepo)
	detectorSvc := service.NewCommunityDetectorService(nodeRepo, commRepo)

	nodeHandler := handler.NewNodeHandler(nodeSvc)
	edgeHandler := handler.NewEdgeHandler(edgeSvc)
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
	app.Get("/api/graph/timeline", nodeHandler.Timeline)

	app.Get("/api/nodes", nodeHandler.List)
	app.Get("/api/nodes/:id", nodeHandler.Get)
	app.Post("/api/nodes", nodeHandler.Create)
	app.Put("/api/nodes/:id", nodeHandler.Update)
	app.Delete("/api/nodes/:id", nodeHandler.Delete)

	app.Get("/api/edges", edgeHandler.List)
	app.Get("/api/edges/:source/:target", edgeHandler.Get)
	app.Post("/api/edges", edgeHandler.Create)
	app.Put("/api/edges/:source/:target", edgeHandler.Update)
	app.Delete("/api/edges/:source/:target", edgeHandler.Delete)

	app.Get("/api/communities", commHandler.List)
	app.Get("/api/communities/tree", commHandler.Tree)
	app.Get("/api/communities/:id", commHandler.Get)
	app.Post("/api/communities", commHandler.Create)
	app.Put("/api/communities/:id", commHandler.Update)
	app.Delete("/api/communities/:id", commHandler.Delete)
	app.Post("/api/communities/detect", commHandler.Detect)

	// Paper and review queue routes (only when MySQL is configured).
	if paperHandler != nil && reviewHandler != nil {
		app.Get("/api/papers", paperHandler.List)
		app.Get("/api/papers/:id", paperHandler.Get)
		app.Post("/api/papers", paperHandler.Create)
		app.Put("/api/papers/:id", paperHandler.Update)
		app.Delete("/api/papers/:id", paperHandler.Delete)

		app.Get("/api/review-queue", reviewHandler.List)
		app.Get("/api/review-queue/:id", reviewHandler.Get)
		app.Post("/api/review-queue", reviewHandler.Submit)
		app.Post("/api/review-queue/:id/approve", reviewHandler.Approve)
		app.Post("/api/review-queue/:id/reject", reviewHandler.Reject)
	}

	slog.Info("core_server_starting", slog.String("port", cfg.Port))

	if err := app.Listen(":" + cfg.Port); err != nil {
		slog.Error("server_failed", slog.Any("error", err))
		os.Exit(1)
	}
}
