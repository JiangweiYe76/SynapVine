package repository

import (
	"testing"

	"core/internal/model"
	"core/internal/testutil"
)

func TestCommunityRepository_CreateAndGet(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	repo := NewCommunityRepository(neo)

	testutil.CleanupAllData(t, neo)

	req := model.CommunityCreateRequest{
		ID:     "c-test-1",
		Name:   "Test Community",
		Color:  "#ff0000",
		Domain: "ai",
	}

	id, err := repo.Create(t.Context(), req)
	if err != nil {
		t.Fatalf("create community failed: %v", err)
	}
	if id != req.ID {
		t.Errorf("returned id = %q, want %q", id, req.ID)
	}

	comm, err := repo.Get(t.Context(), req.ID)
	if err != nil {
		t.Fatalf("get community failed: %v", err)
	}
	if comm == nil {
		t.Fatal("expected community to exist")
	}
	if comm.Name != req.Name {
		t.Errorf("name = %q, want %q", comm.Name, req.Name)
	}
	if comm.Color != req.Color {
		t.Errorf("color = %q, want %q", comm.Color, req.Color)
	}
}

func TestCommunityRepository_List(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	repo := NewCommunityRepository(neo)

	testutil.CleanupAllData(t, neo)

	for _, c := range []model.CommunityCreateRequest{
		{ID: "c-a", Name: "Comm A", Color: "#111111", Domain: "ai"},
		{ID: "c-b", Name: "Comm B", Color: "#222222", Domain: "ai"},
		{ID: "c-c", Name: "Comm C", Color: "#333333", Domain: "ai"},
	} {
		if _, err := repo.Create(t.Context(), c); err != nil {
			t.Fatalf("create community failed: %v", err)
		}
	}

	comms, err := repo.List(t.Context())
	if err != nil {
		t.Fatalf("list communities failed: %v", err)
	}
	if len(comms) != 3 {
		t.Errorf("len(comms) = %d, want 3", len(comms))
	}
}

func TestCommunityRepository_Update(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	repo := NewCommunityRepository(neo)

	testutil.CleanupAllData(t, neo)

	req := model.CommunityCreateRequest{
		ID:     "c-update-1",
		Name:   "Original",
		Color:  "#000000",
		Domain: "ai",
	}
	if _, err := repo.Create(t.Context(), req); err != nil {
		t.Fatalf("create community failed: %v", err)
	}

	newName := "Updated"
	update := model.CommunityUpdateRequest{
		Name: &newName,
	}
	comm, err := repo.Update(t.Context(), req.ID, update)
	if err != nil {
		t.Fatalf("update community failed: %v", err)
	}
	if comm == nil {
		t.Fatal("expected updated community")
	}
	if comm.Name != "Updated" {
		t.Errorf("name = %q, want %q", comm.Name, "Updated")
	}
	if comm.Color != req.Color {
		t.Errorf("color changed unexpectedly to %q", comm.Color)
	}
}

func TestCommunityRepository_Delete(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	repo := NewCommunityRepository(neo)

	testutil.CleanupAllData(t, neo)

	req := model.CommunityCreateRequest{
		ID:     "c-del-1",
		Name:   "ToDelete",
		Color:  "#ffffff",
		Domain: "ai",
	}
	if _, err := repo.Create(t.Context(), req); err != nil {
		t.Fatalf("create community failed: %v", err)
	}

	ok, err := repo.Delete(t.Context(), req.ID)
	if err != nil {
		t.Fatalf("delete community failed: %v", err)
	}
	if !ok {
		t.Fatal("expected delete to succeed")
	}

	comm, err := repo.Get(t.Context(), req.ID)
	if err != nil {
		t.Fatalf("get community after delete failed: %v", err)
	}
	if comm != nil {
		t.Fatal("expected community to be deleted")
	}
}

func TestCommunityRepository_Exists(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	repo := NewCommunityRepository(neo)

	testutil.CleanupAllData(t, neo)

	exists, err := repo.Exists(t.Context(), "c-nonexistent")
	if err != nil {
		t.Fatalf("exists check failed: %v", err)
	}
	if exists {
		t.Error("expected nonexistent community to not exist")
	}

	req := model.CommunityCreateRequest{
		ID:     "c-exist-1",
		Name:   "Exist",
		Color:  "#123456",
		Domain: "ai",
	}
	if _, err := repo.Create(t.Context(), req); err != nil {
		t.Fatalf("create community failed: %v", err)
	}

	exists, err = repo.Exists(t.Context(), req.ID)
	if err != nil {
		t.Fatalf("exists check failed: %v", err)
	}
	if !exists {
		t.Error("expected existing community to exist")
	}
}
