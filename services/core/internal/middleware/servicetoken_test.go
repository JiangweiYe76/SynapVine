package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

var testTokens = map[string]string{
	"portal":    "portal-token",
	"console":   "console-token",
	"discovery": "discovery-token",
}

// newTestApp builds a Fiber app whose single route is protected by
// RequireServiceToken for the given allowed services.
func newTestApp(tokens map[string]string, allowed ...string) *fiber.App {
	app := fiber.New()
	app.Get("/api/test", RequireServiceToken(tokens, allowed...), func(c *fiber.Ctx) error {
		name, _ := c.Locals("service_name").(string)
		return c.JSON(fiber.Map{"service": name})
	})
	return app
}

func TestRequireServiceToken(t *testing.T) {
	tests := []struct {
		name       string
		tokens     map[string]string
		allowed    []string
		header     string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid token of allowed service",
			tokens:     testTokens,
			allowed:    []string{"portal", "console", "discovery"},
			header:     "console-token",
			wantStatus: 200,
			wantBody:   `{"service":"console"}`,
		},
		{
			name:       "read tier accepts portal token",
			tokens:     testTokens,
			allowed:    []string{"portal", "console", "discovery"},
			header:     "portal-token",
			wantStatus: 200,
			wantBody:   `{"service":"portal"}`,
		},
		{
			name:       "write tier rejects portal token",
			tokens:     testTokens,
			allowed:    []string{"console", "discovery"},
			header:     "portal-token",
			wantStatus: 401,
			wantBody:   `{"error":"invalid_service_token","message":"A valid service token is required"}`,
		},
		{
			name:       "internal tier rejects console token",
			tokens:     testTokens,
			allowed:    []string{"discovery"},
			header:     "console-token",
			wantStatus: 401,
			wantBody:   `{"error":"invalid_service_token","message":"A valid service token is required"}`,
		},
		{
			name:       "internal tier accepts discovery token",
			tokens:     testTokens,
			allowed:    []string{"discovery"},
			header:     "discovery-token",
			wantStatus: 200,
			wantBody:   `{"service":"discovery"}`,
		},
		{
			name:       "missing header",
			tokens:     testTokens,
			allowed:    []string{"console"},
			wantStatus: 401,
			wantBody:   `{"error":"invalid_service_token","message":"A valid service token is required"}`,
		},
		{
			name:       "empty header",
			tokens:     testTokens,
			allowed:    []string{"console"},
			header:     " ",
			wantStatus: 401,
			wantBody:   `{"error":"invalid_service_token","message":"A valid service token is required"}`,
		},
		{
			name:       "unknown token value",
			tokens:     testTokens,
			allowed:    []string{"console"},
			header:     "attacker-guess",
			wantStatus: 401,
			wantBody:   `{"error":"invalid_service_token","message":"A valid service token is required"}`,
		},
		{
			name:       "empty tokens map denies everything",
			tokens:     map[string]string{},
			allowed:    []string{"console"},
			header:     "console-token",
			wantStatus: 401,
			wantBody:   `{"error":"invalid_service_token","message":"A valid service token is required"}`,
		},
		{
			name:       "token not present in the tokens map",
			tokens:     testTokens,
			allowed:    []string{"worker"},
			header:     "console-token",
			wantStatus: 401,
			wantBody:   `{"error":"invalid_service_token","message":"A valid service token is required"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp(tt.tokens, tt.allowed...)
			req := httptest.NewRequest("GET", "/api/test", nil)
			if tt.header != "" {
				req.Header.Set(ServiceTokenHeader, tt.header)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}
			if string(body) != tt.wantBody {
				t.Errorf("body = %q, want %q", string(body), tt.wantBody)
			}
		})
	}
}
