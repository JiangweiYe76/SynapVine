package repository

import (
	"context"
	"fmt"
	"strings"

	"core/internal/db"
	"core/internal/model"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// NodeRepository provides CRUD operations for Concept nodes.
type NodeRepository struct {
	neo *db.Neo4j
}

// NewNodeRepository creates a new NodeRepository.
func NewNodeRepository(neo *db.Neo4j) *NodeRepository {
	return &NodeRepository{neo: neo}
}

// nodeSelectFields is the standard RETURN clause used by all node read paths.
// community_id is computed via the BELONGS_TO relationship to a Community.
const nodeSelectFields = `
		n.id AS id, n.name AS name,
		n.description AS description, n.influence_score AS influence_score,
		n.first_appeared AS first_appeared, n.milestones AS milestones,
		[(n)-[:BELONGS_TO]->(c:Community) | c.id][0] AS community_id
	`

// List returns paginated nodes.
func (r *NodeRepository) List(ctx context.Context, offset, limit int) ([]model.Node, int, error) {
	countQuery := "MATCH (n:Concept) RETURN count(n) AS total"
	countResult, err := r.neo.Query(ctx, countQuery, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count nodes: %w", err)
	}
	total := int(countResult[0].Values[0].(int64))

	query := fmt.Sprintf(`
		MATCH (n:Concept)
		RETURN %s
		SKIP $offset LIMIT $limit
	`, nodeSelectFields)
	records, err := r.neo.Query(ctx, query, map[string]any{
		"offset": offset,
		"limit":  limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodes := make([]model.Node, 0, len(records))
	for _, rec := range records {
		nodes = append(nodes, recordToNode(rec))
	}
	return nodes, total, nil
}

// Search returns nodes matching the query string.
func (r *NodeRepository) Search(ctx context.Context, q string, offset, limit int) ([]model.Node, int, error) {
	countQuery := `
		MATCH (n:Concept)
		WHERE toLower(n.name) CONTAINS $q OR toLower(n.id) CONTAINS $q OR toLower(n.description) CONTAINS $q
		RETURN count(n) AS total
	`
	countResult, err := r.neo.Query(ctx, countQuery, map[string]any{"q": strings.ToLower(q)})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}
	total := int(countResult[0].Values[0].(int64))

	query := fmt.Sprintf(`
		MATCH (n:Concept)
		WHERE toLower(n.name) CONTAINS $q OR toLower(n.id) CONTAINS $q OR toLower(n.description) CONTAINS $q
		RETURN %s
		SKIP $offset LIMIT $limit
	`, nodeSelectFields)
	records, err := r.neo.Query(ctx, query, map[string]any{
		"q":      strings.ToLower(q),
		"offset": offset,
		"limit":  limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search nodes: %w", err)
	}

	nodes := make([]model.Node, 0, len(records))
	for _, rec := range records {
		nodes = append(nodes, recordToNode(rec))
	}
	return nodes, total, nil
}

// Get returns a node by ID.
func (r *NodeRepository) Get(ctx context.Context, id string) (*model.Node, error) {
	query := fmt.Sprintf(`
		MATCH (n:Concept {id: $id})
		RETURN %s
	`, nodeSelectFields)
	records, err := r.neo.Query(ctx, query, map[string]any{"id": id})
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	node := recordToNode(records[0])
	return &node, nil
}

// Create creates a new Concept node and optionally attaches it to a community.
//
// If req.ID is empty a fresh UUID is generated server-side and used as the
// node's identifier. The repository returns the resulting id via the
// returned string so the caller can persist/display it.
func (r *NodeRepository) Create(ctx context.Context, req model.NodeCreateRequest) (string, error) {
	id := req.ID
	if id == "" {
		id = uuid.NewString()
	}
	cypher := `
		CREATE (n:Concept {
			id: $id,
			name: $name,
			description: $description,
			influence_score: $influence_score,
			first_appeared: $first_appeared,
			milestones: $milestones,
			source: 'manual',
			status: 'active',
			created_at: datetime()
		})
		WITH n
		OPTIONAL MATCH (c:Community {id: $community_id})
		FOREACH (_ IN CASE WHEN c IS NULL THEN [] ELSE [1] END |
			MERGE (n)-[:BELONGS_TO]->(c)
		)
	`
	return id, r.neo.Execute(ctx, cypher, map[string]any{
		"id":              id,
		"name":            req.Name,
		"description":     req.Description,
		"influence_score": req.InfluenceScore,
		"first_appeared":  req.FirstAppeared,
		"milestones":      req.Milestones,
		"community_id":    req.CommunityID,
	})
}

// Update updates an existing Concept node.
//
// CommunityID is a tri-state pointer-to-pointer field: nil means "leave
// unchanged", a non-nil pointer to a nil means "remove from community", and
// a non-nil pointer to a non-nil int means "assign to the given community".
// The tri-state is required because JSON cannot otherwise distinguish
// "absent" from "null".
func (r *NodeRepository) Update(ctx context.Context, id string, req model.NodeUpdateRequest) (*model.Node, error) {
	setClauses := []string{}
	params := map[string]any{"id": id}

	if req.Name != nil {
		setClauses = append(setClauses, "n.name = $name")
		params["name"] = *req.Name
	}
	if req.Description != nil {
		setClauses = append(setClauses, "n.description = $description")
		params["description"] = *req.Description
	}
	if req.InfluenceScore != nil {
		setClauses = append(setClauses, "n.influence_score = $influence_score")
		params["influence_score"] = *req.InfluenceScore
	}
	if req.FirstAppeared != nil {
		setClauses = append(setClauses, "n.first_appeared = $first_appeared")
		params["first_appeared"] = *req.FirstAppeared
	}
	if req.Milestones != nil {
		setClauses = append(setClauses, "n.milestones = $milestones")
		params["milestones"] = *req.Milestones
	}

	// Build the post-update community reconciliation. We always run a MATCH
	// so the RETURN clause can project community_id consistently.
	communityCypher := `
		MATCH (n:Concept {id: $id})
		OPTIONAL MATCH (n)-[old:BELONGS_TO]->(:Community)
		DELETE old
		WITH n
		OPTIONAL MATCH (c:Community {id: $new_community_id})
		FOREACH (_ IN CASE WHEN c IS NULL THEN [] ELSE [1] END |
			MERGE (n)-[:BELONGS_TO]->(c)
		)
		WITH n
	`
	params["new_community_id"] = nil
	if req.CommunityID != nil {
		if *req.CommunityID != nil {
			params["new_community_id"] = **req.CommunityID
		}
		// else: explicit null -> community will be removed by the DELETE above
	}

	returnCypher := fmt.Sprintf("RETURN %s", nodeSelectFields)

	var cypher string
	if len(setClauses) > 0 {
		cypher = fmt.Sprintf(`
			MATCH (n:Concept {id: $id})
			SET %s
			WITH n
			%s
			%s
		`, strings.Join(setClauses, ", "), communityCypher, returnCypher)
	} else {
		cypher = fmt.Sprintf(`
			%s
			%s
		`, communityCypher, returnCypher)
	}

	records, err := r.neo.QueryWrite(ctx, cypher, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update node: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	node := recordToNode(records[0])
	return &node, nil
}

// Delete deletes a Concept node by ID.
func (r *NodeRepository) Delete(ctx context.Context, id string) (bool, error) {
	cypher := `MATCH (n:Concept {id: $id}) DETACH DELETE n`
	err := r.neo.Execute(ctx, cypher, map[string]any{"id": id})
	if err != nil {
		return false, fmt.Errorf("failed to delete node: %w", err)
	}
	return true, nil
}

// Exists checks if a node exists.
func (r *NodeRepository) Exists(ctx context.Context, id string) (bool, error) {
	query := `MATCH (n:Concept {id: $id}) RETURN count(n) AS count`
	records, err := r.neo.Query(ctx, query, map[string]any{"id": id})
	if err != nil {
		return false, fmt.Errorf("failed to check node existence: %w", err)
	}
	count := records[0].Values[0].(int64)
	return count > 0, nil
}

// TimelineRange returns the [minYear, maxYear] span covered by the
// `first_appeared` field of every Concept node in the graph. Nodes with
// a missing or empty `first_appeared` are ignored.
//
// When the graph has no nodes with a valid `first_appeared` the returned
// range is the zero value (0, 0); callers should treat that as "no data".
//
// The query uses `substring(n.first_appeared, 0, 4)` to extract the year
// prefix from the stored "YYYY-MM" string. This is computed in Cypher
// rather than in Go so that the database does the aggregation, which
// keeps the cost O(N) on the server and constant on the client.
func (r *NodeRepository) TimelineRange(ctx context.Context) (model.TimelineRange, error) {
	query := `
		MATCH (n:Concept)
		WHERE n.first_appeared IS NOT NULL AND n.first_appeared <> ''
		WITH min(toInteger(substring(n.first_appeared, 0, 4))) AS minYear,
		     max(toInteger(substring(n.first_appeared, 0, 4))) AS maxYear
		RETURN minYear, maxYear
	`
	records, err := r.neo.Query(ctx, query, nil)
	if err != nil {
		return model.TimelineRange{}, fmt.Errorf("failed to compute timeline range: %w", err)
	}
	if len(records) == 0 {
		return model.TimelineRange{}, nil
	}
	rec := records[0]
	minYear := valueOrDefault(rec, "minYear", int64(0))
	maxYear := valueOrDefault(rec, "maxYear", int64(0))
	return model.TimelineRange{
		MinYear: int(minYear),
		MaxYear: int(maxYear),
	}, nil
}

// GetAll returns all nodes and edges in the graph.
func (r *NodeRepository) GetAll(ctx context.Context) ([]model.Node, []model.Edge, error) {
	nodeQuery := fmt.Sprintf(`
		MATCH (n:Concept)
		RETURN %s
	`, nodeSelectFields)
	nodeRecords, err := r.neo.Query(ctx, nodeQuery, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get all nodes: %w", err)
	}

	nodes := make([]model.Node, 0, len(nodeRecords))
	for _, rec := range nodeRecords {
		nodes = append(nodes, recordToNode(rec))
	}

	edgeQuery := `
		MATCH (s:Concept)-[r:RELATES_TO]->(t:Concept)
		RETURN s.id AS source, t.id AS target, r.weight AS weight, r.relation AS relation
	`
	edgeRecords, err := r.neo.Query(ctx, edgeQuery, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get all edges: %w", err)
	}

	edges := make([]model.Edge, 0, len(edgeRecords))
	for _, rec := range edgeRecords {
		edges = append(edges, model.Edge{
			Source:   valueOrEmpty[string](rec, "source"),
			Target:   valueOrEmpty[string](rec, "target"),
			Weight:   valueOrDefault(rec, "weight", 0.0),
			Relation: valueOrEmpty[string](rec, "relation"),
		})
	}

	return nodes, edges, nil
}

func recordToNode(rec *neo4j.Record) model.Node {
	node := model.Node{
		ID:             valueOrEmpty[string](rec, "id"),
		Name:           valueOrEmpty[string](rec, "name"),
		Description:    valueOrEmpty[string](rec, "description"),
		InfluenceScore: valueOrDefault(rec, "influence_score", 0.0),
		FirstAppeared:  valueOrDefault(rec, "first_appeared", ""),
		CommunityID:    valueOrNilString(rec, "community_id"),
	}
	if milestones, ok := rec.Get("milestones"); ok && milestones != nil {
		if arr, ok := milestones.([]any); ok {
			node.Milestones = make([]string, 0, len(arr))
			for _, v := range arr {
				if s, ok := v.(string); ok {
					node.Milestones = append(node.Milestones, s)
				}
			}
		}
	}
	return node
}

func valueOrEmpty[T any](rec *neo4j.Record, key string) T {
	var zero T
	if v, ok := rec.Get(key); ok && v != nil {
		if t, ok := v.(T); ok {
			return t
		}
	}
	return zero
}

func valueOrDefault[T any](rec *neo4j.Record, key string, def T) T {
	if v, ok := rec.Get(key); ok && v != nil {
		if t, ok := v.(T); ok {
			return t
		}
	}
	return def
}
