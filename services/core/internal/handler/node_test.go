package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"core/internal/model"

	"github.com/gofiber/fiber/v2"
)

// graphDataStub is a stub NodeService used by the GraphData tests.
type graphDataStub struct {
	nodes []model.Node
	edges []model.Edge
	err   error
}

func (s *graphDataStub) List(_ context.Context, offset, limit int, search string) (*model.NodesListResponse, error) {
	return nil, nil
}

func (s *graphDataStub) Get(_ context.Context, id string) (*model.Node, error) {
	return nil, nil
}

func (s *graphDataStub) Create(_ context.Context, req model.NodeCreateRequest) (*model.Node, error) {
	return nil, nil
}

func (s *graphDataStub) Update(_ context.Context, id string, req model.NodeUpdateRequest) (*model.Node, error) {
	return nil, nil
}

func (s *graphDataStub) Delete(_ context.Context, id string) error { return nil }

func (s *graphDataStub) GetAll(_ context.Context) ([]model.Node, []model.Edge, error) {
	if s.err != nil {
		return nil, nil, s.err
	}
	return s.nodes, s.edges, nil
}

// Compile-time check that graphDataStub satisfies NodeService.
var _ NodeService = (*graphDataStub)(nil)

func newGraphDataApp(svc NodeService) *fiber.App {
	app := fiber.New()
	h := NewNodeHandler(svc)
	app.Get("/api/graph/data", h.GraphData)
	return app
}

func TestNodeHandler_GraphData_Success(t *testing.T) {
	svc := &graphDataStub{
		nodes: []model.Node{
			{ID: "n1", Name: "Alpha", Category: "cat", Description: "d", InfluenceScore: 8.5, FirstAppeared: 2020},
			{ID: "n2", Name: "Beta", Category: "cat", Description: "d", InfluenceScore: 7.0, FirstAppeared: 2021},
		},
		edges: []model.Edge{
			{Source: "n1", Target: "n2", Weight: 0.9, Relation: "based_on"},
		},
	}

	app := newGraphDataApp(svc)
	req := httptest.NewRequest("GET", "/api/graph/data", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}

	var payload struct {
		Nodes []model.Node `json:"nodes"`
		Edges []model.Edge `json:"edges"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response failed: %v\nbody: %s", err, string(body))
	}

	if len(payload.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(payload.Nodes))
	}
	if payload.Nodes[0].ID != "n1" || payload.Nodes[0].Name != "Alpha" {
		t.Errorf("unexpected first node: %+v", payload.Nodes[0])
	}
	if len(payload.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(payload.Edges))
	}
	if payload.Edges[0].Source != "n1" || payload.Edges[0].Relation != "based_on" {
		t.Errorf("unexpected edge: %+v", payload.Edges[0])
	}
}

func TestNodeHandler_GraphData_Empty(t *testing.T) {
	svc := &graphDataStub{}

	app := newGraphDataApp(svc)
	req := httptest.NewRequest("GET", "/api/graph/data", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var payload struct {
		Nodes []model.Node `json:"nodes"`
		Edges []model.Edge `json:"edges"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response failed: %v\nbody: %s", err, string(body))
	}
	if len(payload.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(payload.Nodes))
	}
	if len(payload.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(payload.Edges))
	}
}

func TestNodeHandler_GraphData_ServiceError(t *testing.T) {
	svc := &graphDataStub{err: errors.New("neo4j unavailable")}

	app := newGraphDataApp(svc)
	req := httptest.NewRequest("GET", "/api/graph/data", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 500 {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var er model.ErrorResponse
	if err := json.Unmarshal(body, &er); err != nil {
		t.Fatalf("decode error response failed: %v\nbody: %s", err, string(body))
	}
	if er.Error != "internal_error" {
		t.Errorf("error code = %q, want internal_error", er.Error)
	}
	if er.Message == "" {
		t.Error("expected non-empty error message")
	}
}
