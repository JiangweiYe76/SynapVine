package model

// Node represents a single concept node in the knowledge graph.
//
// CommunityID is the community the node belongs to (via the BELONGS_TO
// relationship in Neo4j). It is nil when the node is not assigned to any
// community. The field is populated by the repository when reading and
// ignored by the standard write path; use NodeUpdateRequest.CommunityID
// to change the assignment through the API.
type Node struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	InfluenceScore float64  `json:"influence_score"`
	FirstAppeared  string   `json:"first_appeared"`
	Milestones     []string `json:"milestones,omitempty"`
	CommunityID    *string  `json:"community_id,omitempty"`
	Degree         int      `json:"degree"`
}

// NodeCreateRequest represents a request to create a new node.
//
// ID is optional. When the caller leaves it empty the core service
// generates a fresh UUID and returns it in the response, so clients do
// not have to fabricate identifiers themselves.
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
// CommunityID uses the "tri-state" convention via a pointer to a pointer:
//   - field absent or nil  -> leave the community assignment unchanged
//   - field present, nil   -> remove the node from its community
//   - field present, *string -> assign the node to the given community
//
// This is necessary because encoding/json maps both "absent" and "null"
// to the same Go value when the field is *string, so we cannot distinguish
// "don't touch" from "clear" without the extra indirection.
type NodeUpdateRequest struct {
	Name           *string   `json:"name,omitempty"`
	Description    *string   `json:"description,omitempty"`
	InfluenceScore *float64  `json:"influence_score,omitempty"`
	FirstAppeared  *string   `json:"first_appeared,omitempty"`
	Milestones     *[]string `json:"milestones,omitempty"`
	CommunityID    **string  `json:"community_id,omitempty"`
}

// Community represents a community in the graph.
//
// The Level field is not stored; it is computed by the service from the
// parent chain. Root communities (ParentID == nil) have level 1; a child
// of a level-N community is level N+1.
type Community struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Color     string   `json:"color"`
	Level     int      `json:"level"`
	Domain    string   `json:"domain"`
	ParentID  *string  `json:"parent_id,omitempty"`
	NodeIDs   []string `json:"node_ids,omitempty"`
	NodeCount int      `json:"node_count"`
}

// CommunityCreateRequest represents a request to create a new community.
//
// Level is intentionally absent: it is derived from ParentID at create time.
// ID is optional. When the caller leaves it empty the core service
// generates a fresh UUID and returns it in the response.
type CommunityCreateRequest struct {
	ID       string  `json:"id,omitempty"`
	Name     string  `json:"name"`
	Color    string  `json:"color"`
	Domain   string  `json:"domain"`
	ParentID *string `json:"parent_id,omitempty"`
}

// CommunityUpdateRequest represents a request to update an existing community.
// Level is intentionally absent: it is re-derived from ParentID on update.
type CommunityUpdateRequest struct {
	Name     *string `json:"name,omitempty"`
	Color    *string `json:"color,omitempty"`
	Domain   *string `json:"domain,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
}

// Pagination provides pagination metadata.
type Pagination struct {
	Offset  int  `json:"offset"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	HasMore bool `json:"has_more"`
}

// NodesListResponse is the response for listing nodes with pagination.
type NodesListResponse struct {
	Nodes      []Node     `json:"nodes"`
	Pagination Pagination `json:"pagination"`
}

// CommunitiesListResponse is the response for listing communities.
type CommunitiesListResponse struct {
	Communities []Community `json:"communities"`
}

// HierarchicalCommunity represents a community in a nested/hierarchical structure.
type HierarchicalCommunity struct {
	ID        string                  `json:"id"`
	ParentID  *string                 `json:"parent_id,omitempty"`
	Name      string                  `json:"name"`
	Color     string                  `json:"color"`
	Level     int                     `json:"level"`
	Domain    string                  `json:"domain"`
	NodeIDs   []string                `json:"node_ids"`
	NodeCount int                     `json:"node_count"`
	Children  []HierarchicalCommunity `json:"children,omitempty"`
}

// Edge represents a relationship between two nodes.
type Edge struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Weight   float64 `json:"weight"`
	Relation string  `json:"relation"`
}

// EdgeCreateRequest represents a request to create a new edge.
//
// (Source, Target) is the composite primary key; both endpoints must already
// exist as Concept nodes. Weight must lie in [0, 1].
type EdgeCreateRequest struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Weight   float64 `json:"weight"`
	Relation string  `json:"relation"`
}

// EdgeUpdateRequest represents a request to update an existing edge.
// Source and Target are intentionally absent: the (source, target) pair is
// the edge identity and is immutable. To move an edge, delete the old one
// and create a new one.
type EdgeUpdateRequest struct {
	Weight   *float64 `json:"weight,omitempty"`
	Relation *string  `json:"relation,omitempty"`
}

// EdgesListResponse is the response for listing edges with pagination.
type EdgesListResponse struct {
	Edges      []Edge     `json:"edges"`
	Pagination Pagination `json:"pagination"`
}

// ErrorResponse is the standard error response format.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// TimelineRange describes the inclusive [minYear, maxYear] span covered by
// the `first_appeared` field of every Concept node in the graph. It is
// returned by the timeline endpoint and is intended to drive UI range
// selectors that need the full extent of the dataset, not just the
// currently visible window.
type TimelineRange struct {
	MinYear int `json:"min_year"`
	MaxYear int `json:"max_year"`
}
