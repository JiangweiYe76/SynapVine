package model

// Node represents a single concept node in the knowledge graph.
type Node struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Category       string   `json:"category"`
	Description    string   `json:"description"`
	InfluenceScore float64  `json:"influence_score"`
	FirstAppeared  int      `json:"first_appeared"`
	Milestones     []string `json:"milestones,omitempty"`
}

// NodeCreateRequest represents a request to create a new node.
type NodeCreateRequest struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Category       string   `json:"category"`
	Description    string   `json:"description"`
	InfluenceScore float64  `json:"influence_score"`
	FirstAppeared  int      `json:"first_appeared"`
	Milestones     []string `json:"milestones,omitempty"`
}

// NodeUpdateRequest represents a request to update an existing node.
type NodeUpdateRequest struct {
	Name           *string   `json:"name,omitempty"`
	Category       *string   `json:"category,omitempty"`
	Description    *string   `json:"description,omitempty"`
	InfluenceScore *float64  `json:"influence_score,omitempty"`
	FirstAppeared  *int      `json:"first_appeared,omitempty"`
	Milestones     *[]string `json:"milestones,omitempty"`
}

// Community represents a community in the graph.
type Community struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Level  int    `json:"level"`
	Domain string `json:"domain"`
}

// CommunityCreateRequest represents a request to create a new community.
type CommunityCreateRequest struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Level  int    `json:"level"`
	Domain string `json:"domain"`
}

// CommunityUpdateRequest represents a request to update an existing community.
type CommunityUpdateRequest struct {
	Name   *string `json:"name,omitempty"`
	Color  *string `json:"color,omitempty"`
	Level  *int    `json:"level,omitempty"`
	Domain *string `json:"domain,omitempty"`
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

// ErrorResponse is the standard error response format.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
