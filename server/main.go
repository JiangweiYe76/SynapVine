package main

import (
	"log"
	"strings"
	"time"

	"ai-graph-server/internal/community"
	"ai-graph-server/internal/config"
	"ai-graph-server/internal/handler"
	"ai-graph-server/internal/loader"
	"ai-graph-server/internal/security"
	"ai-graph-server/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Load graph data from JSON file
	graphData, err := loader.LoadGraphData(cfg.DataPath)
	if err != nil {
		log.Fatalf("Failed to load graph data from %s: %v", cfg.DataPath, err)
	}

	// Detect flat communities using Louvain algorithm
	communities := community.Detect(graphData.Nodes, graphData.Edges)
	community.AssignCommunities(graphData.Nodes, communities)
	community.ComputeDegrees(&graphData.Nodes, graphData.Edges)

	// Detect hierarchical communities for multi-level view
	communityConfig := community.CommunityConfig{
		MaxLevels:        3,
		MinCommunitySize: 3,
	}
	hierarchicalCommunities, maxLevel := community.DetectHierarchical(graphData.Nodes, graphData.Edges, communityConfig)
	community.AssignHierarchicalCommunities(graphData.Nodes, hierarchicalCommunities)

	// Initialize service and handler
	svc := service.New(graphData.Nodes, graphData.Edges, communities, hierarchicalCommunities, maxLevel)
	gh := handler.NewGraphHandler(svc)

	// Initialize token store for API authentication
	tokenStore := security.NewTokenStore()

	// Start background goroutine to clean expired tokens every minute
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			tokenStore.CleanExpired()
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
				"message": "无效的客户端",
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
				"message": "token 无效或已过期",
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
	log.Printf("🚀 AI-Graph Server starting on :%s", cfg.Port)
	log.Printf("📊 Loaded %d nodes, %d edges, %d hierarchical communities (max level %d)",
		len(graphData.Nodes), len(graphData.Edges),
		community.CountAllCommunities(hierarchicalCommunities), maxLevel)

	// Start HTTP server
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
