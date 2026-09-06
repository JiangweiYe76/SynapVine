package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"ai-graph-server/internal/coreclient"
)

// fakeCore counts hits per endpoint so tests can assert that cached
// reads do not hit the core service again.
func fakeCore(t *testing.T, treeHits, dataHits *atomic.Int32) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/communities/tree", func(w http.ResponseWriter, r *http.Request) {
		treeHits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(coreclient.CoreCommunitiesResponse{
			Communities: []coreclient.CoreCommunity{
				{ID: "c1", Name: "Root", Color: "#fff", Level: 0, NodeCount: 1},
			},
		})
	})
	mux.HandleFunc("/api/graph/data", func(w http.ResponseWriter, r *http.Request) {
		dataHits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(coreclient.CoreGraphData{
			Nodes: []coreclient.CoreNode{
				{ID: "n1", Name: "Transformer", CommunityID: strPtr("c1")},
			},
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func strPtr(s string) *string { return &s }

func TestSummaryCachesCommunityTree(t *testing.T) {
	var treeHits, dataHits atomic.Int32
	srv := fakeCore(t, &treeHits, &dataHits)
	svc := New(coreclient.New(srv.URL, ""))
	ctx := context.Background()

	if _, err := svc.Summary(ctx, 10); err != nil {
		t.Fatalf("first Summary failed: %v", err)
	}
	if _, err := svc.Summary(ctx, 10); err != nil {
		t.Fatalf("second Summary failed: %v", err)
	}

	if got := treeHits.Load(); got != 1 {
		t.Errorf("community tree fetched %d times, want 1 (cache miss on second call)", got)
	}
	if got := dataHits.Load(); got != 1 {
		t.Errorf("graph data fetched %d times, want 1", got)
	}
}

func TestNodesCachesCommunityTree(t *testing.T) {
	var treeHits, dataHits atomic.Int32
	srv := fakeCore(t, &treeHits, &dataHits)
	svc := New(coreclient.New(srv.URL, ""))
	ctx := context.Background()

	if _, err := svc.Nodes(ctx, 0, 10, "", "", nil); err != nil {
		t.Fatalf("first Nodes failed: %v", err)
	}
	if _, err := svc.Nodes(ctx, 0, 10, "", "", nil); err != nil {
		t.Fatalf("second Nodes failed: %v", err)
	}

	if got := treeHits.Load(); got != 1 {
		t.Errorf("community tree fetched %d times, want 1", got)
	}
}
