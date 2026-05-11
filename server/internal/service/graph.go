package service

import (
	"sort"
	"strings"

	"ai-graph-server/internal/model"
)

// GraphService provides graph-related business logic
type GraphService struct {
	nodes                    []model.Node
	edges                    []model.Edge
	communities              []model.Community
	hierarchicalCommunities *model.HierarchicalCommunity
	maxLevel                 int
	nodeMap                  map[string]*model.Node
	edgeIndex                map[string][]model.Edge
	neighborMap              map[string][]model.Neighbor
}

// New creates and initializes a new GraphService
func New(nodes []model.Node, edges []model.Edge, communities []model.Community, hierarchical *model.HierarchicalCommunity, maxLevel int) *GraphService {
	svc := &GraphService{
		nodes:                    nodes,
		edges:                    edges,
		communities:              communities,
		hierarchicalCommunities: hierarchical,
		maxLevel:                 maxLevel,
		nodeMap:                  make(map[string]*model.Node),
		edgeIndex:                make(map[string][]model.Edge),
		neighborMap:              make(map[string][]model.Neighbor),
	}

	// Build node map for quick lookup
	for i := range nodes {
		svc.nodeMap[nodes[i].ID] = &nodes[i]
	}

	// Build edge index for quick neighbor lookup
	for _, e := range edges {
		svc.edgeIndex[e.Source] = append(svc.edgeIndex[e.Source], e)
		svc.edgeIndex[e.Target] = append(svc.edgeIndex[e.Target], e)
	}

	// Build valid node set
	nodeSet := make(map[string]bool)
	for _, n := range nodes {
		nodeSet[n.ID] = true
	}

	// Build neighbor map with relationship info
	for _, e := range edges {
		// Only include edges between valid nodes
		if nodeSet[e.Source] && nodeSet[e.Target] {
			src := svc.nodeMap[e.Source]
			tgt := svc.nodeMap[e.Target]
			if src != nil && tgt != nil {
				// Add target to source's neighbors
				svc.neighborMap[e.Source] = append(svc.neighborMap[e.Source], model.Neighbor{
					ID:             tgt.ID,
					Name:           tgt.Name,
					CommunityID:    tgt.CommunityID,
					InfluenceScore: tgt.InfluenceScore,
					Weight:         e.Weight,
					Relation:       e.Relation,
				})
				// Add source to target's neighbors
				svc.neighborMap[e.Target] = append(svc.neighborMap[e.Target], model.Neighbor{
					ID:             src.ID,
					Name:           src.Name,
					CommunityID:    src.CommunityID,
					InfluenceScore: src.InfluenceScore,
					Weight:         e.Weight,
					Relation:       e.Relation,
				})
			}
		}
	}

	// Sort neighbors by weight (highest first)
	for id := range svc.neighborMap {
		neighbors := svc.neighborMap[id]
		sort.Slice(neighbors, func(i, j int) bool {
			return neighbors[i].Weight > neighbors[j].Weight
		})
		svc.neighborMap[id] = neighbors
	}

	return svc
}

// Summary returns a summary of the graph including communities and stats
func (s *GraphService) Summary(topN int) model.SummaryResponse {
	// Default to 20 if topN is invalid
	if topN <= 0 {
		topN = 20
	}
	// Sort nodes by influence score
	top := make([]model.Node, len(s.nodes))
	copy(top, s.nodes)
	sort.Slice(top, func(i, j int) bool {
		return top[i].InfluenceScore > top[j].InfluenceScore
	})
	if topN > len(top) {
		topN = len(top)
	}

	// Prepare hierarchical communities
	var communities []model.HierarchicalCommunity
	if s.hierarchicalCommunities != nil {
		communities = []model.HierarchicalCommunity{*s.hierarchicalCommunities}
	}

	return model.SummaryResponse{
		Communities: communities,
		Stats: model.GraphStats{
			TotalNodes:     len(s.nodes),
			TotalEdges:     len(s.edges),
			CommunityCount: s.countTotalCommunities(),
			MaxLevel:       s.maxLevel,
		},
		TopNodes: top[:topN],
	}
}

// countTotalCommunities returns the total number of communities in the hierarchy
func (s *GraphService) countTotalCommunities() int {
	if s.hierarchicalCommunities == nil {
		// Fallback to flat communities count
		return len(s.communities)
	}
	return s.countCommunitiesRecursive(s.hierarchicalCommunities)
}

// countCommunitiesRecursive recursively counts all communities in the hierarchy
func (s *GraphService) countCommunitiesRecursive(c *model.HierarchicalCommunity) int {
	count := 1 // Count current community
	for _, child := range c.Children {
		count += s.countCommunitiesRecursive(&child)
	}
	return count
}

