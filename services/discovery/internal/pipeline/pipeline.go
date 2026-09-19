// Package pipeline orchestrates the full paper extraction flow:
// fetch paper from core, mark it analyzing, load the default LLM
// provider, run the extraction, submit the result to the review queue,
// and mark the paper analyzed. Both the HTTP handler and the ArXiv
// scheduler push papers through this service so the two entry points
// cannot drift apart.
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"discovery/internal/coreclient"
	"discovery/internal/extractor"
	"discovery/internal/llm"
	"discovery/internal/model"
)

// Stage sentinel errors. Callers (the HTTP handler) map these to
// response codes; the wrapped cause carries the details.
var (
	ErrPaperFetch    = errors.New("paper fetch failed")
	ErrProviderFetch = errors.New("provider fetch failed")
	ErrExtraction    = errors.New("extraction failed")
	ErrReviewSubmit  = errors.New("review submit failed")
)

// Service runs papers through the extraction pipeline.
type Service struct {
	core      *coreclient.Client
	extractor *extractor.Service
}

// NewService creates a pipeline Service.
func NewService(core *coreclient.Client, ext *extractor.Service) *Service {
	return &Service{core: core, extractor: ext}
}

// RunPaper executes the pipeline for one already-registered paper. A
// failure at the fetch/provider/extraction/submit stage rolls the paper
// status back to "uploaded" (or marks it "analyzed" when the review
// submission failed after a successful extraction) so no paper is stuck
// in a transient state.
func (s *Service) RunPaper(ctx context.Context, paperID string) error {
	// 1. Fetch paper from core.
	slog.Info("analyze_fetch_paper", slog.String("paper_id", paperID))
	paper, err := s.core.GetPaper(ctx, paperID)
	if err != nil {
		slog.Error("analyze_fetch_paper_failed", slog.String("paper_id", paperID), slog.Any("error", err))
		return fmt.Errorf("%w: paper %s: %v", ErrPaperFetch, paperID, err)
	}

	// 2. Update paper status to "analyzing".
	if err := s.core.UpdatePaperStatus(ctx, paperID, "analyzing"); err != nil {
		slog.Warn("analyze_status_update_failed", slog.String("paper_id", paperID), slog.Any("error", err))
		// Non-fatal: continue with analysis.
	}

	// 3. Get default LLM provider from core.
	slog.Info("analyze_fetch_provider")
	provider, err := s.core.GetDefaultLLMProvider(ctx)
	if err != nil {
		slog.Error("analyze_fetch_provider_failed", slog.Any("error", err))
		s.rollbackStatus(ctx, paperID, "uploaded")
		return fmt.Errorf("%w: %v", ErrProviderFetch, err)
	}

	// 4. Run extraction pipeline.
	llmClient := llm.NewClient(provider)
	result, err := s.extractor.Extract(ctx, llmClient, paper)
	if err != nil {
		slog.Error("analyze_extraction_failed", slog.String("paper_id", paperID), slog.Any("error", err))
		// Unusable input is terminal: the stored text will not improve on
		// a retry, so mark the paper "failed" instead of "uploaded" to
		// stop it from looking retryable. Transient failures (network,
		// provider errors, malformed output) still roll back to
		// "uploaded" so a retry can succeed.
		if errors.Is(err, extractor.ErrInputUnusable) {
			s.rollbackStatus(ctx, paperID, "failed")
		} else {
			s.rollbackStatus(ctx, paperID, "uploaded")
		}
		// Wrap both sentinels so callers can classify the stage
		// (ErrExtraction) and the cause (ErrInputUnusable) independently.
		return fmt.Errorf("%w: %w", ErrExtraction, err)
	}

	// 4b. Record token usage. Non-fatal: accounting failures must not
	// affect the analysis outcome.
	usage := coreclient.UsageRecord{
		PaperID:          paperID,
		ProviderID:       provider.ID,
		Model:            provider.Model,
		PromptTokens:     result.PromptTokens,
		CompletionTokens: result.CompletionTokens,
		TotalTokens:      result.TotalTokens,
	}
	if err := s.core.RecordUsage(ctx, usage); err != nil {
		slog.Warn("analyze_usage_record_failed",
			slog.String("paper_id", paperID),
			slog.Int("total_tokens", usage.TotalTokens),
			slog.Any("error", err))
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
	if err := s.core.SubmitReviewItem(ctx, reviewItem); err != nil {
		slog.Error("analyze_submit_review_failed", slog.String("paper_id", paperID), slog.Any("error", err))
		s.rollbackStatus(ctx, paperID, "analyzed") // Mark as analyzed but not submitted.
		return fmt.Errorf("%w: %v", ErrReviewSubmit, err)
	}

	// 6. Update paper status to "analyzed".
	if err := s.core.UpdatePaperStatus(ctx, paperID, "analyzed"); err != nil {
		slog.Warn("analyze_final_status_update_failed", slog.String("paper_id", paperID), slog.Any("error", err))
	}

	slog.Info("analyze_completed",
		slog.String("paper_id", paperID),
		slog.Int("nodes", len(result.Nodes)),
		slog.Int("edges", len(result.Edges)),
	)
	return nil
}

// rollbackStatus best-effort resets a paper's status after a failure.
func (s *Service) rollbackStatus(ctx context.Context, paperID, status string) {
	if err := s.core.UpdatePaperStatus(ctx, paperID, status); err != nil {
		slog.Warn("analyze_status_rollback_failed",
			slog.String("paper_id", paperID),
			slog.String("status", status),
			slog.Any("error", err))
	}
}
