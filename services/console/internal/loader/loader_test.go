package loader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"console/internal/model"
)

func createTestGraphStore(t *testing.T) *GraphStore {
	t.Helper()

	data := model.GraphData{
		Nodes: []model.Node{
			{ID: "transformer", Name: "Transformer", Category: "dl_arch", Description: "Self-attention", InfluenceScore: 9.8, FirstAppeared: 2017},
			{ID: "bert", Name: "BERT", Category: "nlp_model", Description: "Bidirectional", InfluenceScore: 9.6, FirstAppeared: 2018},
			{ID: "gpt", Name: "GPT", Category: "nlp_model", Description: "Generative", InfluenceScore: 9.7, FirstAppeared: 2018},
			{ID: "cnn", Name: "CNN", Category: "dl_arch", Description: "Convolutional", InfluenceScore: 9.0, FirstAppeared: 1989},
		},
		Edges: []model.Edge{
			{Source: "transformer", Target: "bert", Weight: 0.96, Relation: "based_on"},
			{Source: "transformer", Target: "gpt", Weight: 0.96, Relation: "based_on"},
			{Source: "gpt", Target: "cnn", Weight: 0.5, Relation: "related_to"},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "graph.json")
	bytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}
	if err := os.WriteFile(path, bytes, 0644); err != nil {
		t.Fatalf("failed to write test data: %v", err)
	}

	store, err := NewGraphStore(path)
	if err != nil {
		t.Fatalf("failed to create GraphStore: %v", err)
	}
	return store
}

// --- Node tests ---

func TestListNodes(t *testing.T) {
	store := createTestGraphStore(t)

	nodes, total := store.ListNodes(0, 2)
	if total != 4 {
		t.Errorf("expected total 4, got %d", total)
	}
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].ID != "transformer" {
		t.Errorf("expected first node transformer, got %s", nodes[0].ID)
	}

	// pagination
	nodes, total = store.ListNodes(2, 10)
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes from offset 2, got %d", len(nodes))
	}

	// empty result
	nodes, total = store.ListNodes(10, 10)
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes from offset 10, got %d", len(nodes))
	}
}

func TestSearchNodes(t *testing.T) {
	store := createTestGraphStore(t)

	nodes, total := store.SearchNodes("bert", 0, 10)
	if total != 1 {
		t.Errorf("expected total 1 for 'bert', got %d", total)
	}
	if len(nodes) != 1 || nodes[0].ID != "bert" {
		t.Errorf("expected bert, got %+v", nodes)
	}

	nodes, total = store.SearchNodes("dl_arch", 0, 10)
	if total != 2 {
		t.Errorf("expected total 2 for category 'dl_arch', got %d", total)
	}

	nodes, total = store.SearchNodes("nonexistent", 0, 10)
	if total != 0 {
		t.Errorf("expected total 0 for 'nonexistent', got %d", total)
	}
}

func TestGetNode(t *testing.T) {
	store := createTestGraphStore(t)

	node := store.GetNode("bert")
	if node == nil {
		t.Fatal("expected node, got nil")
	}
	if node.Name != "BERT" {
		t.Errorf("expected name BERT, got %s", node.Name)
	}

	node = store.GetNode("notfound")
	if node != nil {
		t.Errorf("expected nil for notfound, got %+v", node)
	}
}

func TestNodeExists(t *testing.T) {
	store := createTestGraphStore(t)

	if !store.NodeExists("transformer") {
		t.Error("expected transformer to exist")
	}
	if store.NodeExists("missing") {
		t.Error("expected missing to not exist")
	}
}

func TestCreateNode(t *testing.T) {
	store := createTestGraphStore(t)

	node := model.Node{ID: "vit", Name: "ViT", Category: "cv_model", Description: "Vision", InfluenceScore: 9.1, FirstAppeared: 2020}
	if err := store.CreateNode(node); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	_, total := store.ListNodes(0, 100)
	if total != 5 {
		t.Errorf("expected total 5 after create, got %d", total)
	}
	if !store.NodeExists("vit") {
		t.Error("expected vit to exist after create")
	}
}

func TestUpdateNode(t *testing.T) {
	store := createTestGraphStore(t)

	newName := "BERT-Updated"
	node, err := store.UpdateNode("bert", model.NodeUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if node == nil {
		t.Fatal("expected updated node, got nil")
	}
	if node.Name != "BERT-Updated" {
		t.Errorf("expected name BERT-Updated, got %s", node.Name)
	}

	// verify original unchanged via reload
	reloaded := store.GetNode("bert")
	if reloaded == nil || reloaded.Name != "BERT-Updated" {
		t.Error("expected persisted update")
	}

	// not found
	node, err = store.UpdateNode("missing", model.NodeUpdateRequest{})
	if err != nil {
		t.Fatalf("expected no error for missing, got %v", err)
	}
	if node != nil {
		t.Errorf("expected nil for missing node, got %+v", node)
	}
}

func TestDeleteNode(t *testing.T) {
	store := createTestGraphStore(t)

	if !store.DeleteNode("bert") {
		t.Error("expected delete to succeed")
	}
	if store.NodeExists("bert") {
		t.Error("expected bert to be deleted")
	}

	if store.DeleteNode("missing") {
		t.Error("expected delete missing to fail")
	}
}

// --- Edge tests ---

func TestListEdges(t *testing.T) {
	store := createTestGraphStore(t)

	edges, total := store.ListEdges(0, 2)
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if len(edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(edges))
	}

	// empty result
	edges, total = store.ListEdges(10, 10)
	if len(edges) != 0 {
		t.Errorf("expected 0 edges from offset 10, got %d", len(edges))
	}
}

