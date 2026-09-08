// Package arxiv provides a minimal client for the ArXiv export API. It
// fetches the most recent submissions for a set of categories so the
// discovery scheduler can ingest new papers into the graph pipeline.
package arxiv

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// defaultBaseURL is the ArXiv export API. The export mirror is the
// supported, rate-limited interface for programmatic access.
const defaultBaseURL = "https://export.arxiv.org/api/query"

// versionSuffix matches the version part of an ArXiv id
// (e.g. "2401.12345v2" -> "v2"). Dedup uses the version-less id so a
// paper updated to v2 is not ingested twice.
var versionSuffix = regexp.MustCompile(`v\d+$`)

// Paper is a single ArXiv submission as returned by the export API.
type Paper struct {
	ArxivID   string    // version-less identifier, e.g. "2401.12345"
	Title     string    // whitespace-normalized title
	Summary   string    // whitespace-normalized abstract
	Authors   []string  // author names in listing order
	Published time.Time // original submission date
	AbsURL    string    // canonical abs page URL
	SourceURL string    // alias of AbsURL for paper records
}

// Client fetches recent submissions from the ArXiv export API.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient creates a Client pointing at the public ArXiv export API.
func NewClient() *Client {
	return &Client{
		baseURL: defaultBaseURL,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetBaseURL overrides the export API endpoint. Used by tests to point
// the client at a stub server.
func (c *Client) SetBaseURL(u string) {
	c.baseURL = u
}

// atomFeed mirrors the subset of the Atom format the export API returns.
type atomFeed struct {
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID        string `xml:"id"`
	Title     string `xml:"title"`
	Summary   string `xml:"summary"`
	Published string `xml:"published"`
	Authors   []struct {
		Name string `xml:"name"`
	} `xml:"author"`
	Links []struct {
		Href string `xml:"href"`
		Type string `xml:"type"`
	} `xml:"link"`
}

// FetchRecent queries the newest submissions in the given categories,
// returning at most max papers sorted by submission date descending.
// The query shape is cat:A+OR+cat:B (AND semantics would require a
// paper tagged with every category, which is not the intent here).
func (c *Client) FetchRecent(ctx context.Context, categories []string, max int) ([]Paper, error) {
	if len(categories) == 0 {
		return nil, fmt.Errorf("no categories configured")
	}
	if max <= 0 {
		max = 5
	}

	terms := make([]string, 0, len(categories))
	for _, cat := range categories {
		terms = append(terms, "cat:"+url.QueryEscape(strings.TrimSpace(cat)))
	}

	q := url.Values{}
	q.Set("search_query", strings.Join(terms, "+OR+"))
	q.Set("sortBy", "submittedDate")
	q.Set("sortOrder", "descending")
	q.Set("max_results", strconv.Itoa(max))
	requestURL := c.baseURL + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build arxiv request: %w", err)
	}
	req.Header.Set("User-Agent", "SynapVine-Discovery/1.0 (knowledge graph ingestion)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("arxiv request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("arxiv returned status %d: %s", resp.StatusCode, string(raw))
	}

	var feed atomFeed
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&feed); err != nil {
		return nil, fmt.Errorf("parse arxiv atom feed: %w", err)
	}

	papers := make([]Paper, 0, len(feed.Entries))
	for _, e := range feed.Entries {
		p, err := toPaper(e)
		if err != nil {
			return nil, fmt.Errorf("normalize arxiv entry %q: %w", e.ID, err)
		}
		papers = append(papers, p)
	}
	return papers, nil
}

// toPaper normalizes one Atom entry: extract the version-less ArXiv id,
// collapse whitespace in title/summary, and keep the canonical abs URL.
func toPaper(e atomEntry) (Paper, error) {
	id := arxivIDFromURL(e.ID)
	if id == "" {
		return Paper{}, fmt.Errorf("no arxiv id in entry id")
	}
	published, err := time.Parse(time.RFC3339, e.Published)
	if err != nil {
		return Paper{}, fmt.Errorf("parse published date: %w", err)
	}

	authors := make([]string, 0, len(e.Authors))
	for _, a := range e.Authors {
		if name := strings.TrimSpace(a.Name); name != "" {
			authors = append(authors, name)
		}
	}

	absURL := fmt.Sprintf("https://arxiv.org/abs/%s", id)
	return Paper{
		ArxivID:   id,
		Title:     normalizeSpace(e.Title),
		Summary:   normalizeSpace(e.Summary),
		Authors:   authors,
		Published: published,
		AbsURL:    absURL,
		SourceURL: absURL,
	}, nil
}

// arxivIDFromURL extracts the version-less ArXiv id from an entry id
// like "http://arxiv.org/abs/2401.12345v2".
func arxivIDFromURL(raw string) string {
	raw = strings.TrimSpace(raw)
	idx := strings.LastIndex(raw, "/abs/")
	if idx < 0 {
		return ""
	}
	id := raw[idx+len("/abs/"):]
	return versionSuffix.ReplaceAllString(id, "")
}

// normalizeSpace collapses all whitespace runs (ArXiv titles and
// summaries wrap across many lines) into single spaces.
func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
