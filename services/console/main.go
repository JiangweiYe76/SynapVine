package main

import (
	"context"
	"log/slog"
	"os"

	"console/internal/config"
	"console/internal/coreclient"
	"console/internal/db"
	"console/internal/handler"
	"console/internal/model"
	"console/internal/store"

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
	authHandler := handler.NewAuthHandler(cfg.JWTSecret, users, refreshTokens, audit)
	nodeHandler := handler.NewNodeHandler(core)
	edgeHandler := handler.NewEdgeHandler(core)
	communityHandler := handler.NewCommunityHandler(core)

	app := fiber.New(fiber.Config{
		AppName: "AI-Graph Console Server",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigin,
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Content-Type, Authorization",
		AllowCredentials: true,
	}))

	// Public auth routes (no JWT required).
	app.Post("/api/auth/login", authHandler.Login)
	app.Post("/api/auth/refresh", authHandler.Refresh)

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

	slog.Info("console_server_starting", slog.String("port", cfg.Port))

	if err := app.Listen(":" + cfg.Port); err != nil {
		slog.Error("server_failed", slog.Any("error", err))
		os.Exit(1)
	}
}
