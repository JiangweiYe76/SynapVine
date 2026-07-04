package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"core/internal/model"
	"core/internal/service"

	"github.com/gofiber/fiber/v2"
)

// edgeSvcStub is a stub EdgeService used by the EdgeHandler tests.
// It records the most recent call so individual tests can assert on it.
type edgeSvcStub struct {
	listResp   *model.EdgesListResponse
	listErr    error
	listByIDsResp []model.Edge
	listByIDsErr  error
	getResp    *model.Edge
	getErr     error
	createResp *model.Edge
	createErr  error
	updateResp *model.Edge
	updateErr  error
	deleteErr  error

	lastCreate model.EdgeCreateRequest
	lastUpdate model.EdgeUpdateRequest
}

func (s *edgeSvcStub) List(_ context.Context, _ int, _ int, _ string) (*model.EdgesListResponse, error) {
	return s.listResp, s.listErr
}

func (s *edgeSvcStub) ListByNodeIDs(_ context.Context, _ []string) ([]model.Edge, error) {
	return s.listByIDsResp, s.listByIDsErr
}

func (s *edgeSvcStub) Get(_ context.Context, _, _ string) (*model.Edge, error) {
	return s.getResp, s.getErr
}

func (s *edgeSvcStub) Create(_ context.Context, req model.EdgeCreateRequest) (*model.Edge, error) {
	s.lastCreate = req
	return s.createResp, s.createErr
}

func (s *edgeSvcStub) Update(_ context.Context, _, _ string, req model.EdgeUpdateRequest) (*model.Edge, error) {
	s.lastUpdate = req
	return s.updateResp, s.updateErr
}

func (s *edgeSvcStub) Delete(_ context.Context, _, _ string) error {
	return s.deleteErr
}

// Compile-time check that edgeSvcStub satisfies EdgeService.
var _ EdgeService = (*edgeSvcStub)(nil)

func newEdgeApp(svc EdgeService) *fiber.App {
	app := fiber.New()
	h := NewEdgeHandler(svc)
	app.Get("/api/edges", h.List)
	app.Get("/api/edges/:source/:target", h.Get)
	app.Post("/api/edges", h.Create)
	app.Put("/api/edges/:source/:target", h.Update)
	app.Delete("/api/edges/:source/:target", h.Delete)
	return app
}

func TestEdgeHandler_List_Success(t *testing.T) {
	svc := &edgeSvcStub{
		listResp: &model.EdgesListResponse{
			Edges: []model.Edge{
				{Source: "a", Target: "b", Weight: 0.5, Relation: "based_on"},
			},
			Pagination: model.Pagination{Offset: 0, Limit: 20, Total: 1, HasMore: false},
		},
	}
	app := newEdgeApp(svc)

	req := httptest.NewRequest("GET", "/api/edges?search=base", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var got model.EdgesListResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode failed: %v\nbody: %s", err, string(body))
	}
	if len(got.Edges) != 1 || got.Edges[0].Source != "a" {
		t.Errorf("unexpected edges: %+v", got.Edges)
	}
	if got.Pagination.Total != 1 {
		t.Errorf("unexpected pagination: %+v", got.Pagination)
	}
}

