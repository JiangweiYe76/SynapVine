package handler

import (
	"errors"
	"log/slog"

	"console/internal/coreclient"
	"console/internal/model"

	"github.com/gofiber/fiber/v2"
)

// LLMHandler proxies LLM provider requests to the core service.
type LLMHandler struct {
	core *coreclient.Client
}

// NewLLMHandler creates a new LLMHandler backed by the given core client.
func NewLLMHandler(core *coreclient.Client) *LLMHandler {
	return &LLMHandler{core: core}
}

// List handles GET /api/llm/providers
func (h *LLMHandler) List(c *fiber.Ctx) error {
	resp, err := h.core.ListLLMProviders(c.Context())
	if err != nil {
		slog.Error("llm_providers_list_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to list LLM providers from core service",
		})
	}
	return c.JSON(resp)
}

// Get handles GET /api/llm/providers/:id
func (h *LLMHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	provider, err := h.core.GetLLMProvider(c.Context(), id)
	if err != nil {
		slog.Error("llm_provider_get_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to get LLM provider from core service",
		})
	}
	if provider == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "provider_not_found",
			Message: "LLM provider not found",
		})
	}
	return c.JSON(provider)
}

// Create handles POST /api/llm/providers
func (h *LLMHandler) Create(c *fiber.Ctx) error {
	var req model.LLMProviderCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	provider, err := h.core.CreateLLMProvider(c.Context(), req)
	if err != nil {
		var httpErr *coreclient.HTTPStatusError
		if errors.As(err, &httpErr) {
			if httpErr.StatusCode == 409 {
				return c.Status(409).JSON(model.ErrorResponse{
					Error:   "provider_exists",
					Message: "A provider with this name already exists",
				})
			}
			if httpErr.StatusCode == 400 {
				return c.Status(400).JSON(model.ErrorResponse{
					Error:   "missing_fields",
					Message: "Name, base_url, api_key, and model are required",
				})
			}
		}
		slog.Error("llm_provider_create_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to create LLM provider in core service",
		})
	}

	slog.Info("llm_provider_created", slog.String("id", provider.ID), slog.String("name", provider.Name))
	return c.Status(201).JSON(provider)
}

// Update handles PUT /api/llm/providers/:id
func (h *LLMHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.LLMProviderUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	provider, err := h.core.UpdateLLMProvider(c.Context(), id, req)
	if err != nil {
		var httpErr *coreclient.HTTPStatusError
		if errors.As(err, &httpErr) {
			if httpErr.StatusCode == 409 {
				return c.Status(409).JSON(model.ErrorResponse{
					Error:   "provider_exists",
					Message: "A provider with this name already exists",
				})
			}
		}
		slog.Error("llm_provider_update_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to update LLM provider in core service",
		})
	}
	if provider == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "provider_not_found",
			Message: "LLM provider not found",
		})
	}

	slog.Info("llm_provider_updated", slog.String("id", provider.ID))
	return c.JSON(provider)
}

// Delete handles DELETE /api/llm/providers/:id
func (h *LLMHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	ok, err := h.core.DeleteLLMProvider(c.Context(), id)
	if err != nil {
		slog.Error("llm_provider_delete_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to delete LLM provider in core service",
		})
	}
	if !ok {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "provider_not_found",
			Message: "LLM provider not found",
		})
	}

	slog.Info("llm_provider_deleted", slog.String("id", id))
	return c.SendStatus(204)
}

// GetDefault handles GET /api/llm/providers/default
func (h *LLMHandler) GetDefault(c *fiber.Ctx) error {
	provider, err := h.core.GetDefaultLLMProvider(c.Context())
	if err != nil {
		slog.Error("llm_default_get_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to get default LLM provider from core service",
		})
	}
	if provider == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "no_default_provider",
			Message: "No default LLM provider configured",
		})
	}
	return c.JSON(provider)
}

// Test handles POST /api/llm/providers/:id/test
func (h *LLMHandler) Test(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.core.TestLLMProvider(c.Context(), id)
	if err != nil {
		var httpErr *coreclient.HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "provider_not_found",
				Message: "LLM provider not found",
			})
		}
		slog.Error("llm_provider_test_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to test LLM provider via core service",
		})
	}
	return c.JSON(resp)
}
