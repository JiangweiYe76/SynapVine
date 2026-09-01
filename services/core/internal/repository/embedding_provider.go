package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"core/internal/model"
	"core/internal/security"
)

// EmbeddingProviderRepository persists EmbeddingProvider rows in MySQL. API
// keys are encrypted at rest via the injected KeyCipher; callers always
// see plaintext.
type EmbeddingProviderRepository struct {
	db     *sql.DB
	cipher *security.KeyCipher
}

// NewEmbeddingProviderRepository returns an EmbeddingProviderRepository
// backed by the given *sql.DB. The cipher encrypts API keys on write and
// decrypts them on read.
func NewEmbeddingProviderRepository(db *sql.DB, cipher *security.KeyCipher) *EmbeddingProviderRepository {
	return &EmbeddingProviderRepository{db: db, cipher: cipher}
}

// Create inserts a new embedding provider. Returns ErrDuplicate when the name
// is already taken.
func (r *EmbeddingProviderRepository) Create(ctx context.Context, p *model.EmbeddingProvider) error {
	encryptedKey, err := r.cipher.Encrypt(p.APIKey)
	if err != nil {
		return fmt.Errorf("encrypt api_key: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO embedding_providers (id, name, base_url, api_key, model, dimensions, is_default, is_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.BaseURL, encryptedKey, p.Model, p.Dimensions, p.IsDefault, p.IsEnabled, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("insert embedding_provider: %w", err)
	}
	return nil
}

// GetByID fetches a provider by primary key. Returns ErrNotFound when no
// row matches.
func (r *EmbeddingProviderRepository) GetByID(ctx context.Context, id string) (*model.EmbeddingProvider, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, base_url, api_key, model, dimensions, is_default, is_enabled, created_at, updated_at
		 FROM embedding_providers WHERE id = ?`, id,
	)
	return scanEmbeddingProvider(row, r.cipher)
}

// GetDefault fetches the provider marked as default. Returns ErrNotFound
// when no default is set.
func (r *EmbeddingProviderRepository) GetDefault(ctx context.Context) (*model.EmbeddingProvider, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, base_url, api_key, model, dimensions, is_default, is_enabled, created_at, updated_at
		 FROM embedding_providers WHERE is_default = TRUE LIMIT 1`,
	)
	return scanEmbeddingProvider(row, r.cipher)
}

// List returns all providers ordered by name.
func (r *EmbeddingProviderRepository) List(ctx context.Context) ([]model.EmbeddingProvider, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, base_url, api_key, model, dimensions, is_default, is_enabled, created_at, updated_at
		 FROM embedding_providers ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list embedding_providers: %w", err)
	}
	defer rows.Close()

	var providers []model.EmbeddingProvider
	for rows.Next() {
		var p model.EmbeddingProvider
		if err := rows.Scan(&p.ID, &p.Name, &p.BaseURL, &p.APIKey, &p.Model, &p.Dimensions, &p.IsDefault, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan embedding_provider: %w", err)
		}
		plaintext, err := r.cipher.Decrypt(p.APIKey)
		if err != nil {
			return nil, fmt.Errorf("decrypt embedding_provider api_key (id=%s): %w", p.ID, err)
		}
		p.APIKey = plaintext
		providers = append(providers, p)
	}
	return providers, rows.Err()
}

// Update applies a partial update. Only non-nil fields in the request are written.
func (r *EmbeddingProviderRepository) Update(ctx context.Context, id string, req *model.EmbeddingProviderUpdateRequest) (*model.EmbeddingProvider, error) {
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
	if req.Dimensions != nil {
		setClauses = append(setClauses, "dimensions = ?")
		args = append(args, *req.Dimensions)
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

	query := "UPDATE embedding_providers SET "
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
		return nil, fmt.Errorf("update embedding_provider: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return nil, ErrNotFound
	}

	return r.GetByID(ctx, id)
}

// Delete removes a provider by ID. Returns ErrNotFound when no row matches.
func (r *EmbeddingProviderRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM embedding_providers WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete embedding_provider: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ClearDefault unsets is_default on all providers. Used before setting a new default.
func (r *EmbeddingProviderRepository) ClearDefault(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `UPDATE embedding_providers SET is_default = FALSE WHERE is_default = TRUE`)
	if err != nil {
		return fmt.Errorf("clear default: %w", err)
	}
	return nil
}

func scanEmbeddingProvider(row *sql.Row, cipher *security.KeyCipher) (*model.EmbeddingProvider, error) {
	var p model.EmbeddingProvider
	err := row.Scan(&p.ID, &p.Name, &p.BaseURL, &p.APIKey, &p.Model, &p.Dimensions, &p.IsDefault, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan embedding_provider: %w", err)
	}
	plaintext, err := cipher.Decrypt(p.APIKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt embedding_provider api_key (id=%s): %w", p.ID, err)
	}
	p.APIKey = plaintext
	return &p, nil
}
