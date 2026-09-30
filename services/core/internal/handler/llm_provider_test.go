package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"core/internal/llmprovider"
	"core/internal/model"
	"core/internal/repository"
	"core/internal/testutil"

	"github.com/gofiber/fiber/v2"
)

// newLLMProviderApp wires the provider routes against the test database.
func newLLMProviderApp(t *testing.T) (*fiber.App, *repository.LLMProviderRepository) {
	t.Helper()
	repo := repository.NewLLMProviderRepository(testutil.NewTestMySQL(t), testutil.TestCipher)
	h := NewLLMProviderHandler(repo)
	app := fiber.New()
	app.Post("/api/llm/providers", h.Create)
	app.Put("/api/llm/providers/:id", h.Update)
	return app, repo
}

// postJSON sends a JSON request body to the test app.
func postJSON(t *testing.T, app *fiber.App, method, target string, body any) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, target, reader)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}
	return resp
}

// createProvider seeds one provider through the API and returns its ID.
func createProvider(t *testing.T, app *fiber.App, maxTokens int) string {
	t.Helper()
	resp := postJSON(t, app, http.MethodPost, "/api/llm/providers", map[string]any{
		"name":       "prov-" + strings.ReplaceAll(t.Name(), "/", "-"),
		"base_url":   "https://api.example.com/v1",
		"api_key":    "sk-test",
		"model":      "test-model",
		"max_tokens": maxTokens,
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("seed provider: status = %d, body = %s", resp.StatusCode, raw)
	}
	var created model.LLMProviderResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created provider: %v", err)
	}
	return created.ID
}

// TestLLMProviderCreateRejectsOversizedBudget covers the failure that
// made every extraction unsendable: max_tokens is the per-request output
// budget, not the model context window, so a value above the ceiling has
// to be rejected at write time instead of failing every LLM call.
func TestLLMProviderCreateRejectsOversizedBudget(t *testing.T) {
	app, _ := newLLMProviderApp(t)

	resp := postJSON(t, app, http.MethodPost, "/api/llm/providers", map[string]any{
		"name":       "oversized",
		"base_url":   "https://api.example.com/v1",
		"api_key":    "sk-test",
		"model":      "test-model",
		"max_tokens": 1000000,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	var errBody model.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if errBody.Error != "invalid_max_tokens" {
		t.Errorf("error = %q, want %q", errBody.Error, "invalid_max_tokens")
	}
}

// TestLLMProviderCreateDefaultsMissingBudget pins the default budget so a
// provider created without max_tokens can actually complete a full-paper
// extraction.
func TestLLMProviderCreateDefaultsMissingBudget(t *testing.T) {
	app, _ := newLLMProviderApp(t)

	resp := postJSON(t, app, http.MethodPost, "/api/llm/providers", map[string]any{
		"name":     "defaulted",
		"base_url": "https://api.example.com/v1",
		"api_key":  "sk-test",
		"model":    "test-model",
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s", resp.StatusCode, raw)
	}
	var created model.LLMProviderResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created provider: %v", err)
	}
	if created.MaxTokens != llmprovider.DefaultMaxTokens {
		t.Errorf("max_tokens = %d, want default %d", created.MaxTokens, llmprovider.DefaultMaxTokens)
	}
}

// TestLLMProviderUpdateValidatesBudget covers max_tokens on the update
// path, which applies no normalization: any out-of-range value supplied
// here would be stored verbatim.
func TestLLMProviderUpdateValidatesBudget(t *testing.T) {
	app, _ := newLLMProviderApp(t)
	id := createProvider(t, app, 8192)

	tests := []struct {
		name       string
		maxTokens  int
		wantStatus int
	}{
		{name: "oversized budget is rejected", maxTokens: 1000000, wantStatus: http.StatusBadRequest},
		{name: "zero budget is rejected", maxTokens: 0, wantStatus: http.StatusBadRequest},
		{name: "negative budget is rejected", maxTokens: -5, wantStatus: http.StatusBadRequest},
		{name: "in-range budget is accepted", maxTokens: 8192, wantStatus: http.StatusOK},
		{name: "ceiling budget is accepted", maxTokens: llmprovider.MaxTokensLimit, wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := postJSON(t, app, http.MethodPut, "/api/llm/providers/"+id, map[string]any{
				"max_tokens": tc.maxTokens,
			})
			defer resp.Body.Close()
			if resp.StatusCode != tc.wantStatus {
				raw, _ := io.ReadAll(resp.Body)
				t.Errorf("status = %d, want %d (body = %s)", resp.StatusCode, tc.wantStatus, raw)
			}
		})
	}
}
