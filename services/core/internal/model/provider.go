package model

import "time"

// ProviderBase contains the common fields shared by all provider types (LLM, Embedding, etc.).
type ProviderBase struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	BaseURL   string    `json:"base_url"`
	APIKey    string    `json:"-"` // Never expose API key in responses
	Model     string    `json:"model"`
	IsDefault bool      `json:"is_default"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProviderResponseBase contains the common fields for API responses.
type ProviderResponseBase struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	BaseURL   string    `json:"base_url"`
	Model     string    `json:"model"`
	IsDefault bool      `json:"is_default"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProviderInternalResponseBase contains the common fields for internal service-to-service responses, including the API key.
type ProviderInternalResponseBase struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	BaseURL   string    `json:"base_url"`
	APIKey    string    `json:"api_key"`
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

// ToResponseBase converts ProviderBase to ProviderResponseBase, hiding the API key.
func (p *ProviderBase) ToResponseBase() ProviderResponseBase {
	return ProviderResponseBase{
		ID:        p.ID,
		Name:      p.Name,
		BaseURL:   p.BaseURL,
		Model:     p.Model,
		IsDefault: p.IsDefault,
		IsEnabled: p.IsEnabled,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// ToInternalResponseBase converts ProviderBase to ProviderInternalResponseBase, including the API key.
func (p *ProviderBase) ToInternalResponseBase() ProviderInternalResponseBase {
	return ProviderInternalResponseBase{
		ID:        p.ID,
		Name:      p.Name,
		BaseURL:   p.BaseURL,
		APIKey:    p.APIKey,
		Model:     p.Model,
		IsDefault: p.IsDefault,
		IsEnabled: p.IsEnabled,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
