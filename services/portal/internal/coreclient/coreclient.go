package coreclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"ai-graph-server/internal/model"
)

// Client provides read-only access to the core service.
type Client struct {
	baseURL string
	client  *http.Client
}

// New creates a new core client.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CoreNode is the node shape returned by the core service. The core service
// stores community identifiers as string UUIDs, while the portal model uses
// integer community IDs for the frontend, so a conversion step is required.
type CoreNode struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	InfluenceScore float64  `json:"influence_score"`
	FirstAppeared  string   `json:"first_appeared"`
	Milestones     []string `json:"milestones,omitempty"`
	CommunityID    *string  `json:"community_id,omitempty"`
	Degree         int      `json:"degree"`
}

// CoreGraphData is the raw graph payload returned by the core service.
type CoreGraphData struct {
	Nodes []CoreNode   `json:"nodes"`
	Edges []model.Edge `json:"edges"`
}

// CoreCommunity is the community shape returned by the core service.
type CoreCommunity struct {
	ID        string          `json:"id"`
	ParentID  *string         `json:"parent_id,omitempty"`
	Name      string          `json:"name"`
	Color     string          `json:"color"`
	Level     int             `json:"level"`
	Domain    string          `json:"domain"`
	NodeCount int             `json:"node_count"`
	Children  []CoreCommunity `json:"children,omitempty"`
}

// CoreCommunitiesResponse is the wrapper used by the core /api/communities/tree endpoint.
type CoreCommunitiesResponse struct {
	Communities []CoreCommunity `json:"communities"`
}

// CorePagination matches the pagination block in core list responses.
type CorePagination struct {
	Offset  int  `json:"offset"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	HasMore bool `json:"has_more"`
}

// CoreNodesResponse matches the core /api/nodes response shape.
type CoreNodesResponse struct {
	Nodes      []CoreNode     `json:"nodes"`
	Pagination CorePagination `json:"pagination"`
}

// CoreEdgesListResponse matches the core /api/edges response shape.
type CoreEdgesListResponse struct {
	Edges      []model.Edge   `json:"edges"`
	Pagination CorePagination `json:"pagination"`
}

// CoreTimelineRange matches the core /api/graph/timeline response shape.
type CoreTimelineRange struct {
	MinYear int `json:"min_year"`
	MaxYear int `json:"max_year"`
}

// FetchGraphData retrieves the full graph (nodes + edges) from the core service.
// The returned nodes use the core's string community identifiers; callers must
// map them to portal integer IDs using the community tree from FetchCommunityTree.
func (c *Client) FetchGraphData(ctx context.Context) (*CoreGraphData, error) {
	resp, err := c.do(ctx, "GET", c.baseURL+"/api/graph/data", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to core: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("core returned status %d", resp.StatusCode)
	}

	var payload CoreGraphData
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode core response: %w", err)
	}

	return &payload, nil
}

// FetchCommunityTree retrieves the hierarchical community tree from the core service.
func (c *Client) FetchCommunityTree(ctx context.Context) ([]CoreCommunity, error) {
	resp, err := c.do(ctx, "GET", c.baseURL+"/api/communities/tree", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to core: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("core returned status %d", resp.StatusCode)
	}

	var payload CoreCommunitiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode core response: %w", err)
	}

	return payload.Communities, nil
}

// ListNodes proxies a paginated node listing request to the core service.
// An empty search value returns all nodes within the requested page.
func (c *Client) ListNodes(ctx context.Context, offset, limit int, search string) (*CoreNodesResponse, error) {
	v := url.Values{}
	v.Set("offset", strconv.Itoa(offset))
	v.Set("limit", strconv.Itoa(limit))
	if search != "" {
		v.Set("search", search)
	}
	resp, err := c.do(ctx, "GET", c.baseURL+"/api/nodes?"+v.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to core: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("core returned status %d", resp.StatusCode)
	}

	var payload CoreNodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode core response: %w", err)
	}

	return &payload, nil
}

// GetNode fetches a single node by id from the core service.
// Returns (nil, nil) when the node does not exist (404).
func (c *Client) GetNode(ctx context.Context, id string) (*CoreNode, error) {
	resp, err := c.do(ctx, "GET", c.baseURL+"/api/nodes/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to core: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("core returned status %d", resp.StatusCode)
	}

	var payload CoreNode
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode core response: %w", err)
	}

	return &payload, nil
}

// ListEdges proxies a paginated edge listing request to the core service.
// The portal always requests the full edge set; the dev graph is small.
func (c *Client) ListEdges(ctx context.Context, offset, limit int) (*CoreEdgesListResponse, error) {
	v := url.Values{}
	v.Set("offset", strconv.Itoa(offset))
	v.Set("limit", strconv.Itoa(limit))
	resp, err := c.do(ctx, "GET", c.baseURL+"/api/edges?"+v.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to core: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("core returned status %d", resp.StatusCode)
	}

	var payload CoreEdgesListResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode core response: %w", err)
	}

	return &payload, nil
}

// FetchTimelineRange retrieves the [minYear, maxYear] span of every
// node's `first_appeared` field from the core service. Core computes the
// range in Cypher over the full graph, so the result is independent of
// which nodes the caller has loaded.
func (c *Client) FetchTimelineRange(ctx context.Context) (*CoreTimelineRange, error) {
	resp, err := c.do(ctx, "GET", c.baseURL+"/api/graph/timeline", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to core: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("core returned status %d", resp.StatusCode)
	}

	var payload CoreTimelineRange
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode core response: %w", err)
	}

	return &payload, nil
}

func (c *Client) do(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}
