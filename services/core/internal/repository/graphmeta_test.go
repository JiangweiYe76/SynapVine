package repository

import (
	"testing"

	"core/internal/testutil"
)

func TestGraphMetaRepository_TouchAndRead(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)
	repo := NewGraphMetaRepository(neo)

	// Before any mutation the metadata node does not exist.
	_, ok, err := repo.LastUpdated(t.Context())
	if err != nil {
		t.Fatalf("last_updated read failed: %v", err)
	}
	if ok {
		t.Fatal("expected last_updated to be absent before first touch")
	}

	if err := repo.Touch(t.Context()); err != nil {
		t.Fatalf("touch failed: %v", err)
	}

	first, ok, err := repo.LastUpdated(t.Context())
	if err != nil || !ok {
		t.Fatalf("last_updated after touch: ok=%v err=%v", ok, err)
	}

	// Touching again must update the same singleton node, not create
	// another one, and the timestamp must move forward.
	if err := repo.Touch(t.Context()); err != nil {
		t.Fatalf("second touch failed: %v", err)
	}
	second, ok, err := repo.LastUpdated(t.Context())
	if err != nil || !ok {
		t.Fatalf("last_updated after second touch: ok=%v err=%v", ok, err)
	}
	if !second.After(first) {
		t.Errorf("expected timestamp to advance after second touch: first=%v second=%v", first, second)
	}

	// The metadata node is a singleton keyed by id.
	recs, err := neo.Query(t.Context(), `MATCH (m:GraphMeta) RETURN count(m) AS c`, nil)
	if err != nil {
		t.Fatalf("count meta nodes failed: %v", err)
	}
	if n := recs[0].Values[0].(int64); n != 1 {
		t.Errorf("expected exactly one GraphMeta node, got %d", n)
	}
}
