package model

import "time"

// Paper represents a research paper uploaded for analysis.
type Paper struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Authors   string    `json:"authors"`
	SourceURL string    `json:"source_url"`
	RawText   string    `json:"raw_text"`
	PDFData   []byte    `json:"-"` // Not included in JSON responses
	HasPDF    bool      `json:"has_pdf"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PaperCreateRequest is the payload for creating a new paper.
type PaperCreateRequest struct {
	Title     string `json:"title"`
	Authors   string `json:"authors"`
	SourceURL string `json:"source_url"`
	RawText   string `json:"raw_text"`
	PDFBase64 string  `json:"pdf_base64,omitempty"` // Base64-encoded PDF
}

// PaperUpdateRequest is the payload for updating a paper.
type PaperUpdateRequest struct {
	Title     *string `json:"title"`
	Authors   *string `json:"authors"`
	SourceURL *string `json:"source_url"`
	Status    *string `json:"status"`
}

// PapersListResponse wraps a list of papers.
type PapersListResponse struct {
	Papers []Paper `json:"papers"`
	Total  int     `json:"total"`
}
