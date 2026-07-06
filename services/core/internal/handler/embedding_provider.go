package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"core/internal/model"
	"core/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// EmbeddingProviderHandler manages embedding provider configuration and testing.
type EmbeddingProviderHandler struct {
	repo *repository.EmbeddingProviderRepository
}

// NewEmbeddingProviderHandler creates a new EmbeddingProviderHandler.
func NewEmbeddingProviderHandler(repo *repository.EmbeddingProviderRepository) *EmbeddingProviderHandler {
	return &EmbeddingProviderHandler{repo: repo}
}

// List handles GET /api/embedding/providers
func (h *EmbeddingProviderHandler) List(c *fiber.Ctx) error {
	providers, err := h.repo.List(c.Context())
	if err != nil {
		slog.Error("embedding_providers_list_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list embedding providers",
		})
	}

	resp := make([]model.EmbeddingProviderResponse, 0, len(providers))
	for _, p := range providers {
		resp = append(resp, p.ToResponse())
	}

	return c.JSON(model.EmbeddingProviderListResponse{
		Providers: resp,
		Total:     len(resp),
	})
}

// Get handles GET /api/embedding/providers/:id
func (h *EmbeddingProviderHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	provider, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "provider_not_found",
				Message: "Embedding provider not found",
			})
		}
		slog.Error("embedding_provider_get_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get embedding provider",
		})
	}

	return c.JSON(provider.ToResponse())
}

// Create handles POST /api/embedding/providers
func (h *EmbeddingProviderHandler) Create(c *fiber.Ctx) error {
	var req model.EmbeddingProviderCreateRequest
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

	if req.Dimensions <= 0 {
		req.Dimensions = 1536
	}

	now := time.Now()
	p := &model.EmbeddingProvider{
		ProviderBase: model.ProviderBase{
			ID:        uuid.New().String(),
			Name:      req.Name,
			BaseURL:   req.BaseURL,
			APIKey:    req.APIKey,
			Model:     req.Model,
			IsDefault: req.IsDefault,
			IsEnabled: true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Dimensions: req.Dimensions,
	}

	// If this provider is marked as default, clear any existing default first.
	if p.IsDefault {
		if err := h.repo.ClearDefault(c.Context()); err != nil {
			slog.Error("embedding_clear_default_failed", slog.Any("error", err))
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
		slog.Error("embedding_provider_create_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create embedding provider",
		})
	}

	slog.Info("embedding_provider_created", slog.String("id", p.ID), slog.String("name", p.Name))
	return c.Status(201).JSON(p.ToResponse())
}

// Update handles PUT /api/embedding/providers/:id
func (h *EmbeddingProviderHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.EmbeddingProviderUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// If setting this provider as default, clear existing default first.
	if req.IsDefault != nil && *req.IsDefault {
		if err := h.repo.ClearDefault(c.Context()); err != nil {
			slog.Error("embedding_clear_default_failed", slog.Any("error", err))
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
				Message: "Embedding provider not found",
			})
		}
		if errors.Is(err, repository.ErrDuplicate) {
			return c.Status(409).JSON(model.ErrorResponse{
				Error:   "provider_exists",
				Message: "A provider with this name already exists",
			})
		}
		slog.Error("embedding_provider_update_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update embedding provider",
		})
	}

	slog.Info("embedding_provider_updated", slog.String("id", p.ID))
	return c.JSON(p.ToResponse())
}

// Delete handles DELETE /api/embedding/providers/:id
func (h *EmbeddingProviderHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.repo.Delete(c.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "provider_not_found",
				Message: "Embedding provider not found",
			})
		}
		slog.Error("embedding_provider_delete_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete embedding provider",
		})
	}

	slog.Info("embedding_provider_deleted", slog.String("id", id))
	return c.SendStatus(204)
}

// GetDefault handles GET /api/embedding/providers/default
func (h *EmbeddingProviderHandler) GetDefault(c *fiber.Ctx) error {
	provider, err := h.repo.GetDefault(c.Context())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "no_default_provider",
				Message: "No default embedding provider configured",
			})
		}
		slog.Error("embedding_default_get_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get default embedding provider",
		})
	}

	return c.JSON(provider.ToResponse())
}

// embeddingTestRequest is the wire format for the OpenAI embeddings endpoint.
type embeddingTestRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// embeddingTestResponse is the wire format for the OpenAI embeddings response.
type embeddingTestResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Test handles POST /api/embedding/providers/:id/test
func (h *EmbeddingProviderHandler) Test(c *fiber.Ctx) error {
	id := c.Params("id")

	provider, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "provider_not_found",
				Message: "Embedding provider not found",
			})
		}
		slog.Error("embedding_provider_test_load_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to load embedding provider",
		})
	}

	if !provider.IsEnabled {
		return c.Status(400).JSON(model.EmbeddingTestResponse{
			OK:    false,
			Error: "provider is disabled",
		})
	}

	start := time.Now()
	dimensions, err := testEmbeddingConnectivity(c.Context(), provider)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		slog.Warn("embedding_test_failed", slog.String("id", id), slog.Any("error", err))
		return c.JSON(model.EmbeddingTestResponse{
			OK:        false,
			LatencyMs: latency,
			Error:     err.Error(),
		})
	}

	slog.Info("embedding_test_passed", slog.String("id", id), slog.Int64("latency_ms", latency))
	return c.JSON(model.EmbeddingTestResponse{
		OK:         true,
		Dimensions: dimensions,
		LatencyMs:  latency,
	})
}

// testEmbeddingConnectivity sends a test embedding request to verify the provider is reachable.
func testEmbeddingConnectivity(ctx context.Context, p *model.EmbeddingProvider) (int, error) {
	reqBody := embeddingTestRequest{
		Model: p.Model,
		Input: []string{"Hello, world!"},
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}

	url := p.BaseURL
	if url != "" && url[len(url)-1] != '/' {
		url += "/"
	}
	url += "embeddings"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("http request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("embedding provider returned %d: %s", httpResp.StatusCode, string(respBody))
	}

	var wire embeddingTestResponse
	if err := json.Unmarshal(respBody, &wire); err != nil {
		return 0, fmt.Errorf("unmarshal response: %w", err)
	}

	if wire.Error != nil {
		return 0, fmt.Errorf("embedding error: %s", wire.Error.Message)
	}

	if len(wire.Data) == 0 || len(wire.Data[0].Embedding) == 0 {
		return 0, fmt.Errorf("embedding provider returned no data")
	}

	return len(wire.Data[0].Embedding), nil
}
