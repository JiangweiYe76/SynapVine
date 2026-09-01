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

const encryptedEnvelopePrefix = "enc:v1:"

func newLLMProvider(name, apiKey string) *model.LLMProvider {
	now := time.Now()
	return &model.LLMProvider{
		ProviderBase: model.ProviderBase{
			ID:        uuid.New().String(),
			Name:      name,
			BaseURL:   "https://api.example.com/v1",
			APIKey:    apiKey,
			Model:     "gpt-test",
			IsEnabled: true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		MaxTokens:   4096,
		Temperature: 0.7,
	}
}

// rawLLMAPIKey reads the stored api_key column directly, bypassing the
// repository's decrypt path, so tests can assert what is actually at
// rest in MySQL.
func rawLLMAPIKey(t *testing.T, conn *sql.DB, id string) string {
	t.Helper()
	var stored string
	err := conn.QueryRowContext(context.Background(),
		`SELECT api_key FROM llm_providers WHERE id = ?`, id).Scan(&stored)
	if err != nil {
		t.Fatalf("raw read api_key: %v", err)
	}
	return stored
}

func TestLLMProviderKeyEncryptedRoundtrip(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewLLMProviderRepository(conn, testutil.TestCipher)
	ctx := context.Background()

	p := newLLMProvider("llm-roundtrip-"+uuid.NewString(), "sk-plaintext-secret")
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// The repository must hand plaintext back to callers.
	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.APIKey != "sk-plaintext-secret" {
		t.Fatalf("decrypted key = %q, want plaintext", got.APIKey)
	}

	// What is actually stored must be ciphertext, never the plaintext.
	stored := rawLLMAPIKey(t, conn, p.ID)
	if !strings.HasPrefix(stored, encryptedEnvelopePrefix) {
		t.Fatalf("stored api_key missing envelope prefix: %q", stored)
	}
	if strings.Contains(stored, "sk-plaintext-secret") {
		t.Fatal("stored api_key contains plaintext secret")
	}
}

func TestLLMProviderKeyUpdateEncrypts(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewLLMProviderRepository(conn, testutil.TestCipher)
	ctx := context.Background()

	p := newLLMProvider("llm-update-"+uuid.NewString(), "sk-old")
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	newKey := "sk-rotated-secret"
	if _, err := repo.Update(ctx, p.ID, &model.LLMProviderUpdateRequest{
		ProviderUpdateRequestBase: model.ProviderUpdateRequestBase{APIKey: &newKey},
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.APIKey != newKey {
		t.Fatalf("decrypted key = %q, want %q", got.APIKey, newKey)
	}

	stored := rawLLMAPIKey(t, conn, p.ID)
	if !strings.HasPrefix(stored, encryptedEnvelopePrefix) || strings.Contains(stored, newKey) {
		t.Fatalf("updated api_key not encrypted at rest: %q", stored)
	}
}

func TestLLMProviderLegacyPlaintextLazyMigration(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewLLMProviderRepository(conn, testutil.TestCipher)
	ctx := context.Background()

	// Simulate a row written before encryption existed: plaintext in
	// the column, no envelope prefix.
	p := newLLMProvider("llm-legacy-"+uuid.NewString(), "sk-legacy-plaintext")
	if _, err := conn.ExecContext(ctx,
		`INSERT INTO llm_providers (id, name, base_url, api_key, model, max_tokens, temperature, is_default, is_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.BaseURL, p.APIKey, p.Model, p.MaxTokens, p.Temperature, false, true, p.CreatedAt, p.UpdatedAt,
	); err != nil {
		t.Fatalf("seed legacy row: %v", err)
	}

	// Reads must transparently return the plaintext.
	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID legacy: %v", err)
	}
	if got.APIKey != "sk-legacy-plaintext" {
		t.Fatalf("legacy read = %q, want plaintext passthrough", got.APIKey)
	}

	// Any update must re-write the key as ciphertext (lazy migration).
	newName := "llm-legacy-renamed-" + uuid.NewString()
	if _, err := repo.Update(ctx, p.ID, &model.LLMProviderUpdateRequest{
		ProviderUpdateRequestBase: model.ProviderUpdateRequestBase{Name: &newName},
	}); err != nil {
		t.Fatalf("Update legacy: %v", err)
	}
	stored := rawLLMAPIKey(t, conn, p.ID)
	if !strings.HasPrefix(stored, encryptedEnvelopePrefix) {
		t.Fatalf("legacy row not encrypted after update: %q", stored)
	}
}

func TestLLMProviderTamperedCiphertextFails(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewLLMProviderRepository(conn, testutil.TestCipher)
	ctx := context.Background()

	p := newLLMProvider("llm-tamper-"+uuid.NewString(), "sk-secret")
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Corrupt the stored ciphertext in place.
	if _, err := conn.ExecContext(ctx,
		`UPDATE llm_providers SET api_key = ? WHERE id = ?`,
		encryptedEnvelopePrefix+"AAAAtamperedAAAA", p.ID); err != nil {
		t.Fatalf("corrupt row: %v", err)
	}

	if _, err := repo.GetByID(ctx, p.ID); err == nil {
		t.Fatal("GetByID on tampered ciphertext succeeded, want decrypt error")
	}
}

func TestLLMProviderListDecryptsAll(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewLLMProviderRepository(conn, testutil.TestCipher)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		p := newLLMProvider("llm-list-"+uuid.NewString(), "sk-list-secret")
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
	}

	providers, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(providers) != 3 {
		t.Fatalf("List returned %d providers, want 3", len(providers))
	}
	for _, p := range providers {
		if p.APIKey != "sk-list-secret" {
			t.Fatalf("provider %s decrypted key = %q", p.ID, p.APIKey)
		}
	}
}
