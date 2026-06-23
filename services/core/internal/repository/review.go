package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"core/internal/model"
)

// ReviewQueueRepository persists ReviewQueueItem rows in MySQL.
type ReviewQueueRepository struct {
	db *sql.DB
}

// NewReviewQueueRepository returns a ReviewQueueRepository backed by the given *sql.DB.
func NewReviewQueueRepository(db *sql.DB) *ReviewQueueRepository {
	return &ReviewQueueRepository{db: db}
}

var ErrReviewItemNotFound = errors.New("review item not found")

// Create inserts a new review queue item.
func (r *ReviewQueueRepository) Create(ctx context.Context, item *model.ReviewQueueItem) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO review_queue (id, paper_id, extracted_nodes, extracted_edges, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		item.ID, item.PaperID, item.ExtractedNodes, item.ExtractedEdges, item.Status, item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert review item: %w", err)
	}
	return nil
}

// GetByID fetches a review item by ID.
func (r *ReviewQueueRepository) GetByID(ctx context.Context, id string) (*model.ReviewQueueItem, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, paper_id, extracted_nodes, extracted_edges, status, reviewer_id, review_notes, created_at, reviewed_at
		 FROM review_queue WHERE id = ?`, id,
	)
	return scanReviewItem(row)
}

// List returns a paginated list of review items, optionally filtered by status.
func (r *ReviewQueueRepository) List(ctx context.Context, offset, limit int, status string) ([]model.ReviewQueueItem, int, error) {
	countQuery := `SELECT COUNT(*) FROM review_queue`
	listQuery := `SELECT id, paper_id, extracted_nodes, extracted_edges, status, reviewer_id, review_notes, created_at, reviewed_at
		 FROM review_queue`

	var args []interface{}
	if status != "" {
		countQuery += ` WHERE status = ?`
		listQuery += ` WHERE status = ?`
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count review items: %w", err)
	}

	listQuery += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset) //nolint:gocritic

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list review items: %w", err)
	}
	defer rows.Close()

	items := make([]model.ReviewQueueItem, 0)
	for rows.Next() {
		var item model.ReviewQueueItem
		if err := rows.Scan(&item.ID, &item.PaperID, &item.ExtractedNodes, &item.ExtractedEdges, &item.Status, &item.ReviewerID, &item.ReviewNotes, &item.CreatedAt, &item.ReviewedAt); err != nil {
			return nil, 0, fmt.Errorf("scan review item: %w", err)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

// UpdateStatus updates the status, reviewer, notes, and reviewed_at of a review item.
func (r *ReviewQueueRepository) UpdateStatus(ctx context.Context, id, status, reviewerID, notes string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE review_queue SET status = ?, reviewer_id = ?, review_notes = ?, reviewed_at = ? WHERE id = ?`,
		status, reviewerID, notes, now, id,
	)
	if err != nil {
		return fmt.Errorf("update review status: %w", err)
	}
	return nil
}

// GetExtractedData returns the extracted nodes and edges for a review item.
func (r *ReviewQueueRepository) GetExtractedData(ctx context.Context, id string) (nodes json.RawMessage, edges json.RawMessage, err error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT extracted_nodes, extracted_edges FROM review_queue WHERE id = ?`, id,
	)
	if err := row.Scan(&nodes, &edges); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrReviewItemNotFound
		}
		return nil, nil, fmt.Errorf("get extracted data: %w", err)
	}
	return nodes, edges, nil
}

func scanReviewItem(row *sql.Row) (*model.ReviewQueueItem, error) {
	var item model.ReviewQueueItem
	err := row.Scan(&item.ID, &item.PaperID, &item.ExtractedNodes, &item.ExtractedEdges, &item.Status, &item.ReviewerID, &item.ReviewNotes, &item.CreatedAt, &item.ReviewedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReviewItemNotFound
		}
		return nil, fmt.Errorf("scan review item: %w", err)
	}
	return &item, nil
}
