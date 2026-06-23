package handler

import (
	"log/slog"
	"strconv"

	"console/internal/coreclient"
	"console/internal/model"

	"github.com/gofiber/fiber/v2"
)

// ReviewQueueHandler proxies review queue requests to the core service.
type ReviewQueueHandler struct {
	core *coreclient.Client
}

// NewReviewQueueHandler creates a new ReviewQueueHandler.
func NewReviewQueueHandler(core *coreclient.Client) *ReviewQueueHandler {
	return &ReviewQueueHandler{core: core}
}

// List handles GET /api/review-queue
func (h *ReviewQueueHandler) List(c *fiber.Ctx) error {
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	status := c.Query("status", "")
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	resp, err := h.core.ListReviewItems(c.Context(), offset, limit, status)
	if err != nil {
		slog.Error("review_list_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch review queue from core service",
		})
	}
	return c.JSON(resp)
}

// Get handles GET /api/review-queue/:id
func (h *ReviewQueueHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	item, err := h.core.GetReviewItem(c.Context(), id)
	if err != nil {
		slog.Error("review_get_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch review item from core service",
		})
	}
	if item == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "review_item_not_found",
			Message: "Review item not found",
		})
	}
	return c.JSON(item)
}

// Approve handles POST /api/review-queue/:id/approve
func (h *ReviewQueueHandler) Approve(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		ReviewerID  string `json:"reviewer_id"`
		ReviewNotes string `json:"review_notes"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	item, err := h.core.ApproveReviewItem(c.Context(), id, req.ReviewerID, req.ReviewNotes)
	if err != nil {
		slog.Error("review_approve_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to approve review item",
		})
	}

	slog.Info("review_item_approved", slog.String("id", id))
	return c.JSON(item)
}

// Reject handles POST /api/review-queue/:id/reject
func (h *ReviewQueueHandler) Reject(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		ReviewerID  string `json:"reviewer_id"`
		ReviewNotes string `json:"review_notes"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := h.core.RejectReviewItem(c.Context(), id, req.ReviewerID, req.ReviewNotes); err != nil {
		slog.Error("review_reject_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to reject review item",
		})
	}

	slog.Info("review_item_rejected", slog.String("id", id))
	return c.SendStatus(204)
}
