package handler

import (
	"errors"
	"log/slog"
	"strconv"

	"console/internal/coreclient"
	"console/internal/model"

	"github.com/gofiber/fiber/v2"
)

// EdgeHandler proxies edge operations through the core service. All
// mutations are persisted in Neo4j by core; the handler only performs
// input validation and translates HTTP error codes from the core response.
type EdgeHandler struct {
	core *coreclient.Client
}

// NewEdgeHandler creates a new EdgeHandler backed by the given core client.
func NewEdgeHandler(core *coreclient.Client) *EdgeHandler {
	return &EdgeHandler{core: core}
}

// List handles GET /api/edges
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

	resp, err := h.core.ListEdges(c.Context(), offset, limit, search)
	if err != nil {
		slog.Error("edges_list_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch edges from core service",
		})
	}
	return c.JSON(resp)
}

// Get handles GET /api/edges/:source/:target
func (h *EdgeHandler) Get(c *fiber.Ctx) error {
	source := c.Params("source")
	target := c.Params("target")
	if source == "" || target == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Source and target are required",
		})
	}

	edge, err := h.core.GetEdge(c.Context(), source, target)
	if err != nil {
		slog.Error("edges_get_core_failed",
			slog.String("source", source),
			slog.String("target", target),
			slog.Any("error", err),
		)
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch edge from core service",
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

// Create handles POST /api/edges
func (h *EdgeHandler) Create(c *fiber.Ctx) error {
	var req model.EdgeCreateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("edge_create_parse_error", slog.Any("error", err))
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.Source == "" || req.Target == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "missing_fields",
			Message: "Source and target are required",
		})
	}
	if req.Source == req.Target {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Source and target must differ",
		})
	}

	edge, err := h.core.CreateEdge(c.Context(), req)
	if err != nil {
		var httpErr *coreclient.HTTPStatusError
		if errors.As(err, &httpErr) {
			switch httpErr.StatusCode {
			case 409:
				return c.Status(409).JSON(model.ErrorResponse{
					Error:   "edge_exists",
					Message: "An edge with this source and target already exists",
				})
			case 400:
				return c.Status(400).JSON(model.ErrorResponse{
					Error:   "invalid_request",
					Message: "Invalid edge create request",
				})
			}
		}
		slog.Error("edge_create_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to create edge in core service",
		})
	}

	slog.Info("edge_created",
		slog.String("source", edge.Source),
		slog.String("target", edge.Target),
	)
	return c.Status(201).JSON(edge)
}

// Update handles PUT /api/edges/:source/:target
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

	edge, err := h.core.UpdateEdge(c.Context(), source, target, req)
	if err != nil {
		var httpErr *coreclient.HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 400 {
			return c.Status(400).JSON(model.ErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid edge update request",
			})
		}
		slog.Error("edge_update_core_failed",
			slog.String("source", source),
			slog.String("target", target),
			slog.Any("error", err),
		)
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to update edge in core service",
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

// Delete handles DELETE /api/edges/:source/:target
func (h *EdgeHandler) Delete(c *fiber.Ctx) error {
	source := c.Params("source")
	target := c.Params("target")
	if source == "" || target == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Source and target are required",
		})
	}

	ok, err := h.core.DeleteEdge(c.Context(), source, target)
	if err != nil {
		slog.Error("edge_delete_core_failed",
			slog.String("source", source),
			slog.String("target", target),
			slog.Any("error", err),
		)
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to delete edge in core service",
		})
	}
	if !ok {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "edge_not_found",
			Message: "Edge with the specified source and target not found",
		})
	}

	slog.Info("edge_deleted",
		slog.String("source", source),
		slog.String("target", target),
	)
	return c.SendStatus(204)
}
