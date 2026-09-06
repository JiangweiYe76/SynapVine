package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"ai-graph-server/internal/config"
	"ai-graph-server/internal/coreclient"
	"ai-graph-server/internal/handler"
	"ai-graph-server/internal/security"
	"ai-graph-server/internal/service"

	"github.com/gofiber/fiber/v2"
)

// fakeCore serves a large /api/graph/data payload (>10KB) plus a minimal
// community tree, counting data hits.
func fakeCore(t *testing.T) *httptest.Server {
	t.Helper()
	nodes := make([]string, 0, 100)
	for i := 0; i < 100; i++ {
		nodes = append(nodes, `{"id":"n`+strconv.Itoa(i)+`","name":"Concept `+strconv.Itoa(i)+
			`","description":"A fairly long description of concept `+strconv.Itoa(i)+
			` used to pad the response well beyond any compression threshold. `+
			strings.Repeat("padding ", 20)+`"}`)
	}
	dataPayload := `{"nodes":[` + strings.Join(nodes, ",") + `],"edges":[]}`

	mux := http.NewServeMux()
	mux.HandleFunc("/api/graph/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, dataPayload)
	})
	mux.HandleFunc("/api/communities/tree", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"communities":[{"id":"1","name":"Root","color":"#fff","level":0,"node_count":100}]}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// newTestApp builds the portal app wired to a fake core server and
// returns the app along with an issued access token.
func newTestApp(t *testing.T) (*fiber.App, string) {
	t.Helper()
	srv := fakeCore(t)

	cfg := &config.Config{
		Port:          "0",
		AllowedOrigin: "http://localhost:5173",
		CoreURL:       srv.URL,
		ServiceToken:  "test-token",
	}
	tokenStore := security.NewTokenStore()
	svc := service.New(coreclient.New(srv.URL, "test-token"))
	gh := handler.NewGraphHandler(svc)
	app := newApp(cfg, gh, tokenStore)

	token, err := tokenStore.Issue()
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return app, token
}

func graphNodesRequest(t *testing.T, app *fiber.App, token string, gzipAccepted bool) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/graph/nodes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if gzipAccepted {
		req.Header.Set("Accept-Encoding", "gzip")
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return resp
}

func TestCompressGzipWhenAccepted(t *testing.T) {
	app, token := newTestApp(t)

	resp := graphNodesRequest(t, app, token, true)
	defer resp.Body.Close()

	if got := resp.Header.Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}

	// The body must be a valid gzip stream that decodes to the JSON
	// payload produced by the fake core.
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("body is not gzip: %v", err)
	}
	decoded, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("gunzip: %v", err)
	}
	if !strings.Contains(string(decoded), `"nodes"`) {
		t.Errorf("decoded body does not contain nodes payload: %.80s", decoded)
	}
	if len(decoded) <= len(raw) {
		t.Errorf("expected compression to shrink the body: raw=%d decoded=%d", len(raw), len(decoded))
	}
}

func TestNoCompressionWithoutAcceptEncoding(t *testing.T) {
	app, token := newTestApp(t)

	resp := graphNodesRequest(t, app, token, false)
	defer resp.Body.Close()

	if got := resp.Header.Get("Content-Encoding"); got != "" {
		t.Fatalf("Content-Encoding = %q, want empty", got)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), `"nodes"`) {
		t.Errorf("unexpected body: %.80s", body)
	}
}
