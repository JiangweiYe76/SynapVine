package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"discovery/internal/arxiv"
	"discovery/internal/config"
	"discovery/internal/coreclient"
	"discovery/internal/extractor"
	"discovery/internal/handler"
	"discovery/internal/middleware"
	"discovery/internal/pipeline"
	"discovery/internal/scheduler"

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
		slog.String("core_url", cfg.CoreURL),
		slog.String("allowed_origin", cfg.AllowedOrigin),
	)

	// Fail closed: /api/analyze triggers paid LLM extraction, so it must
	// not serve unauthenticated traffic. Without accepted tokens every
	// analyze request would be rejected, so refuse to start with a clear
	// hint instead of failing per-request at runtime.
	if len(cfg.ServiceTokens) == 0 {
		slog.Error("service_tokens_not_configured",
			slog.String("hint", "Set SERVICE_TOKENS=console=<token> (the console token) to authorize analyze callers"))
		os.Exit(1)
	}

	// The service token authenticates discovery to core; without it
	// every core request is rejected with 401.
	if cfg.ServiceToken == "" {
		slog.Warn("service_token_not_configured",
			slog.String("hint", "Set SERVICE_TOKEN to the discovery token configured in core's SERVICE_TOKENS"))
	}

	// Core service: health check.
	core := coreclient.New(cfg.CoreURL, cfg.ServiceToken)
	if err := core.Health(context.Background()); err != nil {
		slog.Error("core_health_check_failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("core_health_check_passed")

	// Extractor service.
	ext := extractor.NewService()

	// Shared extraction pipeline: the HTTP handler and the ArXiv
	// scheduler both push papers through it.
	pipe := pipeline.NewService(core, ext)

	// Handler.
	analyzeHandler := handler.NewAnalyzeHandler(pipe)

	app := fiber.New(fiber.Config{
		AppName: "AI-Graph Discovery Server",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.AllowedOrigin,
		AllowMethods: "GET, POST, OPTIONS",
		AllowHeaders: "Content-Type, Authorization",
	}))

	app.Get("/health", analyzeHandler.Health)
	// /api/analyze is only called by the console backend (auto-trigger
	// after paper upload), so it requires a valid console service token.
	app.Post("/api/analyze", middleware.RequireServiceToken(cfg.ServiceTokens, "console"), analyzeHandler.Analyze)

	slog.Info("discovery_server_starting", slog.String("port", cfg.Port))

	// Scheduler stop handles; nil when ArXiv ingestion is disabled.
	var schedulerCancel context.CancelFunc
	var schedulerDone chan struct{}

	// ArXiv ingestion scheduler (opt-in via ARXIV_ENABLED=true). It runs
	// its first cycle immediately and then once per interval; shutdown
	// cancels it and waits for the in-flight cycle to finish before the
	// server is considered stopped.
	if cfg.ArxivEnabled {
		// The scheduler talks to core as discovery; without a token its
		// paper creation and lookups would be rejected on every cycle.
		if cfg.ServiceToken == "" {
			slog.Error("arxiv_scheduler_requires_service_token",
				slog.String("hint", "Set SERVICE_TOKEN to the discovery token configured in core's SERVICE_TOKENS, or disable ARXIV_ENABLED"))
			os.Exit(1)
		}
		sched := scheduler.New(arxiv.NewClient(), core, pipe, scheduler.Config{
			Categories: cfg.ArxivCategories,
			Interval:   cfg.ArxivInterval,
			MaxPerRun:  cfg.ArxivMaxPerRun,
		})
		schedCtx, schedCancel := context.WithCancel(context.Background())
		schedDone := make(chan struct{})
		go func() {
			defer close(schedDone)
			sched.Run(schedCtx)
		}()
		defer schedCancel()
		schedulerDone = schedDone
		schedulerCancel = schedCancel
	}

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
		// Stop the ingestion scheduler, if any, and wait for the
		// current cycle to wrap up before declaring shutdown complete.
		if schedulerCancel != nil {
			schedulerCancel()
			<-schedulerDone
		}
		close(idleConnsClosed)
	}()

	if err := app.Listen(":" + cfg.Port); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("server_failed", slog.Any("error", err))
		os.Exit(1)
	}

	<-idleConnsClosed
}
