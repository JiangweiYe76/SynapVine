package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"console/internal/config"
	"console/internal/coreclient"
	"console/internal/db"
	"console/internal/handler"
	"console/internal/model"
	"console/internal/store"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
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
		slog.String("core_url", cfg.CoreURL),
	)

	// Mandatory checks: fail fast when required config is missing.
	if cfg.CoreURL == "" {
		slog.Error("core_url_required",
			slog.String("hint", "Set the CORE_URL environment variable to the core service base URL"),
		)
		os.Exit(1)
	}
	if cfg.MySQLDSN == "" {
		slog.Error("mysql_dsn_required",
			slog.String("hint", "Set the MYSQL_DSN environment variable (e.g. synapvine:synapvine123@tcp(localhost:3306)/synapvine_console?parseTime=true)"),
		)
		os.Exit(1)
	}
	if cfg.JWTSecret == "" {
		slog.Error("jwt_secret_required",
			slog.String("hint", "Set the JWT_SECRET environment variable to a long random string"),
		)
		os.Exit(1)
	}

	// MySQL: connect and run migrations.
	dbConn, err := db.Open(cfg.MySQLDSN)
	if err != nil {
		slog.Error("mysql_connect_failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer dbConn.Close()

	if err := db.Migrate(context.Background(), dbConn); err != nil {
		slog.Error("mysql_migrate_failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("mysql_connected_and_migrated")

	// Core service: health check.
	core := coreclient.New(cfg.CoreURL)
	if err := core.Health(context.Background()); err != nil {
		slog.Error("core_health_check_failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("core_health_check_passed")

	// Stores: each wraps the shared *sql.DB.
	users := store.NewUserStore(dbConn)
	refreshTokens := store.NewRefreshTokenStore(dbConn)
	audit := store.NewAuditStore(dbConn)

	// Handlers.
	authHandler := handler.NewAuthHandler(cfg.JWTSecret, cfg.CookieSecure, users, refreshTokens, audit)
	nodeHandler := handler.NewNodeHandler(core)
	edgeHandler := handler.NewEdgeHandler(core)
	communityHandler := handler.NewCommunityHandler(core)
	llmHandler := handler.NewLLMHandler(core)
	embeddingHandler := handler.NewEmbeddingHandler(core)
	paperHandler := handler.NewPaperHandler(core)
	reviewHandler := handler.NewReviewQueueHandler(core)

	app := fiber.New(fiber.Config{
		AppName:   "AI-Graph Console Server",
		BodyLimit: 50 * 1024 * 1024, // 50 MB — accommodate PDF uploads (base64-encoded)
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigin,
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Content-Type, Authorization",
		AllowCredentials: true,
	}))

	// Rate-limit /api/auth/login to mitigate brute-force and CPU/memory
	// amplification: each call runs argon2id password verification
	// (~64 MiB, ~0.5 s per attempt, see internal/auth/password.go).
	// Other public routes (/auth/refresh, /api/health) are cheap and
	// left unthrottled so legitimate sessions are not affected. Uses
	// c.IP() directly (the limiter default) to prevent clients from
	// bypassing the limit by spoofing X-Forwarded-For.
	loginLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(model.ErrorResponse{
				Error:   "rate_limit_exceeded",
				Message: "Too many login attempts, please try again later",
			})
		},
	})

	// Public routes (no JWT required).
	app.Post("/api/auth/login", loginLimiter, authHandler.Login)
	app.Post("/api/auth/refresh", authHandler.Refresh)

	// Health check endpoint (public, no auth required).
	app.Get("/api/health", func(c *fiber.Ctx) error {
		consoleStatus := "operational"
		coreStatus := "operational"

		if err := dbConn.PingContext(c.Context()); err != nil {
			consoleStatus = "down"
		}
		if err := core.Health(c.Context()); err != nil {
			coreStatus = "down"
		}

		return c.JSON(fiber.Map{
			"console": consoleStatus,
			"core":    coreStatus,
		})
	})

	// Protected routes (JWT required).
	api := app.Group("/api", authHandler.JWTMiddleware())

	// Self.
	api.Get("/me", authHandler.Me)
	api.Post("/auth/logout", authHandler.Logout)

	// Read (viewer+).
	api.Get("/nodes", nodeHandler.List)
	api.Get("/nodes/:id", nodeHandler.Get)
	api.Get("/stats", nodeHandler.Stats)

	// Mutations (editor+).
	editorOnly := handler.RequireRole(model.RoleAdmin, model.RoleEditor)
	api.Post("/nodes", editorOnly, nodeHandler.Create)
	api.Put("/nodes/:id", editorOnly, nodeHandler.Update)
	api.Delete("/nodes/:id", editorOnly, nodeHandler.Delete)

	// Edge read (viewer+).
	api.Get("/edges", edgeHandler.List)
	api.Get("/edges/:source/:target", edgeHandler.Get)

	// Edge mutations (editor+).
	api.Post("/edges", editorOnly, edgeHandler.Create)
	api.Put("/edges/:source/:target", editorOnly, edgeHandler.Update)
	api.Delete("/edges/:source/:target", editorOnly, edgeHandler.Delete)

	// Community read (viewer+).
	api.Get("/communities", communityHandler.List)
	api.Get("/communities/tree", communityHandler.Tree)
	api.Get("/communities/:id", communityHandler.Get)

	// Community mutations (editor+).
	api.Post("/communities", editorOnly, communityHandler.Create)
	api.Put("/communities/:id", editorOnly, communityHandler.Update)
	api.Delete("/communities/:id", editorOnly, communityHandler.Delete)

	// LLM provider read (viewer+).
	api.Get("/llm/providers", llmHandler.List)
	api.Get("/llm/providers/default", llmHandler.GetDefault)
	api.Get("/llm/providers/:id", llmHandler.Get)

	// LLM provider mutations (editor+).
	api.Post("/llm/providers", editorOnly, llmHandler.Create)
	api.Put("/llm/providers/:id", editorOnly, llmHandler.Update)
	api.Post("/llm/providers/:id/test", editorOnly, llmHandler.Test)

	// LLM provider delete (admin+).
	adminOnly := handler.RequireRole(model.RoleAdmin)
	api.Delete("/llm/providers/:id", adminOnly, llmHandler.Delete)

	// Embedding provider read (viewer+).
	api.Get("/embedding/providers", embeddingHandler.List)
	api.Get("/embedding/providers/default", embeddingHandler.GetDefault)
	api.Get("/embedding/providers/:id", embeddingHandler.Get)

	// Embedding provider mutations (editor+).
	api.Post("/embedding/providers", editorOnly, embeddingHandler.Create)
	api.Put("/embedding/providers/:id", editorOnly, embeddingHandler.Update)
	api.Post("/embedding/providers/:id/test", editorOnly, embeddingHandler.Test)

	// Embedding provider delete (admin+).
	api.Delete("/embedding/providers/:id", adminOnly, embeddingHandler.Delete)

	// Paper read (viewer+).
	api.Get("/papers", paperHandler.List)
	api.Get("/papers/stats", paperHandler.Stats)
	api.Get("/papers/:id", paperHandler.Get)
	api.Get("/papers/:id/pdf", paperHandler.GetPDF)

	// Paper mutations (editor+).
	api.Post("/papers", editorOnly, paperHandler.Create)
	api.Put("/papers/:id", editorOnly, paperHandler.Update)
	api.Delete("/papers/:id", adminOnly, paperHandler.Delete)

	// Review queue (viewer+ for read, editor+ for actions).
	api.Get("/review-queue", reviewHandler.List)
	api.Get("/review-queue/:id", reviewHandler.Get)
	api.Post("/review-queue/:id/approve", editorOnly, reviewHandler.Approve)
	api.Post("/review-queue/:id/reject", editorOnly, reviewHandler.Reject)

	slog.Info("console_server_starting", slog.String("port", cfg.Port))

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
