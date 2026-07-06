package model

// LLMProviderResponse is the safe representation returned in API responses.
type LLMProviderResponse struct {
	ProviderBase
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// LLMProviderCreateRequest is the payload for creating a new provider.
type LLMProviderCreateRequest struct {
	ProviderCreateRequestBase
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// LLMProviderUpdateRequest is the payload for updating an existing provider.
type LLMProviderUpdateRequest struct {
	ProviderUpdateRequestBase
	MaxTokens   *int     `json:"max_tokens"`
	Temperature *float64 `json:"temperature"`
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
