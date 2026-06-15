package handler

import (
	"errors"
	"log/slog"
	"strconv"

	"console/internal/coreclient"
	"console/internal/model"

	"github.com/gofiber/fiber/v2"
)

// NodeHandler proxies node-related HTTP requests to the core service.
type NodeHandler struct {
	core *coreclient.Client
}

// NewNodeHandler creates a new NodeHandler backed by the given core client.
func NewNodeHandler(core *coreclient.Client) *NodeHandler {
	return &NodeHandler{core: core}
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

	resp, err := h.core.ListNodes(c.Context(), offset, limit, search)
	if err != nil {
		slog.Error("node_list_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch nodes from core service",
		})
	}
	return c.JSON(resp)
}

// Get handles GET /api/nodes/:id
func (h *NodeHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	node, err := h.core.GetNode(c.Context(), id)
	if err != nil {
		slog.Error("node_get_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch node from core service",
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

	if req.Name == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "missing_fields",
			Message: "Name is required",
		})
	}

	node, err := h.core.CreateNode(c.Context(), req)
	if err != nil {
		var httpErr *coreclient.HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 409 {
			return c.Status(409).JSON(model.ErrorResponse{
				Error:   "node_exists",
				Message: "A node with this ID already exists",
			})
		}
		slog.Error("node_create_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to create node in core service",
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

	node, err := h.core.UpdateNode(c.Context(), id, req)
	if err != nil {
		slog.Error("node_update_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to update node in core service",
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

	ok, err := h.core.DeleteNode(c.Context(), id)
	if err != nil {
		slog.Error("node_delete_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to delete node in core service",
		})
	}
	if !ok {
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
	// Stats are derived from the authoritative graph snapshot in core,
	// which is the single source of truth for nodes and edges.
	graph, err := h.core.GraphData(c.Context())
	if err != nil {
		slog.Error("graph_data_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to load graph data from core service",
		})
	}
	return c.JSON(computeStats(graph))
}

// computeStats derives the dashboard stats response from a graph snapshot.
func computeStats(graph *model.GraphData) model.StatsResponse {
	var totalInfluence float64

	for _, n := range graph.Nodes {
		totalInfluence += n.InfluenceScore
	}

	var avgInfluence float64
	if len(graph.Nodes) > 0 {
		avgInfluence = totalInfluence / float64(len(graph.Nodes))
	}

	return model.StatsResponse{
		TotalNodes:   len(graph.Nodes),
		TotalEdges:   len(graph.Edges),
		AvgInfluence: avgInfluence,
	}
}
