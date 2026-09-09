package model

import "time"

// LLMUsage is a single recorded LLM consumption event: one analysis
// call against one provider.
type LLMUsage struct {
	ID               string    `json:"id"`
	PaperID          string    `json:"paper_id"`
	ProviderID       string    `json:"provider_id"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	CreatedAt        time.Time `json:"created_at"`
}

// LLMUsageRecordRequest is the payload services submit to record a
// usage event.
type LLMUsageRecordRequest struct {
	PaperID          string `json:"paper_id"`
	ProviderID       string `json:"provider_id"`
	Model            string `json:"model"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
}

// LLMUsageProviderSummary aggregates usage for one provider.
type LLMUsageProviderSummary struct {
	ProviderID       string `json:"provider_id"`
	Model            string `json:"model"`
	Calls            int    `json:"calls"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
}

// LLMUsageDaySummary aggregates usage for one calendar day.
type LLMUsageDaySummary struct {
	Date        string `json:"date"`
	Calls       int    `json:"calls"`
	TotalTokens int64  `json:"total_tokens"`
}

// LLMUsageSummaryResponse is the aggregate usage report served to the
// console.
type LLMUsageSummaryResponse struct {
	Totals     LLMUsageProviderSummary   `json:"totals"`
	ByProvider []LLMUsageProviderSummary `json:"by_provider"`
	ByDay      []LLMUsageDaySummary      `json:"by_day"`
}
