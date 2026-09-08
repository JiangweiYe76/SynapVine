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
	"net/url"
	"time"

	"discovery/internal/model"
)

// Client is an HTTP client for the core service REST API.
type Client struct {
	baseURL      string
	serviceToken string
	http         *http.Client
}

// CreatePaperRequest is the payload for creating a paper in core.
type CreatePaperRequest struct {
	Title     string `json:"title"`
	Authors   string `json:"authors"`
	SourceURL string `json:"source_url"`
	ArxivID   string `json:"arxiv_id,omitempty"`
	RawText   string `json:"raw_text"`
}

// New creates a new core client targeting the given base URL. The
// serviceToken is presented to core via the X-Service-Token header on
// every request; it identifies discovery as the caller and grants
// write-tier and internal-tier access.
func New(baseURL, serviceToken string) *Client {
	return &Client{
		baseURL:      baseURL,
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
	if err := c.doJSON(ctx, http.MethodGet, "/api/papers/"+url.PathEscape(id), nil, &paper); err != nil {
		return nil, err
	}
	return &paper, nil
}

// GetPaperByArxivID looks up a paper by its ArXiv identifier. The
// second return value is false when core reports paper_not_found (404).
func (c *Client) GetPaperByArxivID(ctx context.Context, arxivID string) (*model.Paper, bool, error) {
	var paper model.Paper
	found, err := c.doJSONAllow404(ctx, http.MethodGet, "/api/papers/by-arxiv/"+url.PathEscape(arxivID), nil, &paper)
	if err != nil || !found {
		return nil, false, err
	}
	return &paper, true, nil
}

// CreatePaper creates a paper record in core and returns the stored
// paper (including its generated ID).
func (c *Client) CreatePaper(ctx context.Context, req CreatePaperRequest) (*model.Paper, error) {
	var paper model.Paper
	if err := c.doJSON(ctx, http.MethodPost, "/api/papers", req, &paper); err != nil {
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
	found, err := c.doJSONAllow404(ctx, method, path, body, out)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("core returned 404 for %s %s", method, path)
	}
	return nil
}

// doJSONAllow404 performs a request like doJSON, but a 404 response is
// not an error: it reports found=false and leaves out untouched.
func (c *Client) doJSONAllow404(ctx context.Context, method, path string, body any, out any) (bool, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return false, fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return false, fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.serviceToken != "" {
		req.Header.Set("X-Service-Token", c.serviceToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return false, fmt.Errorf("core request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("core returned status %d: %s", resp.StatusCode, string(raw))
	}

	if out == nil {
		return true, nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return false, fmt.Errorf("decode core response: %w", err)
	}
	return true, nil
}
