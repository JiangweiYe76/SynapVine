package handler

import (
	"log/slog"
	"strconv"

	"console/internal/loader"
	"console/internal/model"

	"github.com/gofiber/fiber/v2"
)

// EdgeHandler handles edge-related HTTP requests
type EdgeHandler struct {
	store *loader.GraphStore
}

// NewEdgeHandler creates a new EdgeHandler
func NewEdgeHandler(store *loader.GraphStore) *EdgeHandler {
	return &EdgeHandler{store: store}
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

	var edges []model.Edge
	var total int
	if search != "" {
		edges, total = h.store.SearchEdges(search, offset, limit)
	} else {
		edges, total = h.store.ListEdges(offset, limit)
	}

	return c.JSON(model.EdgesListResponse{
		Edges: edges,
		Pagination: model.Pagination{
			Offset:  offset,
			Limit:   limit,
			Total:   total,
			HasMore: offset+limit < total,
		},
	})
}

// Get handles GET /api/edges/:source/:target
func (h *EdgeHandler) Get(c *fiber.Ctx) error {
	source := c.Params("source")
	target := c.Params("target")

	edge := h.store.GetEdge(source, target)
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

	if !h.store.NodeExists(req.Source) {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "source_not_found",
			Message: "Source node does not exist",
		})
	}

	if !h.store.NodeExists(req.Target) {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "target_not_found",
			Message: "Target node does not exist",
		})
	}

	if h.store.EdgeExists(req.Source, req.Target) {
		return c.Status(409).JSON(model.ErrorResponse{
			Error:   "edge_exists",
			Message: "An edge between these nodes already exists",
		})
	}

	edge := model.Edge{
		Source:   req.Source,
		Target:   req.Target,
		Weight:   req.Weight,
		Relation: req.Relation,
	}

	if err := h.store.CreateEdge(edge); err != nil {
		slog.Error("edge_create_save_error", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "save_failed",
			Message: "Failed to save edge",
		})
	}

	slog.Info("edge_created", slog.String("source", edge.Source), slog.String("target", edge.Target))

	return c.Status(201).JSON(edge)
}

// Update handles PUT /api/edges/:source/:target
func (h *EdgeHandler) Update(c *fiber.Ctx) error {
	source := c.Params("source")
	target := c.Params("target")

	var req model.EdgeUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("edge_update_parse_error", slog.Any("error", err))
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	edge, err := h.store.UpdateEdge(source, target, req)
	if err != nil {
		slog.Error("edge_update_save_error", slog.Any("error", err))
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

	slog.Info("edge_updated", slog.String("source", edge.Source), slog.String("target", edge.Target))

	return c.JSON(edge)
}

// Delete handles DELETE /api/edges/:source/:target
func (h *EdgeHandler) Delete(c *fiber.Ctx) error {
	source := c.Params("source")
	target := c.Params("target")

	if !h.store.DeleteEdge(source, target) {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "edge_not_found",
			Message: "Edge with the specified source and target not found",
		})
	}

	slog.Info("edge_deleted", slog.String("source", source), slog.String("target", target))

	return c.SendStatus(204)
}
