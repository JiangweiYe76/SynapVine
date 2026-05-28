package handler

import (
	"log/slog"
	"strconv"
	"strings"

	"ai-graph-server/internal/model"
	"ai-graph-server/internal/service"

	"github.com/gofiber/fiber/v2"
)

// GraphHandler handles HTTP requests for graph-related endpoints
type GraphHandler struct {
	svc *service.GraphService
}

// NewGraphHandler creates a new GraphHandler with the given service
func NewGraphHandler(svc *service.GraphService) *GraphHandler {
	return &GraphHandler{svc: svc}
}

// Summary handles GET /api/graph/summary
// Returns a summary of the graph including communities, stats, and top nodes
func (h *GraphHandler) Summary(c *fiber.Ctx) error {
	return c.JSON(h.svc.Summary(20))
}

// Nodes handles GET /api/graph/nodes
// Returns a paginated list of nodes with optional filtering
func (h *GraphHandler) Nodes(c *fiber.Ctx) error {
	// Parse query parameters
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	sortBy := c.Query("sort", "influence")
	communityFilter := c.Query("community_id", "")
	idsStr := c.Query("ids", "")

	// Parse comma-separated node IDs if provided
	var ids []string
	if idsStr != "" {
		ids = strings.Split(idsStr, ",")
		for i := range ids {
			ids[i] = strings.TrimSpace(ids[i])
		}
	}

	// Validate and cap the limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	return c.JSON(h.svc.Nodes(offset, limit, sortBy, communityFilter, ids))
}

// NodeDetail handles GET /api/graph/nodes/:id
// Returns detailed information about a specific node
func (h *GraphHandler) NodeDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	detail, ok := h.svc.NodeDetail(id)
	if !ok {
		slog.Warn("node_not_found", slog.String("node_id", id), slog.String("ip", c.IP()))
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "node_not_found",
			Message: "Node " + id + " does not exist",
		})
	}
	return c.JSON(detail)
}

// NodeEdges handles GET /api/graph/nodes/:id/edges
// Returns edges connected to a specific node
func (h *GraphHandler) NodeEdges(c *fiber.Ctx) error {
	id := c.Params("id")
	direction := c.Query("direction", "both")

	// Validate direction parameter
	if direction != "in" && direction != "out" && direction != "both" {
		direction = "both"
	}

	result, ok := h.svc.NodeEdges(id, direction)
	if !ok {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "node_not_found",
			Message: "Node " + id + " does not exist",
		})
	}
	return c.JSON(result)
}

// Search handles GET /api/graph/search
// Searches for nodes by name or description
func (h *GraphHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q", "")
	if query == "" {
		slog.Warn("search_missing_query", slog.String("ip", c.IP()))
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "missing_query",
			Message: "Please provide a search query via the 'q' parameter",
		})
	}
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	result := h.svc.Search(query, limit)
	slog.Info("search_executed",
		slog.String("query", query),
		slog.Int("results", len(result.Results)),
		slog.String("ip", c.IP()),
	)
	return c.JSON(result)
}

// Expand handles GET /api/graph/expand
// Expands a set of nodes to include their neighbors and connecting edges
func (h *GraphHandler) Expand(c *fiber.Ctx) error {
	idsStr := c.Query("ids", "")
	if idsStr == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "missing_ids",
			Message: "Please provide node IDs via the 'ids' parameter",
		})
	}

	// Parse comma-separated node IDs
	ids := strings.Split(idsStr, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	// Parse boolean flags
	includeEdges := c.Query("include_edges", "true") == "true"
	includeNeighbors := c.Query("include_neighbors", "false") == "true"

	return c.JSON(h.svc.Expand(ids, includeEdges, includeNeighbors))
}
