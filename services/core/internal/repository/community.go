package repository

import (
	"context"
	"fmt"
	"strings"

	"core/internal/db"
	"core/internal/model"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// CommunityRepository provides CRUD operations for Community nodes.
type CommunityRepository struct {
	neo *db.Neo4j
}

// NewCommunityRepository creates a new CommunityRepository.
func NewCommunityRepository(neo *db.Neo4j) *CommunityRepository {
	return &CommunityRepository{neo: neo}
}

// List returns all communities.
func (r *CommunityRepository) List(ctx context.Context) ([]model.Community, error) {
	query := `
		MATCH (c:Community)
		RETURN c.id AS id, c.name AS name, c.color AS color, c.level AS level, c.domain AS domain
		ORDER BY c.id
	`
	records, err := r.neo.Query(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list communities: %w", err)
	}

	communities := make([]model.Community, 0, len(records))
	for _, rec := range records {
		communities = append(communities, recordToCommunity(rec))
	}
	return communities, nil
}

// Get returns a community by ID.
func (r *CommunityRepository) Get(ctx context.Context, id int) (*model.Community, error) {
	query := `
		MATCH (c:Community {id: $id})
		RETURN c.id AS id, c.name AS name, c.color AS color, c.level AS level, c.domain AS domain
	`
	records, err := r.neo.Query(ctx, query, map[string]any{"id": id})
	if err != nil {
		return nil, fmt.Errorf("failed to get community: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	comm := recordToCommunity(records[0])
	return &comm, nil
}

// Create creates a new Community node.
func (r *CommunityRepository) Create(ctx context.Context, req model.CommunityCreateRequest) error {
	cypher := `
		CREATE (c:Community {
			id: $id,
			name: $name,
			color: $color,
			level: $level,
			domain: $domain
		})
	`
	return r.neo.Execute(ctx, cypher, map[string]any{
		"id":     req.ID,
		"name":   req.Name,
		"color":  req.Color,
		"level":  req.Level,
		"domain": req.Domain,
	})
}

// Update updates an existing Community node.
func (r *CommunityRepository) Update(ctx context.Context, id int, req model.CommunityUpdateRequest) (*model.Community, error) {
	setClauses := []string{}
	params := map[string]any{"id": id}

	if req.Name != nil {
		setClauses = append(setClauses, "c.name = $name")
		params["name"] = *req.Name
	}
	if req.Color != nil {
		setClauses = append(setClauses, "c.color = $color")
		params["color"] = *req.Color
	}
	if req.Level != nil {
		setClauses = append(setClauses, "c.level = $level")
		params["level"] = *req.Level
	}
	if req.Domain != nil {
		setClauses = append(setClauses, "c.domain = $domain")
		params["domain"] = *req.Domain
	}

	if len(setClauses) == 0 {
		return r.Get(ctx, id)
	}

	cypher := fmt.Sprintf(`
		MATCH (c:Community {id: $id})
		SET %s
		RETURN c.id AS id, c.name AS name, c.color AS color, c.level AS level, c.domain AS domain
	`, strings.Join(setClauses, ", "))

	records, err := r.neo.QueryWrite(ctx, cypher, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update community: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	comm := recordToCommunity(records[0])
	return &comm, nil
}

// Delete deletes a Community node by ID.
func (r *CommunityRepository) Delete(ctx context.Context, id int) (bool, error) {
	cypher := `MATCH (c:Community {id: $id}) DETACH DELETE c`
	err := r.neo.Execute(ctx, cypher, map[string]any{"id": id})
	if err != nil {
		return false, fmt.Errorf("failed to delete community: %w", err)
	}
	return true, nil
}

// Exists checks if a community exists.
func (r *CommunityRepository) Exists(ctx context.Context, id int) (bool, error) {
	query := `MATCH (c:Community {id: $id}) RETURN count(c) AS count`
	records, err := r.neo.Query(ctx, query, map[string]any{"id": id})
	if err != nil {
		return false, fmt.Errorf("failed to check community existence: %w", err)
	}
	count := records[0].Values[0].(int64)
	return count > 0, nil
}

// ClearAll removes all Community nodes and BELONGS_TO relationships.
func (r *CommunityRepository) ClearAll(ctx context.Context) error {
	cypher := `
		MATCH (c:Community)
		OPTIONAL MATCH (c)<-[b:BELONGS_TO]-(:Concept)
		DELETE b, c
	`
	return r.neo.Execute(ctx, cypher, nil)
}

// CreateBatch creates multiple Community nodes in a single transaction.
func (r *CommunityRepository) CreateBatch(ctx context.Context, communities []model.Community) error {
	cypher := `
		UNWIND $communities AS comm
		CREATE (c:Community {
			id: comm.id,
			name: comm.name,
			color: comm.color,
			level: comm.level,
			domain: comm.domain
		})
	`
	params := make([]map[string]any, 0, len(communities))
	for _, comm := range communities {
		params = append(params, map[string]any{
			"id":     comm.ID,
			"name":   comm.Name,
			"color":  comm.Color,
			"level":  comm.Level,
			"domain": comm.Domain,
		})
	}
	return r.neo.Execute(ctx, cypher, map[string]any{"communities": params})
}

// AssignNodesBatch creates BELONGS_TO relationships between Concepts and Communities.
func (r *CommunityRepository) AssignNodesBatch(ctx context.Context, assignments []struct {
	NodeID      string `json:"node_id"`
	CommunityID int    `json:"community_id"`
}) error {
	cypher := `
		UNWIND $assignments AS a
		MATCH (n:Concept {id: a.node_id}), (c:Community {id: a.community_id})
		MERGE (n)-[:BELONGS_TO]->(c)
	`
	return r.neo.Execute(ctx, cypher, map[string]any{"assignments": assignments})
}

func recordToCommunity(rec *neo4j.Record) model.Community {
	return model.Community{
		ID:     int(valueOrDefault(rec, "id", int64(0))),
		Name:   valueOrEmpty[string](rec, "name"),
		Color:  valueOrEmpty[string](rec, "color"),
		Level:  int(valueOrDefault(rec, "level", int64(0))),
		Domain: valueOrEmpty[string](rec, "domain"),
	}
}