func TestSearchEdges(t *testing.T) {
	store := createTestGraphStore(t)

	_, total := store.SearchEdges("based_on", 0, 10)
	if total != 2 {
		t.Errorf("expected total 2 for 'based_on', got %d", total)
	}

	_, total = store.SearchEdges("transformer", 0, 10)
	if total != 2 {
		t.Errorf("expected total 2 for 'transformer', got %d", total)
	}

	_, total = store.SearchEdges("nonexistent", 0, 10)
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
}

func TestGetEdge(t *testing.T) {
	store := createTestGraphStore(t)

	edge := store.GetEdge("transformer", "bert")
	if edge == nil {
		t.Fatal("expected edge, got nil")
	}
	if edge.Relation != "based_on" {
		t.Errorf("expected relation based_on, got %s", edge.Relation)
	}

	edge = store.GetEdge("missing", "node")
	if edge != nil {
		t.Errorf("expected nil for missing edge, got %+v", edge)
	}
}

func TestEdgeExists(t *testing.T) {
	store := createTestGraphStore(t)

	if !store.EdgeExists("transformer", "bert") {
		t.Error("expected edge to exist")
	}
	if store.EdgeExists("transformer", "missing") {
		t.Error("expected edge to not exist")
	}
}

func TestCreateEdge(t *testing.T) {
	store := createTestGraphStore(t)

	edge := model.Edge{Source: "transformer", Target: "cnn", Weight: 0.7, Relation: "hybrid"}
	if err := store.CreateEdge(edge); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	_, total := store.ListEdges(0, 100)
	if total != 4 {
		t.Errorf("expected total 4 after create, got %d", total)
	}
	if !store.EdgeExists("transformer", "cnn") {
		t.Error("expected new edge to exist")
	}
}

func TestUpdateEdge(t *testing.T) {
	store := createTestGraphStore(t)

	newWeight := 0.99
	newRelation := "heavily_based_on"
	edge, err := store.UpdateEdge("transformer", "bert", model.EdgeUpdateRequest{Weight: &newWeight, Relation: &newRelation})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if edge == nil {
		t.Fatal("expected updated edge, got nil")
	}
	if edge.Weight != 0.99 {
		t.Errorf("expected weight 0.99, got %f", edge.Weight)
	}
	if edge.Relation != "heavily_based_on" {
		t.Errorf("expected relation heavily_based_on, got %s", edge.Relation)
	}

	// not found
	edge, err = store.UpdateEdge("missing", "node", model.EdgeUpdateRequest{})
	if err != nil {
		t.Fatalf("expected no error for missing, got %v", err)
	}
	if edge != nil {
		t.Errorf("expected nil for missing edge, got %+v", edge)
	}
}

func TestDeleteEdge(t *testing.T) {
	store := createTestGraphStore(t)

	if !store.DeleteEdge("transformer", "bert") {
		t.Error("expected delete to succeed")
	}
	if store.EdgeExists("transformer", "bert") {
		t.Error("expected edge to be deleted")
	}

	if store.DeleteEdge("missing", "node") {
		t.Error("expected delete missing to fail")
	}
}

// --- Stats tests ---

func TestStats(t *testing.T) {
	store := createTestGraphStore(t)

	stats := store.Stats()
	if stats.TotalNodes != 4 {
		t.Errorf("expected total_nodes 4, got %d", stats.TotalNodes)
	}
	if stats.TotalEdges != 3 {
		t.Errorf("expected total_edges 3, got %d", stats.TotalEdges)
	}
	if stats.CategoryCount != 2 {
		t.Errorf("expected category_count 2, got %d", stats.CategoryCount)
	}
	if stats.AvgInfluence <= 0 {
		t.Errorf("expected positive avg_influence, got %f", stats.AvgInfluence)
	}
	if stats.Categories["dl_arch"] != 2 {
		t.Errorf("expected dl_arch count 2, got %d", stats.Categories["dl_arch"])
	}
	if stats.Categories["nlp_model"] != 2 {
		t.Errorf("expected nlp_model count 2, got %d", stats.Categories["nlp_model"])
	}
}
