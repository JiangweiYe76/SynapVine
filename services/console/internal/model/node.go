package model

// Node represents a single node in the knowledge graph
type Node struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	InfluenceScore float64  `json:"influence_score"`
	FirstAppeared  string   `json:"first_appeared"`
	Milestones     []string `json:"milestones,omitempty"`
	CommunityID    *string  `json:"community_id,omitempty"`
}

// Edge represents a relationship between two nodes
type Edge struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Weight   float64 `json:"weight"`
	Relation string  `json:"relation"`
}

// GraphData contains the complete graph with nodes and edges
type GraphData struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// NodeCreateRequest represents a request to create a new node.
//
// ID is optional. When empty, the core service generates a fresh UUID
// and returns it as part of the created resource.
type NodeCreateRequest struct {
	ID             string   `json:"id,omitempty"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	InfluenceScore float64  `json:"influence_score"`
	FirstAppeared  string   `json:"first_appeared"`
	Milestones     []string `json:"milestones,omitempty"`
	CommunityID    *string  `json:"community_id,omitempty"`
}

// NodeUpdateRequest represents a request to update an existing node.
//
// CommunityID uses a tri-state pointer-to-pointer to distinguish
// "field absent" (leave unchanged), "field null" (clear community),
// and "field value" (assign community) in JSON payloads. See the core
// model for the full rationale.
type NodeUpdateRequest struct {
	Name           *string   `json:"name,omitempty"`
	Description    *string   `json:"description,omitempty"`
	InfluenceScore *float64  `json:"influence_score,omitempty"`
	FirstAppeared  *string   `json:"first_appeared,omitempty"`
	Milestones     *[]string `json:"milestones,omitempty"`
	CommunityID    **string  `json:"community_id,omitempty"`
}

// NodesListResponse is the response for listing nodes with pagination
type NodesListResponse struct {
	Nodes      []Node     `json:"nodes"`
	Pagination Pagination `json:"pagination"`
}

// Pagination provides pagination metadata
type Pagination struct {
	Offset  int  `json:"offset"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	HasMore bool `json:"has_more"`
}

// EdgeCreateRequest represents a request to create a new edge
type EdgeCreateRequest struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Weight   float64 `json:"weight"`
	Relation string  `json:"relation"`
}

// EdgeUpdateRequest represents a request to update an existing edge
type EdgeUpdateRequest struct {
	Weight   *float64 `json:"weight,omitempty"`
	Relation *string  `json:"relation,omitempty"`
}

// EdgesListResponse is the response for listing edges with pagination
type EdgesListResponse struct {
	Edges      []Edge     `json:"edges"`
	Pagination Pagination `json:"pagination"`
}

// StatsResponse provides graph statistics for the dashboard
type StatsResponse struct {
	TotalNodes   int     `json:"total_nodes"`
	TotalEdges   int     `json:"total_edges"`
	AvgInfluence float64 `json:"avg_influence"`
}
