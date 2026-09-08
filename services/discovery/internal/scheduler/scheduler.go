// Package scheduler periodically fetches new papers from ArXiv and
// pushes them through the extraction pipeline. Each cycle: fetch the
// most recent submissions for the configured categories, skip papers
// that were already ingested (dedup by ArXiv id), register the rest in
// core, and run them through the pipeline into the review queue.
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"discovery/internal/arxiv"
	"discovery/internal/coreclient"
	"discovery/internal/pipeline"
)

// Config controls the ingestion loop.
type Config struct {
	Categories []string      // ArXiv categories to watch, e.g. cs.AI
	Interval   time.Duration // time between cycles
	MaxPerRun  int           // max papers fetched per cycle
}

// Scheduler ingests new ArXiv papers on a fixed interval.
type Scheduler struct {
	arxiv    *arxiv.Client
	core     *coreclient.Client
	pipeline *pipeline.Service
	cfg      Config
}

// New creates a Scheduler.
func New(a *arxiv.Client, core *coreclient.Client, pipe *pipeline.Service, cfg Config) *Scheduler {
	if cfg.Interval <= 0 {
		cfg.Interval = time.Hour
	}
	if cfg.MaxPerRun <= 0 {
		cfg.MaxPerRun = 5
	}
	return &Scheduler{arxiv: a, core: core, pipeline: pipe, cfg: cfg}
}

// Run blocks until ctx is cancelled, running one ingestion cycle
// immediately and then after every interval tick. Cycles are strictly
// sequential: a slow cycle delays the next one instead of overlapping.
func (s *Scheduler) Run(ctx context.Context) {
	slog.Info("arxiv_scheduler_started",
		slog.String("categories", strings.Join(s.cfg.Categories, ",")),
		slog.String("interval", s.cfg.Interval.String()),
		slog.Int("max_per_run", s.cfg.MaxPerRun),
	)

	s.RunOnce(ctx) // initial cycle at startup

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("arxiv_scheduler_stopped")
			return
		case <-ticker.C:
			s.RunOnce(ctx)
		}
	}
}

// RunOnce performs a single ingestion cycle. Failures are logged and
// do not stop the scheduler; per-paper failures do not stop the cycle.
func (s *Scheduler) RunOnce(ctx context.Context) {
	papers, err := s.arxiv.FetchRecent(ctx, s.cfg.Categories, s.cfg.MaxPerRun)
	if err != nil {
		slog.Error("arxiv_fetch_failed", slog.Any("error", err))
		return
	}
	slog.Info("arxiv_fetch_completed", slog.Int("fetched", len(papers)))

	for _, p := range papers {
		if ctx.Err() != nil {
			return
		}
		if err := s.ingest(ctx, p); err != nil {
			slog.Error("arxiv_ingest_failed",
				slog.String("arxiv_id", p.ArxivID),
				slog.String("title", p.Title),
				slog.Any("error", err))
		}
	}
}

// ingest registers one paper in core (unless already present) and runs
// it through the extraction pipeline.
func (s *Scheduler) ingest(ctx context.Context, p arxiv.Paper) error {
	if _, found, err := s.core.GetPaperByArxivID(ctx, p.ArxivID); err != nil {
		return fmt.Errorf("dedup lookup failed: %w", err)
	} else if found {
		slog.Debug("arxiv_paper_already_ingested", slog.String("arxiv_id", p.ArxivID))
		return nil
	}

	rawText := p.Title + "\n\n" + p.Summary
	created, err := s.core.CreatePaper(ctx, coreclient.CreatePaperRequest{
		Title:     p.Title,
		Authors:   strings.Join(p.Authors, ", "),
		SourceURL: p.SourceURL,
		ArxivID:   p.ArxivID,
		RawText:   rawText,
	})
	if err != nil {
		return fmt.Errorf("create paper record: %w", err)
	}
	slog.Info("arxiv_paper_created",
		slog.String("arxiv_id", p.ArxivID),
		slog.String("paper_id", created.ID),
		slog.String("title", p.Title),
	)

	if err := s.pipeline.RunPaper(ctx, created.ID); err != nil {
		return fmt.Errorf("pipeline failed: %w", err)
	}
	return nil
}
