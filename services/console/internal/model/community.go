package model

// Community represents a community in the knowledge graph.
//
// IDs are server-generated UUIDs when not supplied on creation. The
// service determines the level from the parent chain.
type Community struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Color     string  `json:"color"`
	Level     int     `json:"level"`
	Domain    string  `json:"domain"`
	ParentID  *string `json:"parent_id,omitempty"`
	NodeCount int     `json:"node_count"`
}

// HierarchicalCommunity represents a community in a nested structure.
type HierarchicalCommunity struct {
	ID        string                  `json:"id"`
	ParentID  *string                 `json:"parent_id,omitempty"`
	Name      string                  `json:"name"`
	Color     string                  `json:"color"`
	Level     int                     `json:"level"`
	Domain    string                  `json:"domain"`
	NodeCount int                     `json:"node_count"`
	Children  []HierarchicalCommunity `json:"children,omitempty"`
}

// CommunitiesTreeResponse wraps the list of root hierarchical communities.
type CommunitiesTreeResponse struct {
	Communities []HierarchicalCommunity `json:"communities"`
}

// CommunitiesListResponse is the response for listing communities.
type CommunitiesListResponse struct {
	Communities []Community `json:"communities"`
}

// CommunityCreateRequest is a request to create a new community.
//
// ID is optional. When empty, the core service generates a fresh UUID
// and returns it in the response.
type CommunityCreateRequest struct {
	ID       string  `json:"id,omitempty"`
	Name     string  `json:"name"`
	Color    string  `json:"color"`
	Domain   string  `json:"domain"`
	ParentID *string `json:"parent_id,omitempty"`
}

// CommunityUpdateRequest is a request to update an existing community.
type CommunityUpdateRequest struct {
	Name     *string `json:"name,omitempty"`
	Color    *string `json:"color,omitempty"`
	Domain   *string `json:"domain,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
}
