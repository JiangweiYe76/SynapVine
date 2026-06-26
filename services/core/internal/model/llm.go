package model

import "time"

// LLMProvider represents a configured LLM provider with OpenAI-compatible API.
type LLMProvider struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	BaseURL     string    `json:"base_url"`
	APIKey      string    `json:"-"` // Never expose API key in responses
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	IsDefault   bool      `json:"is_default"`
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

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

// LLMProviderInternalResponse includes the API key for internal service-to-service use.
type LLMProviderInternalResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	BaseURL     string    `json:"base_url"`
	APIKey      string    `json:"api_key"`
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	IsDefault   bool      `json:"is_default"`
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToInternalResponse converts LLMProvider to LLMProviderInternalResponse, including the API key.
func (p *LLMProvider) ToInternalResponse() LLMProviderInternalResponse {
	return LLMProviderInternalResponse{
		ID:          p.ID,
		Name:        p.Name,
		BaseURL:     p.BaseURL,
		APIKey:      p.APIKey,
		Model:       p.Model,
		MaxTokens:   p.MaxTokens,
		Temperature: p.Temperature,
		IsDefault:   p.IsDefault,
		IsEnabled:   p.IsEnabled,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ToResponse converts LLMProvider to LLMProviderResponse, hiding the API key.
func (p *LLMProvider) ToResponse() LLMProviderResponse {
	return LLMProviderResponse{
		ID:          p.ID,
		Name:        p.Name,
		BaseURL:     p.BaseURL,
		Model:       p.Model,
		MaxTokens:   p.MaxTokens,
		Temperature: p.Temperature,
		IsDefault:   p.IsDefault,
		IsEnabled:   p.IsEnabled,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
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