// Nodes returns a paginated list of nodes with optional filtering
func (s *GraphService) Nodes(offset, limit int, sortBy, communityFilter string, ids []string) model.NodesResponse {
	// Set default limits
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	var filtered []model.Node

	// Filter by node IDs if provided
	if len(ids) > 0 {
		idSet := make(map[string]bool)
		for _, id := range ids {
			idSet[id] = true
		}
		for _, n := range s.nodes {
			if idSet[n.ID] {
				filtered = append(filtered, n)
			}
		}
	} else if communityFilter != "" {
		// Filter by community name
		var cid int
		for _, c := range s.communities {
			if c.Name == communityFilter || strings.EqualFold(c.Name, communityFilter) {
				cid = c.ID
				break
			}
		}
		for _, n := range s.nodes {
			if n.CommunityID == cid {
				filtered = append(filtered, n)
			}
		}
	} else {
		// No filtering - return all nodes
		filtered = make([]model.Node, len(s.nodes))
		copy(filtered, s.nodes)
	}

	// Sort the filtered nodes
	switch sortBy {
	case "name":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Name < filtered[j].Name
		})
	default:
		// Default: sort by influence score descending
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].InfluenceScore > filtered[j].InfluenceScore
		})
	}

	total := len(filtered)
	// Handle case where offset is beyond total
	if offset >= total {
		return model.NodesResponse{
			Nodes:      []model.Node{},
			Pagination: model.Pagination{Offset: offset, Limit: limit, Total: total, HasMore: false},
		}
	}

	// Calculate end index
	end := offset + limit
	if end > total {
		end = total
	}

	return model.NodesResponse{
		Nodes: filtered[offset:end],
		Pagination: model.Pagination{
			Offset:  offset,
			Limit:   limit,
			Total:   total,
			HasMore: end < total,
		},
	}
}

// NodeDetail returns detailed information about a specific node and its neighbors
func (s *GraphService) NodeDetail(id string) (*model.NodeDetail, bool) {
	node, ok := s.nodeMap[id]
	if !ok {
		return nil, false
	}

	neighbors := s.neighborMap[id]
	if neighbors == nil {
		neighbors = []model.Neighbor{}
	}

	return &model.NodeDetail{
		Node:      *node,
		Neighbors: neighbors,
	}, true
}

// NodeEdges returns edges for a specific node with optional direction filtering
func (s *GraphService) NodeEdges(id, direction string) (*model.EdgesResponse, bool) {
	if _, ok := s.nodeMap[id]; !ok {
		return nil, false
	}

	allEdges := s.edgeIndex[id]

	var edges []model.Edge
	for _, e := range allEdges {
		switch direction {
		case "in":
			// Only edges where this node is the target
			if e.Target == id {
				edges = append(edges, e)
			}
		case "out":
			// Only edges where this node is the source
			if e.Source == id {
				edges = append(edges, e)
			}
		default:
			// All edges
			edges = append(edges, e)
		}
	}

	// Ensure we never return nil
	if edges == nil {
		edges = []model.Edge{}
	}

	return &model.EdgesResponse{
		NodeID: id,
		Edges:  edges,
	}, true
}

// Search searches for nodes by name or description
func (s *GraphService) Search(query string, limit int) model.SearchResponse {
	// Set default limits
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	queryLower := strings.ToLower(query)

	var results []model.SearchResult
	for _, n := range s.nodes {
		// Check if query matches name or description
		if strings.Contains(strings.ToLower(n.Name), queryLower) ||
			strings.Contains(strings.ToLower(n.Description), queryLower) {
			if len(results) >= limit {
				break
			}
			results = append(results, model.SearchResult{
				ID:             n.ID,
				Name:           n.Name,
				CommunityID:    n.CommunityID,
				InfluenceScore: n.InfluenceScore,
			})
		}
	}

	// Sort results by influence score
	sort.Slice(results, func(i, j int) bool {
		return results[i].InfluenceScore > results[j].InfluenceScore
	})

	// Ensure we never return nil
	if results == nil {
		results = []model.SearchResult{}
	}

	return model.SearchResponse{
		Query:   query,
		Results: results,
	}
}

// Expand expands a set of nodes to include their neighbors and connecting edges
func (s *GraphService) Expand(ids []string, includeEdges, includeNeighbors bool) model.ExpandResponse {
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	nodeSet := make(map[string]bool)
	edgeSet := make(map[[2]string]model.Edge)

	for _, id := range ids {
		// Add the requested node
		if n, ok := s.nodeMap[id]; ok {
			nodeSet[id] = true
			_ = n
		}

		// Add edges between requested nodes
		if includeEdges {
			for _, e := range s.edgeIndex[id] {
				if idSet[e.Source] && idSet[e.Target] {
					// Use sorted key to avoid duplicates (source-target vs target-source)
					key := [2]string{e.Source, e.Target}
					if _, exists := edgeSet[key]; !exists {
						edgeSet[key] = e
					}
				}
			}
		}

		// Expand to include neighbors
		if includeNeighbors {
			// Add all neighbor nodes
			for _, nb := range s.neighborMap[id] {
				nodeSet[nb.ID] = true
			}
			// Add edges from original node to neighbors
			for _, e := range s.edgeIndex[id] {
				if nodeSet[e.Source] || nodeSet[e.Target] {
					key := [2]string{e.Source, e.Target}
					if _, exists := edgeSet[key]; !exists {
						edgeSet[key] = e
					}
				}
			}
		}
	}

	// Convert nodeSet to actual nodes
	nodes := make([]model.Node, 0, len(nodeSet))
	for _, n := range s.nodes {
		if nodeSet[n.ID] {
			nodes = append(nodes, n)
		}
	}

	// Convert edgeSet to actual edges
	edges := make([]model.Edge, 0, len(edgeSet))
	for _, e := range edgeSet {
		edges = append(edges, e)
	}

	// Ensure we never return nil
	if nodes == nil {
		nodes = []model.Node{}
	}
	if edges == nil {
		edges = []model.Edge{}
	}

	return model.ExpandResponse{
		Nodes: nodes,
		Edges: edges,
	}
}
