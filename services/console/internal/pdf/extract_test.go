package pdf

import (
	"bytes"
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
