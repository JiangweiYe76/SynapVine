package arxiv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sampleFeed = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>ArXiv Query</title>
  <entry>
    <id>http://arxiv.org/abs/2401.12345v2</id>
    <title>
      A  Study  of
      Attention
    </title>
    <summary>  This paper
        studies attention.  </summary>
    <published>2026-09-01T17:59:59Z</published>
    <author><name>Ada Lovelace</name></author>
    <author><name>Alan Turing</name></author>
    <link href="http://arxiv.org/abs/2401.12345v2" rel="alternate" type="text/html"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/2402.00601</id>
    <title>Another Paper</title>
    <summary>Another abstract.</summary>
    <published>2026-09-02T00:00:00Z</published>
    <author><name>Grace Hopper</name></author>
  </entry>
</feed>`

func TestClient_FetchRecent_ParsesAndNormalizes(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/atom+xml")
		_, _ = w.Write([]byte(sampleFeed))
	}))
	defer srv.Close()

	c := NewClient()
	c.baseURL = srv.URL

	papers, err := c.FetchRecent(context.Background(), []string{"cs.AI", "cs.CL"}, 10)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	if len(papers) != 2 {
		t.Fatalf("papers = %d, want 2", len(papers))
	}

	first := papers[0]
	// Version suffix must be stripped for stable dedup.
	if first.ArxivID != "2401.12345" {
		t.Errorf("arxiv id = %q, want 2401.12345", first.ArxivID)
	}
	if first.Title != "A Study of Attention" {
		t.Errorf("title = %q, want whitespace-collapsed", first.Title)
	}
	if first.Summary != "This paper studies attention." {
		t.Errorf("summary = %q, want whitespace-collapsed", first.Summary)
	}
	if len(first.Authors) != 2 || first.Authors[0] != "Ada Lovelace" {
		t.Errorf("authors = %v, want 2 names", first.Authors)
	}
	if first.AbsURL != "https://arxiv.org/abs/2401.12345" {
		t.Errorf("abs url = %q", first.AbsURL)
	}
	if want := "2026-09-01T17:59:59Z"; first.Published.UTC().Format("2006-01-02T15:04:05Z") != want {
		t.Errorf("published = %v, want %s", first.Published, want)
	}

	// The query must OR the categories together and request descending
	// submission order.
	if !strings.Contains(gotQuery, "cat%3Acs.AI%2BOR%2Bcat%3Acs.CL") {
		t.Errorf("query = %q, want OR-joined category filter", gotQuery)
	}
	if !strings.Contains(gotQuery, "sortOrder=descending") || !strings.Contains(gotQuery, "max_results=10") {
		t.Errorf("query = %q, want sort and limit params", gotQuery)
	}
}

func TestClient_FetchRecent_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := NewClient()
	c.baseURL = srv.URL

	if _, err := c.FetchRecent(context.Background(), []string{"cs.AI"}, 5); err == nil {
		t.Fatal("expected error on non-200 response")
	}
}
