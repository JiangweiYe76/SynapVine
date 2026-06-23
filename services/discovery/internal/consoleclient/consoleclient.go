// Package consoleclient provides an HTTP client for the console service.
// Discovery uses it to fetch LLM provider configuration.
package consoleclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"discovery/internal/model"
)

// Client is an HTTP client for the console service REST API.
type Client struct {
	baseURL string
	http    *http.Client
}

// New creates a new console client targeting the given base URL.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Health checks that the console service is reachable.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/llm/providers", nil)
	if err != nil {
		return fmt.Errorf("build health request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("console health check failed: %w", err)
	}
	defer resp.Body.Close()
	// Console requires JWT, so a 401 is also a valid "reachable" signal.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("console health check returned status %d", resp.StatusCode)
	}
	return nil
}

// providerListResponse mirrors the console API response shape.
type providerListResponse struct {
	Providers []model.LLMProvider `json:"providers"`
}

// GetDefaultProvider fetches the default LLM provider from the console service.
// Note: This requires the console to expose an unauthenticated endpoint for
// internal service discovery, or a shared internal token. For the MVP, we
// assume the console's /api/llm/providers/default endpoint is accessible
// internally without JWT.
func (c *Client) GetDefaultProvider(ctx context.Context) (*model.LLMProvider, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/llm/providers/default", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("console request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no default LLM provider configured")
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("console requires authentication; configure an internal token or expose the default provider endpoint")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("console returned status %d: %s", resp.StatusCode, string(raw))
	}

	var provider model.LLMProvider
	if err := json.NewDecoder(resp.Body).Decode(&provider); err != nil {
		return nil, fmt.Errorf("decode console response: %w", err)
	}
	return &provider, nil
}

// GetProviderByID fetches a specific LLM provider by ID.
func (c *Client) GetProviderByID(ctx context.Context, id string) (*model.LLMProvider, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/llm/providers/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("console request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("LLM provider %s not found", id)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("console returned status %d: %s", resp.StatusCode, string(raw))
	}

	var provider model.LLMProvider
	if err := json.NewDecoder(resp.Body).Decode(&provider); err != nil {
		return nil, fmt.Errorf("decode console response: %w", err)
	}
	return &provider, nil
}
