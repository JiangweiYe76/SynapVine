package handler

import (
	"log/slog"
	"strconv"

	"core/internal/model"
	"core/internal/service"

	"github.com/gofiber/fiber/v2"
)

// CommunityHandler handles HTTP requests for community operations.
type CommunityHandler struct {
	svc *service.CommunityService
}

// NewCommunityHandler creates a new CommunityHandler.
func NewCommunityHandler(svc *service.CommunityService) *CommunityHandler {
	return &CommunityHandler{svc: svc}
}

// List handles GET /api/communities.
func (h *CommunityHandler) List(c *fiber.Ctx) error {
	resp, err := h.svc.List(c.Context())
	if err != nil {
		slog.Error("community_list_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list communities",
		})
	}
	return c.JSON(resp)
}

// Get handles GET /api/communities/:id.
func (h *CommunityHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Community ID must be an integer",
		})
	}

	comm, err := h.svc.Get(c.Context(), id)
	if err != nil {
		slog.Error("community_get_failed", slog.Int("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get community",
		})
	}
	if comm == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "community_not_found",
			Message: "Community with the specified ID not found",
		})
	}
	return c.JSON(comm)
}

// Create handles POST /api/communities.
func (h *CommunityHandler) Create(c *fiber.Ctx) error {
	var req model.CommunityCreateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("community_create_parse_error", slog.Any("error", err))
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

	comm, err := h.svc.Create(c.Context(), req)
	if err != nil {
		if err.Error() == "community already exists" {
			return c.Status(409).JSON(model.ErrorResponse{
				Error:   "community_exists",
				Message: "A community with this ID already exists",
			})
		}
		slog.Error("community_create_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "save_failed",
			Message: "Failed to create community",
		})
	}

	slog.Info("community_created", slog.Int("id", comm.ID), slog.String("name", comm.Name))
	return c.Status(201).JSON(comm)
}

// Update handles PUT /api/communities/:id.
func (h *CommunityHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Community ID must be an integer",
		})
	}

	var req model.CommunityUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("community_update_parse_error", slog.Any("error", err))
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	comm, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		slog.Error("community_update_failed", slog.Int("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "save_failed",
			Message: "Failed to update community",
		})
	}
	if comm == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "community_not_found",
			Message: "Community with the specified ID not found",
		})
	}

	slog.Info("community_updated", slog.Int("id", comm.ID))
	return c.JSON(comm)
}

// Delete handles DELETE /api/communities/:id.
func (h *CommunityHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Community ID must be an integer",
		})
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		slog.Error("community_delete_failed", slog.Int("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "delete_failed",
			Message: "Failed to delete community",
		})
	}

	slog.Info("community_deleted", slog.Int("id", id))
	return c.SendStatus(204)
}
