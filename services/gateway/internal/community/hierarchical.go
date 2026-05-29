package community

import (
	"sync"

	"ai-graph-server/internal/model"

	"gonum.org/v1/gonum/graph/community"
	"gonum.org/v1/gonum/graph/simple"
)

// LevelPalettes defines color schemes for different hierarchy levels
var LevelPalettes = map[int][]string{
	0: {"#666666"},
	1: {"#4C78A8", "#F58518", "#E45756", "#72B7B2", "#54A24B"},
	2: {"#9D755D", "#BAB0AC", "#FF9DA6", "#EECA3B", "#B279A2"},
	3: {"#1f77b4", "#aec7e8", "#ff7f0e", "#ffbb78", "#2ca02c"},
}

// CommunityConfig contains configuration for hierarchical community detection
type CommunityConfig struct {
	MaxLevels        int // Maximum hierarchy depth (default: 3)
	MinCommunitySize int // Minimum number of nodes per community (default: 3)
}

var (
	communityIDCounter int        // Counter for generating unique community IDs
	communityIDMutex   sync.Mutex // Mutex for thread-safe ID generation
)

// nextCommunityID generates a unique community ID in a thread-safe manner
func nextCommunityID() int {
	communityIDMutex.Lock()
	defer communityIDMutex.Unlock()
	communityIDCounter++
	return communityIDCounter
}

// DetectHierarchical performs hierarchical community detection using recursive Louvain algorithm
// Returns the root community and the maximum depth detected
func DetectHierarchical(nodes []model.Node, edges []model.Edge, config CommunityConfig) (*model.HierarchicalCommunity, int) {
	// Set default values if not provided
	if config.MaxLevels <= 0 {
		config.MaxLevels = 3
	}
	if config.MinCommunitySize <= 0 {
		config.MinCommunitySize = 3
	}

	// Reset ID counter for new detection
	communityIDCounter = 0

	// Collect all node IDs for the root community
	allNodeIDs := make([]string, 0, len(nodes))
	for _, n := range nodes {
		allNodeIDs = append(allNodeIDs, n.ID)
	}

	// Create root community (level 0)
	root := &model.HierarchicalCommunity{
		ID:        nextCommunityID(),
		ParentID:  nil,
		Name:      "全部",
		Color:     getColor(0, 0),
		Level:     0,
		NodeIDs:   allNodeIDs,
		NodeCount: len(allNodeIDs),
	}

	// Recursively detect sub-communities
	maxLevel := 0
	detectSubCommunities(root, nodes, edges, 1, config, &maxLevel)

	return root, maxLevel
}

// detectSubCommunities recursively detects communities within a parent community
func detectSubCommunities(parent *model.HierarchicalCommunity, allNodes []model.Node, allEdges []model.Edge, currentLevel int, config CommunityConfig, maxLevel *int) {
	// Stop if we've reached maximum depth
	if currentLevel > config.MaxLevels {
		return
	}

	// Update max level counter
	if currentLevel > *maxLevel {
		*maxLevel = currentLevel
	}

	// Filter to get only nodes and edges within this parent community
	subNodes := filterNodesByIDs(allNodes, parent.NodeIDs)
	subEdges := filterEdgesContainingNodes(allEdges, parent.NodeIDs)

	// Need enough nodes to potentially split into multiple communities
	if len(subNodes) < config.MinCommunitySize*2 {
		return
	}

	// Run Louvain algorithm on the subgraph
	subCommunities := runLouvainOnSubgraph(subNodes, subEdges)

	// If only one community found, no need to go deeper
	if len(subCommunities) <= 1 {
		return
	}

	// Create child communities for each detected sub-community
	for i, subCommunityNodeIDs := range subCommunities {
		// Skip communities that are too small
		if len(subCommunityNodeIDs) < config.MinCommunitySize {
			continue
		}

		// Create child community
		childID := nextCommunityID()
		child := model.HierarchicalCommunity{
			ID:        childID,
			ParentID:  &parent.ID,
			Name:      nameCommunity(subCommunityNodeIDs, allNodes),
			Color:     getColor(currentLevel, i),
			Level:     currentLevel,
			NodeIDs:   subCommunityNodeIDs,
			NodeCount: len(subCommunityNodeIDs),
		}

		// Add child to parent
		parent.Children = append(parent.Children, child)

		// Recursively detect sub-communities within this child
		detectSubCommunities(&parent.Children[len(parent.Children)-1], allNodes, allEdges, currentLevel+1, config, maxLevel)
	}
}

