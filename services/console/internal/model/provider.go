package model

import "time"

// ProviderBase contains the common fields shared by all provider types (LLM, Embedding, etc.).
type ProviderBase struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	BaseURL   string    `json:"base_url"`
	Model     string    `json:"model"`
	IsDefault bool      `json:"is_default"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProviderCreateRequestBase contains the common fields for creating a new provider.
type ProviderCreateRequestBase struct {
	Name      string `json:"name"`
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key"`
	Model     string `json:"model"`
	IsDefault bool   `json:"is_default"`
}

// ProviderUpdateRequestBase contains the common fields for updating an existing provider.
type ProviderUpdateRequestBase struct {
	Name      *string `json:"name"`
	BaseURL   *string `json:"base_url"`
	APIKey    *string `json:"api_key"`
	Model     *string `json:"model"`
	IsDefault *bool   `json:"is_default"`
	IsEnabled *bool   `json:"is_enabled"`
}
