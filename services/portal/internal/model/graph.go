package model

// Node represents a single node in the knowledge graph
type Node struct {
	ID             string   `json:"id"`              // Unique identifier for the node
	Name           string   `json:"name"`            // Display name of the node
	Description    string   `json:"description"`     // Brief description of the node
	InfluenceScore float64  `json:"influence_score"` // Influence/popularity score (0-10)
	CommunityID    int      `json:"community_id"`    // ID of the community this node belongs to
	Degree         int      `json:"degree"`          // Number of connected edges
	FirstAppeared  string   `json:"first_appeared"`  // Year and month first appeared (YYYY-MM)
	Milestones     []string `json:"milestones,omitempty"` // Key milestones
}

// Edge represents a relationship between two nodes
type Edge struct {
	Source   string  `json:"source"`   // Source node ID
	Target   string  `json:"target"`   // Target node ID
	Weight   float64 `json:"weight"`   // Connection strength (0-1)
	Relation string  `json:"relation"` // Description of the relationship
}

// GraphData contains the complete graph with nodes and edges
type GraphData struct {
	Nodes []Node `json:"nodes"` // List of all nodes
	Edges []Edge `json:"edges"` // List of all edges
}

// Community represents a flat community structure
type Community struct {
	ID        int      `json:"id"`         // Unique community ID
	Name      string   `json:"name"`       // Community name
	Color     string   `json:"color"`      // Display color (hex)
	NodeIDs   []string `json:"node_ids"`   // List of node IDs in this community
	NodeCount int      `json:"node_count"` // Number of nodes in this community
}

// HierarchicalCommunity represents a community in a nested/hierarchical structure
type HierarchicalCommunity struct {
	ID        int                   `json:"id"`           // Unique community ID
	ParentID  *int                  `json:"parent_id"`    // Parent community ID (null for root)
	Name      string                `json:"name"`         // Community name
	Color     string                `json:"color"`        // Display color (hex)
	Level     int                   `json:"level"`        // Depth level in the hierarchy (0 for root)
	NodeIDs   []string              `json:"node_ids"`     // List of node IDs in this community
	NodeCount int                   `json:"node_count"`   // Number of nodes in this community
	Children  []HierarchicalCommunity `json:"children,omitempty"` // Sub-communities
}

// GraphStats provides statistical information about the graph
type GraphStats struct {
	TotalNodes     int `json:"total_nodes"`     // Total number of nodes
	TotalEdges     int `json:"total_edges"`     // Total number of edges
	CommunityCount int `json:"community_count"` // Total number of communities
	MaxLevel       int `json:"max_level"`       // Maximum hierarchy depth
}

// SummaryResponse is the response for the graph summary endpoint
type SummaryResponse struct {
	Communities []HierarchicalCommunity `json:"communities"` // Hierarchical community tree
	Stats       GraphStats              `json:"stats"`       // Graph statistics
	TopNodes    []Node                  `json:"top_nodes"`   // Top nodes by influence
}

// Pagination provides pagination metadata
type Pagination struct {
	Offset  int  `json:"offset"`   // Current offset
	Limit   int  `json:"limit"`    // Items per page
	Total   int  `json:"total"`    // Total items
	HasMore bool `json:"has_more"` // Whether more items exist
}

// NodesResponse is the response for the nodes endpoint
type NodesResponse struct {
	Nodes      []Node     `json:"nodes"`      // List of nodes
	Pagination Pagination `json:"pagination"` // Pagination metadata
}

// NodeDetail provides detailed information about a single node
type NodeDetail struct {
	Node      Node       `json:"node"`      // The node itself
	Neighbors []Neighbor `json:"neighbors"` // Connected neighbor nodes
}

// Neighbor represents a connected neighbor node
type Neighbor struct {
	ID             string  `json:"id"`              // Neighbor node ID
	Name           string  `json:"name"`            // Neighbor name
	CommunityID    int     `json:"community_id"`    // Neighbor's community
	InfluenceScore float64 `json:"influence_score"` // Neighbor's influence
	Weight         float64 `json:"weight"`          // Connection weight
	Relation       string  `json:"relation"`        // Relationship description
}

// EdgesResponse is the response for the edges endpoint
type EdgesResponse struct {
	NodeID string `json:"node_id"` // The node we're looking at
	Edges  []Edge `json:"edges"`   // Connected edges
}

// ExpandResponse is the response for expanding nodes
type ExpandResponse struct {
	Nodes []Node `json:"nodes"` // Expanded nodes
	Edges []Edge `json:"edges"` // Connected edges
}

// SearchResult represents a single search result
type SearchResult struct {
	ID             string  `json:"id"`              // Node ID
	Name           string  `json:"name"`            // Node name
	CommunityID    int     `json:"community_id"`    // Community ID
	InfluenceScore float64 `json:"influence_score"` // Influence score
	Highlight      string  `json:"highlight,omitempty"` // Highlighted text snippet
}

// SearchResponse is the response for search queries
type SearchResponse struct {
	Query   string         `json:"query"`   // Search query
	Results []SearchResult `json:"results"` // Matching nodes
}

// TokenResponse is the response for token requests
type TokenResponse struct {
	Token string `json:"token"` // Temporary access token
}

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error   string `json:"error"`   // Error code
	Message string `json:"message"` // Human-readable error message
}
