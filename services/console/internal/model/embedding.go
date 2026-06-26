package model

import "time"

// EmbeddingProviderResponse is the safe representation returned in API responses.
type EmbeddingProviderResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	BaseURL    string    `json:"base_url"`
	Model      string    `json:"model"`
	Dimensions int       `json:"dimensions"`
	IsDefault  bool      `json:"is_default"`
	IsEnabled  bool      `json:"is_enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// EmbeddingProviderCreateRequest is the payload for creating a new embedding provider.
type EmbeddingProviderCreateRequest struct {
	Name       string `json:"name"`
	BaseURL    string `json:"base_url"`
	APIKey     string `json:"api_key"`
	Model      string `json:"model"`
	Dimensions int    `json:"dimensions"`
	IsDefault  bool   `json:"is_default"`
}

// EmbeddingProviderUpdateRequest is the payload for updating an existing embedding provider.
type EmbeddingProviderUpdateRequest struct {
	Name       *string `json:"name"`
	BaseURL    *string `json:"base_url"`
	APIKey     *string `json:"api_key"`
	Model      *string `json:"model"`
	Dimensions *int    `json:"dimensions"`
	IsDefault  *bool   `json:"is_default"`
	IsEnabled  *bool   `json:"is_enabled"`
}

// EmbeddingProviderListResponse wraps a list of embedding providers.
type EmbeddingProviderListResponse struct {
	Providers []EmbeddingProviderResponse `json:"providers"`
	Total     int                         `json:"total"`
}

// EmbeddingTestResponse is the result of a connectivity test.
type EmbeddingTestResponse struct {
	OK         bool   `json:"ok"`
	Dimensions int    `json:"dimensions,omitempty"`
	LatencyMs  int64  `json:"latency_ms,omitempty"`
	Error      string `json:"error,omitempty"`
}
