// Package llm provides an OpenAI-compatible HTTP client for calling LLM
// providers. This is a standalone copy of the console's llm client,
// keeping the discovery service self-contained.
package llm

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

// Message represents a single chat message in OpenAI format.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest is the payload sent to the LLM provider.
type CompletionRequest struct {
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	JSONMode    bool      `json:"json_mode,omitempty"`
}

// CompletionResponse is the parsed LLM response.
type CompletionResponse struct {
	Content      string `json:"content"`
	TokensUsed   int    `json:"tokens_used"`
	FinishReason string `json:"finish_reason"`
}

// Client calls a single OpenAI-compatible LLM provider.
type Client struct {
	httpClient  *http.Client
	baseURL     string
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
}

// NewClient creates a Client from an LLMProvider model.
func NewClient(p *model.LLMProvider) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		baseURL:     p.BaseURL,
		apiKey:      p.APIKey,
		model:       p.Model,
		maxTokens:   p.MaxTokens,
		temperature: p.Temperature,
	}
}

type wireFormat struct {
	Model       string          `json:"model"`
	Messages    []Message       `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Format      *responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type wireResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Complete sends a chat completion request and returns the response.
func (c *Client) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	body := wireFormat{
		Model:       c.model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
	}
	if body.Temperature == 0 {
		body.Temperature = c.temperature
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = c.maxTokens
	}
	body.MaxTokens = maxTokens

	if req.JSONMode {
		body.Format = &responseFormat{Type: "json_object"}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL
	if url != "" && url[len(url)-1] != '/' {
		url += "/"
	}
	url += "chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llm provider returned %d: %s", httpResp.StatusCode, string(respBody))
	}

	var wire wireResponse
	if err := json.Unmarshal(respBody, &wire); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if wire.Error != nil {
		return nil, fmt.Errorf("llm error: %s", wire.Error.Message)
	}

	if len(wire.Choices) == 0 {
		return nil, fmt.Errorf("llm returned no choices")
	}

	return &CompletionResponse{
		Content:      wire.Choices[0].Message.Content,
		TokensUsed:   wire.Usage.TotalTokens,
		FinishReason: wire.Choices[0].FinishReason,
	}, nil
}
