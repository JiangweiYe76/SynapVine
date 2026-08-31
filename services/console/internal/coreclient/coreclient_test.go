package coreclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"console/internal/model"
)

func newTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	return New(server.URL, ""), server
}

func TestHealth_OK(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	if err := c.Health(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestHealth_NonOK(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	if err := c.Health(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListNodes_Success(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/nodes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("offset"); got != "0" {
			t.Errorf("expected offset=0, got %s", got)
		}
		if got := r.URL.Query().Get("limit"); got != "20" {
			t.Errorf("expected limit=20, got %s", got)
		}
		if got := r.URL.Query().Get("search"); got != "bert" {
			t.Errorf("expected search=bert, got %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(model.NodesListResponse{
			Nodes: []model.Node{
				{ID: "bert", Name: "BERT", Description: "d", InfluenceScore: 9.6, FirstAppeared: "2018-01"},
			},
			Pagination: model.Pagination{Offset: 0, Limit: 20, Total: 1, HasMore: false},
		})
	})
	defer server.Close()

	resp, err := c.ListNodes(context.Background(), 0, 20, "bert")
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}
	if len(resp.Nodes) != 1 || resp.Nodes[0].ID != "bert" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestListNodes_ServerError(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	})
	defer server.Close()

	_, err := c.ListNodes(context.Background(), 0, 20, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var httpErr *HTTPStatusError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPStatusError, got %T", err)
	}
	if httpErr.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", httpErr.StatusCode)
	}
}

func TestGetNode_Success(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/nodes/bert" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(model.Node{ID: "bert", Name: "BERT"})
	})
	defer server.Close()

	node, err := c.GetNode(context.Background(), "bert")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if node == nil || node.ID != "bert" {
		t.Errorf("unexpected node: %+v", node)
	}
}

func TestGetNode_NotFound(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"node_not_found"}`))
	})
	defer server.Close()

	node, err := c.GetNode(context.Background(), "missing")
	if err != nil {
		t.Fatalf("expected nil error for 404, got %v", err)
	}
	if node != nil {
		t.Errorf("expected nil node, got %+v", node)
	}
}

func TestCreateNode_Success(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/nodes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body model.NodeCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.ID != "vit" {
			t.Errorf("expected id=vit, got %s", body.ID)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(model.Node{ID: "vit", Name: "ViT"})
	})
	defer server.Close()

	node, err := c.CreateNode(context.Background(), model.NodeCreateRequest{
		ID: "vit", Name: "ViT", Description: "Vision", InfluenceScore: 9.1, FirstAppeared: "2020-01",
	})
	if err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}
	if node == nil || node.ID != "vit" {
		t.Errorf("unexpected node: %+v", node)
	}
}

func TestCreateNode_Conflict(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"node_exists"}`))
	})
	defer server.Close()

	_, err := c.CreateNode(context.Background(), model.NodeCreateRequest{ID: "vit", Name: "ViT"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var httpErr *HTTPStatusError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPStatusError, got %T", err)
	}
	if httpErr.StatusCode != 409 {
		t.Errorf("expected status 409, got %d", httpErr.StatusCode)
	}
}

func TestUpdateNode_NotFound(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"node_not_found"}`))
	})
	defer server.Close()

	newName := "X"
	node, err := c.UpdateNode(context.Background(), "missing", model.NodeUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("expected nil error for 404, got %v", err)
	}
	if node != nil {
		t.Errorf("expected nil node, got %+v", node)
	}
}

func TestDeleteNode_Success(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	ok, err := c.DeleteNode(context.Background(), "bert")
	if err != nil {
		t.Fatalf("DeleteNode failed: %v", err)
	}
	if !ok {
		t.Errorf("expected ok=true, got false")
	}
}

func TestDeleteNode_NotFound(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"node_not_found"}`))
	})
	defer server.Close()

	ok, err := c.DeleteNode(context.Background(), "missing")
	if err != nil {
		t.Fatalf("expected nil error for 404, got %v", err)
	}
	if ok {
		t.Errorf("expected ok=false for 404, got true")
	}
}

