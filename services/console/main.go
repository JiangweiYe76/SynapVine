package main

import (
	"context"
	"log/slog"
	"os"

	"console/internal/config"
	"console/internal/coreclient"
	"console/internal/handler"

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

	// Core service is mandatory: fail fast when CORE_URL is missing.
	if cfg.CoreURL == "" {
		slog.Error("core_url_required",
			slog.String("hint", "Set the CORE_URL environment variable to the core service base URL"),
		)
		os.Exit(1)
	}

	core := coreclient.New(cfg.CoreURL)
	if err := core.Health(context.Background()); err != nil {
		slog.Error("core_health_check_failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("core_health_check_passed")

	authHandler := handler.NewAuthHandler(cfg.JWTSecret)
	nodeHandler := handler.NewNodeHandler(core)
	edgeHandler := handler.NewEdgeHandler(core)

	app := fiber.New(fiber.Config{
		AppName: "AI-Graph Console Server",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigin,
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Content-Type, Authorization",
		AllowCredentials: true,
	}))

	app.Post("/api/auth/login", authHandler.Login)

	api := app.Group("/api", authHandler.JWTMiddleware())
	api.Get("/me", authHandler.Me)

	api.Get("/nodes", nodeHandler.List)
	api.Get("/nodes/:id", nodeHandler.Get)
	api.Post("/nodes", nodeHandler.Create)
	api.Put("/nodes/:id", nodeHandler.Update)
	api.Delete("/nodes/:id", nodeHandler.Delete)

	api.Get("/edges", edgeHandler.List)
	api.Get("/edges/:source/:target", edgeHandler.Get)
	api.Post("/edges", edgeHandler.Create)
	api.Put("/edges/:source/:target", edgeHandler.Update)
	api.Delete("/edges/:source/:target", edgeHandler.Delete)

	api.Get("/stats", nodeHandler.Stats)

	slog.Info("console_server_starting", slog.String("port", cfg.Port))

	if err := app.Listen(":" + cfg.Port); err != nil {
		slog.Error("server_failed", slog.Any("error", err))
		os.Exit(1)
	}
}
