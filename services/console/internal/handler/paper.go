package handler

import (
	"log/slog"
	"strconv"

	"console/internal/coreclient"
	"console/internal/model"

	"github.com/gofiber/fiber/v2"
)

// PaperHandler proxies paper-related HTTP requests to the core service.
type PaperHandler struct {
	core *coreclient.Client
}

// NewPaperHandler creates a new PaperHandler.
func NewPaperHandler(core *coreclient.Client) *PaperHandler {
	return &PaperHandler{core: core}
}

// List handles GET /api/papers
func (h *PaperHandler) List(c *fiber.Ctx) error {
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	resp, err := h.core.ListPapers(c.Context(), offset, limit)
	if err != nil {
		slog.Error("paper_list_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch papers from core service",
		})
	}
	return c.JSON(resp)
}

// Get handles GET /api/papers/:id
func (h *PaperHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	paper, err := h.core.GetPaper(c.Context(), id)
	if err != nil {
		slog.Error("paper_get_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch paper from core service",
		})
	}
	if paper == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "paper_not_found",
			Message: "Paper not found",
		})
	}
	return c.JSON(paper)
}

// Create handles POST /api/papers
func (h *PaperHandler) Create(c *fiber.Ctx) error {
	var req model.PaperCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}
	if req.Title == "" || req.RawText == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "missing_fields",
			Message: "Title and raw_text are required",
		})
	}

	paper, err := h.core.CreatePaper(c.Context(), req)
	if err != nil {
		slog.Error("paper_create_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to create paper in core service",
		})
	}

	slog.Info("paper_created", slog.String("id", paper.ID), slog.String("title", paper.Title))
	return c.Status(201).JSON(paper)
}

// Update handles PUT /api/papers/:id
func (h *PaperHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req model.PaperUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	paper, err := h.core.UpdatePaper(c.Context(), id, req)
	if err != nil {
		slog.Error("paper_update_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to update paper in core service",
		})
	}
	if paper == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "paper_not_found",
			Message: "Paper not found",
		})
	}

	slog.Info("paper_updated", slog.String("id", paper.ID))
	return c.JSON(paper)
}

// Delete handles DELETE /api/papers/:id
func (h *PaperHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	ok, err := h.core.DeletePaper(c.Context(), id)
	if err != nil {
		slog.Error("paper_delete_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to delete paper in core service",
		})
	}
	if !ok {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "paper_not_found",
			Message: "Paper not found",
		})
	}

	slog.Info("paper_deleted", slog.String("id", id))
	return c.SendStatus(204)
}

// Stats handles GET /api/papers/stats — returns paper counts by status.
func (h *PaperHandler) Stats(c *fiber.Ctx) error {
	// Fetch all papers (small dataset) and aggregate.
	resp, err := h.core.ListPapers(c.Context(), 0, 10000)
	if err != nil {
		slog.Error("paper_stats_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch papers for stats",
		})
	}

	stats := map[string]int{
		"total":     resp.Total,
		"uploaded":  0,
		"analyzing": 0,
		"analyzed":  0,
		"merged":    0,
	}
	for _, p := range resp.Papers {
		stats[p.Status]++
	}

	return c.JSON(stats)
}
