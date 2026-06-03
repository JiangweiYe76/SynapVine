package loader

import (
	"encoding/json"
	"os"
	"strings"
	"sync"

	"console/internal/model"
)

// GraphStore provides thread-safe access to graph data
type GraphStore struct {
	mu   sync.RWMutex
	data *model.GraphData
	path string
}

// NewGraphStore loads graph data from the specified path and returns a GraphStore
func NewGraphStore(path string) (*GraphStore, error) {
	data, err := loadGraphData(path)
	if err != nil {
		return nil, err
	}
	return &GraphStore{data: data, path: path}, nil
}

func loadGraphData(path string) (*model.GraphData, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var graph model.GraphData
	if err := json.Unmarshal(bytes, &graph); err != nil {
		return nil, err
	}
	return &graph, nil
}

func (s *GraphStore) save() error {
	bytes, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, bytes, 0644)
}

// ListNodes returns a paginated slice of nodes
func (s *GraphStore) ListNodes(offset, limit int) ([]model.Node, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.data.Nodes)
	if offset >= total {
		return []model.Node{}, total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return s.data.Nodes[offset:end], total
}

// SearchNodes returns a paginated slice of nodes matching the query string
func (s *GraphStore) SearchNodes(query string, offset, limit int) ([]model.Node, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matched []model.Node
	q := strings.ToLower(query)
	for _, n := range s.data.Nodes {
		if strings.Contains(strings.ToLower(n.ID), q) ||
			strings.Contains(strings.ToLower(n.Name), q) ||
			strings.Contains(strings.ToLower(n.Category), q) ||
			strings.Contains(strings.ToLower(n.Description), q) {
			matched = append(matched, n)
		}
	}

	total := len(matched)
	if offset >= total {
		return []model.Node{}, total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return matched[offset:end], total
}

// GetNode returns a node by ID, or nil if not found
func (s *GraphStore) GetNode(id string) *model.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.data.Nodes {
		if s.data.Nodes[i].ID == id {
			node := s.data.Nodes[i]
			return &node
		}
	}
	return nil
}

// CreateNode adds a new node and persists to disk
func (s *GraphStore) CreateNode(node model.Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Nodes = append(s.data.Nodes, node)
	return s.save()
}

// UpdateNode updates an existing node by ID and persists to disk
func (s *GraphStore) UpdateNode(id string, update model.NodeUpdateRequest) (*model.Node, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.data.Nodes {
		if s.data.Nodes[i].ID == id {
			node := &s.data.Nodes[i]
			if update.Name != nil {
				node.Name = *update.Name
			}
			if update.Category != nil {
				node.Category = *update.Category
			}
			if update.Description != nil {
				node.Description = *update.Description
			}
			if update.InfluenceScore != nil {
				node.InfluenceScore = *update.InfluenceScore
			}
			if update.FirstAppeared != nil {
				node.FirstAppeared = *update.FirstAppeared
			}
			if update.Milestones != nil {
				node.Milestones = *update.Milestones
			}
			if err := s.save(); err != nil {
				return nil, err
			}
			result := *node
			return &result, nil
		}
	}
	return nil, nil
}

// DeleteNode removes a node by ID and persists to disk
func (s *GraphStore) DeleteNode(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.data.Nodes {
		if s.data.Nodes[i].ID == id {
			s.data.Nodes = append(s.data.Nodes[:i], s.data.Nodes[i+1:]...)
			if err := s.save(); err != nil {
				return false
			}
			return true
		}
	}
	return false
}

// NodeExists checks if a node with the given ID exists
func (s *GraphStore) NodeExists(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, n := range s.data.Nodes {
		if n.ID == id {
			return true
		}
	}
	return false
}

// Stats returns graph statistics
func (s *GraphStore) Stats() model.StatsResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	categories := make(map[string]int)
	var totalInfluence float64

	for _, n := range s.data.Nodes {
		categories[n.Category]++
		totalInfluence += n.InfluenceScore
	}

	var avgInfluence float64
	if len(s.data.Nodes) > 0 {
		avgInfluence = totalInfluence / float64(len(s.data.Nodes))
	}

	return model.StatsResponse{
		TotalNodes:    len(s.data.Nodes),
		TotalEdges:    len(s.data.Edges),
		CategoryCount: len(categories),
		Categories:    categories,
		AvgInfluence:  avgInfluence,
	}
}
