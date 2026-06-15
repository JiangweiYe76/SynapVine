package coreclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchGraphData_Success(t *testing.T) {
	body := `{
		"nodes": [
			{"id": "n1", "name": "Node One", "description": "d",
			 "influence_score": 1.5, "community_id": "comm-1", "degree": 2, "first_appeared": "2020-01"}
		],
		"edges": [
			{"source": "n1", "target": "n2", "weight": 0.8, "relation": "based_on"}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/graph/data" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	c := New(server.URL)
	data, err := c.FetchGraphData(context.Background())
	if err != nil {
		t.Fatalf("FetchGraphData failed: %v", err)
	}
	if len(data.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(data.Nodes))
	}
	if data.Nodes[0].ID != "n1" || data.Nodes[0].Name != "Node One" {
		t.Errorf("unexpected node: %+v", data.Nodes[0])
	}
	if len(data.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(data.Edges))
	}
	if data.Edges[0].Source != "n1" || data.Edges[0].Relation != "based_on" {
		t.Errorf("unexpected edge: %+v", data.Edges[0])
	}
}

func TestFetchGraphData_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"nodes":[], "edges":[]}`))
	}))
	defer server.Close()

	c := New(server.URL)
	data, err := c.FetchGraphData(context.Background())
	if err != nil {
		t.Fatalf("FetchGraphData failed: %v", err)
	}
	if len(data.Nodes) != 0 || len(data.Edges) != 0 {
		t.Errorf("expected empty graph, got %+v", data)
	}
}

func TestFetchGraphData_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer server.Close()

	c := New(server.URL)
	_, err := c.FetchGraphData(context.Background())
	if err == nil {
		t.Fatal("expected error on 500 response, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to mention status 500, got: %v", err)
	}
}

func TestFetchGraphData_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not-json`))
	}))
	defer server.Close()

	c := New(server.URL)
	_, err := c.FetchGraphData(context.Background())
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("expected decode error, got: %v", err)
	}
}

func TestFetchGraphData_ConnectionError(t *testing.T) {
	c := New("http://127.0.0.1:1")
	_, err := c.FetchGraphData(context.Background())
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
}
