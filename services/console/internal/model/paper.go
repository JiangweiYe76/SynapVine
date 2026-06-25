package model

import (
	"encoding/json"
	"time"
)

// Paper represents a research paper (mirrors core model).
type Paper struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Authors   string    `json:"authors"`
	SourceURL string    `json:"source_url"`
	RawText   string    `json:"raw_text"`
	HasPDF    bool      `json:"has_pdf"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PaperCreateRequest is the payload for creating a paper via core.
type PaperCreateRequest struct {
	Title     string `json:"title"`
	Authors   string `json:"authors"`
	SourceURL string `json:"source_url"`
	RawText   string `json:"raw_text"`
	PDFBase64 string  `json:"pdf_base64,omitempty"`
}

// PaperUpdateRequest is the payload for updating a paper via core.
type PaperUpdateRequest struct {
	Title     *string `json:"title,omitempty"`
	Authors   *string `json:"authors,omitempty"`
	SourceURL *string `json:"source_url,omitempty"`
	Status    *string `json:"status,omitempty"`
}

// PapersListResponse wraps a list of papers.
type PapersListResponse struct {
	Papers []Paper `json:"papers"`
	Total  int     `json:"total"`
}

// ReviewQueueItem represents a review queue entry (mirrors core model).
type ReviewQueueItem struct {
	ID             string          `json:"id"`
	PaperID        string          `json:"paper_id"`
	ExtractedNodes json.RawMessage `json:"extracted_nodes"`
	ExtractedEdges json.RawMessage `json:"extracted_edges"`
	Status         string          `json:"status"`
	ReviewerID     *string         `json:"reviewer_id"`
	ReviewNotes    *string         `json:"review_notes"`
	CreatedAt      time.Time       `json:"created_at"`
	ReviewedAt     *time.Time      `json:"reviewed_at"`
}

// ReviewQueueListResponse wraps a list of review items.
type ReviewQueueListResponse struct {
	Items []ReviewQueueItem `json:"items"`
	Total int               `json:"total"`
}
