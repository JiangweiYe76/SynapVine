// Package pdf provides PDF text extraction utilities.
package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	pdflib "github.com/ledongthuc/pdf"
)

// ErrNoExtractableText reports that a PDF yielded too little text to be
// trusted as the document's content. It covers genuinely text-free files
// (scanned documents) and files whose fonts the extraction backend
// cannot decode. The backend signals the latter case by returning empty
// pages rather than an error, so ExtractText has to detect the shortfall
// itself.
var ErrNoExtractableText = errors.New("no extractable text layer in PDF")

// minTextPageRatio is the smallest share of pages that must yield text
// for a PDF to be considered readable. When a document's fonts cannot be
// decoded the backend still returns the few runs of text it does
// understand — page stamps, running headers — which is enough to make
// the result non-empty and therefore look like a success.
const minTextPageRatio = 0.5

// minCharsPerPage is a floor on extraction density. Research papers sit
// far above 1000 characters per page, so a document averaging fewer than
// this has been decoded only in fragments.
const minCharsPerPage = 50

// ExtractText reads a PDF from the given reader and returns the extracted
// plain-text content. The caller can pass an io.ReadCloser obtained from
// a multipart file header.
//
// It returns ErrNoExtractableText when too little text could be read to
// be trusted as the document's content. Callers must treat that as a
// failure instead of storing the fragment: a partially decoded PDF
// otherwise reaches the LLM and produces fabricated concepts.
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

	totalPages := reader.NumPage()
	if totalPages == 0 {
		return "", fmt.Errorf("%w: document has no pages", ErrNoExtractableText)
	}

	var sb strings.Builder
	textPages := 0
	for i := 1; i <= totalPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		// A page that errors, or that decodes to nothing, contributes no
		// text. Undecodable fonts surface as empty pages rather than as
		// errors, so both cases are tallied the same way.
		text, err := page.GetPlainText(nil)
		if err != nil || strings.TrimSpace(text) == "" {
			continue
		}
		textPages++
		sb.WriteString(text)
		if i < totalPages {
			sb.WriteString("\n")
		}
	}

	result := strings.TrimSpace(sb.String())
	if err := checkUsable(result, textPages, totalPages); err != nil {
		return "", err
	}
	return result, nil
}

// checkUsable rejects extraction results too thin to be the document's
// real content. The measurements appear in the error so a rejected
// upload can be diagnosed from the logs alone.
func checkUsable(result string, textPages, totalPages int) error {
	chars := len([]rune(result))
	if textPages == 0 {
		return fmt.Errorf("%w: 0 of %d pages yielded text", ErrNoExtractableText, totalPages)
	}
	if ratio := float64(textPages) / float64(totalPages); ratio < minTextPageRatio {
		return fmt.Errorf("%w: only %d of %d pages yielded text (%.0f%%, need at least %.0f%%)",
			ErrNoExtractableText, textPages, totalPages, ratio*100, minTextPageRatio*100)
	}
	if density := chars / totalPages; density < minCharsPerPage {
		return fmt.Errorf("%w: %d characters over %d pages (%d per page, need at least %d)",
			ErrNoExtractableText, chars, totalPages, density, minCharsPerPage)
	}
	return nil
}
