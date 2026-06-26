package handler

import (
	"errors"
	"log/slog"

	"console/internal/coreclient"
	"console/internal/model"

	"github.com/gofiber/fiber/v2"
)

// EmbeddingHandler proxies embedding provider requests to the core service.
type EmbeddingHandler struct {
	core *coreclient.Client
}

// NewEmbeddingHandler creates a new EmbeddingHandler backed by the given core client.
func NewEmbeddingHandler(core *coreclient.Client) *EmbeddingHandler {
	return &EmbeddingHandler{core: core}
}

// List handles GET /api/embedding/providers
func (h *EmbeddingHandler) List(c *fiber.Ctx) error {
	resp, err := h.core.ListEmbeddingProviders(c.Context())
	if err != nil {
		slog.Error("embedding_providers_list_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to list embedding providers from core service",
		})
	}
	return c.JSON(resp)
}

// Get handles GET /api/embedding/providers/:id
func (h *EmbeddingHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	provider, err := h.core.GetEmbeddingProvider(c.Context(), id)
	if err != nil {
		slog.Error("embedding_provider_get_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to get embedding provider from core service",
		})
	}
	if provider == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "provider_not_found",
			Message: "Embedding provider not found",
		})
	}
	return c.JSON(provider)
}

// Create handles POST /api/embedding/providers
func (h *EmbeddingHandler) Create(c *fiber.Ctx) error {
	var req model.EmbeddingProviderCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	provider, err := h.core.CreateEmbeddingProvider(c.Context(), req)
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
		slog.Error("embedding_provider_create_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to create embedding provider in core service",
		})
	}

	slog.Info("embedding_provider_created", slog.String("id", provider.ID), slog.String("name", provider.Name))
	return c.Status(201).JSON(provider)
}

// Update handles PUT /api/embedding/providers/:id
func (h *EmbeddingHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.EmbeddingProviderUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	provider, err := h.core.UpdateEmbeddingProvider(c.Context(), id, req)
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
		slog.Error("embedding_provider_update_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to update embedding provider in core service",
		})
	}
	if provider == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "provider_not_found",
			Message: "Embedding provider not found",
		})
	}

	slog.Info("embedding_provider_updated", slog.String("id", provider.ID))
	return c.JSON(provider)
}

// Delete handles DELETE /api/embedding/providers/:id
func (h *EmbeddingHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	ok, err := h.core.DeleteEmbeddingProvider(c.Context(), id)
	if err != nil {
		slog.Error("embedding_provider_delete_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to delete embedding provider in core service",
		})
	}
	if !ok {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "provider_not_found",
			Message: "Embedding provider not found",
		})
	}

	slog.Info("embedding_provider_deleted", slog.String("id", id))
	return c.SendStatus(204)
}

// GetDefault handles GET /api/embedding/providers/default
func (h *EmbeddingHandler) GetDefault(c *fiber.Ctx) error {
	provider, err := h.core.GetDefaultEmbeddingProvider(c.Context())
	if err != nil {
		slog.Error("embedding_default_get_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to get default embedding provider from core service",
		})
	}
	if provider == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "no_default_provider",
			Message: "No default embedding provider configured",
		})
	}
	return c.JSON(provider)
}

// Test handles POST /api/embedding/providers/:id/test
func (h *EmbeddingHandler) Test(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.core.TestEmbeddingProvider(c.Context(), id)
	if err != nil {
		var httpErr *coreclient.HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "provider_not_found",
				Message: "Embedding provider not found",
			})
		}
		slog.Error("embedding_provider_test_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to test embedding provider via core service",
		})
	}
	return c.JSON(resp)
}
