package service

import (
	"testing"

	"core/internal/db"
	"core/internal/model"
	"core/internal/repository"
	"core/internal/testutil"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// countConcepts returns the total number of Concept nodes in the graph.
func countConcepts(t *testing.T, neo *db.Neo4j) int {
	t.Helper()
	recs, err := neo.Query(t.Context(), `MATCH (n:Concept) RETURN count(n) AS c`, nil)
	if err != nil {
		t.Fatalf("count concepts failed: %v", err)
	}
	return int(recordInt(recs[0], "c"))
}

// countEdges returns the total number of RELATES_TO relationships.
func countEdges(t *testing.T, neo *db.Neo4j) int {
	t.Helper()
	recs, err := neo.Query(t.Context(), `MATCH ()-[r:RELATES_TO]->() RETURN count(r) AS c`, nil)
	if err != nil {
		t.Fatalf("count edges failed: %v", err)
	}
	return int(recordInt(recs[0], "c"))
}

func recordInt(rec *neo4j.Record, key string) int64 {
	v, ok := rec.Get(key)
	if !ok || v == nil {
		return 0
	}
	if n, ok := v.(int64); ok {
		return n
	}
	return 0
}

func recordFloat(rec *neo4j.Record, key string) float64 {
	v, ok := rec.Get(key)
	if !ok || v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

// TestMerge_CreatesNewNodesAndEdges verifies that a fresh merge creates
// the nodes and edges with the expected provenance properties.
func TestMerge_CreatesNewNodesAndEdges(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)
	svc := NewMergeService(neo)

	nodes := []model.ExtractedNode{
		{Name: "Transformer", Description: "Attention-based model", Relevance: 9.0},
		{Name: "BERT", Description: "Encoder-only transformer", Relevance: 8.0},
	}
	edges := []model.ExtractedEdge{
		{Source: "Transformer", Target: "BERT", Relation: "inspired", Weight: 0.8},
	}

	res, err := svc.Merge(t.Context(), "paper-1", nodes, edges)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if res.CreatedNodes != 2 {
		t.Errorf("created_nodes = %d, want 2", res.CreatedNodes)
	}
	if res.ReusedNodes != 0 {
		t.Errorf("reused_nodes = %d, want 0", res.ReusedNodes)
	}
	if res.CreatedEdges != 1 {
		t.Errorf("created_edges = %d, want 1", res.CreatedEdges)
	}
	if res.SkippedEdges != 0 {
		t.Errorf("skipped_edges = %d, want 0", res.SkippedEdges)
	}

	// Nodes carry source='extraction' and source_paper_id.
	recs, err := neo.Query(t.Context(), `
		MATCH (n:Concept {source: 'extraction'})
		RETURN n.name AS name, n.source_paper_id AS pid, n.influence_score AS score
		ORDER BY name`, nil)
	if err != nil {
		t.Fatalf("query extraction nodes failed: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 extraction nodes, got %d", len(recs))
	}
	for _, rec := range recs {
		if got := recordString(rec, "pid"); got != "paper-1" {
			t.Errorf("source_paper_id = %q, want paper-1", got)
		}
	}
	// Transformer relevance 9.0 -> influence_score 0.9.
	if got := recordFloat(recs[1], "score"); got < 0.89 || got > 0.91 {
		t.Errorf("Transformer influence_score = %v, want ~0.9", got)
	}

	// Edge connects the two nodes with the given relation.
	erecs, err := neo.Query(t.Context(), `
		MATCH (s:Concept)-[r:RELATES_TO]->(t:Concept)
		RETURN s.name AS s, t.name AS t, r.relation AS rel, r.weight AS w`, nil)
	if err != nil {
		t.Fatalf("query edge failed: %v", err)
	}
	if len(erecs) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(erecs))
	}
	if got := recordString(erecs[0], "s"); got != "Transformer" {
		t.Errorf("edge source = %q, want Transformer", got)
	}
	if got := recordString(erecs[0], "t"); got != "BERT" {
		t.Errorf("edge target = %q, want BERT", got)
	}
	if got := recordString(erecs[0], "rel"); got != "inspired" {
		t.Errorf("edge relation = %q, want inspired", got)
	}
}

