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
	"core/internal/middleware"
	"core/internal/repository"
	"core/internal/security"
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
		slog.Int("service_token_count", len(cfg.ServiceTokens)),
	)

	// Fail closed: core is the authoritative data store and must never
	// serve unauthenticated traffic. Without service tokens every API
	// request would be rejected, so refuse to start with a clear hint
	// instead of failing per-request at runtime.
	if len(cfg.ServiceTokens) == 0 {
		slog.Error("service_tokens_not_configured",
			slog.String("hint", "Set SERVICE_TOKENS=portal=<token>,console=<token>,discovery=<token>"))
		os.Exit(1)
	}

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
		// Fail closed: provider API keys are stored in MySQL and must be
		// encrypted at rest. Without a key the only options would be to
		// store plaintext or refuse provider features at runtime; refuse
		// to start instead so the misconfiguration is explicit.
		if cfg.ProviderEncryptionKey == "" {
			slog.Error("provider_encryption_key_not_configured",
				slog.String("hint", "Set PROVIDER_ENCRYPTION_KEY to a base64-encoded 32-byte key (required when MYSQL_DSN is set)"))
			os.Exit(1)
		}
		keyCipher, err := security.NewKeyCipherFromBase64(cfg.ProviderEncryptionKey)
		if err != nil {
			slog.Error("invalid_provider_encryption_key", slog.Any("error", err))
			os.Exit(1)
		}

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
		mergeSvc := service.NewMergeService(neo)
		reviewHandler = handler.NewReviewQueueHandler(reviewRepo, paperRepo, mergeSvc)

		llmProviderRepo := repository.NewLLMProviderRepository(mysqlDB, keyCipher)
		embeddingProviderRepo := repository.NewEmbeddingProviderRepository(mysqlDB, keyCipher)
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

	// Service-token middleware for the three permission tiers. Every
	// /api route requires a valid token; the tier determines which
	// services may call it.
	//
	//   read     any first-party service (graph data is public via portal)
	//   write    console (user-facing RBAC proxy) and discovery (review
	//            queue submissions, paper status updates)
	//   internal discovery only (endpoints that return secrets, e.g. the
	//            LLM provider API key)
	readAuth := middleware.RequireServiceToken(cfg.ServiceTokens, "portal", "console", "discovery")
	writeAuth := middleware.RequireServiceToken(cfg.ServiceTokens, "console", "discovery")
	internalAuth := middleware.RequireServiceToken(cfg.ServiceTokens, "discovery")

	app.Get("/api/graph/data", readAuth, nodeHandler.GraphData)
	app.Get("/api/graph/timeline", readAuth, nodeHandler.Timeline)

	app.Get("/api/nodes", readAuth, nodeHandler.List)
	app.Get("/api/nodes/:id", readAuth, nodeHandler.Get)
	app.Post("/api/nodes", writeAuth, nodeHandler.Create)
	app.Put("/api/nodes/:id", writeAuth, nodeHandler.Update)
	app.Delete("/api/nodes/:id", writeAuth, nodeHandler.Delete)

	app.Get("/api/edges", readAuth, edgeHandler.List)
	app.Get("/api/edges/:source/:target", readAuth, edgeHandler.Get)
	app.Post("/api/edges", writeAuth, edgeHandler.Create)
	app.Put("/api/edges/:source/:target", writeAuth, edgeHandler.Update)
	app.Delete("/api/edges/:source/:target", writeAuth, edgeHandler.Delete)

	app.Get("/api/communities", readAuth, commHandler.List)
	app.Get("/api/communities/tree", readAuth, commHandler.Tree)
	app.Get("/api/communities/:id", readAuth, commHandler.Get)
	app.Post("/api/communities", writeAuth, commHandler.Create)
	app.Put("/api/communities/:id", writeAuth, commHandler.Update)
	app.Delete("/api/communities/:id", writeAuth, commHandler.Delete)
	app.Post("/api/communities/detect", writeAuth, commHandler.Detect)

	// Paper and review queue routes (only when MySQL is configured).
	if paperHandler != nil && reviewHandler != nil {
		app.Get("/api/papers", readAuth, paperHandler.List)
		app.Get("/api/papers/:id", readAuth, paperHandler.Get)
		app.Get("/api/papers/:id/pdf", readAuth, paperHandler.GetPDF)
		app.Post("/api/papers", writeAuth, paperHandler.Create)
		app.Put("/api/papers/:id", writeAuth, paperHandler.Update)
		app.Delete("/api/papers/:id", writeAuth, paperHandler.Delete)

		app.Get("/api/review-queue", readAuth, reviewHandler.List)
		app.Get("/api/review-queue/:id", readAuth, reviewHandler.Get)
		app.Post("/api/review-queue", writeAuth, reviewHandler.Submit)
		app.Post("/api/review-queue/:id/approve", writeAuth, reviewHandler.Approve)
		app.Post("/api/review-queue/:id/reject", writeAuth, reviewHandler.Reject)
	}

	// LLM and embedding provider routes (only when MySQL is configured).
	if provHandlers != nil {
		// LLM provider routes
		app.Get("/api/llm/providers", readAuth, provHandlers.llm.List)
		app.Get("/api/llm/providers/default", readAuth, provHandlers.llm.GetDefault)
		app.Get("/api/llm/providers/:id", readAuth, provHandlers.llm.Get)
		app.Post("/api/llm/providers", writeAuth, provHandlers.llm.Create)
		app.Put("/api/llm/providers/:id", writeAuth, provHandlers.llm.Update)
		app.Delete("/api/llm/providers/:id", writeAuth, provHandlers.llm.Delete)
		app.Post("/api/llm/providers/:id/test", writeAuth, provHandlers.llm.Test)

		// Internal LLM provider route (includes API key for service-to-service use)
		app.Get("/api/internal/llm/providers/default", internalAuth, provHandlers.llm.GetDefaultInternal)

		// Embedding provider routes
		app.Get("/api/embedding/providers", readAuth, provHandlers.embedding.List)
		app.Get("/api/embedding/providers/default", readAuth, provHandlers.embedding.GetDefault)
		app.Get("/api/embedding/providers/:id", readAuth, provHandlers.embedding.Get)
		app.Post("/api/embedding/providers", writeAuth, provHandlers.embedding.Create)
		app.Put("/api/embedding/providers/:id", writeAuth, provHandlers.embedding.Update)
		app.Delete("/api/embedding/providers/:id", writeAuth, provHandlers.embedding.Delete)
		app.Post("/api/embedding/providers/:id/test", writeAuth, provHandlers.embedding.Test)
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
