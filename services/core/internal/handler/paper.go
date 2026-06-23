package handler

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"core/internal/model"
	"core/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// PaperHandler manages paper CRUD operations.
type PaperHandler struct {
	repo *repository.PaperRepository
}

// NewPaperHandler creates a new PaperHandler.
func NewPaperHandler(repo *repository.PaperRepository) *PaperHandler {
	return &PaperHandler{repo: repo}
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

	papers, total, err := h.repo.List(c.Context(), offset, limit)
	if err != nil {
		slog.Error("paper_list_failed", slog.Any("error", err))
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to list papers"))
	}

	return c.JSON(model.PapersListResponse{
		Papers: papers,
		Total:  total,
	})
}

// Get handles GET /api/papers/:id
func (h *PaperHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	paper, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrPaperNotFound) {
			return c.Status(404).JSON(errorResponse("paper_not_found", "Paper not found"))
		}
		slog.Error("paper_get_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to get paper"))
	}
	return c.JSON(paper)
}

// Create handles POST /api/papers
func (h *PaperHandler) Create(c *fiber.Ctx) error {
	var req model.PaperCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errorResponse("invalid_request", "Invalid request body"))
	}
	if req.Title == "" || req.RawText == "" {
		return c.Status(400).JSON(errorResponse("missing_fields", "Title and raw_text are required"))
	}

	now := time.Now()
	paper := &model.Paper{
		ID:        uuid.New().String(),
		Title:     req.Title,
		Authors:   req.Authors,
		SourceURL: req.SourceURL,
		RawText:   req.RawText,
		Status:    "uploaded",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.repo.Create(c.Context(), paper); err != nil {
		slog.Error("paper_create_failed", slog.Any("error", err))
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to create paper"))
	}

	slog.Info("paper_created", slog.String("id", paper.ID), slog.String("title", paper.Title))
	return c.Status(201).JSON(paper)
}

// Update handles PUT /api/papers/:id
func (h *PaperHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req model.PaperUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errorResponse("invalid_request", "Invalid request body"))
	}

	paper, err := h.repo.Update(c.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrPaperNotFound) {
			return c.Status(404).JSON(errorResponse("paper_not_found", "Paper not found"))
		}
		slog.Error("paper_update_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to update paper"))
	}

	slog.Info("paper_updated", slog.String("id", paper.ID))
	return c.JSON(paper)
}

// Delete handles DELETE /api/papers/:id
func (h *PaperHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.repo.Delete(c.Context(), id); err != nil {
		if errors.Is(err, repository.ErrPaperNotFound) {
			return c.Status(404).JSON(errorResponse("paper_not_found", "Paper not found"))
		}
		slog.Error("paper_delete_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(errorResponse("internal_error", "Failed to delete paper"))
	}
	slog.Info("paper_deleted", slog.String("id", id))
	return c.SendStatus(204)
}
