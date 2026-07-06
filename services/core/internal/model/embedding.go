package model

// EmbeddingProvider represents a configured embedding provider with OpenAI-compatible API.
type EmbeddingProvider struct {
	ProviderBase
	Dimensions int `json:"dimensions"`
}

// EmbeddingProviderResponse is the safe representation returned in API responses.
type EmbeddingProviderResponse struct {
	ProviderResponseBase
	Dimensions int `json:"dimensions"`
}

// ToResponse converts EmbeddingProvider to EmbeddingProviderResponse, hiding the API key.
func (p *EmbeddingProvider) ToResponse() EmbeddingProviderResponse {
	return EmbeddingProviderResponse{
		ProviderResponseBase: p.ToResponseBase(),
		Dimensions:           p.Dimensions,
	}
}

// EmbeddingProviderCreateRequest is the payload for creating a new embedding provider.
type EmbeddingProviderCreateRequest struct {
	ProviderCreateRequestBase
	Dimensions int `json:"dimensions"`
}

// EmbeddingProviderUpdateRequest is the payload for updating an existing embedding provider.
type EmbeddingProviderUpdateRequest struct {
	ProviderUpdateRequestBase
	Dimensions *int `json:"dimensions"`
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
