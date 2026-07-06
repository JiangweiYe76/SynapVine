package model

// EmbeddingProviderResponse is the safe representation returned in API responses.
type EmbeddingProviderResponse struct {
	ProviderBase
	Dimensions int `json:"dimensions"`
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