func TestGraphData_Success(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/graph/data" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(model.GraphData{
			Nodes: []model.Node{{ID: "n1", Name: "N1", Description: "d", InfluenceScore: 1.5, FirstAppeared: "2020-01"}},
			Edges: []model.Edge{{Source: "n1", Target: "n2", Weight: 0.8, Relation: "based_on"}},
		})
	})
	defer server.Close()

	data, err := c.GraphData(context.Background())
	if err != nil {
		t.Fatalf("GraphData failed: %v", err)
	}
	if len(data.Nodes) != 1 || data.Nodes[0].ID != "n1" {
		t.Errorf("unexpected nodes: %+v", data.Nodes)
	}
	if len(data.Edges) != 1 || data.Edges[0].Source != "n1" {
		t.Errorf("unexpected edges: %+v", data.Edges)
	}
}

func TestGraphData_InvalidJSON(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not-json`))
	})
	defer server.Close()

	_, err := c.GraphData(context.Background())
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("expected decode error, got: %v", err)
	}
}

func TestGraphData_ConnectionError(t *testing.T) {
	c := New("http://127.0.0.1:1", "")
	_, err := c.GraphData(context.Background())
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
}

func TestListEdges_Success(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/edges" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("search"); got != "based" {
			t.Errorf("expected search=based, got %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(model.EdgesListResponse{
			Edges: []model.Edge{
				{Source: "a", Target: "b", Weight: 0.4, Relation: "based_on"},
			},
			Pagination: model.Pagination{Offset: 0, Limit: 20, Total: 1, HasMore: false},
		})
	})
	defer server.Close()

	resp, err := c.ListEdges(context.Background(), 0, 20, "based")
	if err != nil {
		t.Fatalf("ListEdges failed: %v", err)
	}
	if len(resp.Edges) != 1 || resp.Edges[0].Source != "a" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetEdge_Success(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/edges/a/b" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(model.Edge{Source: "a", Target: "b", Weight: 0.5, Relation: "r"})
	})
	defer server.Close()

	edge, err := c.GetEdge(context.Background(), "a", "b")
	if err != nil {
		t.Fatalf("GetEdge failed: %v", err)
	}
	if edge == nil || edge.Target != "b" {
		t.Errorf("unexpected edge: %+v", edge)
	}
}

func TestGetEdge_NotFound(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"edge_not_found"}`))
	})
	defer server.Close()

	edge, err := c.GetEdge(context.Background(), "missing", "x")
	if err != nil {
		t.Fatalf("expected nil error for 404, got %v", err)
	}
	if edge != nil {
		t.Errorf("expected nil edge, got %+v", edge)
	}
}

func TestCreateEdge_Conflict(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"edge_exists"}`))
	})
	defer server.Close()

	_, err := c.CreateEdge(context.Background(), model.EdgeCreateRequest{
		Source: "a", Target: "b", Weight: 0.5, Relation: "r",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var httpErr *HTTPStatusError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPStatusError, got %T", err)
	}
	if httpErr.StatusCode != 409 {
		t.Errorf("expected status 409, got %d", httpErr.StatusCode)
	}
}

func TestUpdateEdge_NotFound(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"edge_not_found"}`))
	})
	defer server.Close()

	w := 0.7
	edge, err := c.UpdateEdge(context.Background(), "missing", "x", model.EdgeUpdateRequest{Weight: &w})
	if err != nil {
		t.Fatalf("expected nil error for 404, got %v", err)
	}
	if edge != nil {
		t.Errorf("expected nil edge, got %+v", edge)
	}
}

func TestDeleteEdge_NotFound(t *testing.T) {
	c, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"edge_not_found"}`))
	})
	defer server.Close()

	ok, err := c.DeleteEdge(context.Background(), "missing", "x")
	if err != nil {
		t.Fatalf("expected nil error for 404, got %v", err)
	}
	if ok {
		t.Errorf("expected ok=false for 404, got true")
	}
}
