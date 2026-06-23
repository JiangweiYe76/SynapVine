package model

import (
	"encoding/json"
	"time"
)

// ReviewQueueItem represents a pending extraction result awaiting human review.
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

// ReviewQueueSubmitRequest is the payload for submitting a new review item.
type ReviewQueueSubmitRequest struct {
	PaperID        string          `json:"paper_id"`
	ExtractedNodes json.RawMessage `json:"extracted_nodes"`
	ExtractedEdges json.RawMessage `json:"extracted_edges"`
}

// ReviewQueueApproveRequest is the payload for approving a review item.
type ReviewQueueApproveRequest struct {
	ReviewerID  string `json:"reviewer_id"`
	ReviewNotes string `json:"review_notes"`
}

// ReviewQueueRejectRequest is the payload for rejecting a review item.
type ReviewQueueRejectRequest struct {
	ReviewerID  string `json:"reviewer_id"`
	ReviewNotes string `json:"review_notes"`
}

// ReviewQueueListResponse wraps a list of review items.
type ReviewQueueListResponse struct {
	Items []ReviewQueueItem `json:"items"`
	Total int               `json:"total"`
}
