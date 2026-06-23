package handler

import (
	"log/slog"

	"discovery/internal/consoleclient"
	"discovery/internal/coreclient"
	"discovery/internal/extractor"
	"discovery/internal/llm"
	"discovery/internal/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ErrorResponse is the standard error JSON shape.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// AnalyzeHandler handles paper analysis requests.
type AnalyzeHandler struct {
	core      *coreclient.Client
	console   *consoleclient.Client
	extractor *extractor.Service
}

// NewAnalyzeHandler creates a new AnalyzeHandler.
func NewAnalyzeHandler(core *coreclient.Client, console *consoleclient.Client, ext *extractor.Service) *AnalyzeHandler {
	return &AnalyzeHandler{
		core:      core,
		console:   console,
		extractor: ext,
	}
}

// Analyze handles POST /api/analyze. It triggers the full extraction
// pipeline: fetch paper → get LLM provider → extract → submit review.
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

	ctx := c.Context()
	paperID := req.PaperID

	// 1. Fetch paper from core.
	slog.Info("analyze_fetch_paper", slog.String("paper_id", paperID))
	paper, err := h.core.GetPaper(ctx, paperID)
	if err != nil {
		slog.Error("analyze_fetch_paper_failed", slog.String("paper_id", paperID), slog.Any("error", err))
		return c.Status(502).JSON(ErrorResponse{
			Error:   "paper_fetch_failed",
			Message: "Failed to fetch paper from core service",
		})
	}

	// 2. Update paper status to "analyzing".
	if err := h.core.UpdatePaperStatus(ctx, paperID, "analyzing"); err != nil {
		slog.Warn("analyze_status_update_failed", slog.String("paper_id", paperID), slog.Any("error", err))
		// Non-fatal: continue with analysis.
	}

	// 3. Get default LLM provider from console.
	slog.Info("analyze_fetch_provider")
	provider, err := h.console.GetDefaultProvider(ctx)
	if err != nil {
		slog.Error("analyze_fetch_provider_failed", slog.Any("error", err))
		h.core.UpdatePaperStatus(ctx, paperID, "uploaded") // Rollback status.
		return c.Status(502).JSON(ErrorResponse{
			Error:   "provider_fetch_failed",
			Message: "Failed to fetch LLM provider configuration",
		})
	}

	// 4. Run extraction pipeline.
	llmClient := llm.NewClient(provider)
	result, err := h.extractor.Extract(ctx, llmClient, paper)
	if err != nil {
		slog.Error("analyze_extraction_failed", slog.String("paper_id", paperID), slog.Any("error", err))
		h.core.UpdatePaperStatus(ctx, paperID, "uploaded") // Rollback status.
		return c.Status(500).JSON(ErrorResponse{
			Error:   "extraction_failed",
			Message: "LLM extraction failed: " + err.Error(),
		})
	}

	// 5. Submit to review queue.
	slog.Info("analyze_submit_review",
		slog.String("paper_id", paperID),
		slog.Int("nodes", len(result.Nodes)),
		slog.Int("edges", len(result.Edges)),
	)
	reviewItem := model.ReviewQueueItem{
		PaperID:        paperID,
		ExtractedNodes: result.Nodes,
		ExtractedEdges: result.Edges,
	}
	if err := h.core.SubmitReviewItem(ctx, reviewItem); err != nil {
		slog.Error("analyze_submit_review_failed", slog.String("paper_id", paperID), slog.Any("error", err))
		h.core.UpdatePaperStatus(ctx, paperID, "analyzed") // Mark as analyzed but not submitted.
		return c.Status(502).JSON(ErrorResponse{
			Error:   "review_submit_failed",
			Message: "Failed to submit extraction result to review queue",
		})
	}

	// 6. Update paper status to "analyzed".
	if err := h.core.UpdatePaperStatus(ctx, paperID, "analyzed"); err != nil {
		slog.Warn("analyze_final_status_update_failed", slog.String("paper_id", paperID), slog.Any("error", err))
	}

	slog.Info("analyze_completed",
		slog.String("paper_id", paperID),
		slog.Int("nodes", len(result.Nodes)),
		slog.Int("edges", len(result.Edges)),
		slog.String("review_item_id", uuid.New().String()), // Placeholder; actual ID comes from core.
	)

	return c.JSON(model.AnalyzeResponse{
		Status:  "completed",
		Message: "Extraction completed successfully",
	})
}

// Health handles GET /health.
func (h *AnalyzeHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
