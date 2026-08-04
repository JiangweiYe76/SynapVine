package model

// ExtractedNode represents a node extracted by the LLM from a paper, as
// stored in the review queue's extracted_nodes JSON column. It mirrors
// the discovery service's extractor.ExtractedNode schema so that core
// can unmarshal review queue payloads without importing discovery.
type ExtractedNode struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Relevance   float64 `json:"relevance"` // 0-10, how central to the paper
}

// ExtractedEdge represents a relationship extracted by the LLM from a
// paper, as stored in the review queue's extracted_edges JSON column.
// Source and Target are node NAMES (as produced by the LLM), not IDs;
// the merge service resolves them to Neo4j node IDs before creating
// the RELATES_TO relationship.
type ExtractedEdge struct {
	Source   string  `json:"source"`   // name of the source node
	Target   string  `json:"target"`   // name of the target node
	Relation string  `json:"relation"` // description of the relationship
	Weight   float64 `json:"weight"`   // 0-1, strength of relationship
}

// MergeResult reports the outcome of merging an approved extraction into
// the Neo4j graph. It is returned by the merge service and surfaced to
// the reviewer so they can see what the approval actually changed.
type MergeResult struct {
	CreatedNodes int `json:"created_nodes"` // nodes newly created this merge
	ReusedNodes  int `json:"reused_nodes"`  // existing nodes reused (by name)
	CreatedEdges int `json:"created_edges"` // edges newly created this merge
	SkippedEdges int `json:"skipped_edges"` // edges skipped (unresolved name or self-loop)
}
