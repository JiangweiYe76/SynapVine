// Package coreclient provides an HTTP client for the core service.
// Discovery uses it to fetch papers and submit review queue items.
package coreclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"discovery/internal/model"
)

// Client is an HTTP client for the core service REST API.
type Client struct {
	baseURL     string
	serviceToken string
	http        *http.Client
}

// New creates a new core client targeting the given base URL. The
// serviceToken is presented to core via the X-Service-Token header on
// every request; it identifies discovery as the caller and grants
// write-tier and internal-tier access.
func New(baseURL, serviceToken string) *Client {
	return &Client{
		baseURL:     baseURL,
		serviceToken: serviceToken,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Health checks that the core service is reachable.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("build health request: %w", err)
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

// GetPaper fetches a paper by ID from the core service.
func (c *Client) GetPaper(ctx context.Context, id string) (*model.Paper, error) {
	var paper model.Paper
	if err := c.doJSON(ctx, http.MethodGet, "/api/papers/"+id, nil, &paper); err != nil {
		return nil, err
	}
	return &paper, nil
}

// SubmitReviewItem submits an extraction result to core's review queue.
func (c *Client) SubmitReviewItem(ctx context.Context, item model.ReviewQueueItem) error {
	return c.doJSON(ctx, http.MethodPost, "/api/review-queue", item, nil)
}

// UpdatePaperStatus updates the status of a paper.
func (c *Client) UpdatePaperStatus(ctx context.Context, id, status string) error {
	body := map[string]string{"status": status}
	return c.doJSON(ctx, http.MethodPut, "/api/papers/"+id, body, nil)
}

// GetDefaultLLMProvider fetches the default LLM provider (including API key) from core.
func (c *Client) GetDefaultLLMProvider(ctx context.Context) (*model.LLMProvider, error) {
	var provider model.LLMProvider
	if err := c.doJSON(ctx, http.MethodGet, "/api/internal/llm/providers/default", nil, &provider); err != nil {
		return nil, err
	}
	return &provider, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.serviceToken != "" {
		req.Header.Set("X-Service-Token", c.serviceToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("core request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("core returned status %d: %s", resp.StatusCode, string(raw))
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode core response: %w", err)
	}
	return nil
}
