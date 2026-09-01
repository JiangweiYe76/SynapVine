package repository

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"core/internal/model"
	"core/internal/testutil"

	"github.com/google/uuid"
)

func newEmbeddingProvider(name, apiKey string) *model.EmbeddingProvider {
	now := time.Now()
	return &model.EmbeddingProvider{
		ProviderBase: model.ProviderBase{
			ID:        uuid.New().String(),
			Name:      name,
			BaseURL:   "https://api.example.com/v1",
			APIKey:    apiKey,
			Model:     "text-embedding-test",
			IsEnabled: true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Dimensions: 1536,
	}
}

// rawEmbeddingAPIKey reads the stored api_key column directly, bypassing
// the repository's decrypt path.
func rawEmbeddingAPIKey(t *testing.T, conn *sql.DB, id string) string {
	t.Helper()
	var stored string
	err := conn.QueryRowContext(context.Background(),
		`SELECT api_key FROM embedding_providers WHERE id = ?`, id).Scan(&stored)
	if err != nil {
		t.Fatalf("raw read api_key: %v", err)
	}
	return stored
}

func TestEmbeddingProviderKeyEncryptedRoundtrip(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewEmbeddingProviderRepository(conn, testutil.TestCipher)
	ctx := context.Background()

	p := newEmbeddingProvider("emb-roundtrip-"+uuid.NewString(), "sk-embedding-secret")
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.APIKey != "sk-embedding-secret" {
		t.Fatalf("decrypted key = %q, want plaintext", got.APIKey)
	}

	stored := rawEmbeddingAPIKey(t, conn, p.ID)
	if !strings.HasPrefix(stored, encryptedEnvelopePrefix) {
		t.Fatalf("stored api_key missing envelope prefix: %q", stored)
	}
	if strings.Contains(stored, "sk-embedding-secret") {
		t.Fatal("stored api_key contains plaintext secret")
	}
}

func TestEmbeddingProviderLegacyPlaintextPassthrough(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewEmbeddingProviderRepository(conn, testutil.TestCipher)
	ctx := context.Background()

	// Seed a legacy plaintext row directly, marked as default.
	p := newEmbeddingProvider("emb-legacy-"+uuid.NewString(), "sk-emb-legacy")
	if _, err := conn.ExecContext(ctx,
		`INSERT INTO embedding_providers (id, name, base_url, api_key, model, dimensions, is_default, is_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.BaseURL, p.APIKey, p.Model, p.Dimensions, true, true, p.CreatedAt, p.UpdatedAt,
	); err != nil {
		t.Fatalf("seed legacy row: %v", err)
	}

	got, err := repo.GetDefault(ctx)
	if err != nil {
		t.Fatalf("GetDefault legacy: %v", err)
	}
	if got.APIKey != "sk-emb-legacy" {
		t.Fatalf("legacy read = %q, want plaintext passthrough", got.APIKey)
	}
}