// TestMerge_ReusesExistingNodesByCaseInsensitiveName verifies that a
// node whose name matches case-insensitively is reused rather than
// duplicated, and that edges connect to the original node.
func TestMerge_ReusesExistingNodesByCaseInsensitiveName(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)

	// Pre-create a manually-authored "Transformer" node.
	repo := repository.NewNodeRepository(neo)
	origID, err := repo.Create(t.Context(), model.NodeCreateRequest{
		Name:           "Transformer",
		Description:    "original",
		InfluenceScore: 9.5,
		FirstAppeared:  "2017-06",
	})
	if err != nil {
		t.Fatalf("pre-create node failed: %v", err)
	}

	svc := NewMergeService(neo)
	nodes := []model.ExtractedNode{
		{Name: "transformer", Description: "from paper", Relevance: 5.0}, // lowercase, should reuse
		{Name: "BERT", Description: "new", Relevance: 7.0},
	}
	edges := []model.ExtractedEdge{
		{Source: "transformer", Target: "BERT", Relation: "extends", Weight: 0.5},
	}

	res, err := svc.Merge(t.Context(), "paper-2", nodes, edges)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if res.CreatedNodes != 1 {
		t.Errorf("created_nodes = %d, want 1 (BERT only)", res.CreatedNodes)
	}
	if res.ReusedNodes != 1 {
		t.Errorf("reused_nodes = %d, want 1 (transformer)", res.ReusedNodes)
	}
	if res.CreatedEdges != 1 {
		t.Errorf("created_edges = %d, want 1", res.CreatedEdges)
	}

	// Total nodes = 2 (original Transformer + new BERT), no duplicate.
	if got := countConcepts(t, neo); got != 2 {
		t.Errorf("total concepts = %d, want 2 (no duplicate)", got)
	}

	// Edge connects the ORIGINAL Transformer (by id) to BERT.
	erecs, err := neo.Query(t.Context(), `
		MATCH (s:Concept {id: $id})-[r:RELATES_TO]->(t:Concept)
		RETURN t.name AS t`, map[string]any{"id": origID})
	if err != nil {
		t.Fatalf("query edge from original node failed: %v", err)
	}
	if len(erecs) != 1 || recordString(erecs[0], "t") != "BERT" {
		t.Errorf("expected edge from original Transformer to BERT, got %v records", len(erecs))
	}
}

// TestMerge_DoesNotOverwriteExistingFields verifies that reusing an
// existing node leaves its description and influence_score untouched.
func TestMerge_DoesNotOverwriteExistingFields(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)

	repo := repository.NewNodeRepository(neo)
	if _, err := repo.Create(t.Context(), model.NodeCreateRequest{
		Name:           "GAN",
		Description:    "original description",
		InfluenceScore: 7.5,
		FirstAppeared:  "2014-06",
	}); err != nil {
		t.Fatalf("pre-create node failed: %v", err)
	}

	svc := NewMergeService(neo)
	nodes := []model.ExtractedNode{
		{Name: "gan", Description: "SHOULD NOT OVERWRITE", Relevance: 1.0},
	}
	if _, err := svc.Merge(t.Context(), "paper-3", nodes, nil); err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	recs, err := neo.Query(t.Context(), `
		MATCH (n:Concept)
		WHERE toLower(n.name) = 'gan'
		RETURN n.description AS d, n.influence_score AS score, n.source AS source`, nil)
	if err != nil {
		t.Fatalf("query node failed: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 gan node, got %d", len(recs))
	}
	if got := recordString(recs[0], "d"); got != "original description" {
		t.Errorf("description = %q, want 'original description' (not overwritten)", got)
	}
	if got := recordFloat(recs[0], "score"); got < 7.49 || got > 7.51 {
		t.Errorf("influence_score = %v, want 7.5 (not overwritten)", got)
	}
	if got := recordString(recs[0], "source"); got != "manual" {
		t.Errorf("source = %q, want manual (reuse should not flip it)", got)
	}
}

