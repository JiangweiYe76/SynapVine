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

// ErrPaperNotFound is returned when a lookup matches no paper row.
var ErrPaperNotFound = errors.New("paper not found")

// Create inserts a new paper. Optional text columns are written as SQL
// NULL rather than "" — see nullableText for why that matters.
func (r *PaperRepository) Create(ctx context.Context, p *model.Paper) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO papers (id, title, authors, source_url, arxiv_id, raw_text, pdf_data, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Title, nullableText(p.Authors), nullableText(p.SourceURL), nullableText(p.ArxivID), p.RawText, p.PDFData, p.Status, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert paper: %w", err)
	}
	return nil
}

// GetByID fetches a paper by ID (without PDF binary).
func (r *PaperRepository) GetByID(ctx context.Context, id string) (*model.Paper, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+paperSelectColumns+` FROM papers WHERE id = ?`, id,
	)
	return scanPaper(row)
}

// GetByArxivID fetches a paper by its ArXiv identifier (without PDF
// binary). Returns ErrPaperNotFound when no paper carries that ID.
func (r *PaperRepository) GetByArxivID(ctx context.Context, arxivID string) (*model.Paper, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+paperSelectColumns+` FROM papers WHERE arxiv_id = ?`, arxivID,
	)
	return scanPaper(row)
}

// GetPDFData returns only the raw PDF bytes for a paper.
func (r *PaperRepository) GetPDFData(ctx context.Context, id string) ([]byte, error) {
	var data []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT pdf_data FROM papers WHERE id = ?`, id,
	).Scan(&data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaperNotFound
		}
		return nil, fmt.Errorf("get pdf data: %w", err)
	}
	if len(data) == 0 {
		return nil, ErrPaperNotFound
	}
	return data, nil
}

// List returns a paginated list of papers.
func (r *PaperRepository) List(ctx context.Context, offset, limit int) ([]model.Paper, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM papers`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count papers: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT `+paperSelectColumns+` FROM papers ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list papers: %w", err)
	}
	defer rows.Close()

	papers := make([]model.Paper, 0)
	for rows.Next() {
		p, err := scanPaperRow(rows)
		if err != nil {
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
		args = append(args, nullableText(*req.Authors))
	}
	if req.SourceURL != nil {
		setClauses = append(setClauses, "source_url = ?")
		args = append(args, nullableText(*req.SourceURL))
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

// paperSelectColumns is the column projection shared by every papers
// read. The PDF blob is projected as a boolean so large binaries never
// leave the database; Model.Paper.PDFData stays empty on reads.
const paperSelectColumns = `id, title, authors, source_url, arxiv_id, raw_text, (pdf_data IS NOT NULL AND LENGTH(pdf_data) > 0), status, created_at, updated_at`

// rowScanner is satisfied by both *sql.Row and *sql.Rows so a single
// scan implementation can back the single-row and list reads.
type rowScanner interface {
	Scan(dest ...any) error
}

// nullableText maps an empty optional string onto SQL NULL. The papers
// table declares authors, source_url and arxiv_id as NULLABLE, and
// arxiv_id carries a UNIQUE index: MySQL allows any number of NULLs but
// only one empty string. Storing an empty string for "no arXiv id"
// would therefore cap manual uploads at a single paper, so absence
// must be stored as NULL.
func nullableText(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func scanPaper(row *sql.Row) (*model.Paper, error) {
	p, err := scanPaperRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaperNotFound
		}
		return nil, fmt.Errorf("scan paper: %w", err)
	}
	return &p, nil
}

// scanPaperRow scans one paperSelectColumns row into a model.Paper.
// authors, source_url and arxiv_id are nullable in the schema, so they
// land in sql.NullString first and collapse to "" for the caller: NULL
// and "not set" mean the same thing for these columns, and no caller
// needs to tell them apart. Scanning them straight into a string fails
// with "converting NULL to string is unsupported".
func scanPaperRow(s rowScanner) (model.Paper, error) {
	var (
		p                           model.Paper
		authors, sourceURL, arxivID sql.NullString
	)
	err := s.Scan(
		&p.ID, &p.Title, &authors, &sourceURL, &arxivID, &p.RawText,
		&p.HasPDF, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return model.Paper{}, err
	}
	p.Authors = authors.String
	p.SourceURL = sourceURL.String
	p.ArxivID = arxivID.String
	return p, nil
}
