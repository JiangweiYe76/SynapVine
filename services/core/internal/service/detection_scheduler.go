package service

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// redetectTimeout bounds a single community re-detection run triggered
// from an approval.
const redetectTimeout = 2 * time.Minute

// Detector is the minimal community-detection capability the scheduler
// needs. *CommunityDetectorService satisfies it.
type Detector interface {
	DetectAndStore(ctx context.Context) error
}

// DetectionScheduler triggers community re-detection asynchronously in
// response to graph mutations (e.g. review approvals). Concurrent
// triggers are coalesced: while a detection is running, additional
// triggers collapse into at most one queued follow-up run, so a burst
// of approvals never runs overlapping detections.
type DetectionScheduler struct {
	detector Detector
	timeout  time.Duration

	mu     sync.Mutex  // held for the duration of a run
	queued atomic.Bool // a follow-up run is requested
}

// NewDetectionScheduler creates a scheduler around the given detector.
func NewDetectionScheduler(detector Detector) *DetectionScheduler {
	return &DetectionScheduler{detector: detector, timeout: redetectTimeout}
}

// Trigger launches a community re-detection in a background goroutine.
// If a detection is already in progress, the request is coalesced into
// one queued follow-up run. Safe to call with a nil receiver (no-op),
// so callers do not need to guard disabled setups.
func (s *DetectionScheduler) Trigger() {
	if s == nil {
		return
	}
	if !s.mu.TryLock() {
		s.queued.Store(true)
		return
	}
	go s.run()
}

// run executes detection runs until no queued trigger remains. The
// scheduler mutex is held for the whole sequence.
func (s *DetectionScheduler) run() {
	defer s.mu.Unlock()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		err := s.detector.DetectAndStore(ctx)
		cancel()
		if err != nil {
			slog.Error("community_redetect_failed", slog.Any("error", err))
		}
		// Collapse triggers that arrived while running into a single
		// follow-up run; stop when nothing is pending.
		if !s.queued.CompareAndSwap(true, false) {
			return
		}
	}
}
