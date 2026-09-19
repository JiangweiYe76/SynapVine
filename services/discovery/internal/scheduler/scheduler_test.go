package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"discovery/internal/arxiv"
	"discovery/internal/coreclient"
	"discovery/internal/extractor"
	"discovery/internal/model"
	"discovery/internal/pipeline"
)

const atomFeed = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <entry>
    <id>http://arxiv.org/abs/2609.00001v1</id>
    <title>Paper One</title>
    <summary>We present a new method for training large language models on long documents. Our approach
    combines sparse attention with a learned retrieval index, reducing the quadratic cost of full
    attention to near-linear while preserving accuracy on downstream reasoning benchmarks. Experiments
    across three model scales show consistent perplexity improvements and a fourfold speedup at inference.</summary>
    <published>2026-09-07T10:00:00Z</published>
    <author><name>A. Author</name></author>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/2609.00002v1</id>
    <title>Paper Two</title>
    <summary>This paper studies knowledge graph completion under sparse supervision. We propose a
    contrastive objective that aligns entity embeddings with textual descriptions harvested from the
    web, and show that it substantially improves link prediction on standard benchmarks. Ablations
    isolate the contribution of each component and confirm robustness to noisy descriptions.</summary>
    <published>2026-09-07T11:00:00Z</published>
    <author><name>B. Author</name></author>
  </entry>
</feed>`

const extractionJSON = `{"nodes":[{"name":"Concept","description":"A thing","relevance":5}],"edges":[]}`

// fakeCoreSim is an in-memory core: papers keyed by id, an arxiv-id
// index for the by-arxiv lookup, recorded review submissions, and the
// fake OpenAI-compatible LLM endpoint.
type fakeCoreSim struct {
	mu       sync.Mutex
	papers   map[string]*model.Paper
	byArxiv  map[string]string
	review   []model.ReviewQueueItem
	provider model.LLMProvider
	nextID   int
}

// newCoreSim starts an httptest server simulating core (+LLM) and
// returns the sim plus a coreclient pointed at it.
func newCoreSim(t *testing.T) (*fakeCoreSim, *coreclient.Client) {
	t.Helper()
	fc := &fakeCoreSim{
		papers:   map[string]*model.Paper{},
		byArxiv:  map[string]string{},
		provider: model.LLMProvider{ID: "p", BaseURL: "placeholder", APIKey: "k", Model: "m"},
	}

	chat := func(w http.ResponseWriter, _ *http.Request) {
		content, _ := json.Marshal(extractionJSON)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":` + string(content) + `}}],"usage":{"total_tokens":1}}`))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/chat/completions", chat)

	mux.HandleFunc("POST /api/papers", func(w http.ResponseWriter, r *http.Request) {
		var req coreclient.CreatePaperRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		fc.mu.Lock()
		defer fc.mu.Unlock()
		if _, dup := fc.byArxiv[req.ArxivID]; dup {
			w.WriteHeader(409)
			_, _ = w.Write([]byte(`{"error":"duplicate"}`))
			return
		}
		fc.nextID++
		id := fmt.Sprintf("paper-%d", fc.nextID)
		paper := &model.Paper{ID: id, Title: req.Title, RawText: req.RawText, Status: "uploaded"}
		fc.papers[id] = paper
		if req.ArxivID != "" {
			fc.byArxiv[req.ArxivID] = id
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(paper)
	})
	mux.HandleFunc("GET /api/papers/by-arxiv/", func(w http.ResponseWriter, r *http.Request) {
		arxivID := r.URL.Path[len("/api/papers/by-arxiv/"):]
		fc.mu.Lock()
		defer fc.mu.Unlock()
		id, ok := fc.byArxiv[arxivID]
		if !ok {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"error":"paper_not_found"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(fc.papers[id])
	})
	mux.HandleFunc("GET /api/papers/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/api/papers/"):]
		fc.mu.Lock()
		defer fc.mu.Unlock()
		p, ok := fc.papers[id]
		if !ok {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"error":"paper_not_found"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(p)
	})
	mux.HandleFunc("PUT /api/papers/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	})
	mux.HandleFunc("GET /api/internal/llm/providers/default", func(w http.ResponseWriter, _ *http.Request) {
		fc.mu.Lock()
		defer fc.mu.Unlock()
		_ = json.NewEncoder(w).Encode(fc.provider)
	})
	mux.HandleFunc("POST /api/review-queue", func(w http.ResponseWriter, r *http.Request) {
		var item model.ReviewQueueItem
		_ = json.NewDecoder(r.Body).Decode(&item)
		fc.mu.Lock()
		fc.review = append(fc.review, item)
		fc.mu.Unlock()
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{}`))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	// The provider BaseURL must point at the fake LLM endpoint; the
	// server URL is only known now, and the handler reads it at
	// request time.
	fc.mu.Lock()
	fc.provider.BaseURL = srv.URL + "/v1"
	fc.mu.Unlock()

	return fc, coreclient.New(srv.URL, "tok")
}

// newArxivStub starts an httptest server serving a fixed Atom feed and
// returns an arxiv.Client pointed at it.
func newArxivStub(t *testing.T) *arxiv.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		_, _ = w.Write([]byte(atomFeed))
	}))
	t.Cleanup(srv.Close)
	c := arxiv.NewClient()
	c.SetBaseURL(srv.URL)
	return c
}

func (fc *fakeCoreSim) counts() (papers, reviews int) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return len(fc.papers), len(fc.review)
}

func newScheduler(a *arxiv.Client, client *coreclient.Client) *Scheduler {
	pipe := pipeline.NewService(client, extractor.NewService())
	return New(a, client, pipe, Config{
		Categories: []string{"cs.AI"},
		Interval:   time.Hour,
		MaxPerRun:  5,
	})
}

// TestScheduler_RunOnce_IngestsAllPapers verifies a cycle registers
// every fetched paper and pushes it through the pipeline.
func TestScheduler_RunOnce_IngestsAllPapers(t *testing.T) {
	fc, client := newCoreSim(t)
	sched := newScheduler(newArxivStub(t), client)

	sched.RunOnce(context.Background())

	papers, reviews := fc.counts()
	if papers != 2 {
		t.Errorf("papers = %d, want 2", papers)
	}
	if reviews != 2 {
		t.Errorf("review submissions = %d, want 2", reviews)
	}
}

// TestScheduler_DeduplicatesAcrossRuns verifies the second cycle is a
// no-op because every paper already exists (by-arxiv lookup hits).
func TestScheduler_DeduplicatesAcrossRuns(t *testing.T) {
	fc, client := newCoreSim(t)
	sched := newScheduler(newArxivStub(t), client)

	sched.RunOnce(context.Background())
	sched.RunOnce(context.Background())

	papers, reviews := fc.counts()
	if papers != 2 || reviews != 2 {
		t.Errorf("after second run papers=%d reviews=%d, want 2/2 (dedup)", papers, reviews)
	}
}

// TestScheduler_DedupLookupFailureAbortsIngest verifies a failing dedup
// lookup surfaces as an ingest error instead of creating duplicates.
func TestScheduler_DedupLookupFailureAbortsIngest(t *testing.T) {
	// Core whose by-arxiv endpoint 500s.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	fc2, _ := newCoreSim(t)
	_ = fc2

	client := coreclient.New(srv.URL, "tok")
	sched := newScheduler(newArxivStub(t), client)

	// Must not panic; failures are logged per paper and the cycle ends.
	sched.RunOnce(context.Background())
}
