package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"time"

	"core/internal/db"
)

// categoryToCommunity maps graph.json category to community ID
type categoryToCommunity struct {
	CommunityID   int
	CommunityName string
}

var categoryMap = map[string]categoryToCommunity{
	"dl_arch":           {1, "深度学习架构"},
	"dl_mechanism":      {1, "深度学习架构"},
	"nlp_model":         {2, "NLP模型"},
	"cv_model":          {3, "计算机视觉"},
	"gen_model":         {4, "生成模型"},
	"multimodal":        {5, "多模态"},
	"speech_model":      {6, "语音模型"},
	"gnn":               {7, "图神经网络"},
	"rl_algorithm":      {8, "强化学习"},
	"dl_technique":      {9, "深度学习技术"},
	"nlp_technique":     {10, "NLP技术"},
	"optimizer":         {11, "优化器"},
	"alignment":         {12, "AI对齐"},
	"platform":          {13, "平台与基础设施"},
	"infrastructure":    {13, "平台与基础设施"},
	"application":       {14, "应用"},
}

type GraphData struct {
	Nodes []struct {
		ID             string  `json:"id"`
		Name           string  `json:"name"`
		Category       string  `json:"category"`
		Description    string  `json:"description"`
		InfluenceScore float64 `json:"influence_score"`
		FirstAppeared  string  `json:"first_appeared"`
	} `json:"nodes"`
	Edges []struct {
		Source   string  `json:"source"`
		Target   string  `json:"target"`
		Weight   float64 `json:"weight"`
		Relation string  `json:"relation"`
	} `json:"edges"`
}

func main() {
	dataPath := flag.String("data", "../../data/graph.json", "Path to graph.json")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	data, err := loadGraphData(*dataPath)
	if err != nil {
		slog.Error("failed_to_load_data", slog.Any("error", err))
		os.Exit(1)
	}

	cfg := db.LoadConfigFromEnv()
	neo, err := db.New(cfg)
	if err != nil {
		slog.Error("failed_to_connect_neo4j", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	defer neo.Close(ctx)

	if err := seedConcepts(ctx, neo, data.Nodes); err != nil {
		slog.Error("seed_concepts_failed", slog.Any("error", err))
		os.Exit(1)
	}

	if err := seedEdges(ctx, neo, data.Edges); err != nil {
		slog.Error("seed_edges_failed", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("seed_completed",
		slog.Int("nodes", len(data.Nodes)),
		slog.Int("edges", len(data.Edges)),
	)
}

func loadGraphData(path string) (*GraphData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data GraphData
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func seedConcepts(ctx context.Context, neo *db.Neo4j, nodes []struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Category       string  `json:"category"`
	Description    string  `json:"description"`
	InfluenceScore float64 `json:"influence_score"`
	FirstAppeared  string  `json:"first_appeared"`
}) error {
	cypher := `
		UNWIND $nodes AS node
		MERGE (c:Concept {id: node.id})
		SET c.name = node.name,
		    c.category = node.category,
		    c.description = node.description,
		    c.influence_score = node.influence_score,
		    c.first_appeared = node.first_appeared,
		    c.source = 'manual',
		    c.status = 'active',
		    c.created_at = datetime()
		WITH c, node
		MATCH (comm:Community {id: node.community_id})
		MERGE (c)-[:BELONGS_TO]->(comm)
	`

	var nodeParams []map[string]any
	for _, n := range nodes {
		comm := categoryMap[n.Category]
		nodeParams = append(nodeParams, map[string]any{
			"id":              n.ID,
			"name":            n.Name,
			"category":        n.Category,
			"description":     n.Description,
			"influence_score": n.InfluenceScore,
			"first_appeared":  n.FirstAppeared,
			"community_id":    comm.CommunityID,
		})
	}

	return neo.Execute(ctx, cypher, map[string]any{"nodes": nodeParams})
}

func seedEdges(ctx context.Context, neo *db.Neo4j, edges []struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Weight   float64 `json:"weight"`
	Relation string  `json:"relation"`
}) error {
	cypher := `
		UNWIND $edges AS edge
		MATCH (s:Concept {id: edge.source}), (t:Concept {id: edge.target})
		MERGE (s)-[r:RELATES_TO]->(t)
		SET r.weight = edge.weight,
		    r.relation = edge.relation,
		    r.source_type = 'manual',
		    r.created_at = datetime()
	`

	var edgeParams []map[string]any
	for _, e := range edges {
		edgeParams = append(edgeParams, map[string]any{
			"source":   e.Source,
			"target":   e.Target,
			"weight":   e.Weight,
			"relation": e.Relation,
		})
	}

	return neo.Execute(ctx, cypher, map[string]any{"edges": edgeParams})
}
