package community

import (
	"sync"

	"core/internal/model"

	"gonum.org/v1/gonum/graph/community"
	"gonum.org/v1/gonum/graph/simple"
)

// LevelPalettes defines color schemes for different hierarchy levels.
var LevelPalettes = map[int][]string{
	0: {"#666666"},
	1: {"#4C78A8", "#F58518", "#E45756", "#72B7B2", "#54A24B"},
	2: {"#9D755D", "#BAB0AC", "#FF9DA6", "#EECA3B", "#B279A2"},
	3: {"#1f77b4", "#aec7e8", "#ff7f0e", "#ffbb78", "#2ca02c"},
}

// CommunityConfig contains configuration for hierarchical community detection.
type CommunityConfig struct {
	MaxLevels        int // Maximum hierarchy depth (default: 3)
	MinCommunitySize int // Minimum number of nodes per community (default: 3)
}

var (
	communityIDCounter int
	communityIDMutex   sync.Mutex
)

func nextCommunityID() int {
	communityIDMutex.Lock()
	defer communityIDMutex.Unlock()
	communityIDCounter++
	return communityIDCounter
}

// DetectHierarchical performs hierarchical community detection using recursive Louvain.
func DetectHierarchical(nodes []model.Node, edges []model.Edge, config CommunityConfig) (*model.HierarchicalCommunity, int) {
	if config.MaxLevels <= 0 {
		config.MaxLevels = 3
	}
	if config.MinCommunitySize <= 0 {
		config.MinCommunitySize = 3
	}

	communityIDCounter = 0

	allNodeIDs := make([]string, 0, len(nodes))
	for _, n := range nodes {
		allNodeIDs = append(allNodeIDs, n.ID)
	}

	root := &model.HierarchicalCommunity{
		ID:        nextCommunityID(),
		ParentID:  nil,
		Name:      "All",
		Color:     getColor(0, 0),
		Level:     0,
		NodeIDs:   allNodeIDs,
		NodeCount: len(allNodeIDs),
	}

	maxLevel := 0
	detectSubCommunities(root, nodes, edges, 1, config, &maxLevel)

	return root, maxLevel
}

func detectSubCommunities(parent *model.HierarchicalCommunity, allNodes []model.Node, allEdges []model.Edge, currentLevel int, config CommunityConfig, maxLevel *int) {
	if currentLevel > config.MaxLevels {
		return
	}
	if currentLevel > *maxLevel {
		*maxLevel = currentLevel
	}

	subNodes := filterNodesByIDs(allNodes, parent.NodeIDs)
	subEdges := filterEdgesContainingNodes(allEdges, parent.NodeIDs)

	if len(subNodes) < config.MinCommunitySize*2 {
		return
	}

	subCommunities := runLouvainOnSubgraph(subNodes, subEdges)
	if len(subCommunities) <= 1 {
		return
	}

	for i, subCommunityNodeIDs := range subCommunities {
		if len(subCommunityNodeIDs) < config.MinCommunitySize {
			continue
		}

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
		parent.Children = append(parent.Children, child)
		detectSubCommunities(&parent.Children[len(parent.Children)-1], allNodes, allEdges, currentLevel+1, config, maxLevel)
	}
}

func runLouvainOnSubgraph(subNodes []model.Node, subEdges []model.Edge) [][]string {
	g := simple.NewUndirectedGraph()

	nodeIDToInt := make(map[string]int64)
	intToNodeID := make(map[int64]string)

	for i, n := range subNodes {
		id := int64(i)
		nodeIDToInt[n.ID] = id
		intToNodeID[id] = n.ID
		g.AddNode(simple.Node(id))
	}

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

	r := community.Modularize(g, 1.0, nil)
	communities := r.Communities()

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

func getColor(level int, index int) string {
	palette, ok := LevelPalettes[level]
	if !ok {
		if level > 3 {
			palette = LevelPalettes[3]
		} else {
			palette = LevelPalettes[1]
		}
	}
	return palette[index%len(palette)]
}

// AssignHierarchicalCommunities assigns the deepest (leaf) community ID to each node.
func AssignHierarchicalCommunities(nodes []model.Node, root *model.HierarchicalCommunity) {
	nodeLeafCommunity := make(map[string]int)

	var traverse func(c *model.HierarchicalCommunity)
	traverse = func(c *model.HierarchicalCommunity) {
		for _, id := range c.NodeIDs {
			if len(c.Children) == 0 {
				nodeLeafCommunity[id] = c.ID
			}
		}
		for _, child := range c.Children {
			traverse(&child)
		}
	}
	traverse(root)

	for i := range nodes {
		if cid, ok := nodeLeafCommunity[nodes[i].ID]; ok {
			nodes[i].CommunityID = cid
		}
	}
}

// CountAllCommunities recursively counts all communities in the hierarchy.
func CountAllCommunities(root *model.HierarchicalCommunity) int {
	count := 1
	for _, child := range root.Children {
		count += CountAllCommunities(&child)
	}
	return count
}

// FlattenHierarchicalCommunities flattens the hierarchical tree into a slice of communities.
func FlattenHierarchicalCommunities(root *model.HierarchicalCommunity) []model.Community {
	var result []model.Community
	var traverse func(c *model.HierarchicalCommunity)
	traverse = func(c *model.HierarchicalCommunity) {
		result = append(result, model.Community{
			ID:        c.ID,
			Name:      c.Name,
			Color:     c.Color,
			Level:     c.Level,
			Domain:    c.Name,
			NodeIDs:   c.NodeIDs,
			NodeCount: c.NodeCount,
		})
		for _, child := range c.Children {
			traverse(&child)
		}
	}
	traverse(root)
	return result
}
