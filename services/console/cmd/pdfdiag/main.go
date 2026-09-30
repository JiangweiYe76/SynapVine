// Command pdfdiag reports why PDF text extraction yields too little text.
// It shows per-page extraction outcomes and whether the result passes the
// library's sanity checks, which is the fastest way to tell a genuinely
// text-free PDF from one whose fonts the extraction backend cannot decode.
//
// Usage:
//
//	go run ./cmd/pdfdiag <file.pdf> [file.pdf ...]
package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"console/internal/pdf"

	pdflib "github.com/ledongthuc/pdf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: pdfdiag <file.pdf> [file.pdf ...]")
		os.Exit(1)
	}
	for _, path := range os.Args[1:] {
		report(path)
	}
}

func report(path string) {
	fmt.Printf("=== %s\n", path)

	buf, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("  read: %v\n\n", err)
		return
	}
	fmt.Printf("  bytes: %d\n", len(buf))

	reader, err := pdflib.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		fmt.Printf("  open: %v\n\n", err)
		return
	}

	totalPages := reader.NumPage()
	if totalPages == 0 {
		fmt.Printf("  pages: 0 (no content)\n\n")
		return
	}

	var sb strings.Builder
	textPages := 0
	shown := 0
	const maxSamples = 3

	for i := 1; i <= totalPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			if shown < maxSamples {
				fmt.Printf("  page %2d: ERROR %v\n", i, err)
				shown++
			}
			continue
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		textPages++
		sb.WriteString(text)
		if i < totalPages {
			sb.WriteString("\n")
		}
		if shown < maxSamples || i == totalPages {
			snippet := strings.TrimSpace(text)
			if len(snippet) > 70 {
				snippet = snippet[:70] + "..."
			}
			fmt.Printf("  page %2d: %5d chars | %q\n", i, len(text), snippet)
			shown++
		}
	}

	result := strings.TrimSpace(sb.String())
	fmt.Printf("  pages: %d, with text: %d (%.0f%%)\n", totalPages, textPages, pct(textPages, totalPages))
	fmt.Printf("  extracted: %d chars (%d per page)\n", len([]rune(result)), len([]rune(result))/totalPages)

	// Reuse the library's own verdict so the diagnostic cannot drift from
	// what an upload would actually do.
	if _, err := pdf.ExtractText(bytes.NewReader(buf)); err != nil {
		fmt.Printf("  verdict: REJECTED — %v\n", err)
		fmt.Printf("          the upload would fail with 422 pdf_text_not_extractable\n\n")
		return
	}
	fmt.Printf("  verdict: accepted — the upload would store this text\n\n")
}

func pct(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
