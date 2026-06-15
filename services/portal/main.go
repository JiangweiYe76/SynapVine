package main

import (
	"log/slog"
	"os"
	"strings"
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

	// Start background goroutine to clean expired tokens every minute
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			cleaned := tokenStore.CleanExpired()
			if cleaned > 0 {
				slog.Info("tokens_cleaned", slog.Int("count", cleaned))
			}
		}
	}()

	// Create Fiber app instance
	app := fiber.New(fiber.Config{
		AppName: "AI-Graph Server",
	})

	// Apply rate limiting middleware (60 requests per minute per IP)
	app.Use(limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use X-Forwarded-For header if present, otherwise use IP
			if xff := c.Get("X-Forwarded-For"); xff != "" {
				return strings.Split(xff, ",")[0]
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error":       "rate_limit_exceeded",
				"retry_after": 60,
			})
		},
	}))

	// Apply request logging middleware
	app.Use(middleware.Logger())

	// Apply CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigin + ", http://localhost:5173",
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
		return c.JSON(fiber.Map{
			"token": tokenStore.Issue(),
		})
	})

	// Protected API routes - require valid token
	api := app.Group("/api/graph", func(c *fiber.Ctx) error {
		// Try to get token from query param or Authorization header
		token := c.Query("token", "")
		if token == "" {
			token = c.Get("Authorization")
			token = strings.TrimPrefix(token, "Bearer ")
		}
		// Validate token
		if token == "" || !tokenStore.Validate(token) {
			return c.Status(401).JSON(fiber.Map{
				"error":   "invalid_token",
				"message": "Token is invalid or expired",
			})
		}
		return c.Next()
	})

	// Register API route handlers
	api.Get("/summary", gh.Summary)
	api.Get("/nodes", gh.Nodes)
	api.Get("/nodes/:id", gh.NodeDetail)
	api.Get("/nodes/:id/edges", gh.NodeEdges)
	api.Get("/search", gh.Search)
	api.Get("/expand", gh.Expand)

	// Log startup information
	slog.Info("server_starting", slog.String("port", cfg.Port))

	// Start HTTP server
	if err := app.Listen(":" + cfg.Port); err != nil {
		slog.Error("server_failed", slog.Any("error", err))
		os.Exit(1)
	}
}
