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

// FetchGraphData retrieves the full graph (nodes + edges) from the core service.
func (c *Client) FetchGraphData() (*model.GraphData, error) {
	url := c.baseURL + "/api/graph/data"
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to core: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("core returned status %d", resp.StatusCode)
	}

	var payload struct {
		Nodes []model.Node `json:"nodes"`
		Edges []model.Edge `json:"edges"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode core response: %w", err)
	}

	return &model.GraphData{
		Nodes: payload.Nodes,
		Edges: payload.Edges,
	}, nil
}
