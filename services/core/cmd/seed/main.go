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

// GraphData is the on-disk format consumed by this seed. Grouping
// happens via Community membership, which is either assigned
// explicitly via community_name or detected later by the Louvain
// community detector at portal startup.
type GraphData struct {
	Nodes []struct {
		ID             string  `json:"id"`
		Name           string  `json:"name"`
		CommunityName  string  `json:"community_name,omitempty"`
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

// seedConcepts creates Concept nodes. Each concept's community_name
// (if present) is matched against an existing Community and a
// BELONGS_TO relationship is created; otherwise the concept is left
// unassigned and Louvain will cluster it at portal startup.
func seedConcepts(ctx context.Context, neo *db.Neo4j, nodes []struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	CommunityName  string  `json:"community_name,omitempty"`
	Description    string  `json:"description"`
	InfluenceScore float64 `json:"influence_score"`
	FirstAppeared  string  `json:"first_appeared"`
}) error {
	cypher := `
		UNWIND $nodes AS node
		MERGE (c:Concept {id: node.id})
		SET c.name = node.name,
		    c.description = node.description,
		    c.influence_score = node.influence_score,
		    c.first_appeared = node.first_appeared,
		    c.source = 'manual',
		    c.status = 'active',
		    c.created_at = datetime()
		WITH c, node
		OPTIONAL MATCH (comm:Community {name: node.community_name})
		FOREACH (_ IN CASE WHEN comm IS NULL THEN [] ELSE [1] END |
			MERGE (c)-[:BELONGS_TO]->(comm)
		)
	`

	params := make([]map[string]any, 0, len(nodes))
	for _, n := range nodes {
		params = append(params, map[string]any{
			"id":              n.ID,
			"name":            n.Name,
			"description":     n.Description,
			"influence_score": n.InfluenceScore,
			"first_appeared":  n.FirstAppeared,
			"community_name":  n.CommunityName,
		})
	}

	return neo.Execute(ctx, cypher, map[string]any{"nodes": params})
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

	edgeParams := make([]map[string]any, 0, len(edges))
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
