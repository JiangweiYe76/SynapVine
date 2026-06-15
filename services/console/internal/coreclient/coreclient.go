// Package coreclient provides an HTTP client for the core service.
//
// The console service treats the core service (which fronts Neo4j) as the
// authoritative source of truth for nodes, edges, communities, and
// statistics.
package coreclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"console/internal/model"
)

// Client is an HTTP client for the core service REST API.
type Client struct {
	baseURL string
	http    *http.Client
}

// New creates a new core client targeting the given base URL (e.g. "http://localhost:8001").
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Health checks that the core service is reachable by issuing a GET /health
// request. It returns nil on a 200 response and an error otherwise.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to build health request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("core health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("core health check returned status %d", resp.StatusCode)
	}
	return nil
}

// ListNodes returns a paginated list of nodes from the core service.
func (c *Client) ListNodes(ctx context.Context, offset, limit int, search string) (*model.NodesListResponse, error) {
	params := url.Values{}
	params.Set("offset", fmt.Sprintf("%d", offset))
	params.Set("limit", fmt.Sprintf("%d", limit))
	if search != "" {
		params.Set("search", search)
	}
	var resp model.NodesListResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/nodes?"+params.Encode(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetNode fetches a single node by ID. It returns (nil, nil) when the core
// responds with 404, so callers can distinguish "not found" from real errors.
func (c *Client) GetNode(ctx context.Context, id string) (*model.Node, error) {
	var node model.Node
	err := c.doJSON(ctx, http.MethodGet, "/api/nodes/"+id, nil, &node)
	if err != nil {
		var httpErr *HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &node, nil
}

// CreateNode creates a new node and returns the created resource.
func (c *Client) CreateNode(ctx context.Context, req model.NodeCreateRequest) (*model.Node, error) {
	var node model.Node
	if err := c.doJSON(ctx, http.MethodPost, "/api/nodes", req, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

// UpdateNode updates an existing node and returns the updated resource.
// It returns (nil, nil) when the core responds with 404.
func (c *Client) UpdateNode(ctx context.Context, id string, req model.NodeUpdateRequest) (*model.Node, error) {
	var node model.Node
	err := c.doJSON(ctx, http.MethodPut, "/api/nodes/"+id, req, &node)
	if err != nil {
		var httpErr *HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &node, nil
}

// DeleteNode removes a node by ID. It returns (false, nil) when the core
// responds with 404.
func (c *Client) DeleteNode(ctx context.Context, id string) (bool, error) {
	err := c.doJSON(ctx, http.MethodDelete, "/api/nodes/"+id, nil, nil)
	if err != nil {
		var httpErr *HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListEdges returns a paginated list of edges from the core service.
func (c *Client) ListEdges(ctx context.Context, offset, limit int, search string) (*model.EdgesListResponse, error) {
	params := url.Values{}
	params.Set("offset", fmt.Sprintf("%d", offset))
	params.Set("limit", fmt.Sprintf("%d", limit))
	if search != "" {
		params.Set("search", search)
	}
	var resp model.EdgesListResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/edges?"+params.Encode(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetEdge fetches a single edge by its (source, target) pair. It returns
// (nil, nil) when the core responds with 404.
func (c *Client) GetEdge(ctx context.Context, source, target string) (*model.Edge, error) {
	var edge model.Edge
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/edges/%s/%s", source, target), nil, &edge)
	if err != nil {
		var httpErr *HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &edge, nil
}

// CreateEdge creates a new edge and returns the created resource.
func (c *Client) CreateEdge(ctx context.Context, req model.EdgeCreateRequest) (*model.Edge, error) {
	var edge model.Edge
	if err := c.doJSON(ctx, http.MethodPost, "/api/edges", req, &edge); err != nil {
		return nil, err
	}
	return &edge, nil
}

// UpdateEdge updates an existing edge and returns the updated resource.
// It returns (nil, nil) when the core responds with 404.
func (c *Client) UpdateEdge(ctx context.Context, source, target string, req model.EdgeUpdateRequest) (*model.Edge, error) {
	var edge model.Edge
	err := c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/edges/%s/%s", source, target), req, &edge)
	if err != nil {
		var httpErr *HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &edge, nil
}

// DeleteEdge removes an edge by its (source, target) pair. It returns
// (false, nil) when the core responds with 404.
func (c *Client) DeleteEdge(ctx context.Context, source, target string) (bool, error) {
	err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/edges/%s/%s", source, target), nil, nil)
	if err != nil {
		var httpErr *HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GraphData fetches the full graph (nodes + edges) from the core service.
// It is used to compute statistics because core has no dedicated /api/stats
// endpoint.
func (c *Client) GraphData(ctx context.Context) (*model.GraphData, error) {
	var data model.GraphData
	if err := c.doJSON(ctx, http.MethodGet, "/api/graph/data", nil, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// ListCommunities returns the flat list of communities from the core service.
func (c *Client) ListCommunities(ctx context.Context) (*model.CommunitiesListResponse, error) {
	var resp model.CommunitiesListResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/communities", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCommunityTree returns the communities assembled as a tree.
func (c *Client) GetCommunityTree(ctx context.Context) (*model.CommunitiesTreeResponse, error) {
	var resp model.CommunitiesTreeResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/communities/tree", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCommunity fetches a single community by ID.
// It returns (nil, nil) when core responds with 404.
func (c *Client) GetCommunity(ctx context.Context, id string) (*model.Community, error) {
	var comm model.Community
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/communities/%s", id), nil, &comm)
	if err != nil {
		var httpErr *HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &comm, nil
}

// CreateCommunity creates a new community and returns the created resource.
func (c *Client) CreateCommunity(ctx context.Context, req model.CommunityCreateRequest) (*model.Community, error) {
	var comm model.Community
	if err := c.doJSON(ctx, http.MethodPost, "/api/communities", req, &comm); err != nil {
		return nil, err
	}
	return &comm, nil
}

// UpdateCommunity updates an existing community and returns the updated resource.
// It returns (nil, nil) when core responds with 404.
func (c *Client) UpdateCommunity(ctx context.Context, id string, req model.CommunityUpdateRequest) (*model.Community, error) {
	var comm model.Community
	err := c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/communities/%s", id), req, &comm)
	if err != nil {
		var httpErr *HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &comm, nil
}

// DeleteCommunity removes a community by ID.
// It returns (false, nil) when core responds with 404.
func (c *Client) DeleteCommunity(ctx context.Context, id string) (bool, error) {
	err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/communities/%s", id), nil, nil)
	if err != nil {
		var httpErr *HTTPStatusError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// --- HTTP helpers ---

// HTTPStatusError carries the HTTP status code returned by the core service
// alongside the underlying error message. Handlers can use errors.As to
// distinguish transient failures from semantically meaningful statuses
// such as 404 or 409.
type HTTPStatusError struct {
	StatusCode int
	Body       string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("core returned status %d: %s", e.StatusCode, e.Body)
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	_, err := c.doJSONStatus(ctx, method, path, body, out)
	return err
}

func (c *Client) doJSONStatus(ctx context.Context, method, path string, body any, out any) (int, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, fmt.Errorf("failed to encode request body: %w", err)
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, fmt.Errorf("failed to build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("core request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, &HTTPStatusError{StatusCode: resp.StatusCode, Body: string(raw)}
	}

	if out == nil {
		return resp.StatusCode, nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return resp.StatusCode, fmt.Errorf("failed to decode core response: %w", err)
	}
	return resp.StatusCode, nil
}
