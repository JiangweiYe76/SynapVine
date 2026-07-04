package service

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ai-graph-server/internal/coreclient"
	"ai-graph-server/internal/model"
)

const graphDataCacheTTL = 30 * time.Second

// graphDataCache holds a cached copy of the full graph data (nodes +
// edges) returned by core's /api/graph/data endpoint. The cache is
// shared across requests so that multiple handlers in the same request
// cycle (e.g. Summary + Expand) don't each trigger a full-graph fetch.
type graphDataCache struct {
	mu        sync.RWMutex
	data      *coreclient.CoreGraphData
	fetchedAt time.Time
}

// GraphService is a read-through adapter in front of the core service.
// It caches the full graph data for a short TTL to avoid repeated
// full-graph fetches within the same request window.
type GraphService struct {
	core  *coreclient.Client
	cache *graphDataCache
}

// New creates a GraphService that proxies reads to the given core client.
func New(core *coreclient.Client) *GraphService {
	return &GraphService{
		core:  core,
		cache: &graphDataCache{},
	}
}

// fetchGraphDataCached returns the full graph data, using a short-lived
// in-memory cache to avoid repeated full-graph fetches from core.
func (s *GraphService) fetchGraphDataCached(ctx context.Context) (*coreclient.CoreGraphData, error) {
	s.cache.mu.RLock()
	if s.cache.data != nil && time.Since(s.cache.fetchedAt) < graphDataCacheTTL {
		data := s.cache.data
		s.cache.mu.RUnlock()
		return data, nil
	}
	s.cache.mu.RUnlock()

	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()
	// Double-check after acquiring write lock.
	if s.cache.data != nil && time.Since(s.cache.fetchedAt) < graphDataCacheTTL {
		return s.cache.data, nil
	}

	data, err := s.core.FetchGraphData(ctx)
	if err != nil {
		return nil, err
	}
	s.cache.data = data
	s.cache.fetchedAt = time.Now()
	return data, nil
}

// Summary returns a summary of the graph including communities, stats,
// and the top N nodes by influence score.
func (s *GraphService) Summary(ctx context.Context, topN int) (model.SummaryResponse, error) {
	if topN <= 0 {
		topN = 20
	}

	data, err := s.fetchGraphDataCached(ctx)
	if err != nil {
		return model.SummaryResponse{}, err
	}
	coreComms, err := s.core.FetchCommunityTree(ctx)
	if err != nil {
		return model.SummaryResponse{}, err
	}

	communityMap := buildCommunityIDMap(coreComms)
	nodes := make([]model.Node, 0, len(data.Nodes))
	for _, cn := range data.Nodes {
		nodes = append(nodes, toPortalNode(cn, communityMap))
	}

	communities := convertCommunities(coreComms, communityMap)

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].InfluenceScore > nodes[j].InfluenceScore
	})
	if topN > len(nodes) {
		topN = len(nodes)
	}

	return model.SummaryResponse{
		Communities: communities,
		Stats: model.GraphStats{
			TotalNodes:     len(nodes),
			TotalEdges:     len(data.Edges),
			CommunityCount: len(communities),
			MaxLevel:       computeMaxLevel(communities),
		},
		TopNodes: nodes[:topN],
	}, nil
}

// Nodes returns a paginated list of nodes with optional filtering.
func (s *GraphService) Nodes(ctx context.Context, offset, limit int, sortBy, communityFilter string, ids []string) (model.NodesResponse, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	data, err := s.fetchGraphDataCached(ctx)
	if err != nil {
		return model.NodesResponse{}, err
	}
	coreComms, err := s.core.FetchCommunityTree(ctx)
	if err != nil {
		return model.NodesResponse{}, err
	}

	communityMap := buildCommunityIDMap(coreComms)
	nodes := make([]model.Node, 0, len(data.Nodes))
	for _, cn := range data.Nodes {
		nodes = append(nodes, toPortalNode(cn, communityMap))
	}

	// Filter
	if len(ids) > 0 {
		idSet := make(map[string]bool, len(ids))
		for _, id := range ids {
			idSet[id] = true
		}
		filtered := nodes[:0]
		for _, n := range nodes {
			if idSet[n.ID] {
				filtered = append(filtered, n)
			}
		}
		nodes = filtered
	} else if communityFilter != "" {
		cid, err := strconv.Atoi(communityFilter)
		if err == nil {
			filtered := nodes[:0]
			for _, n := range nodes {
				if n.CommunityID == cid {
					filtered = append(filtered, n)
				}
			}
			nodes = filtered
		}
	}

	switch sortBy {
	case "name":
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Name < nodes[j].Name
		})
	default:
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].InfluenceScore > nodes[j].InfluenceScore
		})
	}

	total := len(nodes)
	if offset >= total {
		return model.NodesResponse{
			Nodes:      []model.Node{},
			Pagination: model.Pagination{Offset: offset, Limit: limit, Total: total, HasMore: false},
		}, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return model.NodesResponse{
		Nodes: nodes[offset:end],
		Pagination: model.Pagination{
			Offset:  offset,
			Limit:   limit,
			Total:   total,
			HasMore: end < total,
		},
	}, nil
}

