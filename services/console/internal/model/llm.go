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

// LLMUsageProviderSummary aggregates token usage for one provider.
type LLMUsageProviderSummary struct {
	ProviderID       string `json:"provider_id"`
	Model            string `json:"model"`
	Calls            int    `json:"calls"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
}

// LLMUsageDaySummary aggregates token usage for one calendar day.
type LLMUsageDaySummary struct {
	Date        string `json:"date"`
	Calls       int    `json:"calls"`
	TotalTokens int64  `json:"total_tokens"`
}

// LLMUsageSummary is the aggregated usage report from core.
type LLMUsageSummary struct {
	Totals     LLMUsageProviderSummary   `json:"totals"`
	ByProvider []LLMUsageProviderSummary `json:"by_provider"`
	ByDay      []LLMUsageDaySummary      `json:"by_day"`
}

// LLMTestResponse is the result of a connectivity test.
type LLMTestResponse struct {
	OK        bool   `json:"ok"`
	Model     string `json:"model,omitempty"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
	Error     string `json:"error,omitempty"`
}