// TestMerge_IdempotentRetry verifies that merging the same payload twice
// does not duplicate nodes or edges.
func TestMerge_IdempotentRetry(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)
	svc := NewMergeService(neo)

	nodes := []model.ExtractedNode{
		{Name: "CNN", Description: "Convolutional", Relevance: 8.0},
		{Name: "ResNet", Description: "Residual", Relevance: 9.0},
	}
	edges := []model.ExtractedEdge{
		{Source: "CNN", Target: "ResNet", Relation: "improves", Weight: 0.7},
	}

	if _, err := svc.Merge(t.Context(), "paper-4", nodes, edges); err != nil {
		t.Fatalf("first merge failed: %v", err)
	}
	firstNodes := countConcepts(t, neo)
	firstEdges := countEdges(t, neo)
	if firstNodes != 2 || firstEdges != 1 {
		t.Fatalf("after first merge: nodes=%d edges=%d, want 2/1", firstNodes, firstEdges)
	}

	// Second merge with identical input.
	res, err := svc.Merge(t.Context(), "paper-4", nodes, edges)
	if err != nil {
		t.Fatalf("second merge failed: %v", err)
	}
	if res.CreatedNodes != 0 {
		t.Errorf("second merge created_nodes = %d, want 0", res.CreatedNodes)
	}
	if res.ReusedNodes != 2 {
		t.Errorf("second merge reused_nodes = %d, want 2", res.ReusedNodes)
	}
	if res.CreatedEdges != 0 {
		t.Errorf("second merge created_edges = %d, want 0 (MERGE skipped existing)", res.CreatedEdges)
	}
	if got := countConcepts(t, neo); got != firstNodes {
		t.Errorf("node count after retry = %d, want %d (no duplicate)", got, firstNodes)
	}
	if got := countEdges(t, neo); got != firstEdges {
		t.Errorf("edge count after retry = %d, want %d (no duplicate)", got, firstEdges)
	}
}

// TestMerge_SkipsEdgesWithUnresolvedEndpointsAndSelfLoops verifies that
// edges referencing a name not in the extracted nodes, and self-loops,
// are skipped while valid edges are still created.
func TestMerge_SkipsEdgesWithUnresolvedEndpointsAndSelfLoops(t *testing.T) {
	neo := testutil.NewTestNeo4j(t)
	testutil.CleanupAllData(t, neo)
	svc := NewMergeService(neo)

	nodes := []model.ExtractedNode{
		{Name: "Alpha", Description: "a", Relevance: 5.0},
		{Name: "Beta", Description: "b", Relevance: 5.0},
	}
	edges := []model.ExtractedEdge{
		{Source: "Alpha", Target: "Beta", Relation: "valid", Weight: 0.5},    // ok
		{Source: "Alpha", Target: "Ghost", Relation: "phantom", Weight: 0.1}, // unresolved target
		{Source: "Alpha", Target: "Alpha", Relation: "self", Weight: 0.2},    // self-loop
	}

	res, err := svc.Merge(t.Context(), "paper-5", nodes, edges)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if res.CreatedNodes != 2 {
		t.Errorf("created_nodes = %d, want 2", res.CreatedNodes)
	}
	if res.CreatedEdges != 1 {
		t.Errorf("created_edges = %d, want 1 (only Alpha->Beta)", res.CreatedEdges)
	}
	if res.SkippedEdges != 2 {
		t.Errorf("skipped_edges = %d, want 2 (unresolved + self-loop)", res.SkippedEdges)
	}
	if got := countEdges(t, neo); got != 1 {
		t.Errorf("total edges = %d, want 1", got)
	}
}
