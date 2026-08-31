package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"ai-graph-server/internal/config"
	"ai-graph-server/internal/coreclient"
	"ai-graph-server/internal/handler"
	"ai-graph-server/internal/middleware"
	"ai-graph-server/internal/security"
	"ai-graph-server/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {
	// Initialize structured logger with JSON output
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration from environment variables
	cfg := config.Load()
	slog.Info("configuration_loaded",
		slog.String("port", cfg.Port),
		slog.String("allowed_origin", cfg.AllowedOrigin),
		slog.String("core_url", cfg.CoreURL),
	)

	// Core is a stateless dependency; the portal reads through it on
	// every request, so console writes are visible without a restart.
	core := coreclient.New(cfg.CoreURL)

	// Initialize service and handler
	svc := service.New(core)
	gh := handler.NewGraphHandler(svc)

	// Initialize token store for API authentication
	tokenStore := security.NewTokenStore()

	// Start background goroutine to clean expired tokens every minute.
	// The goroutine exits when the shutdown signal closes cleanerDone.
	cleanerDone := make(chan struct{})
	go func() {
		for {
			select {
			case <-cleanerDone:
				return
			case <-time.After(1 * time.Minute):
				cleaned := tokenStore.CleanExpired()
				if cleaned > 0 {
					slog.Info("tokens_cleaned", slog.Int("count", cleaned))
				}
			}
		}
	}()

	// Create Fiber app instance
	app := fiber.New(fiber.Config{
		AppName: "AI-Graph Server",
	})

	// Apply rate limiting middleware (60 requests per minute per IP).
	// Uses c.IP() directly to prevent clients from bypassing rate limits
	// by spoofing the X-Forwarded-For header.
	app.Use(limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error":       "rate_limit_exceeded",
				"retry_after": 60,
			})
		},
	}))

	// Apply request logging middleware
	app.Use(middleware.Logger())

	// Apply CORS middleware.
	// Configure allowed origins via ALLOWED_ORIGIN env var (default: http://localhost:5173).
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigin,
		AllowMethods:     "GET, POST",
		AllowCredentials: true,
	}))

	// Token endpoint - issues temporary access tokens
	app.Get("/api/token", func(c *fiber.Ctx) error {
		ua := c.Get("User-Agent")
		// Basic bot detection - reject requests with missing or short User-Agent
		if ua == "" || len(ua) < 20 {
			return c.Status(403).JSON(fiber.Map{
				"error":   "forbidden",
				"message": "Invalid client",
			})
		}
		token, err := tokenStore.Issue()
		if err != nil {
			slog.Error("token_issue_failed", slog.Any("error", err))
			return c.Status(500).JSON(fiber.Map{
				"error":   "token_error",
				"message": "Failed to generate token",
			})
		}
		return c.JSON(fiber.Map{
			"token": token,
		})
	})

	// Protected API routes - require valid token
	api := app.Group("/api/graph", func(c *fiber.Ctx) error {
		// Authenticate exclusively via the Authorization: Bearer <token> header
		authToken := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		// Validate token
		if authToken == "" || !tokenStore.Validate(authToken) {
			return c.Status(401).JSON(fiber.Map{
				"error":   "invalid_token",
				"message": "Token is invalid or expired",
			})
		}
		return c.Next()
	})

	// Register API route handlers
	api.Get("/summary", gh.Summary)
	api.Get("/timeline", gh.Timeline)
	api.Get("/nodes", gh.Nodes)
	api.Get("/nodes/:id", gh.NodeDetail)
	api.Get("/nodes/:id/edges", gh.NodeEdges)
	api.Get("/search", gh.Search)
	api.Get("/expand", gh.Expand)

	// Log startup information
	slog.Info("server_starting", slog.String("port", cfg.Port))

	// Graceful shutdown on SIGINT/SIGTERM, mirroring the pattern used by
	// the core service.
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
		close(cleanerDone)
	}()

	// Start HTTP server
	if err := app.Listen(":" + cfg.Port); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("server_failed", slog.Any("error", err))
		os.Exit(1)
	}

	<-idleConnsClosed
}
