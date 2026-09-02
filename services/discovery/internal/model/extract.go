package model

// ExtractedNode represents a node extracted by the LLM from a paper.
type ExtractedNode struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Relevance   float64 `json:"relevance"` // 0-10, how central to the paper
}

// ExtractedEdge represents a relationship extracted by the LLM from a paper.
type ExtractedEdge struct {
	Source   string  `json:"source"`   // Name of the source node
	Target   string  `json:"target"`   // Name of the target node
	Relation string  `json:"relation"` // Description of the relationship
	Weight   float64 `json:"weight"`   // 0-1, strength of relationship
}

// ExtractionResult is the structured output of the LLM extraction pipeline.
type ExtractionResult struct {
	Nodes []ExtractedNode `json:"nodes"`
	Edges []ExtractedEdge `json:"edges"`
}

// Paper is a minimal representation of a paper fetched from the core service.
type Paper struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Authors  string `json:"authors"`
	RawText  string `json:"raw_text"`
	Status   string `json:"status"`
}

// ReviewQueueItem is the payload submitted to core's review queue.
type ReviewQueueItem struct {
	PaperID        string          `json:"paper_id"`
	ExtractedNodes []ExtractedNode `json:"extracted_nodes"`
	ExtractedEdges []ExtractedEdge `json:"extracted_edges"`
}

// LLMProvider is the minimal representation fetched from the core service.
type LLMProvider struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	BaseURL     string  `json:"base_url"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// AnalyzeRequest is the payload for POST /api/analyze.
type AnalyzeRequest struct {
	PaperID string `json:"paper_id"`
}

// AnalyzeResponse is the response for POST /api/analyze.
type AnalyzeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
