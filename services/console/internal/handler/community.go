package handler

import (
	"errors"
	"log/slog"

	"console/internal/coreclient"
	"console/internal/model"

	"github.com/gofiber/fiber/v2"
)

// CommunityHandler proxies community HTTP requests to the core service.
type CommunityHandler struct {
	core *coreclient.Client
}

// NewCommunityHandler creates a new CommunityHandler backed by the given core client.
func NewCommunityHandler(core *coreclient.Client) *CommunityHandler {
	return &CommunityHandler{core: core}
}

// List handles GET /api/communities
func (h *CommunityHandler) List(c *fiber.Ctx) error {
	resp, err := h.core.ListCommunities(c.Context())
	if err != nil {
		slog.Error("community_list_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch communities from core service",
		})
	}
	return c.JSON(resp)
}

// Tree handles GET /api/communities/tree
func (h *CommunityHandler) Tree(c *fiber.Ctx) error {
	resp, err := h.core.GetCommunityTree(c.Context())
	if err != nil {
		slog.Error("community_tree_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch community tree from core service",
		})
	}
	return c.JSON(resp)
}

// Get handles GET /api/communities/:id
func (h *CommunityHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Community ID is required",
		})
	}

	comm, err := h.core.GetCommunity(c.Context(), id)
	if err != nil {
		slog.Error("community_get_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to fetch community from core service",
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

// Create handles POST /api/communities
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

	comm, err := h.core.CreateCommunity(c.Context(), req)
	if err != nil {
		var httpErr *coreclient.HTTPStatusError
		if errors.As(err, &httpErr) {
			switch httpErr.StatusCode {
			case 409:
				return c.Status(409).JSON(model.ErrorResponse{
					Error:   "community_exists",
					Message: "A community with this ID already exists",
				})
			case 400:
				return c.Status(400).JSON(model.ErrorResponse{
					Error:   "invalid_request",
					Message: "Invalid community create request",
				})
			}
		}
		slog.Error("community_create_core_failed", slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to create community in core service",
		})
	}

	slog.Info("community_created", slog.String("id", comm.ID), slog.String("name", comm.Name))
	return c.Status(201).JSON(comm)
}

// Update handles PUT /api/communities/:id
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

	comm, err := h.core.UpdateCommunity(c.Context(), id, req)
	if err != nil {
		var httpErr *coreclient.HTTPStatusError
		if errors.As(err, &httpErr) {
			switch httpErr.StatusCode {
			case 400:
				return c.Status(400).JSON(model.ErrorResponse{
					Error:   "invalid_request",
					Message: "Invalid community update request",
				})
			}
		}
		slog.Error("community_update_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to update community in core service",
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

// Delete handles DELETE /api/communities/:id
func (h *CommunityHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Community ID is required",
		})
	}

	ok, err := h.core.DeleteCommunity(c.Context(), id)
	if err != nil {
		var httpErr *coreclient.HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 409 {
			return c.Status(409).JSON(model.ErrorResponse{
				Error:   "community_has_children",
				Message: "Cannot delete a community that has child communities",
			})
		}
		slog.Error("community_delete_core_failed", slog.String("id", id), slog.Any("error", err))
		return c.Status(502).JSON(model.ErrorResponse{
			Error:   "core_unavailable",
			Message: "Failed to delete community in core service",
		})
	}
	if !ok {
		return c.Status(404).JSON(model.ErrorResponse{
			Error:   "community_not_found",
			Message: "Community with the specified ID not found",
		})
	}

	slog.Info("community_deleted", slog.String("id", id))
	return c.SendStatus(204)
}
