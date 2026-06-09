package community

import (
	"sort"

	"core/internal/model"

	"gonum.org/v1/gonum/graph/community"
	"gonum.org/v1/gonum/graph/simple"
)

// Palette defines default colors for flat communities.
var Palette = []string{
	"#4C78A8", "#F58518", "#E45756", "#72B7B2",
	"#54A24B", "#EECA3B", "#B279A2", "#FF9DA6",
	"#9D755D", "#BAB0AC",
}

// Detect runs the Louvain algorithm and returns flat communities.
func Detect(nodes []model.Node, edges []model.Edge) []model.Community {
	g := simple.NewUndirectedGraph()

	nodeIDToInt := make(map[string]int64)
	intToNodeID := make(map[int64]string)

	for i, n := range nodes {
		id := int64(i)
		nodeIDToInt[n.ID] = id
		intToNodeID[id] = n.ID
		g.AddNode(simple.Node(id))
	}

	for _, e := range edges {
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

	result := make([]model.Community, 0, len(communities))
	for i, c := range communities {
		nodeIDs := make([]string, 0, len(c))
		for _, n := range c {
			nodeIDs = append(nodeIDs, intToNodeID[n.ID()])
		}

		name := nameCommunity(nodeIDs, nodes)

		color := ""
		if i < len(Palette) {
			color = Palette[i]
		}

		result = append(result, model.Community{
			ID:     i,
			Name:   name,
			Color:  color,
			Level:  1,
			Domain: name,
		})
	}

	return result
}

func nameCommunity(memberIDs []string, nodes []model.Node) string {
	counts := make(map[string]int)
	for _, id := range memberIDs {
		for _, n := range nodes {
			if n.ID == id {
				counts[n.Category]++
				break
			}
		}
	}
	maxCat, maxCount := "", 0
	for cat, count := range counts {
		if count > maxCount {
			maxCat, maxCount = cat, count
		}
	}
	return maxCat
}

// AssignCommunities updates each node with its flat community ID.
func AssignCommunities(nodes []model.Node, communities []model.Community) {
	nodeCommunity := make(map[string]int)
	for _, c := range communities {
		for _, id := range c.NodeIDs {
			nodeCommunity[id] = c.ID
		}
	}
	for i := range nodes {
		if cid, ok := nodeCommunity[nodes[i].ID]; ok {
			nodes[i].CommunityID = cid
		}
	}
}

// ComputeDegrees calculates the degree of each node.
func ComputeDegrees(nodes *[]model.Node, edges []model.Edge) {
	degree := make(map[string]int)
	for _, n := range *nodes {
		degree[n.ID] = 0
	}
	for _, e := range edges {
		degree[e.Source]++
		degree[e.Target]++
	}
	for i := range *nodes {
		(*nodes)[i].Degree = degree[(*nodes)[i].ID]
	}
}

// GetTopNodes returns the top N nodes by influence score.
func GetTopNodes(nodes []model.Node, limit int) []model.Node {
	sorted := make([]model.Node, len(nodes))
	copy(sorted, nodes)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].InfluenceScore > sorted[j].InfluenceScore
	})
	if limit > len(sorted) {
		limit = len(sorted)
	}
	return sorted[:limit]
}
