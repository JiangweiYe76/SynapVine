package coreclient

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	Nodes []CoreNode `json:"nodes"`
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

// FetchGraphData retrieves the full graph (nodes + edges) from the core service.
// The returned nodes use the core's string community identifiers; callers must
// map them to portal integer IDs using the community tree from FetchCommunityTree.
func (c *Client) FetchGraphData() (*CoreGraphData, error) {
	url := c.baseURL + "/api/graph/data"
	resp, err := c.client.Get(url)
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
func (c *Client) FetchCommunityTree() ([]CoreCommunity, error) {
	url := c.baseURL + "/api/communities/tree"
	resp, err := c.client.Get(url)
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
