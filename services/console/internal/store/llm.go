package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"console/internal/model"
)

// LLMProviderStore persists LLMProvider rows in MySQL.
type LLMProviderStore struct {
	db *sql.DB
}

// NewLLMProviderStore returns an LLMProviderStore backed by the given *sql.DB.
func NewLLMProviderStore(db *sql.DB) *LLMProviderStore {
	return &LLMProviderStore{db: db}
}

// Create inserts a new LLM provider. Returns ErrDuplicate when the name
// is already taken.
func (s *LLMProviderStore) Create(ctx context.Context, p *model.LLMProvider) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO llm_providers (id, name, base_url, api_key, model, max_tokens, temperature, is_default, is_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.BaseURL, p.APIKey, p.Model, p.MaxTokens, p.Temperature, p.IsDefault, p.IsEnabled, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("insert llm_provider: %w", err)
	}
	return nil
}

// GetByID fetches a provider by primary key. Returns ErrNotFound when no
// row matches.
func (s *LLMProviderStore) GetByID(ctx context.Context, id string) (*model.LLMProvider, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, base_url, api_key, model, max_tokens, temperature, is_default, is_enabled, created_at, updated_at
		 FROM llm_providers WHERE id = ?`, id,
	)
	return scanLLMProvider(row)
}

// GetDefault fetches the provider marked as default. Returns ErrNotFound
// when no default is set.
func (s *LLMProviderStore) GetDefault(ctx context.Context) (*model.LLMProvider, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, base_url, api_key, model, max_tokens, temperature, is_default, is_enabled, created_at, updated_at
		 FROM llm_providers WHERE is_default = TRUE LIMIT 1`,
	)
	return scanLLMProvider(row)
}

// List returns all providers ordered by name.
func (s *LLMProviderStore) List(ctx context.Context) ([]model.LLMProvider, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, base_url, api_key, model, max_tokens, temperature, is_default, is_enabled, created_at, updated_at
		 FROM llm_providers ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list llm_providers: %w", err)
	}
	defer rows.Close()

	var providers []model.LLMProvider
	for rows.Next() {
		var p model.LLMProvider
		if err := rows.Scan(&p.ID, &p.Name, &p.BaseURL, &p.APIKey, &p.Model, &p.MaxTokens, &p.Temperature, &p.IsDefault, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan llm_provider: %w", err)
		}
		providers = append(providers, p)
	}
	return providers, rows.Err()
}

// Update applies a partial update. Only non-nil fields in the request are written.
func (s *LLMProviderStore) Update(ctx context.Context, id string, req *model.LLMProviderUpdateRequest) (*model.LLMProvider, error) {
	setClauses := []string{}
	args := []interface{}{}

	if req.Name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *req.Name)
	}
	if req.BaseURL != nil {
		setClauses = append(setClauses, "base_url = ?")
		args = append(args, *req.BaseURL)
	}
	if req.APIKey != nil {
		setClauses = append(setClauses, "api_key = ?")
		args = append(args, *req.APIKey)
	}
	if req.Model != nil {
		setClauses = append(setClauses, "model = ?")
		args = append(args, *req.Model)
	}
	if req.MaxTokens != nil {
		setClauses = append(setClauses, "max_tokens = ?")
		args = append(args, *req.MaxTokens)
	}
	if req.Temperature != nil {
		setClauses = append(setClauses, "temperature = ?")
		args = append(args, *req.Temperature)
	}
	if req.IsDefault != nil {
		setClauses = append(setClauses, "is_default = ?")
		args = append(args, *req.IsDefault)
	}
	if req.IsEnabled != nil {
		setClauses = append(setClauses, "is_enabled = ?")
		args = append(args, *req.IsEnabled)
	}

	if len(setClauses) == 0 {
		return s.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, id)

	query := "UPDATE llm_providers SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += " WHERE id = ?"

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		if isDuplicateKey(err) {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("update llm_provider: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return nil, ErrNotFound
	}

	return s.GetByID(ctx, id)
}

// Delete removes a provider by ID. Returns ErrNotFound when no row matches.
func (s *LLMProviderStore) Delete(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM llm_providers WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete llm_provider: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ClearDefault unsets is_default on all providers. Used before setting a new default.
func (s *LLMProviderStore) ClearDefault(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `UPDATE llm_providers SET is_default = FALSE WHERE is_default = TRUE`)
	if err != nil {
		return fmt.Errorf("clear default: %w", err)
	}
	return nil
}

func scanLLMProvider(row *sql.Row) (*model.LLMProvider, error) {
	var p model.LLMProvider
	err := row.Scan(&p.ID, &p.Name, &p.BaseURL, &p.APIKey, &p.Model, &p.MaxTokens, &p.Temperature, &p.IsDefault, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan llm_provider: %w", err)
	}
	return &p, nil
}