// NodeDetail returns a single node plus the neighbors connected to it.
func (s *GraphService) NodeDetail(ctx context.Context, id string) (model.NodeDetail, bool, error) {
	coreNode, err := s.core.GetNode(ctx, id)
	if err != nil {
		return model.NodeDetail{}, false, err
	}
	if coreNode == nil {
		return model.NodeDetail{}, false, nil
	}

	coreComms, err := s.core.FetchCommunityTree(ctx)
	if err != nil {
		return model.NodeDetail{}, false, err
	}
	communityMap := buildCommunityIDMap(coreComms)

	node := toPortalNode(*coreNode, communityMap)

	edges, err := s.fetchEdgesByNodeIDs(ctx, []string{id})
	if err != nil {
		return model.NodeDetail{}, false, err
	}
	nodesByID, err := s.fetchNodeIndex(ctx, communityMap)
	if err != nil {
		return model.NodeDetail{}, false, err
	}

	neighbors := []model.Neighbor{}
	seen := make(map[string]bool)
	for _, e := range edges {
		var neighborID string
		if e.Source == id {
			neighborID = e.Target
		} else if e.Target == id {
			neighborID = e.Source
		} else {
			continue
		}
		if seen[neighborID] {
			continue
		}
		seen[neighborID] = true

		n, ok := nodesByID[neighborID]
		if !ok {
			continue
		}
		neighbors = append(neighbors, model.Neighbor{
			ID:             n.ID,
			Name:           n.Name,
			CommunityID:    n.CommunityID,
			InfluenceScore: n.InfluenceScore,
			Weight:         e.Weight,
			Relation:       e.Relation,
		})
	}
	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].Weight > neighbors[j].Weight
	})

	return model.NodeDetail{Node: node, Neighbors: neighbors}, true, nil
}

// NodeEdges returns edges connected to a node, optionally filtered by direction.
func (s *GraphService) NodeEdges(ctx context.Context, id, direction string) (model.EdgesResponse, bool, error) {
	coreNode, err := s.core.GetNode(ctx, id)
	if err != nil {
		return model.EdgesResponse{}, false, err
	}
	if coreNode == nil {
		return model.EdgesResponse{}, false, nil
	}

	edges, err := s.fetchEdgesByNodeIDs(ctx, []string{id})
	if err != nil {
		return model.EdgesResponse{}, false, err
	}

	var matched []model.Edge
	for _, e := range edges {
		switch direction {
		case "in":
			if e.Target == id {
				matched = append(matched, e)
			}
		case "out":
			if e.Source == id {
				matched = append(matched, e)
			}
		default:
			matched = append(matched, e)
		}
	}
	if matched == nil {
		matched = []model.Edge{}
	}

	return model.EdgesResponse{NodeID: id, Edges: matched}, true, nil
}

// Search searches for nodes by name or description.
func (s *GraphService) Search(ctx context.Context, query string, limit int) (model.SearchResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	data, err := s.fetchGraphDataCached(ctx)
	if err != nil {
		return model.SearchResponse{}, err
	}
	coreComms, err := s.core.FetchCommunityTree(ctx)
	if err != nil {
		return model.SearchResponse{}, err
	}
	communityMap := buildCommunityIDMap(coreComms)

	q := strings.ToLower(query)
	results := make([]model.SearchResult, 0)
	for _, cn := range data.Nodes {
		if !strings.Contains(strings.ToLower(cn.Name), q) &&
			!strings.Contains(strings.ToLower(cn.Description), q) {
			continue
		}
		cid := 0
		if cn.CommunityID != nil {
			if mapped, ok := communityMap[*cn.CommunityID]; ok {
				cid = mapped
			}
		}
		results = append(results, model.SearchResult{
			ID:             cn.ID,
			Name:           cn.Name,
			CommunityID:    cid,
			InfluenceScore: cn.InfluenceScore,
		})
		if len(results) >= limit {
			break
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].InfluenceScore > results[j].InfluenceScore
	})
	if results == nil {
		results = []model.SearchResult{}
	}
	return model.SearchResponse{Query: query, Results: results}, nil
}

