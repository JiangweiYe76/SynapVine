package handler

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"core/internal/model"
	"core/internal/repository"

	"github.com/gofiber/fiber/v2"
)

// LLMUsageHandler records and aggregates LLM token usage.
type LLMUsageHandler struct {
	repo *repository.LLMUsageRepository
}

// NewLLMUsageHandler creates a new LLMUsageHandler.
func NewLLMUsageHandler(repo *repository.LLMUsageRepository) *LLMUsageHandler {
	return &LLMUsageHandler{repo: repo}
}

// Record handles POST /api/internal/llm/usage — service-to-service
// usage submission after a successful extraction.
func (h *LLMUsageHandler) Record(c *fiber.Ctx) error {
	var req model.LLMUsageRecordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_json",
			Message: "Failed to parse request body",
		})
	}
	if req.PaperID == "" || req.ProviderID == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "missing_fields",
			Message: "paper_id and provider_id are required",
		})
	}

	if err := h.repo.Insert(c.Context(), &req); err != nil {
		slog.Error("llm_usage_record_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to record LLM usage",
		})
	}

	return c.Status(201).JSON(map[string]bool{"ok": true})
}

// Summary handles GET /api/internal/llm/usage/summary — aggregated
// usage consumed by the console overview. Optional from/to query
// params (YYYY-MM-DD) bound the window; to is inclusive.
func (h *LLMUsageHandler) Summary(c *fiber.Ctx) error {
	filter := repository.UsageFilter{}
	if raw := c.Query("from"); raw != "" {
		from, err := parseUsageDate(raw, false)
		if err != nil {
			return c.Status(400).JSON(model.ErrorResponse{
				Error:   "invalid_from",
				Message: err.Error(),
			})
		}
		filter.From = from
	}
	if raw := c.Query("to"); raw != "" {
		to, err := parseUsageDate(raw, true)
		if err != nil {
			return c.Status(400).JSON(model.ErrorResponse{
				Error:   "invalid_to",
				Message: err.Error(),
			})
		}
		filter.To = to
	}
	if filter.From != nil && filter.To != nil && filter.To.Before(*filter.From) {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_range",
			Message: "to must not be before from",
		})
	}

	summary, err := h.repo.Summary(c.Context(), filter)
	if err != nil {
		slog.Error("llm_usage_summary_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to aggregate LLM usage",
		})
	}

	// Expose nothing extra: the ByDay series already bounds the window.
	return c.JSON(summary)
}

// parseUsageDate parses a YYYY-MM-DD query parameter into a time.Time
// bound (start: beginning of day; end: exclusive start of next day).
func parseUsageDate(s string, endOfDay bool) (*time.Time, error) {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(s))
	if err != nil {
		return nil, fmt.Errorf("invalid date %q, want YYYY-MM-DD", s)
	}
	if endOfDay {
		t = t.AddDate(0, 0, 1)
	}
	return &t, nil
}
