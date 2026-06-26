package handler

import (
	"errors"
	"log/slog"
	"time"

	"core/internal/llm"
	"core/internal/model"
	"core/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// LLMProviderHandler manages LLM provider configuration and testing.
type LLMProviderHandler struct {
	repo    *repository.LLMProviderRepository
	manager *llm.Manager
}

// NewLLMProviderHandler creates a new LLMProviderHandler.
func NewLLMProviderHandler(repo *repository.LLMProviderRepository) *LLMProviderHandler {
	return &LLMProviderHandler{
		repo:    repo,
		manager: llm.NewManager(repo),
	}
}

// List handles GET /api/llm/providers
func (h *LLMProviderHandler) List(c *fiber.Ctx) error {
	providers, err := h.repo.List(c.Context())
	if err != nil {
		slog.Error("llm_providers_list_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list LLM providers",
		})
	}

	resp := make([]model.LLMProviderResponse, 0, len(providers))
	for _, p := range providers {
		resp = append(resp, p.ToResponse())
	}

	return c.JSON(model.LLMProviderListResponse{
		Providers: resp,
		Total:     len(resp),
	})
}

// Get handles GET /api/llm/providers/:id
func (h *LLMProviderHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	provider, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "provider_not_found",
				Message: "LLM provider not found",
			})
		}
		slog.Error("llm_provider_get_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get LLM provider",
		})
	}

	return c.JSON(provider.ToResponse())
}

// Create handles POST /api/llm/providers
func (h *LLMProviderHandler) Create(c *fiber.Ctx) error {
	var req model.LLMProviderCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.Name == "" || req.BaseURL == "" || req.APIKey == "" || req.Model == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "missing_fields",
			Message: "Name, base_url, api_key, and model are required",
		})
	}

	if req.MaxTokens <= 0 {
		req.MaxTokens = 4096
	}
	if req.Temperature <= 0 {
		req.Temperature = 0.7
	}

	now := time.Now()
	p := &model.LLMProvider{
		ID:          uuid.New().String(),
		Name:        req.Name,
		BaseURL:     req.BaseURL,
		APIKey:      req.APIKey,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		IsDefault:   req.IsDefault,
		IsEnabled:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// If this provider is marked as default, clear any existing default first.
	if p.IsDefault {
		if err := h.repo.ClearDefault(c.Context()); err != nil {
			slog.Error("llm_clear_default_failed", slog.Any("error", err))
			return c.Status(500).JSON(model.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to update default provider",
			})
		}
	}

	if err := h.repo.Create(c.Context(), p); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return c.Status(409).JSON(model.ErrorResponse{
				Error:   "provider_exists",
				Message: "A provider with this name already exists",
			})
		}
		slog.Error("llm_provider_create_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create LLM provider",
		})
	}

	slog.Info("llm_provider_created", slog.String("id", p.ID), slog.String("name", p.Name))
	return c.Status(201).JSON(p.ToResponse())
}

// Update handles PUT /api/llm/providers/:id
func (h *LLMProviderHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.LLMProviderUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// If setting this provider as default, clear existing default first.
	if req.IsDefault != nil && *req.IsDefault {
		if err := h.repo.ClearDefault(c.Context()); err != nil {
			slog.Error("llm_clear_default_failed", slog.Any("error", err))
			return c.Status(500).JSON(model.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to update default provider",
			})
		}
	}

	p, err := h.repo.Update(c.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "provider_not_found",
				Message: "LLM provider not found",
			})
		}
		if errors.Is(err, repository.ErrDuplicate) {
			return c.Status(409).JSON(model.ErrorResponse{
				Error:   "provider_exists",
				Message: "A provider with this name already exists",
			})
		}
		slog.Error("llm_provider_update_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update LLM provider",
		})
	}

	slog.Info("llm_provider_updated", slog.String("id", p.ID))
	return c.JSON(p.ToResponse())
}

// Delete handles DELETE /api/llm/providers/:id
func (h *LLMProviderHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.repo.Delete(c.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "provider_not_found",
				Message: "LLM provider not found",
			})
		}
		slog.Error("llm_provider_delete_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete LLM provider",
		})
	}

	slog.Info("llm_provider_deleted", slog.String("id", id))
	return c.SendStatus(204)
}

// GetDefault handles GET /api/llm/providers/default
func (h *LLMProviderHandler) GetDefault(c *fiber.Ctx) error {
	provider, err := h.repo.GetDefault(c.Context())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "no_default_provider",
				Message: "No default LLM provider configured",
			})
		}
		slog.Error("llm_default_get_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get default LLM provider",
		})
	}

	return c.JSON(provider.ToResponse())
}

// GetDefaultInternal handles GET /api/internal/llm/providers/default
// Returns the default LLM provider including the API key.
// This endpoint is for internal service-to-service communication only.
func (h *LLMProviderHandler) GetDefaultInternal(c *fiber.Ctx) error {
	provider, err := h.repo.GetDefault(c.Context())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "no_default_provider",
				Message: "No default LLM provider configured",
			})
		}
		slog.Error("llm_default_get_internal_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get default LLM provider",
		})
	}

	return c.JSON(provider.ToInternalResponse())
}

// Test handles POST /api/llm/providers/:id/test
func (h *LLMProviderHandler) Test(c *fiber.Ctx) error {
	id := c.Params("id")

	client, err := h.manager.ClientByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "provider_not_found",
				Message: "LLM provider not found",
			})
		}
		return c.Status(400).JSON(model.LLMTestResponse{
			OK:    false,
			Error: err.Error(),
		})
	}

	start := time.Now()
	resp, err := client.TestConnectivity(c.Context())
	latency := time.Since(start).Milliseconds()

	if err != nil {
		slog.Warn("llm_test_failed", slog.String("id", id), slog.Any("error", err))
		return c.JSON(model.LLMTestResponse{
			OK:        false,
			LatencyMs: latency,
			Error:     err.Error(),
		})
	}

	slog.Info("llm_test_passed", slog.String("id", id), slog.Int64("latency_ms", latency))
	return c.JSON(model.LLMTestResponse{
		OK:        true,
		Model:     resp.Content,
		LatencyMs: latency,
	})
}
