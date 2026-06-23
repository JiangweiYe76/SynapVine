package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"core/internal/model"
)

// PaperRepository persists Paper rows in MySQL.
type PaperRepository struct {
	db *sql.DB
}

// NewPaperRepository returns a PaperRepository backed by the given *sql.DB.
func NewPaperRepository(db *sql.DB) *PaperRepository {
	return &PaperRepository{db: db}
}

var ErrPaperNotFound = errors.New("paper not found")

// Create inserts a new paper.
func (r *PaperRepository) Create(ctx context.Context, p *model.Paper) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO papers (id, title, authors, source_url, raw_text, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Title, p.Authors, p.SourceURL, p.RawText, p.Status, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert paper: %w", err)
	}
	return nil
}

// GetByID fetches a paper by ID.
func (r *PaperRepository) GetByID(ctx context.Context, id string) (*model.Paper, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, title, authors, source_url, raw_text, status, created_at, updated_at
		 FROM papers WHERE id = ?`, id,
	)
	return scanPaper(row)
}

// List returns a paginated list of papers.
func (r *PaperRepository) List(ctx context.Context, offset, limit int) ([]model.Paper, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM papers`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count papers: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, authors, source_url, raw_text, status, created_at, updated_at
		 FROM papers ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list papers: %w", err)
	}
	defer rows.Close()

	papers := make([]model.Paper, 0)
	for rows.Next() {
		var p model.Paper
		if err := rows.Scan(&p.ID, &p.Title, &p.Authors, &p.SourceURL, &p.RawText, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan paper: %w", err)
		}
		papers = append(papers, p)
	}
	return papers, total, rows.Err()
}

// Update applies a partial update.
func (r *PaperRepository) Update(ctx context.Context, id string, req *model.PaperUpdateRequest) (*model.Paper, error) {
	setClauses := []string{}
	args := []interface{}{}

	if req.Title != nil {
		setClauses = append(setClauses, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Authors != nil {
		setClauses = append(setClauses, "authors = ?")
		args = append(args, *req.Authors)
	}
	if req.SourceURL != nil {
		setClauses = append(setClauses, "source_url = ?")
		args = append(args, *req.SourceURL)
	}
	if req.Status != nil {
		setClauses = append(setClauses, "status = ?")
		args = append(args, *req.Status)
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, id)

	query := "UPDATE papers SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += " WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update paper: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return nil, ErrPaperNotFound
	}
	return r.GetByID(ctx, id)
}

// Delete removes a paper by ID.
func (r *PaperRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM papers WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete paper: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrPaperNotFound
	}
	return nil
}

func scanPaper(row *sql.Row) (*model.Paper, error) {
	var p model.Paper
	err := row.Scan(&p.ID, &p.Title, &p.Authors, &p.SourceURL, &p.RawText, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaperNotFound
		}
		return nil, fmt.Errorf("scan paper: %w", err)
	}
	return &p, nil
}
