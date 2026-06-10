package handler

import (
	"log/slog"
	"strconv"
	"strings"

	"console/internal/coreclient"
	"console/internal/model"

	"github.com/gofiber/fiber/v2"
)

// EdgeHandler proxies edge reads through the core service. Edge writes are
// not yet supported because the core service does not expose edge CRUD
// endpoints; write requests return 501 Not Implemented.
type EdgeHandler struct {
	core *coreclient.Client
}

// NewEdgeHandler creates a new EdgeHandler backed by the given core client.
func NewEdgeHandler(core *coreclient.Client) *EdgeHandler {
	return &EdgeHandler{core: core}
}

// List handles GET /api/edges
func (h *EdgeHandler) List(c *fiber.Ctx) error {
	search := c.Query("search", "")
	graph, err := h.core.GraphData(c.Context())
	if err != nil {
		slog.Error("edges_list_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch edges from core service",
		})
	}

	edges := graph.Edges
	if search != "" {
		edges = filterEdges(graph.Edges, search)
	}

	offset, limit, err := parsePagination(c)
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	total := len(edges)
	if offset >= total {
		return c.JSON(model.EdgesListResponse{
			Edges: []model.Edge{},
			Pagination: model.Pagination{
				Offset:  offset,
				Limit:   limit,
				Total:   total,
				HasMore: false,
			},
		})
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return c.JSON(model.EdgesListResponse{
		Edges: edges[offset:end],
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

	graph, err := h.core.GraphData(c.Context())
	if err != nil {
		slog.Error("edges_get_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch edges from core service",
		})
	}

	for _, e := range graph.Edges {
		if e.Source == source && e.Target == target {
			return c.JSON(e)
		}
	}
	return c.Status(404).JSON(model.ErrorResponse{
		Error:   "edge_not_found",
		Message: "Edge with the specified source and target not found",
	})
}

// Create handles POST /api/edges. Edge writes are not yet supported.
func (h *EdgeHandler) Create(c *fiber.Ctx) error {
	return notImplemented(c, "edge_create")
}

// Update handles PUT /api/edges/:source/:target. Edge writes are not yet supported.
func (h *EdgeHandler) Update(c *fiber.Ctx) error {
	return notImplemented(c, "edge_update")
}

// Delete handles DELETE /api/edges/:source/:target. Edge writes are not yet supported.
func (h *EdgeHandler) Delete(c *fiber.Ctx) error {
	return notImplemented(c, "edge_delete")
}

// notImplemented responds with 501 and logs the event. The reason is that
// the core service does not yet expose edge CRUD endpoints; once it does,
// the corresponding handler will be implemented.
func notImplemented(c *fiber.Ctx, op string) error {
	slog.Warn("edge_operation_not_implemented", slog.String("op", op))
	return c.Status(501).JSON(model.ErrorResponse{
		Error:   "not_implemented",
		Message: "Edge mutations are not yet supported; the core service needs edge CRUD endpoints first",
	})
}

// filterEdges returns the subset of edges whose source, target, or relation
// contains the search query (case-insensitive).
func filterEdges(edges []model.Edge, query string) []model.Edge {
	q := strings.ToLower(query)
	var matched []model.Edge
	for _, e := range edges {
		if strings.Contains(strings.ToLower(e.Source), q) ||
			strings.Contains(strings.ToLower(e.Target), q) ||
			strings.Contains(strings.ToLower(e.Relation), q) {
			matched = append(matched, e)
		}
	}
	return matched
}

// parsePagination extracts and validates the offset/limit query parameters.
func parsePagination(c *fiber.Ctx) (int, int, error) {
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return offset, limit, nil
}
