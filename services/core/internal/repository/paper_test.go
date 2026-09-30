package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"core/internal/model"
	"core/internal/testutil"

	"github.com/google/uuid"
)

// newPaper builds a Paper fixture. Empty optional fields are passed
// through verbatim so tests can exercise the "absent" case.
func newPaper(title string) *model.Paper {
	now := time.Now()
	return &model.Paper{
		ID:        uuid.New().String(),
		Title:     title,
		Status:    "uploaded",
		RawText:   "raw text for " + title,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// insertPaperWithNullOptionals writes a papers row bypassing the
// repository, leaving authors, source_url and arxiv_id as SQL NULL.
// NULL is the state the schema itself produces: the migration that adds
// arxiv_id as a NULLABLE column backfills NULL into the rows it does not
// supply a value for, so a NULL optional is ordinary data rather than an
// anomaly.
func insertPaperWithNullOptionals(t *testing.T, conn *sql.DB, title string) string {
	t.Helper()
	id := uuid.New().String()
	_, err := conn.ExecContext(context.Background(),
		`INSERT INTO papers (id, title, raw_text, status) VALUES (?, ?, ?, 'uploaded')`,
		id, title, "raw text for "+title,
	)
	if err != nil {
		t.Fatalf("insert paper with NULL optionals: %v", err)
	}
	return id
}

// rawPaperColumn reads one papers column directly so tests can assert
// what is actually stored in MySQL rather than what the repository
// reports back.
func rawPaperColumn(t *testing.T, conn *sql.DB, id, column string) sql.NullString {
	t.Helper()
	var v sql.NullString
	err := conn.QueryRowContext(context.Background(),
		`SELECT `+column+` FROM papers WHERE id = ?`, id).Scan(&v)
	if err != nil {
		t.Fatalf("raw read papers.%s: %v", column, err)
	}
	return v
}

// TestPaperListHandlesNullOptionalColumns covers reading papers whose
// authors, source_url and arxiv_id columns are NULL. Those columns are
// NULLABLE in the schema, and scanning a NULL into a plain Go string
// fails, so a single NULL row must not fail the whole listing.
func TestPaperListHandlesNullOptionalColumns(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewPaperRepository(conn)
	ctx := context.Background()

	sparseID := insertPaperWithNullOptionals(t, conn, "sparse-row")

	// A fully populated row must still be readable alongside the NULL one.
	populated := newPaper("populated-row")
	populated.Authors = "A. Author"
	populated.SourceURL = "https://example.org/paper"
	populated.ArxivID = "2501.00001"
	if err := repo.Create(ctx, populated); err != nil {
		t.Fatalf("Create populated: %v", err)
	}

	papers, total, err := repo.List(ctx, 0, 50)
	if err != nil {
		t.Fatalf("List with a NULL row present: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(papers) != 2 {
		t.Fatalf("len(papers) = %d, want 2", len(papers))
	}

	byID := make(map[string]model.Paper, len(papers))
	for _, p := range papers {
		byID[p.ID] = p
	}

	// The NULL row must surface as empty strings, not an error.
	sparse, ok := byID[sparseID]
	if !ok {
		t.Fatalf("row with NULL optionals missing from List result")
	}
	if sparse.Authors != "" || sparse.SourceURL != "" || sparse.ArxivID != "" {
		t.Errorf("NULL optionals = (%q, %q, %q), want all empty",
			sparse.Authors, sparse.SourceURL, sparse.ArxivID)
	}
	if sparse.Title != "sparse-row" {
		t.Errorf("Title = %q, want %q", sparse.Title, "sparse-row")
	}

	// The populated row must keep its values.
	if got := byID[populated.ID].ArxivID; got != "2501.00001" {
		t.Errorf("ArxivID = %q, want %q", got, "2501.00001")
	}
	if got := byID[populated.ID].Authors; got != "A. Author" {
		t.Errorf("Authors = %q, want %q", got, "A. Author")
	}

	// GetByID goes through a different code path and must not regress.
	got, err := repo.GetByID(ctx, sparseID)
	if err != nil {
		t.Fatalf("GetByID on a NULL row: %v", err)
	}
	if got.ArxivID != "" || got.Authors != "" || got.SourceURL != "" {
		t.Errorf("GetByID optionals = (%q, %q, %q), want all empty",
			got.ArxivID, got.Authors, got.SourceURL)
	}
}

// TestPaperCreateStoresNullForAbsentOptionals pins the write side: an
// absent optional value must land as SQL NULL, not "".
func TestPaperCreateStoresNullForAbsentOptionals(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewPaperRepository(conn)
	ctx := context.Background()

	p := newPaper("no-optionals")
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	for _, column := range []string{"authors", "source_url", "arxiv_id"} {
		if v := rawPaperColumn(t, conn, p.ID, column); v.Valid {
			t.Errorf("papers.%s = %q, want SQL NULL", column, v.String)
		}
	}

	// Reading it back must yield empty strings.
	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Authors != "" || got.SourceURL != "" || got.ArxivID != "" {
		t.Errorf("optionals = (%q, %q, %q), want all empty",
			got.Authors, got.SourceURL, got.ArxivID)
	}
}

// TestPaperCreateAllowsManyPapersWithoutArxivID covers inserting several
// papers that carry no arXiv identifier. uq_papers_arxiv_id is a UNIQUE
// index and MySQL permits any number of NULLs but only one empty string,
// so absence must be stored as NULL to keep manual uploads working.
func TestPaperCreateAllowsManyPapersWithoutArxivID(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewPaperRepository(conn)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		p := newPaper("manual-upload")
		p.Authors = ""
		p.SourceURL = ""
		p.ArxivID = ""
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("Create manual upload %d: %v", i, err)
		}
	}

	_, total, err := repo.List(ctx, 0, 50)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
}

// TestPaperArxivIDStillDeduplicates guards the behaviour the UNIQUE
// index exists for: normalizing NULL must not weaken arXiv dedup.
func TestPaperArxivIDStillDeduplicates(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewPaperRepository(conn)
	ctx := context.Background()

	first := newPaper("arxiv-first")
	first.ArxivID = "2501.00002"
	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("Create first: %v", err)
	}

	dup := newPaper("arxiv-duplicate")
	dup.ArxivID = "2501.00002"
	if err := repo.Create(ctx, dup); err == nil {
		t.Errorf("Create duplicate arxiv_id: expected a unique-constraint error, got nil")
	}

	found, err := repo.GetByArxivID(ctx, "2501.00002")
	if err != nil {
		t.Fatalf("GetByArxivID: %v", err)
	}
	if found.ID != first.ID {
		t.Errorf("GetByArxivID returned %q, want %q", found.ID, first.ID)
	}

	if _, err := repo.GetByArxivID(ctx, "does-not-exist"); err == nil {
		t.Errorf("GetByArxivID for an unknown id: expected ErrPaperNotFound, got nil")
	}
}

// TestPaperUpdateClearsOptionalToNull checks the partial-update path
// stores a cleared value as NULL too, so repeated edits cannot walk
// authors/source_url back to "".
func TestPaperUpdateClearsOptionalToNull(t *testing.T) {
	conn := testutil.NewTestMySQL(t)
	repo := NewPaperRepository(conn)
	ctx := context.Background()

	p := newPaper("update-optionals")
	p.Authors = "A. Author"
	p.SourceURL = "https://example.org/before"
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	empty := ""
	got, err := repo.Update(ctx, p.ID, &model.PaperUpdateRequest{
		Authors:   &empty,
		SourceURL: &empty,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.Authors != "" || got.SourceURL != "" {
		t.Errorf("updated optionals = (%q, %q), want all empty", got.Authors, got.SourceURL)
	}

	for _, column := range []string{"authors", "source_url"} {
		if v := rawPaperColumn(t, conn, p.ID, column); v.Valid {
			t.Errorf("papers.%s = %q, want SQL NULL", column, v.String)
		}
	}
}
