package handler

import (
	"errors"
	"log/slog"

	"core/internal/model"
	"core/internal/service"

	"github.com/gofiber/fiber/v2"
)

// CommunityHandler handles HTTP requests for community operations.
type CommunityHandler struct {
	svc      *service.CommunityService
	detector *service.CommunityDetectorService
}

// NewCommunityHandler creates a new CommunityHandler.
func NewCommunityHandler(svc *service.CommunityService, detector *service.CommunityDetectorService) *CommunityHandler {
	return &CommunityHandler{svc: svc, detector: detector}
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

// Tree handles GET /api/communities/tree.
func (h *CommunityHandler) Tree(c *fiber.Ctx) error {
	tree, err := h.svc.GetTree(c.Context())
	if err != nil {
		slog.Error("community_tree_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to build community tree",
		})
	}
	return c.JSON(fiber.Map{"communities": tree})
}

// Get handles GET /api/communities/:id.
func (h *CommunityHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Community ID is required",
		})
	}

	comm, err := h.svc.Get(c.Context(), id)
	if err != nil {
		slog.Error("community_get_failed", slog.String("id", id), slog.Any("error", err))
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
		switch {
		case errors.Is(err, service.ErrCommunityExists):
			return c.Status(409).JSON(model.ErrorResponse{
				Error:   "community_exists",
				Message: "A community with this ID already exists",
			})
		case errors.Is(err, service.ErrParentNotFound):
			return c.Status(400).JSON(model.ErrorResponse{
				Error:   "parent_not_found",
				Message: "Parent community does not exist",
			})
		}
		slog.Error("community_create_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "save_failed",
			Message: "Failed to create community",
		})
	}

	slog.Info("community_created", slog.String("id", comm.ID), slog.String("name", comm.Name))
	return c.Status(201).JSON(comm)
}

// Update handles PUT /api/communities/:id.
func (h *CommunityHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Community ID is required",
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
		switch {
		case errors.Is(err, service.ErrParentNotFound):
			return c.Status(400).JSON(model.ErrorResponse{
				Error:   "parent_not_found",
				Message: "Parent community does not exist",
			})
		case errors.Is(err, service.ErrCycle):
			return c.Status(400).JSON(model.ErrorResponse{
				Error:   "cycle_detected",
				Message: "The chosen parent would create a cycle",
			})
		}
		slog.Error("community_update_failed", slog.String("id", id), slog.Any("error", err))
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

	slog.Info("community_updated", slog.String("id", comm.ID))
	return c.JSON(comm)
}

// Delete handles DELETE /api/communities/:id.
func (h *CommunityHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Community ID is required",
		})
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		switch {
		case errors.Is(err, service.ErrCommunityNotFound):
			return c.Status(404).JSON(model.ErrorResponse{
				Error:   "community_not_found",
				Message: "Community with the specified ID not found",
			})
		case errors.Is(err, service.ErrHasChildren):
			return c.Status(409).JSON(model.ErrorResponse{
				Error:   "community_has_children",
				Message: "Cannot delete a community that has child communities",
			})
		}
		slog.Error("community_delete_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "delete_failed",
			Message: "Failed to delete community",
		})
	}

	slog.Info("community_deleted", slog.String("id", id))
	return c.SendStatus(204)
}

// Detect triggers community detection and writes results back to Neo4j.
func (h *CommunityHandler) Detect(c *fiber.Ctx) error {
	if h.detector == nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "not_configured",
			Message: "Community detector is not configured",
		})
	}

	if err := h.detector.DetectAndStore(c.Context()); err != nil {
		slog.Error("community_detect_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "detect_failed",
			Message: "Failed to run community detection",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
	})
}
