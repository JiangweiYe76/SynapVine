package repository

import (
	"testing"

	"core/internal/db"
	"core/internal/model"
	"core/internal/testutil"
)

// seedConcept inserts a single Concept node so edges can target it.
func seedConcept(t *testing.T, neo *db.Neo4j, id, name string) {
	t.Helper()
	if err := neo.Execute(t.Context(),
		`CREATE (:Concept {id: $id, name: $name, source: 'manual', status: 'active', created_at: datetime()})`,
		map[string]any{"id": id, "name": name},
	); err != nil {
		t.Fatalf("seed concept %q failed: %v", id, err)
	}
}

func TestEdgeRepository_CreateAndGet(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)

	repo := NewEdgeRepository(neo)
	seedConcept(t, neo, "src", "Source")
	seedConcept(t, neo, "dst", "Target")

	req := model.EdgeCreateRequest{
		Source: "src", Target: "dst", Weight: 0.5, Relation: "based_on",
	}
	if err := repo.Create(t.Context(), req); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}

	edge, err := repo.Get(t.Context(), "src", "dst")
	if err != nil {
		t.Fatalf("get edge failed: %v", err)
	}
	if edge == nil {
		t.Fatal("expected edge to exist")
	}
	if edge.Relation != "based_on" || edge.Weight != 0.5 {
		t.Errorf("unexpected edge: %+v", edge)
	}
}

func TestEdgeRepository_Get_NotFound(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)

	repo := NewEdgeRepository(neo)
	edge, err := repo.Get(t.Context(), "missing", "alsomissing")
	if err != nil {
		t.Fatalf("get edge failed: %v", err)
	}
	if edge != nil {
		t.Errorf("expected nil edge, got %+v", edge)
	}
}

func TestEdgeRepository_List(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)

	repo := NewEdgeRepository(neo)
	seedConcept(t, neo, "a", "A")
	seedConcept(t, neo, "b", "B")
	seedConcept(t, neo, "c", "C")

	for _, e := range []model.EdgeCreateRequest{
		{Source: "a", Target: "b", Weight: 0.1, Relation: "r1"},
		{Source: "a", Target: "c", Weight: 0.2, Relation: "r2"},
		{Source: "b", Target: "c", Weight: 0.3, Relation: "r3"},
	} {
		if err := repo.Create(t.Context(), e); err != nil {
			t.Fatalf("create edge failed: %v", err)
		}
	}

	edges, total, err := repo.List(t.Context(), 0, 10)
	if err != nil {
		t.Fatalf("list edges failed: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(edges) != 3 {
		t.Errorf("len(edges) = %d, want 3", len(edges))
	}

	// Pagination
	edges, total, err = repo.List(t.Context(), 1, 1)
	if err != nil {
		t.Fatalf("list with pagination failed: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(edges) != 1 {
		t.Errorf("len(edges) = %d, want 1", len(edges))
	}
}

func TestEdgeRepository_Search(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)

	repo := NewEdgeRepository(neo)
	seedConcept(t, neo, "bert", "BERT")
	seedConcept(t, neo, "gpt", "GPT")

	if err := repo.Create(t.Context(), model.EdgeCreateRequest{
		Source: "bert", Target: "gpt", Weight: 0.7, Relation: "influenced",
	}); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}
	if err := repo.Create(t.Context(), model.EdgeCreateRequest{
		Source: "gpt", Target: "bert", Weight: 0.4, Relation: "evolved_to",
	}); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}

	edges, total, err := repo.Search(t.Context(), "influ", 0, 10)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if total != 1 || len(edges) != 1 {
		t.Errorf("expected 1 match, got total=%d len=%d", total, len(edges))
	}
}

func TestEdgeRepository_Update(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)

	repo := NewEdgeRepository(neo)
	seedConcept(t, neo, "a", "A")
	seedConcept(t, neo, "b", "B")
	if err := repo.Create(t.Context(), model.EdgeCreateRequest{
		Source: "a", Target: "b", Weight: 0.1, Relation: "old",
	}); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}

	newWeight := 0.9
	newRel := "new"
	edge, err := repo.Update(t.Context(), "a", "b", model.EdgeUpdateRequest{
		Weight: &newWeight, Relation: &newRel,
	})
	if err != nil {
		t.Fatalf("update edge failed: %v", err)
	}
	if edge == nil {
		t.Fatal("expected updated edge")
	}
	if edge.Weight != 0.9 || edge.Relation != "new" {
		t.Errorf("unexpected updated edge: %+v", edge)
	}
}

func TestEdgeRepository_Update_NotFound(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)

	repo := NewEdgeRepository(neo)
	seedConcept(t, neo, "a", "A")
	seedConcept(t, neo, "b", "B")

	newWeight := 0.9
	edge, err := repo.Update(t.Context(), "a", "b", model.EdgeUpdateRequest{Weight: &newWeight})
	if err != nil {
		t.Fatalf("update edge failed: %v", err)
	}
	if edge != nil {
		t.Errorf("expected nil edge, got %+v", edge)
	}
}

func TestEdgeRepository_Delete(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)

	repo := NewEdgeRepository(neo)
	seedConcept(t, neo, "a", "A")
	seedConcept(t, neo, "b", "B")
	if err := repo.Create(t.Context(), model.EdgeCreateRequest{
		Source: "a", Target: "b", Weight: 0.5, Relation: "r",
	}); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}

	ok, err := repo.Delete(t.Context(), "a", "b")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if !ok {
		t.Fatal("expected delete to succeed")
	}

	edge, err := repo.Get(t.Context(), "a", "b")
	if err != nil {
		t.Fatalf("get after delete failed: %v", err)
	}
	if edge != nil {
		t.Errorf("expected edge to be gone, got %+v", edge)
	}
}

func TestEdgeRepository_Delete_NotFound(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)

	repo := NewEdgeRepository(neo)
	ok, err := repo.Delete(t.Context(), "missing", "x")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if ok {
		t.Error("expected ok=false for missing edge")
	}
}

func TestEdgeRepository_Exists(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)

	repo := NewEdgeRepository(neo)
	seedConcept(t, neo, "a", "A")
	seedConcept(t, neo, "b", "B")
	if err := repo.Create(t.Context(), model.EdgeCreateRequest{
		Source: "a", Target: "b", Weight: 0.5, Relation: "r",
	}); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}

	exists, err := repo.Exists(t.Context(), "a", "b")
	if err != nil {
		t.Fatalf("exists failed: %v", err)
	}
	if !exists {
		t.Error("expected edge to exist")
	}

	exists, err = repo.Exists(t.Context(), "a", "missing")
	if err != nil {
		t.Fatalf("exists failed: %v", err)
	}
	if exists {
		t.Error("expected edge to not exist")
	}
}
