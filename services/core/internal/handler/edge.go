package handler

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"core/internal/model"
	"core/internal/service"

	"github.com/gofiber/fiber/v2"
)

// Compile-time check that *service.EdgeService satisfies EdgeService.
var _ EdgeService = (*service.EdgeService)(nil)

// EdgeService is the subset of *service.EdgeService used by EdgeHandler.
// Declared as an interface to enable stubbing in unit tests.
type EdgeService interface {
	List(ctx context.Context, offset, limit int, search string) (*model.EdgesListResponse, error)
	Get(ctx context.Context, source, target string) (*model.Edge, error)
	Create(ctx context.Context, req model.EdgeCreateRequest) (*model.Edge, error)
	Update(ctx context.Context, source, target string, req model.EdgeUpdateRequest) (*model.Edge, error)
	Delete(ctx context.Context, source, target string) error
}

// EdgeHandler handles HTTP requests for RELATES_TO edge operations.
type EdgeHandler struct {
	svc EdgeService
}

// NewEdgeHandler creates a new EdgeHandler.
func NewEdgeHandler(svc EdgeService) *EdgeHandler {
	return &EdgeHandler{svc: svc}
}

// List handles GET /api/edges.
func (h *EdgeHandler) List(c *fiber.Ctx) error {
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	resp, err := h.svc.List(c.Context(), offset, limit, search)
	if err != nil {
		slog.Error("edge_list_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list edges",
		})
	}
	return c.JSON(resp)
}

// Get handles GET /api/edges/:source/:target.
func (h *EdgeHandler) Get(c *fiber.Ctx) error {
	source := c.Params("source")
	target := c.Params("target")
	if source == "" || target == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Source and target are required",
		})
	}

	edge, err := h.svc.Get(c.Context(), source, target)
	if err != nil {
		slog.Error("edge_get_failed",
			slog.String("source", source),
			slog.String("target", target),
			slog.Any("error", err),
		)
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get edge",
		})
	}
	if edge == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "edge_not_found",
			Message: "Edge with the specified source and target not found",
		})
	}
	return c.JSON(edge)
}

// Create handles POST /api/edges.
func (h *EdgeHandler) Create(c *fiber.Ctx) error {
	var req model.EdgeCreateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("edge_create_parse_error", slog.Any("error", err))
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	edge, err := h.svc.Create(c.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNodeNotFound):
			return c.Status(400).JSON(model.ErrorResponse{
				Error:   "node_not_found",
				Message: "Source or target node does not exist",
			})
		case errors.Is(err, service.ErrEdgeExists):
			return c.Status(409).JSON(model.ErrorResponse{
				Error:   "edge_exists",
				Message: "An edge with this source and target already exists",
			})
		case errors.Is(err, service.ErrSameEndpoints):
			return c.Status(400).JSON(model.ErrorResponse{
				Error:   "invalid_request",
				Message: "Source and target must differ",
			})
		case errors.Is(err, service.ErrInvalidWeight):
			return c.Status(400).JSON(model.ErrorResponse{
				Error:   "invalid_weight",
				Message: "Weight must be in [0, 1]",
			})
		}
		slog.Error("edge_create_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "save_failed",
			Message: "Failed to create edge",
		})
	}

	slog.Info("edge_created",
		slog.String("source", edge.Source),
		slog.String("target", edge.Target),
		slog.String("relation", edge.Relation),
	)
	return c.Status(201).JSON(edge)
}

// Update handles PUT /api/edges/:source/:target.
func (h *EdgeHandler) Update(c *fiber.Ctx) error {
	source := c.Params("source")
	target := c.Params("target")
	if source == "" || target == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Source and target are required",
		})
	}

	var req model.EdgeUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("edge_update_parse_error", slog.Any("error", err))
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	edge, err := h.svc.Update(c.Context(), source, target, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidWeight) {
			return c.Status(400).JSON(model.ErrorResponse{
				Error:   "invalid_weight",
				Message: "Weight must be in [0, 1]",
			})
		}
		slog.Error("edge_update_failed",
			slog.String("source", source),
			slog.String("target", target),
			slog.Any("error", err),
		)
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "save_failed",
			Message: "Failed to update edge",
		})
	}
	if edge == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "edge_not_found",
			Message: "Edge with the specified source and target not found",
		})
	}

	slog.Info("edge_updated",
		slog.String("source", source),
		slog.String("target", target),
	)
	return c.JSON(edge)
}

// Delete handles DELETE /api/edges/:source/:target.
func (h *EdgeHandler) Delete(c *fiber.Ctx) error {
	source := c.Params("source")
	target := c.Params("target")
	if source == "" || target == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Source and target are required",
		})
	}

	if err := h.svc.Delete(c.Context(), source, target); err != nil {
		if errors.Is(err, service.ErrEdgeNotFound) {
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "edge_not_found",
				Message: "Edge with the specified source and target not found",
			})
		}
		slog.Error("edge_delete_failed",
			slog.String("source", source),
			slog.String("target", target),
			slog.Any("error", err),
		)
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "delete_failed",
			Message: "Failed to delete edge",
		})
	}

	slog.Info("edge_deleted",
		slog.String("source", source),
		slog.String("target", target),
	)
	return c.SendStatus(204)
}
