// Package discoveryclient provides an HTTP client for the discovery service.
// The console service uses it to asynchronously trigger paper analysis
// after a paper is uploaded.
package discoveryclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for the discovery service REST API.
type Client struct {
	baseURL      string
	serviceToken string
	http         *http.Client
}

// New creates a new discovery client targeting the given base URL (e.g.
// "http://localhost:8003"). The serviceToken is presented to discovery
// via the X-Service-Token header; discovery requires the console token
// on /api/analyze because it triggers paid LLM extraction.
// Returns nil if baseURL is empty (auto-trigger disabled).
func New(baseURL, serviceToken string) *Client {
	if baseURL == "" {
		return nil
	}
	return &Client{
		baseURL:      baseURL,
		serviceToken: serviceToken,
		http: &http.Client{
			Timeout: 120 * time.Second, // LLM extraction can take a while
		},
	}
}

// AnalyzeRequest is the payload sent to POST /api/analyze.
type AnalyzeRequest struct {
	PaperID string `json:"paper_id"`
}

// AnalyzeResponse is the response from POST /api/analyze.
type AnalyzeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// TriggerAnalyze sends an asynchronous analysis request to the discovery service.
// It returns immediately after sending the request; the actual LLM extraction
// runs in the background on the discovery service.
func (c *Client) TriggerAnalyze(ctx context.Context, paperID string) error {
	if c == nil {
		return nil // Auto-trigger disabled (no discovery URL configured)
	}

	body, err := json.Marshal(AnalyzeRequest{PaperID: paperID})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/analyze", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.serviceToken != "" {
		req.Header.Set("X-Service-Token", c.serviceToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("discovery request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return newStatusError(resp.StatusCode, resp.Body)
	}

	return nil
}

// maxErrorBodyBytes caps how much of an error response is read back for
// the error message.
const maxErrorBodyBytes = 512

// newStatusError builds an error for a non-2xx response. Discovery's
// error responses carry a JSON {"error","message"} body; surfacing the
// message keeps a 422 such as "paper_text_unusable" legible in the logs
// instead of reducing it to a bare status code.
func newStatusError(status int, body io.Reader) error {
	raw, err := io.ReadAll(io.LimitReader(body, maxErrorBodyBytes))
	if err != nil || len(bytes.TrimSpace(raw)) == 0 {
		return fmt.Errorf("discovery returned status %d", status)
	}

	var payload struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &payload); err == nil && payload.Message != "" {
		if payload.Error != "" {
			return fmt.Errorf("discovery returned status %d (%s): %s", status, payload.Error, payload.Message)
		}
		return fmt.Errorf("discovery returned status %d: %s", status, payload.Message)
	}
	return fmt.Errorf("discovery returned status %d: %s", status, bytes.TrimSpace(raw))
}

// Health checks that the discovery service is reachable.
func (c *Client) Health(ctx context.Context) error {
	if c == nil {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("build health request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("discovery health check failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discovery health check returned status %d", resp.StatusCode)
	}
	return nil
}