func TestEdgeHandler_List_ServiceError(t *testing.T) {
	svc := &edgeSvcStub{listErr: errors.New("neo4j down")}
	app := newEdgeApp(svc)

	req := httptest.NewRequest("GET", "/api/edges", nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 500 {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestEdgeHandler_Get_Success(t *testing.T) {
	svc := &edgeSvcStub{
		getResp: &model.Edge{Source: "a", Target: "b", Weight: 0.7, Relation: "evolved_to"},
	}
	app := newEdgeApp(svc)

	req := httptest.NewRequest("GET", "/api/edges/a/b", nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var got model.Edge
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if got.Source != "a" || got.Target != "b" {
		t.Errorf("unexpected edge: %+v", got)
	}
}

func TestEdgeHandler_Get_NotFound(t *testing.T) {
	svc := &edgeSvcStub{getResp: nil}
	app := newEdgeApp(svc)

	req := httptest.NewRequest("GET", "/api/edges/missing/x", nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestEdgeHandler_Create_Success(t *testing.T) {
	svc := &edgeSvcStub{
		createResp: &model.Edge{Source: "a", Target: "b", Weight: 0.3, Relation: "uses"},
	}
	app := newEdgeApp(svc)

	body, _ := json.Marshal(model.EdgeCreateRequest{
		Source: "a", Target: "b", Weight: 0.3, Relation: "uses",
	})
	req := httptest.NewRequest("POST", "/api/edges", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if svc.lastCreate.Source != "a" || svc.lastCreate.Weight != 0.3 {
		t.Errorf("svc did not receive expected create payload: %+v", svc.lastCreate)
	}
}

func TestEdgeHandler_Create_BadJSON(t *testing.T) {
	svc := &edgeSvcStub{}
	app := newEdgeApp(svc)

	req := httptest.NewRequest("POST", "/api/edges", bytes.NewReader([]byte(`{not-json`)))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestEdgeHandler_Create_Conflict(t *testing.T) {
	svc := &edgeSvcStub{createErr: service.ErrEdgeExists}
	app := newEdgeApp(svc)

	body, _ := json.Marshal(model.EdgeCreateRequest{Source: "a", Target: "b", Weight: 0.3, Relation: "uses"})
	req := httptest.NewRequest("POST", "/api/edges", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 409 {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestEdgeHandler_Create_NodeNotFound(t *testing.T) {
	svc := &edgeSvcStub{createErr: service.ErrNodeNotFound}
	app := newEdgeApp(svc)

	body, _ := json.Marshal(model.EdgeCreateRequest{Source: "ghost", Target: "b", Weight: 0.3, Relation: "uses"})
	req := httptest.NewRequest("POST", "/api/edges", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestEdgeHandler_Create_InvalidWeight(t *testing.T) {
	svc := &edgeSvcStub{createErr: service.ErrInvalidWeight}
	app := newEdgeApp(svc)

	body, _ := json.Marshal(model.EdgeCreateRequest{Source: "a", Target: "b", Weight: 5, Relation: "uses"})
	req := httptest.NewRequest("POST", "/api/edges", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestEdgeHandler_Create_SameEndpoints(t *testing.T) {
	svc := &edgeSvcStub{createErr: service.ErrSameEndpoints}
	app := newEdgeApp(svc)

	body, _ := json.Marshal(model.EdgeCreateRequest{Source: "a", Target: "a", Weight: 0.3, Relation: "uses"})
	req := httptest.NewRequest("POST", "/api/edges", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestEdgeHandler_Update_Success(t *testing.T) {
	svc := &edgeSvcStub{
		updateResp: &model.Edge{Source: "a", Target: "b", Weight: 0.9, Relation: "based_on"},
	}
	app := newEdgeApp(svc)

	newWeight := 0.9
	body, _ := json.Marshal(model.EdgeUpdateRequest{Weight: &newWeight})
	req := httptest.NewRequest("PUT", "/api/edges/a/b", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if svc.lastUpdate.Weight == nil || *svc.lastUpdate.Weight != 0.9 {
		t.Errorf("svc did not receive expected update payload: %+v", svc.lastUpdate)
	}
}

func TestEdgeHandler_Update_NotFound(t *testing.T) {
	svc := &edgeSvcStub{updateResp: nil}
	app := newEdgeApp(svc)

	body, _ := json.Marshal(model.EdgeUpdateRequest{})
	req := httptest.NewRequest("PUT", "/api/edges/missing/x", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestEdgeHandler_Update_InvalidWeight(t *testing.T) {
	svc := &edgeSvcStub{updateErr: service.ErrInvalidWeight}
	app := newEdgeApp(svc)

	bad := 2.0
	body, _ := json.Marshal(model.EdgeUpdateRequest{Weight: &bad})
	req := httptest.NewRequest("PUT", "/api/edges/a/b", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestEdgeHandler_Delete_Success(t *testing.T) {
	svc := &edgeSvcStub{}
	app := newEdgeApp(svc)

	req := httptest.NewRequest("DELETE", "/api/edges/a/b", nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestEdgeHandler_Delete_NotFound(t *testing.T) {
	svc := &edgeSvcStub{deleteErr: service.ErrEdgeNotFound}
	app := newEdgeApp(svc)

	req := httptest.NewRequest("DELETE", "/api/edges/missing/x", nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}
