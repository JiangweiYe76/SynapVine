package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// newFiberRequest builds a request for app.Test.
func newFiberRequest(method, target string) *http.Request {
	return httptest.NewRequest(method, target, nil)
}

// readBody reads and returns the full response body.
func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}
	return body
}

// fakeMetaReader is a controllable MetaReader for handler tests.
type fakeMetaReader struct {
	ts  time.Time
	ok  bool
	err error
}

func (f *fakeMetaReader) LastUpdated(context.Context) (time.Time, bool, error) {
	return f.ts, f.ok, f.err
}

func TestGraphMetaHandler_Get_WhenNeverUpdated(t *testing.T) {
	app := fiber.New()
	app.Get("/api/graph/meta", NewGraphMetaHandler(&fakeMetaReader{}).Get)

	resp, err := app.Test(newFiberRequest("GET", "/api/graph/meta"))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body := readBody(t, resp)
	if want := `{"last_updated":null}`; string(body) != want {
		t.Errorf("body = %s, want %s", body, want)
	}
}

func TestGraphMetaHandler_Get_ReturnsTimestamp(t *testing.T) {
	ts := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	app := fiber.New()
	app.Get("/api/graph/meta", NewGraphMetaHandler(&fakeMetaReader{ts: ts, ok: true}).Get)

	resp, err := app.Test(newFiberRequest("GET", "/api/graph/meta"))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body := readBody(t, resp)
	if want := `{"last_updated":"` + ts.Format(time.RFC3339) + `"}`; string(body) != want {
		t.Errorf("body = %s, want %s", body, want)
	}
}

func TestGraphMetaHandler_Get_RepositoryError(t *testing.T) {
	app := fiber.New()
	app.Get("/api/graph/meta", NewGraphMetaHandler(&fakeMetaReader{err: errors.New("boom")}).Get)

	resp, err := app.Test(newFiberRequest("GET", "/api/graph/meta"))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
