package handler

import (
	"errors"
	"log/slog"

	"discovery/internal/model"
	"discovery/internal/pipeline"

	"github.com/gofiber/fiber/v2"
)

// ErrorResponse is the standard error JSON shape.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// AnalyzeHandler handles paper analysis requests.
type AnalyzeHandler struct {
	pipeline *pipeline.Service
}

// NewAnalyzeHandler creates a new AnalyzeHandler.
func NewAnalyzeHandler(pipe *pipeline.Service) *AnalyzeHandler {
	return &AnalyzeHandler{pipeline: pipe}
}

// Analyze handles POST /api/analyze. It runs the extraction pipeline
// (fetch paper → LLM extract → submit review) for the given paper; the
// pipeline is shared with the ArXiv scheduler.
func (h *AnalyzeHandler) Analyze(c *fiber.Ctx) error {
	var req model.AnalyzeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}
	if req.PaperID == "" {
		return c.Status(400).JSON(ErrorResponse{
			Error:   "missing_fields",
			Message: "paper_id is required",
		})
	}

	if err := h.pipeline.RunPaper(c.Context(), req.PaperID); err != nil {
		return analyzeErrorResponse(c, err)
	}

	return c.JSON(model.AnalyzeResponse{
		Status:  "completed",
		Message: "Extraction completed successfully",
	})
}

// analyzeErrorResponse maps pipeline stage errors to HTTP codes.
func analyzeErrorResponse(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, pipeline.ErrPaperFetch):
		return c.Status(502).JSON(ErrorResponse{
			Error:   "paper_fetch_failed",
			Message: "Failed to fetch paper from core service",
		})
	case errors.Is(err, pipeline.ErrProviderFetch):
		return c.Status(502).JSON(ErrorResponse{
			Error:   "provider_fetch_failed",
			Message: "Failed to fetch LLM provider configuration",
		})
	case errors.Is(err, pipeline.ErrExtraction):
		return c.Status(500).JSON(ErrorResponse{
			Error:   "extraction_failed",
			Message: "LLM extraction failed: " + err.Error(),
		})
	case errors.Is(err, pipeline.ErrReviewSubmit):
		return c.Status(502).JSON(ErrorResponse{
			Error:   "review_submit_failed",
			Message: "Failed to submit extraction result to review queue",
		})
	default:
		slog.Error("analyze_unexpected_error", slog.Any("error", err))
		return c.Status(500).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Unexpected analysis failure",
		})
	}
}

// Health handles GET /health.
func (h *AnalyzeHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
