package pdf

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestExtractText_InvalidInput(t *testing.T) {
	_, err := ExtractText(strings.NewReader("this is not a pdf"))
	if err == nil {
		t.Fatal("expected error for non-PDF input")
	}
}

func TestExtractText_EmptyInput(t *testing.T) {
	_, err := ExtractText(bytes.NewReader(nil))
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestExtractText_TruncatedPDF(t *testing.T) {
	// Minimal PDF header followed by garbage — should fail gracefully.
	broken := []byte("%PDF-1.4\ntrailer garbage\n")
	_, err := ExtractText(bytes.NewReader(broken))
	if err == nil {
		t.Fatal("expected error for truncated PDF")
	}
}

// TestCheckUsable_RejectsThinExtraction covers a PDF whose fonts the
// extraction backend cannot decode: it returns the few runs of text it does
// understand (a page stamp, a running header) and nothing for every other
// page. The result is non-empty, so a "did we get any text" check would
// accept it as a successful extraction.
func TestCheckUsable_RejectsThinExtraction(t *testing.T) {
	tests := []struct {
		name       string
		result     string
		textPages  int
		totalPages int
	}{
		{
			name:       "one page stamp out of twelve",
			result:     "arXiv:0000.00000v1  [cs.LG]  1 Jan 2000",
			textPages:  1,
			totalPages: 12,
		},
		{
			name:       "running header only",
			result:     "Journal of Example Results, Vol 12, 2000",
			textPages:  1,
			totalPages: 20,
		},
		{
			name:       "no pages yielded text",
			result:     "",
			textPages:  0,
			totalPages: 8,
		},
		{
			name:       "just under the page ratio",
			result:     strings.Repeat("body text ", 200),
			textPages:  4,
			totalPages: 9,
		},
		{
			name:       "too few characters per page",
			result:     "short",
			textPages:  1,
			totalPages: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := checkUsable(tc.result, tc.textPages, tc.totalPages)
			if err == nil {
				t.Fatal("expected the extraction to be rejected")
			}
			if !errors.Is(err, ErrNoExtractableText) {
				t.Errorf("error = %v, want it to wrap ErrNoExtractableText", err)
			}
		})
	}
}

// TestCheckUsable_AcceptsRealPaper guards against the sanity checks
// rejecting genuine extractions, which would break working uploads.
func TestCheckUsable_AcceptsRealPaper(t *testing.T) {
	tests := []struct {
		name       string
		result     string
		textPages  int
		totalPages int
	}{
		{
			name:       "typical research paper",
			result:     strings.Repeat("reinforcement learning policy gradient ", 800),
			textPages:  12,
			totalPages: 12,
		},
		{
			name:       "single page with real content",
			result:     strings.Repeat("abstract body text ", 20),
			textPages:  1,
			totalPages: 1,
		},
		{
			name:       "exactly at the page ratio",
			result:     strings.Repeat("content ", 100),
			textPages:  5,
			totalPages: 10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := checkUsable(tc.result, tc.textPages, tc.totalPages); err != nil {
				t.Errorf("checkUsable rejected a usable extraction: %v", err)
			}
		})
	}
}

// TestCheckUsable_ErrorMentionsMeasurements keeps a rejected upload
// diagnosable from the logs alone.
func TestCheckUsable_ErrorMentionsMeasurements(t *testing.T) {
	err := checkUsable("running header text", 1, 12)
	if err == nil {
		t.Fatal("expected an error")
	}
	msg := err.Error()
	for _, want := range []string{"1", "12"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q does not mention %q", msg, want)
		}
	}
}
