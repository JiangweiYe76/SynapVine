package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"core/internal/config"
	"core/internal/db"
	"core/internal/handler"
	"core/internal/repository"
	"core/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// providerHandlers holds LLM and embedding provider handlers when MySQL is configured.
type providerHandlers struct {
	llm       *handler.LLMProviderHandler
	embedding *handler.EmbeddingProviderHandler
}

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

	// MySQL: optional (for papers, review queue, and provider management).
	var paperHandler *handler.PaperHandler
	var reviewHandler *handler.ReviewQueueHandler
	var provHandlers *providerHandlers
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

		llmProviderRepo := repository.NewLLMProviderRepository(mysqlDB)
		embeddingProviderRepo := repository.NewEmbeddingProviderRepository(mysqlDB)
		provHandlers = &providerHandlers{
			llm:       handler.NewLLMProviderHandler(llmProviderRepo),
			embedding: handler.NewEmbeddingProviderHandler(embeddingProviderRepo),
		}
	} else {
		slog.Warn("mysql_not_configured", slog.String("hint", "Set MYSQL_DSN to enable papers, review queue, and provider management"))
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
		AppName:   "AI-Graph Core Server",
		BodyLimit: 50 * 1024 * 1024, // 50 MB — accommodate PDF uploads (base64-encoded)
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
		app.Get("/api/papers/:id/pdf", paperHandler.GetPDF)
		app.Post("/api/papers", paperHandler.Create)
		app.Put("/api/papers/:id", paperHandler.Update)
		app.Delete("/api/papers/:id", paperHandler.Delete)

		app.Get("/api/review-queue", reviewHandler.List)
		app.Get("/api/review-queue/:id", reviewHandler.Get)
		app.Post("/api/review-queue", reviewHandler.Submit)
		app.Post("/api/review-queue/:id/approve", reviewHandler.Approve)
		app.Post("/api/review-queue/:id/reject", reviewHandler.Reject)
	}

	// LLM and embedding provider routes (only when MySQL is configured).
	if provHandlers != nil {
		// LLM provider routes
		app.Get("/api/llm/providers", provHandlers.llm.List)
		app.Get("/api/llm/providers/default", provHandlers.llm.GetDefault)
		app.Get("/api/llm/providers/:id", provHandlers.llm.Get)
		app.Post("/api/llm/providers", provHandlers.llm.Create)
		app.Put("/api/llm/providers/:id", provHandlers.llm.Update)
		app.Delete("/api/llm/providers/:id", provHandlers.llm.Delete)
		app.Post("/api/llm/providers/:id/test", provHandlers.llm.Test)

		// Internal LLM provider route (includes API key for service-to-service use)
		app.Get("/api/internal/llm/providers/default", provHandlers.llm.GetDefaultInternal)

		// Embedding provider routes
		app.Get("/api/embedding/providers", provHandlers.embedding.List)
		app.Get("/api/embedding/providers/default", provHandlers.embedding.GetDefault)
		app.Get("/api/embedding/providers/:id", provHandlers.embedding.Get)
		app.Post("/api/embedding/providers", provHandlers.embedding.Create)
		app.Put("/api/embedding/providers/:id", provHandlers.embedding.Update)
		app.Delete("/api/embedding/providers/:id", provHandlers.embedding.Delete)
		app.Post("/api/embedding/providers/:id/test", provHandlers.embedding.Test)
	}

	slog.Info("core_server_starting", slog.String("port", cfg.Port))

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		slog.Info("shutdown_signal_received", slog.String("action", "gracefully_stopping_server"))
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			slog.Error("server_shutdown_failed", slog.Any("error", err))
		}
		close(idleConnsClosed)
	}()

	if err := app.Listen(":" + cfg.Port); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("server_failed", slog.Any("error", err))
		os.Exit(1)
	}

	<-idleConnsClosed
}
