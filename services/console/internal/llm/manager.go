package llm

import (
	"context"
	"fmt"

	"console/internal/store"
)

// Manager provides high-level access to LLM providers. It loads
// configuration from the database and creates clients on demand.
type Manager struct {
	store *store.LLMProviderStore
}

// NewManager creates a Manager backed by the given store.
func NewManager(s *store.LLMProviderStore) *Manager {
	return &Manager{store: s}
}

// DefaultClient returns a Client configured with the default provider.
// Returns an error if no default provider is configured.
func (m *Manager) DefaultClient(ctx context.Context) (*Client, error) {
	p, err := m.store.GetDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("load default provider: %w", err)
	}
	if !p.IsEnabled {
		return nil, fmt.Errorf("default provider %q is disabled", p.Name)
	}
	return NewClient(p), nil
}

// ClientByID returns a Client configured with the provider identified by id.
func (m *Manager) ClientByID(ctx context.Context, id string) (*Client, error) {
	p, err := m.store.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load provider %s: %w", id, err)
	}
	if !p.IsEnabled {
		return nil, fmt.Errorf("provider %q is disabled", p.Name)
	}
	return NewClient(p), nil
}
