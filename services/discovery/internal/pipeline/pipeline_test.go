package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"discovery/internal/coreclient"
	"discovery/internal/extractor"
	"discovery/internal/model"
)

// fakeCore is an httptest core service covering the endpoints the
// pipeline touches, with failure injection for the provider and review
// stages.
type fakeCore struct {
	mu sync.Mutex

	paper       model.Paper
	provider    model.LLMProvider
	providerErr bool
	reviewErr   bool

	statusLog []string // ordered UpdatePaperStatus payloads
	review    []model.ReviewQueueItem
}

const extractionJSON = `{"nodes":[{"name":"Attention","description":"Mechanism","relevance":9}],"edges":[]}`

// newEnv builds a pipeline wired to a fake core; the fake also serves
// the OpenAI-compatible chat completions endpoint the provider points at.
func newEnv(t *testing.T) (*fakeCore, *Service) {
	t.Helper()
	fc := &fakeCore{
		paper: model.Paper{ID: "p1", Title: "Attention", RawText: "text", Status: "uploaded"},
		provider: model.LLMProvider{
			ID: "prov1", BaseURL: "placeholder", APIKey: "k", Model: "m",
		},
	}

	chat := func(w http.ResponseWriter, _ *http.Request) {
		// The LLM returns the extraction result as a JSON string inside
		// the message content field, matching the real client protocol.
		content, _ := json.Marshal(extractionJSON)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":` + string(content) + `}}],"usage":{"total_tokens":10}}`))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/papers/p1", func(w http.ResponseWriter, _ *http.Request) {
		fc.mu.Lock()
		defer fc.mu.Unlock()
		_ = json.NewEncoder(w).Encode(fc.paper)
	})
	mux.HandleFunc("PUT /api/papers/p1", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Status string `json:"status"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		fc.mu.Lock()
		fc.statusLog = append(fc.statusLog, body.Status)
		fc.mu.Unlock()
		_, _ = w.Write([]byte(`{}`))
	})
	mux.HandleFunc("GET /api/internal/llm/providers/default", func(w http.ResponseWriter, _ *http.Request) {
		fc.mu.Lock()
		defer fc.mu.Unlock()
		if fc.providerErr {
			w.WriteHeader(502)
			_, _ = w.Write([]byte(`{"error":"no provider"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(fc.provider)
	})
	mux.HandleFunc("POST /api/review-queue", func(w http.ResponseWriter, r *http.Request) {
		var item model.ReviewQueueItem
		_ = json.NewDecoder(r.Body).Decode(&item)
		fc.mu.Lock()
		defer fc.mu.Unlock()
		if fc.reviewErr {
			w.WriteHeader(502)
			_, _ = w.Write([]byte(`{"error":"queue down"}`))
			return
		}
		fc.review = append(fc.review, item)
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{}`))
	})
	mux.HandleFunc("POST /v1/chat/completions", chat)
	mux.HandleFunc("POST /chat/completions", chat)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	fc.mu.Lock()
	fc.provider.BaseURL = srv.URL + "/v1"
	fc.mu.Unlock()

	client := coreclient.New(srv.URL, "tok")
	return fc, NewService(client, extractor.NewService())
}

func (fc *fakeCore) statuses() []string {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return append([]string(nil), fc.statusLog...)
}

func (fc *fakeCore) reviews() []model.ReviewQueueItem {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return append([]model.ReviewQueueItem(nil), fc.review...)
}

// TestRunPaper_HappyPath verifies the full sequence: analyze status,
// review submission carrying the extracted payload, analyzed status.
func TestRunPaper_HappyPath(t *testing.T) {
	fc, svc := newEnv(t)

	if err := svc.RunPaper(context.Background(), "p1"); err != nil {
		t.Fatalf("RunPaper failed: %v", err)
	}

	if got := fc.statuses(); len(got) != 2 || got[0] != "analyzing" || got[1] != "analyzed" {
		t.Errorf("statuses = %v, want [analyzing analyzed]", got)
	}
	reviews := fc.reviews()
	if len(reviews) != 1 {
		t.Fatalf("review submissions = %d, want 1", len(reviews))
	}
	if reviews[0].PaperID != "p1" {
		t.Errorf("review paper_id = %q, want p1", reviews[0].PaperID)
	}
	if len(reviews[0].ExtractedNodes) != 1 || reviews[0].ExtractedNodes[0].Name != "Attention" {
		t.Errorf("review nodes = %+v, want one 'Attention' node", reviews[0].ExtractedNodes)
	}
}

// TestRunPaper_ProviderFailureRollsBackStatus verifies the paper is
// reset to "uploaded" when the provider cannot be fetched.
func TestRunPaper_ProviderFailureRollsBackStatus(t *testing.T) {
	fc, svc := newEnv(t)
	fc.providerErr = true

	err := svc.RunPaper(context.Background(), "p1")
	if !errors.Is(err, ErrProviderFetch) {
		t.Fatalf("err = %v, want ErrProviderFetch", err)
	}
	if got := fc.statuses(); len(got) != 2 || got[0] != "analyzing" || got[1] != "uploaded" {
		t.Errorf("statuses = %v, want rollback to uploaded", got)
	}
	if len(fc.reviews()) != 0 {
		t.Error("no review should be submitted on provider failure")
	}
}

// TestRunPaper_ExtractionFailureRollsBackStatus verifies the rollback
// when the LLM call fails.
func TestRunPaper_ExtractionFailureRollsBackStatus(t *testing.T) {
	fc, svc := newEnv(t)
	// Point the provider at a dead endpoint so the extraction call fails.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()
	fc.mu.Lock()
	fc.provider.BaseURL = srv.URL
	fc.mu.Unlock()

	err := svc.RunPaper(context.Background(), "p1")
	if !errors.Is(err, ErrExtraction) {
		t.Fatalf("err = %v, want ErrExtraction", err)
	}
	if got := fc.statuses(); len(got) != 2 || got[1] != "uploaded" {
		t.Errorf("statuses = %v, want rollback to uploaded", got)
	}
}

// TestRunPaper_ReviewSubmitFailureKeepsAnalyzed verifies that a failed
// review submission leaves the paper as "analyzed" (not "uploaded"),
// preserving the original handler semantics.
func TestRunPaper_ReviewSubmitFailureKeepsAnalyzed(t *testing.T) {
	fc, svc := newEnv(t)
	fc.reviewErr = true

	err := svc.RunPaper(context.Background(), "p1")
	if !errors.Is(err, ErrReviewSubmit) {
		t.Fatalf("err = %v, want ErrReviewSubmit", err)
	}
	if got := fc.statuses(); len(got) != 2 || got[0] != "analyzing" || got[1] != "analyzed" {
		t.Errorf("statuses = %v, want [analyzing analyzed]", got)
	}
}