// runLouvainOnSubgraph runs the Louvain community detection algorithm on a subset of the graph
func runLouvainOnSubgraph(subNodes []model.Node, subEdges []model.Edge) [][]string {
	// Create a new undirected graph
	g := simple.NewUndirectedGraph()

	// Maps for converting between string node IDs and int64 IDs (required by gonum)
	nodeIDToInt := make(map[string]int64)
	intToNodeID := make(map[int64]string)

	// Add nodes to the graph
	for i, n := range subNodes {
		id := int64(i)
		nodeIDToInt[n.ID] = id
		intToNodeID[id] = n.ID
		g.AddNode(simple.Node(id))
	}

	// Add edges to the graph
	for _, e := range subEdges {
		srcID, sok := nodeIDToInt[e.Source]
		tgtID, tok := nodeIDToInt[e.Target]
		if !sok || !tok {
			continue
		}
		from := g.Node(srcID)
		to := g.Node(tgtID)
		if from == nil || to == nil {
			continue
		}
		g.SetEdge(simple.WeightedEdge{F: from, T: to, W: e.Weight})
	}

	// Run Louvain algorithm to detect communities
	r := community.Modularize(g, 1.0, nil)
	communities := r.Communities()

	// Convert back to string node IDs
	result := make([][]string, 0, len(communities))
	for _, c := range communities {
		nodeIDs := make([]string, 0, len(c))
		for _, n := range c {
			nodeIDs = append(nodeIDs, intToNodeID[n.ID()])
		}
		result = append(result, nodeIDs)
	}

	return result
}

// filterNodesByIDs returns only the nodes with IDs in the given list
func filterNodesByIDs(allNodes []model.Node, ids []string) []model.Node {
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var filtered []model.Node
	for _, n := range allNodes {
		if idSet[n.ID] {
			filtered = append(filtered, n)
		}
	}
	return filtered
}

// filterEdgesContainingNodes returns edges where both source and target are in the ID list
func filterEdgesContainingNodes(allEdges []model.Edge, ids []string) []model.Edge {
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var filtered []model.Edge
	for _, e := range allEdges {
		if idSet[e.Source] && idSet[e.Target] {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// getColor returns a color from the palette based on hierarchy level and index
func getColor(level int, index int) string {
	palette, ok := LevelPalettes[level]
	if !ok {
		// Fallback to appropriate palette if level not found
		if level > 3 {
			palette = LevelPalettes[3]
		} else {
			palette = LevelPalettes[1]
		}
	}
	// Cycle through palette using modulo
	return palette[index%len(palette)]
}

// AssignHierarchicalCommunities assigns community IDs to nodes based on the hierarchy
// Each node gets the ID of the deepest (leaf) community it belongs to
func AssignHierarchicalCommunities(nodes []model.Node, root *model.HierarchicalCommunity) {
	nodeCommunityPath := make(map[string][]int) // Full path from root to leaf
	nodeLeafCommunity := make(map[string]int)   // Just the leaf community ID

	// Traverse the hierarchy recursively
	var traverse func(c *model.HierarchicalCommunity, path []int)
	traverse = func(c *model.HierarchicalCommunity, path []int) {
		// Build current path
		currentPath := make([]int, len(path)+1)
		copy(currentPath, path)
		currentPath[len(path)] = c.ID

		// Assign community info to nodes
		for _, id := range c.NodeIDs {
			nodeCommunityPath[id] = currentPath
			// If this is a leaf community (no children), assign the ID
			if len(c.Children) == 0 {
				nodeLeafCommunity[id] = c.ID
			}
		}

		// Recurse on children
		for _, child := range c.Children {
			traverse(&child, currentPath)
		}
	}

	traverse(root, []int{})

	// Update nodes with their leaf community ID
	for i := range nodes {
		if cid, ok := nodeLeafCommunity[nodes[i].ID]; ok {
			nodes[i].CommunityID = cid
		}
	}
}

// CountAllCommunities recursively counts all communities in the hierarchy
func CountAllCommunities(root *model.HierarchicalCommunity) int {
	count := 1 // Count this community
	for _, child := range root.Children {
		count += CountAllCommunities(&child)
	}
	return count
}