// Expand expands a set of nodes to include their neighbors and connecting edges.
// Only edges connected to the requested nodes are fetched from core (via the
// node_ids filter), instead of loading the entire edge set.
func (s *GraphService) Expand(ctx context.Context, ids []string, includeEdges, includeNeighbors bool) (model.ExpandResponse, error) {
	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}

	// Fetch edges connected to the requested nodes only.
	edges, err := s.fetchEdgesByNodeIDs(ctx, ids)
	if err != nil {
		return model.ExpandResponse{}, err
	}

	// Determine which node IDs we need to resolve.
	nodeIDsNeeded := make(map[string]bool)
	for _, id := range ids {
		nodeIDsNeeded[id] = true
	}
	if includeNeighbors {
		for _, e := range edges {
			if idSet[e.Source] {
				nodeIDsNeeded[e.Target] = true
			}
			if idSet[e.Target] {
				nodeIDsNeeded[e.Source] = true
			}
		}
	}

	// Fetch only the nodes we need.
	coreComms, err := s.core.FetchCommunityTree(ctx)
	if err != nil {
		return model.ExpandResponse{}, err
	}
	communityMap := buildCommunityIDMap(coreComms)

	data, err := s.fetchGraphDataCached(ctx)
	if err != nil {
		return model.ExpandResponse{}, err
	}

	nodesByID := make(map[string]model.Node, len(nodeIDsNeeded))
	for _, cn := range data.Nodes {
		if nodeIDsNeeded[cn.ID] {
			nodesByID[cn.ID] = toPortalNode(cn, communityMap)
		}
	}

	// Build edge set.
	edgeSet := make(map[[2]string]model.Edge)
	for _, e := range edges {
		if includeEdges && idSet[e.Source] && idSet[e.Target] {
			edgeSet[[2]string{e.Source, e.Target}] = e
		}
		if includeNeighbors && (nodeIDsNeeded[e.Source] || nodeIDsNeeded[e.Target]) {
			edgeSet[[2]string{e.Source, e.Target}] = e
		}
	}

	expanded := make([]model.Node, 0, len(nodesByID))
	for _, n := range nodesByID {
		expanded = append(expanded, n)
	}
	expandedEdges := make([]model.Edge, 0, len(edgeSet))
	for _, e := range edgeSet {
		expandedEdges = append(expandedEdges, e)
	}
	if expanded == nil {
		expanded = []model.Node{}
	}
	if expandedEdges == nil {
		expandedEdges = []model.Edge{}
	}

	return model.ExpandResponse{Nodes: expanded, Edges: expandedEdges}, nil
}

// TimelineRange returns the [minYear, maxYear] span of every node's
// `first_appeared` field, computed by core over the full graph. The
// portal does not aggregate locally; it forwards the call to core so
// the result reflects the entire dataset, not just the nodes the caller
// has loaded.
func (s *GraphService) TimelineRange(ctx context.Context) (model.TimelineRange, error) {
	core, err := s.core.FetchTimelineRange(ctx)
	if err != nil {
		return model.TimelineRange{}, err
	}
	return model.TimelineRange{
		MinYear: core.MinYear,
		MaxYear: core.MaxYear,
	}, nil
}

// fetchEdgesByNodeIDs fetches edges connected to the given node IDs.
// Filtering is pushed down to core (and ultimately Neo4j) so we no
// longer need to load the entire edge set into memory.
func (s *GraphService) fetchEdgesByNodeIDs(ctx context.Context, nodeIDs []string) ([]model.Edge, error) {
	return s.core.ListEdgesByNodeIDs(ctx, nodeIDs)
}

func (s *GraphService) fetchNodeIndex(ctx context.Context, communityMap map[string]int) (map[string]model.Node, error) {
	data, err := s.fetchGraphDataCached(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]model.Node, len(data.Nodes))
	for _, cn := range data.Nodes {
		out[cn.ID] = toPortalNode(cn, communityMap)
	}
	return out, nil
}

// toPortalNode converts a core node (string community id) into the
// portal's representation (integer community id).
func toPortalNode(cn coreclient.CoreNode, communityMap map[string]int) model.Node {
	cid := 0
	if cn.CommunityID != nil {
		if mapped, ok := communityMap[*cn.CommunityID]; ok {
			cid = mapped
		}
	}
	return model.Node{
		ID:             cn.ID,
		Name:           cn.Name,
		Description:    cn.Description,
		InfluenceScore: cn.InfluenceScore,
		FirstAppeared:  cn.FirstAppeared,
		Milestones:     cn.Milestones,
		CommunityID:    cid,
		Degree:         cn.Degree,
	}
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
// communities (integer IDs) using the provided mapping. The result is
// wrapped in a synthetic root node (id=0) so the frontend sidebar can
// render "All" as a container with the real communities as
// children. This matches the shape the mock data provides.
func convertCommunities(coreCommunities []coreclient.CoreCommunity, mapping map[string]int) []model.HierarchicalCommunity {
	children := make([]model.HierarchicalCommunity, 0, len(coreCommunities))
	totalNodes := 0
	for _, c := range coreCommunities {
		children = append(children, convertCommunity(c, mapping))
		totalNodes += c.NodeCount
	}
	root := model.HierarchicalCommunity{
		ID:        0,
		ParentID:  nil,
		Name:      "All",
		Color:     "#888888",
		Level:     0,
		NodeIDs:   []string{},
		NodeCount: totalNodes,
		Children:  children,
	}
	return []model.HierarchicalCommunity{root}
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
