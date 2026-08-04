package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"core/internal/model"
	"core/internal/repository"
	"core/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ReviewQueueHandler manages the review queue.
type ReviewQueueHandler struct {
	repo      *repository.ReviewQueueRepository
	paperRepo *repository.PaperRepository
	merge     *service.MergeService
}

// NewReviewQueueHandler creates a new ReviewQueueHandler. The merge
// service is used by Approve to persist the extracted nodes/edges into
// Neo4j before the review item is marked as approved.
func NewReviewQueueHandler(repo *repository.ReviewQueueRepository, paperRepo *repository.PaperRepository, merge *service.MergeService) *ReviewQueueHandler {
	return &ReviewQueueHandler{repo: repo, paperRepo: paperRepo, merge: merge}
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

	items, total, err := h.repo.List(c.Context(), offset, limit, status)
	if err != nil {
		slog.Error("review_list_failed", slog.Any("error", err))
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to list review queue"))
	}

	return c.JSON(model.ReviewQueueListResponse{
		Items: items,
		Total: total,
	})
}

// Get handles GET /api/review-queue/:id
func (h *ReviewQueueHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	item, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrReviewItemNotFound) {
			return c.Status(404).JSON(errorResponse("review_item_not_found", "Review item not found"))
		}
		slog.Error("review_get_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to get review item"))
	}
	return c.JSON(item)
}

// Submit handles POST /api/review-queue. Called by discovery service.
func (h *ReviewQueueHandler) Submit(c *fiber.Ctx) error {
	var req model.ReviewQueueSubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errorResponse("invalid_request", "Invalid request body"))
	}
	if req.PaperID == "" {
		return c.Status(400).JSON(errorResponse("missing_fields", "paper_id is required"))
	}

	// Validate extracted data is valid JSON.
	if !json.Valid(req.ExtractedNodes) || !json.Valid(req.ExtractedEdges) {
		return c.Status(400).JSON(errorResponse("invalid_json", "extracted_nodes and extracted_edges must be valid JSON"))
	}

	// Verify paper exists.
	_, err := h.paperRepo.GetByID(c.Context(), req.PaperID)
	if err != nil {
		if errors.Is(err, repository.ErrPaperNotFound) {
			return c.Status(404).JSON(errorResponse("paper_not_found", "Paper not found"))
		}
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to verify paper"))
	}

	now := time.Now()
	item := &model.ReviewQueueItem{
		ID:             uuid.New().String(),
		PaperID:        req.PaperID,
		ExtractedNodes: req.ExtractedNodes,
		ExtractedEdges: req.ExtractedEdges,
		Status:         "pending",
		CreatedAt:      now,
	}

	if err := h.repo.Create(c.Context(), item); err != nil {
		slog.Error("review_create_failed", slog.Any("error", err))
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to create review item"))
	}

	slog.Info("review_item_created",
		slog.String("id", item.ID),
		slog.String("paper_id", item.PaperID),
	)
	return c.Status(201).JSON(item)
}

// Approve handles POST /api/review-queue/:id/approve
//
// The extracted nodes/edges are merged into the Neo4j graph BEFORE the
// review item's status is flipped to "approved". On merge failure the
// item stays pending so the reviewer can retry; the merge runs in a
// single transaction and is idempotent, so retrying after a partial
// failure is safe.
func (h *ReviewQueueHandler) Approve(c *fiber.Ctx) error {
	id := c.Params("id")
	var req model.ReviewQueueApproveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errorResponse("invalid_request", "Invalid request body"))
	}

	item, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrReviewItemNotFound) {
			return c.Status(404).JSON(errorResponse("review_item_not_found", "Review item not found"))
		}
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to get review item"))
	}

	if item.Status != "pending" {
		return c.Status(400).JSON(errorResponse("invalid_status", "Only pending items can be approved"))
	}

	// Unmarshal the extracted payload stored in MySQL.
	var nodes []model.ExtractedNode
	var edges []model.ExtractedEdge
	if err := json.Unmarshal(item.ExtractedNodes, &nodes); err != nil {
		slog.Error("review_approve_unmarshal_nodes_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(400).JSON(errorResponse("invalid_extracted_data", "extracted_nodes is not valid JSON"))
	}
	if err := json.Unmarshal(item.ExtractedEdges, &edges); err != nil {
		slog.Error("review_approve_unmarshal_edges_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(400).JSON(errorResponse("invalid_extracted_data", "extracted_edges is not valid JSON"))
	}

	// Merge into Neo4j. Failure leaves the item pending for retry.
	mergeResult, err := h.merge.Merge(c.Context(), item.PaperID, nodes, edges)
	if err != nil {
		slog.Error("review_approve_merge_failed",
			slog.String("id", id),
			slog.String("paper_id", item.PaperID),
			slog.Any("error", err))
		return c.Status(500).JSON(errorResponse("merge_failed", "Failed to merge extracted data into the graph"))
	}

	if err := h.repo.UpdateStatus(c.Context(), id, "approved", req.ReviewerID, req.ReviewNotes); err != nil {
		slog.Error("review_approve_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to approve review item"))
	}

	// Update paper status to "merged".
	paperStatus := "merged"
	h.paperRepo.Update(c.Context(), item.PaperID, &model.PaperUpdateRequest{Status: &paperStatus})

	slog.Info("review_item_approved",
		slog.String("id", id),
		slog.String("reviewer", req.ReviewerID),
		slog.Int("created_nodes", mergeResult.CreatedNodes),
		slog.Int("reused_nodes", mergeResult.ReusedNodes),
		slog.Int("created_edges", mergeResult.CreatedEdges),
		slog.Int("skipped_edges", mergeResult.SkippedEdges))

	// Return updated item.
	updated, _ := h.repo.GetByID(c.Context(), id)
	return c.JSON(updated)
}

// Reject handles POST /api/review-queue/:id/reject
func (h *ReviewQueueHandler) Reject(c *fiber.Ctx) error {
	id := c.Params("id")
	var req model.ReviewQueueRejectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errorResponse("invalid_request", "Invalid request body"))
	}

	item, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrReviewItemNotFound) {
			return c.Status(404).JSON(errorResponse("review_item_not_found", "Review item not found"))
		}
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to get review item"))
	}

	if item.Status != "pending" {
		return c.Status(400).JSON(errorResponse("invalid_status", "Only pending items can be rejected"))
	}

	if err := h.repo.UpdateStatus(c.Context(), id, "rejected", req.ReviewerID, req.ReviewNotes); err != nil {
		slog.Error("review_reject_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to reject review item"))
	}

	slog.Info("review_item_rejected", slog.String("id", id), slog.String("reviewer", req.ReviewerID))
	return c.SendStatus(204)
}
