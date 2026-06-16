package service

import (
	"context"
	"testing"

	"core/internal/model"
)

// TestTimelineRange_CacheShortCircuitsRepo verifies that once the range
// has been computed and cached, subsequent calls do not invoke the
// underlying repository. The test relies on the repository being nil;
// a non-nil cache must prevent the call from reaching the repo.
func TestTimelineRange_CacheShortCircuitsRepo(t *testing.T) {
	svc := &NodeService{
		repo:          nil, // would panic if the cache were missed
		timelineCache: &model.TimelineRange{MinYear: 1957, MaxYear: 2024},
	}

	tr, err := svc.TimelineRange(context.Background())
	if err != nil {
		t.Fatalf("TimelineRange failed: %v", err)
	}
	if tr.MinYear != 1957 || tr.MaxYear != 2024 {
		t.Errorf("got %+v, want {1957, 2024}", tr)
	}
}

// TestTimelineRange_InvalidateClearsCache verifies that the cache
// invalidation hook called by the write paths actually drops the cache,
// forcing the next read to recompute.
func TestTimelineRange_InvalidateClearsCache(t *testing.T) {
	svc := &NodeService{
		repo:          nil,
		timelineCache: &model.TimelineRange{MinYear: 1957, MaxYear: 2024},
	}

	svc.invalidateTimelineRange()

	if svc.timelineCache != nil {
		t.Fatalf("expected cache to be nil after invalidation, got %+v", svc.timelineCache)
	}
}

// TestTimelineRange_ConcurrentReadsAreSafe verifies that many concurrent
// readers do not race on the cache. A data race would be reported by
// `go test -race`.
func TestTimelineRange_ConcurrentReadsAreSafe(t *testing.T) {
	svc := &NodeService{
		repo:          nil,
		timelineCache: &model.TimelineRange{MinYear: 1957, MaxYear: 2024},
	}

	const goroutines = 64
	done := make(chan struct{}, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			tr, err := svc.TimelineRange(context.Background())
			if err != nil {
				t.Errorf("TimelineRange failed: %v", err)
				return
			}
			if tr.MinYear != 1957 || tr.MaxYear != 2024 {
				t.Errorf("got %+v, want {1957, 2024}", tr)
			}
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}
}
