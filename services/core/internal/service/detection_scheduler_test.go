package service

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeDetector counts DetectAndStore calls and can block the first call
// until released, simulating a long-running detection.
type fakeDetector struct {
	mu    sync.Mutex
	calls atomic.Int32
	gate  chan struct{} // when non-nil, the first call blocks until closed
}

func (f *fakeDetector) DetectAndStore(ctx context.Context) error {
	if f.gate != nil {
		f.mu.Lock()
		g := f.gate
		f.mu.Unlock()
		if g != nil {
			<-g
			f.mu.Lock()
			f.gate = nil // only the first call blocks
			f.mu.Unlock()
		}
	}
	f.calls.Add(1)
	return nil
}

func (f *fakeDetector) callCount() int32 { return f.calls.Load() }

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("timed out waiting for condition")
}

// TestDetectionScheduler_TriggerRunsDetection verifies a single trigger
// runs the detector exactly once.
func TestDetectionScheduler_TriggerRunsDetection(t *testing.T) {
	det := &fakeDetector{}
	s := NewDetectionScheduler(det)

	s.Trigger()
	waitFor(t, 2*time.Second, func() bool { return det.callCount() == 1 })

	// Give a potential duplicate run time to (wrongly) start.
	time.Sleep(50 * time.Millisecond)
	if got := det.callCount(); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}

// TestDetectionScheduler_NilReceiverIsNoop verifies the nil-safety
// contract used by handlers with optional schedulers.
func TestDetectionScheduler_NilReceiverIsNoop(t *testing.T) {
	var s *DetectionScheduler
	s.Trigger() // must not panic
}

// TestDetectionScheduler_CoalescesConcurrentTriggers verifies that
// triggers arriving while a detection is running collapse into exactly
// one queued follow-up run instead of N overlapping runs.
func TestDetectionScheduler_CoalescesConcurrentTriggers(t *testing.T) {
	det := &fakeDetector{gate: make(chan struct{})}
	s := NewDetectionScheduler(det)

	s.Trigger() // acquires the scheduler and blocks inside DetectAndStore

	// Wait until the first run is actually inside the detector.
	waitFor(t, 2*time.Second, func() bool { return s.mu.TryLock() == false })

	// A burst of triggers while the first run is in flight: each one
	// should only set the queued flag.
	for i := 0; i < 5; i++ {
		s.Trigger()
	}

	close(det.gate) // let the first run finish

	// Exactly one follow-up run should execute, so the total is 2.
	waitFor(t, 2*time.Second, func() bool { return det.callCount() == 2 })
	time.Sleep(50 * time.Millisecond)
	if got := det.callCount(); got != 2 {
		t.Errorf("calls = %d, want 2 (1 initial + 1 coalesced follow-up)", got)
	}
}
