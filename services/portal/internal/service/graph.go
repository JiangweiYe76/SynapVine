package service

import (
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"ai-graph-server/internal/coreclient"
	"ai-graph-server/internal/model"
)

// GraphService provides graph-related business logic
type GraphService struct {
	nodes        []model.Node
	edges        []model.Edge
	communities  []model.HierarchicalCommunity
	maxLevel     int
	nodeMap      map[string]*model.Node
	edgeIndex    map[string][]model.Edge
	neighborMap  map[string][]model.Neighbor
}

// New creates and initializes a new GraphService from raw core data.
// Core stores community IDs as string UUIDs; the service maps them to
// sequential integer IDs so the portal frontend can use them directly.
func New(coreNodes []coreclient.CoreNode, edges []model.Edge, coreCommunities []coreclient.CoreCommunity) *GraphService {
	// Build a deterministic string -> int community ID map. Roots are listed
	// first, then children, which keeps parent IDs lower than child IDs.
	communityMap := buildCommunityIDMap(coreCommunities)

	// Convert core nodes into portal nodes using the integer community mapping.
	nodes := make([]model.Node, 0, len(coreNodes))
	for _, cn := range coreNodes {
		communityID := 0
		if cn.CommunityID != nil {
			if id, ok := communityMap[*cn.CommunityID]; ok {
				communityID = id
			}
		}
		nodes = append(nodes, model.Node{
			ID:             cn.ID,
			Name:           cn.Name,
			Description:    cn.Description,
			InfluenceScore: cn.InfluenceScore,
			FirstAppeared:  cn.FirstAppeared,
			Milestones:     cn.Milestones,
			CommunityID:    communityID,
			Degree:         cn.Degree,
		})
	}

	// Convert core communities into portal hierarchical communities.
	communities := convertCommunities(coreCommunities, communityMap)
	maxLevel := computeMaxLevel(communities)

	svc := &GraphService{
		nodes:        nodes,
		edges:        edges,
		communities:  communities,
		maxLevel:     maxLevel,
		nodeMap:      make(map[string]*model.Node),
		edgeIndex:    make(map[string][]model.Edge),
		neighborMap:  make(map[string][]model.Neighbor),
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

	slog.Info("graph_service_initialized",
		slog.Int("nodes", len(nodes)),
		slog.Int("edges", len(edges)),
		slog.Int("communities", len(communities)),
		slog.Int("max_level", maxLevel),
	)

	return svc
}

// buildCommunityIDMap walks the core community tree and assigns each unique
// string community ID a sequential integer starting at 1. Unassigned nodes
// keep the default community ID of 0.
func buildCommunityIDMap(coreCommunities []coreclient.CoreCommunity) map[string]int {
	mapping := make(map[string]int)
	nextID := 1
	var walk func(comms []coreclient.CoreCommunity)
	walk = func(comms []coreclient.CoreCommunity) {
		for _, c := range comms {
			if _, exists := mapping[c.ID]; !exists {
				mapping[c.ID] = nextID
				nextID++
			}
			if len(c.Children) > 0 {
				walk(c.Children)
			}
		}
	}
	walk(coreCommunities)
	return mapping
}

// convertCommunities transforms core communities (string IDs) into portal
// communities (integer IDs) using the provided mapping.
func convertCommunities(coreCommunities []coreclient.CoreCommunity, mapping map[string]int) []model.HierarchicalCommunity {
	result := make([]model.HierarchicalCommunity, 0, len(coreCommunities))
	for _, c := range coreCommunities {
		result = append(result, convertCommunity(c, mapping))
	}
	return result
}

func convertCommunity(c coreclient.CoreCommunity, mapping map[string]int) model.HierarchicalCommunity {
	id := mapping[c.ID]
	var parentID *int
	if c.ParentID != nil {
		if pid, ok := mapping[*c.ParentID]; ok {
			parentID = &pid
		}
	}
	children := make([]model.HierarchicalCommunity, 0, len(c.Children))
	for _, child := range c.Children {
		children = append(children, convertCommunity(child, mapping))
	}
	return model.HierarchicalCommunity{
		ID:        id,
		ParentID:  parentID,
		Name:      c.Name,
		Color:     c.Color,
		Level:     c.Level,
		NodeIDs:   []string{},
		NodeCount: c.NodeCount,
		Children:  children,
	}
}

// computeMaxLevel returns the deepest level found in the community tree.
func computeMaxLevel(communities []model.HierarchicalCommunity) int {
	max := 0
	var walk func(comms []model.HierarchicalCommunity)
	walk = func(comms []model.HierarchicalCommunity) {
		for _, c := range comms {
			if c.Level > max {
				max = c.Level
			}
			if len(c.Children) > 0 {
				walk(c.Children)
			}
		}
	}
	walk(communities)
	return max
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

	return model.SummaryResponse{
		Communities: s.communities,
		Stats: model.GraphStats{
			TotalNodes:     len(s.nodes),
			TotalEdges:     len(s.edges),
			CommunityCount: len(s.communities),
			MaxLevel:       s.maxLevel,
		},
		TopNodes: top[:topN],
	}
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
		// Filter by community ID
		cid, err := strconv.Atoi(communityFilter)
		if err == nil {
			for _, n := range s.nodes {
				if n.CommunityID == cid {
					filtered = append(filtered, n)
				}
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
