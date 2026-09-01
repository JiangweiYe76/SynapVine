package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"core/internal/model"
	"core/internal/security"
)

// LLMProviderRepository persists LLMProvider rows in MySQL. API keys are
// encrypted at rest via the injected KeyCipher; callers always see
// plaintext.
type LLMProviderRepository struct {
	db     *sql.DB
	cipher *security.KeyCipher
}

// NewLLMProviderRepository returns an LLMProviderRepository backed by the
// given *sql.DB. The cipher encrypts API keys on write and decrypts them
// on read.
func NewLLMProviderRepository(db *sql.DB, cipher *security.KeyCipher) *LLMProviderRepository {
	return &LLMProviderRepository{db: db, cipher: cipher}
}

// Create inserts a new LLM provider. Returns ErrDuplicate when the name
// is already taken.
func (r *LLMProviderRepository) Create(ctx context.Context, p *model.LLMProvider) error {
	encryptedKey, err := r.cipher.Encrypt(p.APIKey)
	if err != nil {
		return fmt.Errorf("encrypt api_key: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO llm_providers (id, name, base_url, api_key, model, max_tokens, temperature, is_default, is_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.BaseURL, encryptedKey, p.Model, p.MaxTokens, p.Temperature, p.IsDefault, p.IsEnabled, p.CreatedAt, p.UpdatedAt,
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
func (r *LLMProviderRepository) GetByID(ctx context.Context, id string) (*model.LLMProvider, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, base_url, api_key, model, max_tokens, temperature, is_default, is_enabled, created_at, updated_at
		 FROM llm_providers WHERE id = ?`, id,
	)
	return scanLLMProvider(row, r.cipher)
}

// GetDefault fetches the provider marked as default. Returns ErrNotFound
// when no default is set.
func (r *LLMProviderRepository) GetDefault(ctx context.Context) (*model.LLMProvider, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, base_url, api_key, model, max_tokens, temperature, is_default, is_enabled, created_at, updated_at
		 FROM llm_providers WHERE is_default = TRUE LIMIT 1`,
	)
	return scanLLMProvider(row, r.cipher)
}

// List returns all providers ordered by name.
func (r *LLMProviderRepository) List(ctx context.Context) ([]model.LLMProvider, error) {
	rows, err := r.db.QueryContext(ctx,
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
		plaintext, err := r.cipher.Decrypt(p.APIKey)
		if err != nil {
			return nil, fmt.Errorf("decrypt llm_provider api_key (id=%s): %w", p.ID, err)
		}
		p.APIKey = plaintext
		providers = append(providers, p)
	}
	return providers, rows.Err()
}

// Update applies a partial update. Only non-nil fields in the request are written.
func (r *LLMProviderRepository) Update(ctx context.Context, id string, req *model.LLMProviderUpdateRequest) (*model.LLMProvider, error) {
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
		encryptedKey, err := r.cipher.Encrypt(*req.APIKey)
		if err != nil {
			return nil, fmt.Errorf("encrypt api_key: %w", err)
		}
		setClauses = append(setClauses, "api_key = ?")
		args = append(args, encryptedKey)
	} else {
		// Re-seal the existing key on every other update so legacy
		// plaintext rows converge to ciphertext on their next write.
		existing, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		encryptedKey, err := r.cipher.Encrypt(existing.APIKey)
		if err != nil {
			return nil, fmt.Errorf("encrypt api_key: %w", err)
		}
		setClauses = append(setClauses, "api_key = ?")
		args = append(args, encryptedKey)
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
		return r.GetByID(ctx, id)
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

	result, err := r.db.ExecContext(ctx, query, args...)
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

	return r.GetByID(ctx, id)
}

// Delete removes a provider by ID. Returns ErrNotFound when no row matches.
func (r *LLMProviderRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM llm_providers WHERE id = ?`, id)
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
func (r *LLMProviderRepository) ClearDefault(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `UPDATE llm_providers SET is_default = FALSE WHERE is_default = TRUE`)
	if err != nil {
		return fmt.Errorf("clear default: %w", err)
	}
	return nil
}

func scanLLMProvider(row *sql.Row, cipher *security.KeyCipher) (*model.LLMProvider, error) {
	var p model.LLMProvider
	err := row.Scan(&p.ID, &p.Name, &p.BaseURL, &p.APIKey, &p.Model, &p.MaxTokens, &p.Temperature, &p.IsDefault, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan llm_provider: %w", err)
	}
	plaintext, err := cipher.Decrypt(p.APIKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt llm_provider api_key (id=%s): %w", p.ID, err)
	}
	p.APIKey = plaintext
	return &p, nil
}

// isDuplicateKey reports whether err is a MySQL 1062 duplicate-key error.
func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	return containsAny(err.Error(), "Error 1062", "Duplicate entry")
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) == 0 {
			continue
		}
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
	}
	return false
}
