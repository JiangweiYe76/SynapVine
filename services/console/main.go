package main

import (
	"log/slog"
	"os"

	"console/internal/config"
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
	)

	app := fiber.New(fiber.Config{
		AppName: "AI-Graph Console Server",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigin,
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Content-Type, Authorization",
		AllowCredentials: true,
	}))

	authHandler := handler.NewAuthHandler(cfg.JWTSecret)

	app.Post("/api/auth/login", authHandler.Login)

	api := app.Group("/api", authHandler.JWTMiddleware())
	api.Get("/me", authHandler.Me)

	slog.Info("console_server_starting", slog.String("port", cfg.Port))

	if err := app.Listen(":" + cfg.Port); err != nil {
		slog.Error("server_failed", slog.Any("error", err))
		os.Exit(1)
	}
}
