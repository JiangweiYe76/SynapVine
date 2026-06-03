package handler

import (
	"log/slog"
	"strconv"

	"console/internal/loader"
	"console/internal/model"

	"github.com/gofiber/fiber/v2"
)

// NodeHandler handles node-related HTTP requests
type NodeHandler struct {
	store *loader.GraphStore
}

// NewNodeHandler creates a new NodeHandler
func NewNodeHandler(store *loader.GraphStore) *NodeHandler {
	return &NodeHandler{store: store}
}

// List handles GET /api/nodes
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

	var nodes []model.Node
	var total int
	if search != "" {
		nodes, total = h.store.SearchNodes(search, offset, limit)
	} else {
		nodes, total = h.store.ListNodes(offset, limit)
	}

	return c.JSON(model.NodesListResponse{
		Nodes: nodes,
		Pagination: model.Pagination{
			Offset:  offset,
			Limit:   limit,
			Total:   total,
			HasMore: offset+limit < total,
		},
	})
}

// Get handles GET /api/nodes/:id
func (h *NodeHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	node := h.store.GetNode(id)
	if node == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "node_not_found",
			Message: "Node with the specified ID not found",
		})
	}

	return c.JSON(node)
}

// Create handles POST /api/nodes
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

	if h.store.NodeExists(req.ID) {
		return c.Status(409).JSON(model.ErrorResponse{
			Error:   "node_exists",
			Message: "A node with this ID already exists",
		})
	}

	node := model.Node{
		ID:             req.ID,
		Name:           req.Name,
		Category:       req.Category,
		Description:    req.Description,
		InfluenceScore: req.InfluenceScore,
		FirstAppeared:  req.FirstAppeared,
		Milestones:     req.Milestones,
	}

	if err := h.store.CreateNode(node); err != nil {
		slog.Error("node_create_save_error", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "save_failed",
			Message: "Failed to save node",
		})
	}

	slog.Info("node_created", slog.String("id", node.ID), slog.String("name", node.Name))

	return c.Status(201).JSON(node)
}

// Update handles PUT /api/nodes/:id
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

	node, err := h.store.UpdateNode(id, req)
	if err != nil {
		slog.Error("node_update_save_error", slog.Any("error", err))
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

// Delete handles DELETE /api/nodes/:id
func (h *NodeHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if !h.store.DeleteNode(id) {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "node_not_found",
			Message: "Node with the specified ID not found",
		})
	}

	slog.Info("node_deleted", slog.String("id", id))

	return c.SendStatus(204)
}

// Stats handles GET /api/stats
func (h *NodeHandler) Stats(c *fiber.Ctx) error {
	stats := h.store.Stats()
	return c.JSON(stats)
}
