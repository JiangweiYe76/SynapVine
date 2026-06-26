package model

import "time"

// LLMProviderResponse is the safe representation returned in API responses.
type LLMProviderResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	BaseURL     string    `json:"base_url"`
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	IsDefault   bool      `json:"is_default"`
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LLMProviderCreateRequest is the payload for creating a new provider.
type LLMProviderCreateRequest struct {
	Name        string  `json:"name"`
	BaseURL     string  `json:"base_url"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	IsDefault   bool    `json:"is_default"`
}

// LLMProviderUpdateRequest is the payload for updating an existing provider.
type LLMProviderUpdateRequest struct {
	Name        *string  `json:"name"`
	BaseURL     *string  `json:"base_url"`
	APIKey      *string  `json:"api_key"`
	Model       *string  `json:"model"`
	MaxTokens   *int     `json:"max_tokens"`
	Temperature *float64 `json:"temperature"`
	IsDefault   *bool    `json:"is_default"`
	IsEnabled   *bool    `json:"is_enabled"`
}

// LLMProviderListResponse wraps a list of providers.
type LLMProviderListResponse struct {
	Providers []LLMProviderResponse `json:"providers"`
	Total     int                   `json:"total"`
}

// LLMTestResponse is the result of a connectivity test.
type LLMTestResponse struct {
	OK        bool   `json:"ok"`
	Model     string `json:"model,omitempty"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
	Error     string `json:"error,omitempty"`
}
