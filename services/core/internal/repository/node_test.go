package repository

import (
	"testing"

	"core/internal/model"
	"core/internal/testutil"
)

func TestNodeRepository_CreateAndGet(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	repo := NewNodeRepository(neo)

	testutil.CleanupAllData(t, neo)

	req := model.NodeCreateRequest{
		ID:             "test-transformer",
		Name:           "Transformer",
		Description:    "Attention is all you need",
		InfluenceScore: 9.5,
		FirstAppeared:  "2017-06",
		Milestones:     []string{"2017-06"},
	}

	if _, err := repo.Create(t.Context(), req); err != nil {
		t.Fatalf("create node failed: %v", err)
	}

	node, err := repo.Get(t.Context(), req.ID)
	if err != nil {
		t.Fatalf("get node failed: %v", err)
	}
	if node == nil {
		t.Fatal("expected node to exist")
	}
	if node.Name != req.Name {
		t.Errorf("name = %q, want %q", node.Name, req.Name)
	}
}

func TestNodeRepository_CreateGeneratesID(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	repo := NewNodeRepository(neo)

	testutil.CleanupAllData(t, neo)

	// Empty ID -> backend mints a fresh UUID and returns it.
	req := model.NodeCreateRequest{
		Name:           "Auto",
		Description:    "no id provided",
		InfluenceScore: 1.0,
		FirstAppeared:  "2024-01",
	}
	id, err := repo.Create(t.Context(), req)
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty id from server-side generation")
	}
	node, err := repo.Get(t.Context(), id)
	if err != nil {
		t.Fatalf("get node failed: %v", err)
	}
	if node == nil || node.ID != id {
		t.Fatalf("expected round-trip via generated id, got %+v", node)
	}
}

func TestNodeRepository_List(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	repo := NewNodeRepository(neo)

	testutil.CleanupAllData(t, neo)

	for _, n := range []model.NodeCreateRequest{
		{ID: "node-a", Name: "Alpha", Description: "First node", InfluenceScore: 1.0, FirstAppeared: "2000-01"},
		{ID: "node-b", Name: "Beta", Description: "Second node", InfluenceScore: 2.0, FirstAppeared: "2001-01"},
		{ID: "node-c", Name: "Gamma", Description: "Third node", InfluenceScore: 3.0, FirstAppeared: "2002-01"},
	} {
		if _, err := repo.Create(t.Context(), n); err != nil {
			t.Fatalf("create node failed: %v", err)
		}
	}

	nodes, total, err := repo.List(t.Context(), 0, 10)
	if err != nil {
		t.Fatalf("list nodes failed: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(nodes) != 3 {
		t.Errorf("len(nodes) = %d, want 3", len(nodes))
	}

	// Test pagination
	nodes, total, err = repo.List(t.Context(), 1, 1)
	if err != nil {
		t.Fatalf("list nodes with pagination failed: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(nodes) != 1 {
		t.Errorf("len(nodes) = %d, want 1", len(nodes))
	}
}

func TestNodeRepository_Search(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	repo := NewNodeRepository(neo)

	testutil.CleanupAllData(t, neo)

	for _, n := range []model.NodeCreateRequest{
		{ID: "search-a", Name: "Alpha Search", Description: "desc a", InfluenceScore: 1.0, FirstAppeared: "2000-01"},
		{ID: "search-b", Name: "Beta", Description: "desc beta", InfluenceScore: 2.0, FirstAppeared: "2001-01"},
	} {
		if _, err := repo.Create(t.Context(), n); err != nil {
			t.Fatalf("create node failed: %v", err)
		}
	}

	nodes, total, err := repo.Search(t.Context(), "alpha", 0, 10)
	if err != nil {
		t.Fatalf("search nodes failed: %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(nodes) != 1 || nodes[0].ID != "search-a" {
		t.Errorf("unexpected search result: %+v", nodes)
	}
}

func TestNodeRepository_Update(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	repo := NewNodeRepository(neo)

	testutil.CleanupAllData(t, neo)

	req := model.NodeCreateRequest{
		ID:             "update-node",
		Name:           "Original",
		Description:    "original desc",
		InfluenceScore: 5.0,
		FirstAppeared:  "2000-01",
	}
	if _, err := repo.Create(t.Context(), req); err != nil {
		t.Fatalf("create node failed: %v", err)
	}

	newName := "Updated"
	update := model.NodeUpdateRequest{
		Name: &newName,
	}
	node, err := repo.Update(t.Context(), req.ID, update)
	if err != nil {
		t.Fatalf("update node failed: %v", err)
	}
	if node == nil {
		t.Fatal("expected updated node")
	}
	if node.Name != "Updated" {
		t.Errorf("name = %q, want %q", node.Name, "Updated")
	}
}

func TestNodeRepository_Delete(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	repo := NewNodeRepository(neo)

	testutil.CleanupAllData(t, neo)

	req := model.NodeCreateRequest{
		ID:             "delete-node",
		Name:           "ToDelete",
		Description:    "delete me",
		InfluenceScore: 1.0,
		FirstAppeared:  "2000-01",
	}
	if _, err := repo.Create(t.Context(), req); err != nil {
		t.Fatalf("create node failed: %v", err)
	}

	ok, err := repo.Delete(t.Context(), req.ID)
	if err != nil {
		t.Fatalf("delete node failed: %v", err)
	}
	if !ok {
		t.Fatal("expected delete to succeed")
	}

	node, err := repo.Get(t.Context(), req.ID)
	if err != nil {
		t.Fatalf("get node after delete failed: %v", err)
	}
	if node != nil {
		t.Fatal("expected node to be deleted")
	}
}

func TestNodeRepository_Exists(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	repo := NewNodeRepository(neo)

	testutil.CleanupAllData(t, neo)

	exists, err := repo.Exists(t.Context(), "nonexistent")
	if err != nil {
		t.Fatalf("exists check failed: %v", err)
	}
	if exists {
		t.Error("expected nonexistent node to not exist")
	}

	req := model.NodeCreateRequest{
		ID:             "exist-node",
		Name:           "Exist",
		Description:    "exists",
		InfluenceScore: 1.0,
		FirstAppeared:  "2000-01",
	}
	if _, err := repo.Create(t.Context(), req); err != nil {
		t.Fatalf("create node failed: %v", err)
	}

	exists, err = repo.Exists(t.Context(), req.ID)
	if err != nil {
		t.Fatalf("exists check failed: %v", err)
	}
	if !exists {
		t.Error("expected existing node to exist")
	}
}
