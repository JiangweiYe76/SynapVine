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

	chatCalls int // number of LLM chat completion requests served

	statusLog []string // ordered UpdatePaperStatus payloads
	review    []model.ReviewQueueItem
	usage     []coreclient.UsageRecord
}

const extractionJSON = `{"nodes":[{"name":"Attention","description":"Mechanism","relevance":9}],"edges":[]}`

// paperText is long enough to pass the extractor's minimum-length guard.
const paperText = "We introduce the Transformer, a network architecture based solely on attention mechanisms, " +
	"dispensing with recurrence and convolutions entirely. Experiments on two machine translation tasks show " +
	"these models to be superior in quality while being more parallelizable and requiring significantly less " +
	"time to train. Our model achieves state-of-the-art results on the WMT 2014 English-to-German translation task."

// newEnv builds a pipeline wired to a fake core; the fake also serves
// the OpenAI-compatible chat completions endpoint the provider points at.
func newEnv(t *testing.T) (*fakeCore, *Service) {
	t.Helper()
	fc := &fakeCore{
		paper: model.Paper{ID: "p1", Title: "Attention", RawText: paperText, Status: "uploaded"},
		provider: model.LLMProvider{
			ID: "prov1", BaseURL: "placeholder", APIKey: "k", Model: "m",
		},
	}

	chat := func(w http.ResponseWriter, _ *http.Request) {
		fc.mu.Lock()
		fc.chatCalls++
		fc.mu.Unlock()
		// The LLM returns the extraction result as a JSON string inside
		// the message content field, matching the real client protocol.
		content, _ := json.Marshal(extractionJSON)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":` + string(content) + `}}],"usage":{"prompt_tokens":100,"completion_tokens":20,"total_tokens":120}}`))
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

	mux.HandleFunc("POST /api/internal/llm/usage", func(w http.ResponseWriter, r *http.Request) {
		var rec coreclient.UsageRecord
		_ = json.NewDecoder(r.Body).Decode(&rec)
		fc.mu.Lock()
		fc.usage = append(fc.usage, rec)
		fc.mu.Unlock()
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

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

func (fc *fakeCore) usages() []coreclient.UsageRecord {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return append([]coreclient.UsageRecord(nil), fc.usage...)
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
	usages := fc.usages()
	if len(usages) != 1 {
		t.Fatalf("usage records = %d, want 1", len(usages))
	}
	u := usages[0]
	if u.PaperID != "p1" || u.ProviderID != "prov1" || u.Model != "m" {
		t.Errorf("usage refs = %+v, want paper p1 / provider prov1 / model m", u)
	}
	if u.PromptTokens != 100 || u.CompletionTokens != 20 || u.TotalTokens != 120 {
		t.Errorf("usage tokens = %d/%d/%d, want 100/20/120", u.PromptTokens, u.CompletionTokens, u.TotalTokens)
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

// TestRunPaper_PlaceholderTextSkipsLLM verifies that a paper whose text
// is a placeholder (failed PDF extraction upstream) never reaches the
// LLM, is marked "failed", and produces no review item or usage record.
func TestRunPaper_PlaceholderTextSkipsLLM(t *testing.T) {
	fc, svc := newEnv(t)
	fc.paper.RawText = "(PDF uploaded — text extraction pending)"

	err := svc.RunPaper(context.Background(), "p1")
	if !errors.Is(err, ErrExtraction) {
		t.Fatalf("err = %v, want ErrExtraction", err)
	}
	if !errors.Is(err, extractor.ErrInputUnusable) {
		t.Errorf("err = %v, want it to wrap extractor.ErrInputUnusable", err)
	}
	if got := fc.statuses(); len(got) != 2 || got[0] != "analyzing" || got[1] != "failed" {
		t.Errorf("statuses = %v, want [analyzing failed]", got)
	}
	if fc.chatCalls != 0 {
		t.Errorf("llm chat calls = %d, want 0 (no tokens spent)", fc.chatCalls)
	}
	if n := len(fc.reviews()); n != 0 {
		t.Errorf("review submissions = %d, want 0 (no fabricated nodes)", n)
	}
	if n := len(fc.usages()); n != 0 {
		t.Errorf("usage records = %d, want 0", n)
	}
}

// TestRunPaper_ShortTextSkipsLLM verifies the minimum-length guard uses
// the same terminal "failed" path as the placeholder guard.
func TestRunPaper_ShortTextSkipsLLM(t *testing.T) {
	fc, svc := newEnv(t)
	fc.paper.RawText = "too short to analyze"

	err := svc.RunPaper(context.Background(), "p1")
	if !errors.Is(err, ErrExtraction) {
		t.Fatalf("err = %v, want ErrExtraction", err)
	}
	if got := fc.statuses(); len(got) != 2 || got[1] != "failed" {
		t.Errorf("statuses = %v, want rollback to failed", got)
	}
	if fc.chatCalls != 0 {
		t.Errorf("llm chat calls = %d, want 0", fc.chatCalls)
	}
}

// TestRunPaper_TransientFailureStillRollsBackToUploaded verifies that a
// real LLM failure keeps the retryable "uploaded" status, distinct from
// the terminal "failed" used for unusable input.
func TestRunPaper_TransientFailureStillRollsBackToUploaded(t *testing.T) {
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
	if errors.Is(err, extractor.ErrInputUnusable) {
		t.Error("transient failure must not be classified as unusable input")
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
	if got := fc.statusLog; len(got) != 2 || got[0] != "analyzing" || got[1] != "analyzed" {
		t.Errorf("statuses = %v, want [analyzing analyzed]", got)
	}
	// Usage is recorded before the review submission, so the spent
	// tokens must still be accounted for.
	if got := fc.usages(); len(got) != 1 {
		t.Errorf("usage records = %d, want 1 (recorded despite review failure)", len(got))
	}
}
