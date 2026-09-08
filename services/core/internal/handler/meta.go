package handler

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

// MetaReader abstracts graph metadata reads for testing.
type MetaReader interface {
	LastUpdated(ctx context.Context) (time.Time, bool, error)
}

// GraphMetaHandler serves graph-level metadata.
type GraphMetaHandler struct {
	meta MetaReader
}

// NewGraphMetaHandler creates a new GraphMetaHandler.
func NewGraphMetaHandler(meta MetaReader) *GraphMetaHandler {
	return &GraphMetaHandler{meta: meta}
}

// Get handles GET /api/graph/meta. It returns the timestamp of the last
// graph mutation (extraction merge or community re-detection), or null
// when the graph has never been mutated.
func (h *GraphMetaHandler) Get(c *fiber.Ctx) error {
	ts, ok, err := h.meta.LastUpdated(c.Context())
	if err != nil {
		slog.Error("graph_meta_read_failed", slog.Any("error", err))
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to read graph metadata"))
	}
	if !ok {
		return c.JSON(fiber.Map{"last_updated": nil})
	}
	return c.JSON(fiber.Map{"last_updated": ts.Format(time.RFC3339)})
}
