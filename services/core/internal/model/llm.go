package model

// LLMProvider represents a configured LLM provider with OpenAI-compatible API.
type LLMProvider struct {
	ProviderBase
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// LLMProviderResponse is the safe representation returned in API responses.
type LLMProviderResponse struct {
	ProviderResponseBase
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// LLMProviderInternalResponse includes the API key for internal service-to-service use.
type LLMProviderInternalResponse struct {
	ProviderInternalResponseBase
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// ToInternalResponse converts LLMProvider to LLMProviderInternalResponse, including the API key.
func (p *LLMProvider) ToInternalResponse() LLMProviderInternalResponse {
	return LLMProviderInternalResponse{
		ProviderInternalResponseBase: p.ToInternalResponseBase(),
		MaxTokens:                    p.MaxTokens,
		Temperature:                  p.Temperature,
	}
}

// ToResponse converts LLMProvider to LLMProviderResponse, hiding the API key.
func (p *LLMProvider) ToResponse() LLMProviderResponse {
	return LLMProviderResponse{
		ProviderResponseBase: p.ToResponseBase(),
		MaxTokens:            p.MaxTokens,
		Temperature:          p.Temperature,
	}
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
