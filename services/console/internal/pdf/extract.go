// Package pdf provides PDF text extraction utilities.
package pdf

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	pdflib "github.com/ledongthuc/pdf"
)

// ExtractText reads a PDF from the given reader and returns the extracted
// plain-text content. The caller can pass an io.ReadCloser obtained from
// a multipart file header.
func ExtractText(r io.Reader) (string, error) {
	// ledongthuc/pdf requires an io.ReadSeeker, so we buffer the entire
	// content first.
	buf, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read pdf content: %w", err)
	}

	reader, err := pdflib.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}

	var sb strings.Builder
	totalPages := reader.NumPage()
	for i := 1; i <= totalPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			// Some pages may fail to extract; skip them rather than
			// aborting the entire upload.
			continue
		}
		sb.WriteString(text)
		if i < totalPages {
			sb.WriteString("\n")
		}
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		return "", fmt.Errorf("no text could be extracted from the PDF")
	}
	return result, nil
}
