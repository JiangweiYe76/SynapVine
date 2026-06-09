package handler

import (
	"log/slog"
	"strconv"

	"core/internal/model"
	"core/internal/service"

	"github.com/gofiber/fiber/v2"
)

// NodeHandler handles HTTP requests for node operations.
type NodeHandler struct {
	svc *service.NodeService
}

// NewNodeHandler creates a new NodeHandler.
func NewNodeHandler(svc *service.NodeService) *NodeHandler {
	return &NodeHandler{svc: svc}
}

// List handles GET /api/nodes.
func (h *NodeHandler) List(c *fiber.Ctx) error {
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
		slog.Error("node_list_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list nodes",
		})
	}
	return c.JSON(resp)
}

// Get handles GET /api/nodes/:id.
func (h *NodeHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	node, err := h.svc.Get(c.Context(), id)
	if err != nil {
		slog.Error("node_get_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get node",
		})
	}
	if node == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "node_not_found",
			Message: "Node with the specified ID not found",
		})
	}
	return c.JSON(node)
}

// Create handles POST /api/nodes.
func (h *NodeHandler) Create(c *fiber.Ctx) error {
	var req model.NodeCreateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("node_create_parse_error", slog.Any("error", err))
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.ID == "" || req.Name == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "missing_fields",
			Message: "ID and name are required",
		})
	}

	node, err := h.svc.Create(c.Context(), req)
	if err != nil {
		if err.Error() == "node already exists" {
			return c.Status(409).JSON(model.ErrorResponse{
				Error:   "node_exists",
				Message: "A node with this ID already exists",
			})
		}
		slog.Error("node_create_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "save_failed",
			Message: "Failed to create node",
		})
	}

	slog.Info("node_created", slog.String("id", node.ID), slog.String("name", node.Name))
	return c.Status(201).JSON(node)
}

// Update handles PUT /api/nodes/:id.
func (h *NodeHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.NodeUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("node_update_parse_error", slog.Any("error", err))
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	node, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		slog.Error("node_update_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "save_failed",
			Message: "Failed to update node",
		})
	}
	if node == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "node_not_found",
			Message: "Node with the specified ID not found",
		})
	}

	slog.Info("node_updated", slog.String("id", node.ID))
	return c.JSON(node)
}

// Delete handles DELETE /api/nodes/:id.
func (h *NodeHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.svc.Delete(c.Context(), id); err != nil {
		slog.Error("node_delete_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "delete_failed",
			Message: "Failed to delete node",
		})
	}

	slog.Info("node_deleted", slog.String("id", id))
	return c.SendStatus(204)
}

// GraphData handles GET /api/graph/data, returning all nodes and edges.
func (h *NodeHandler) GraphData(c *fiber.Ctx) error {
	nodes, edges, err := h.svc.GetAll(c.Context())
	if err != nil {
		slog.Error("graph_data_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to load graph data",
		})
	}
	return c.JSON(fiber.Map{
		"nodes": nodes,
		"edges": edges,
	})
}
