package repository

import (
	"context"
	"fmt"
	"strings"

	"core/internal/db"
	"core/internal/model"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// edgeSelectFields is the standard RETURN clause used by edge read paths.
const edgeSelectFields = `
		s.id AS source, t.id AS target,
		r.weight AS weight, r.relation AS relation
	`

// EdgeRepository provides CRUD operations for RELATES_TO relationships
// between Concept nodes. Edges are addressed by the composite
// (source, target) pair.
type EdgeRepository struct {
	neo *db.Neo4j
}

// NewEdgeRepository creates a new EdgeRepository.
func NewEdgeRepository(neo *db.Neo4j) *EdgeRepository {
	return &EdgeRepository{neo: neo}
}

// List returns paginated RELATES_TO edges.
func (r *EdgeRepository) List(ctx context.Context, offset, limit int) ([]model.Edge, int, error) {
	countQuery := `MATCH ()-[r:RELATES_TO]->() RETURN count(r) AS total`
	countResult, err := r.neo.Query(ctx, countQuery, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count edges: %w", err)
	}
	total := int(recordCount(countResult))

	query := fmt.Sprintf(`
		MATCH (s:Concept)-[r:RELATES_TO]->(t:Concept)
		RETURN %s
		ORDER BY s.id, t.id
		SKIP $offset LIMIT $limit
	`, edgeSelectFields)
	records, err := r.neo.Query(ctx, query, map[string]any{
		"offset": offset,
		"limit":  limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list edges: %w", err)
	}

	edges := make([]model.Edge, 0, len(records))
	for _, rec := range records {
		edges = append(edges, recordToEdge(rec))
	}
	return edges, total, nil
}

// Search returns edges whose source id, target id, or relation label
// contains the query string (case-insensitive).
func (r *EdgeRepository) Search(ctx context.Context, q string, offset, limit int) ([]model.Edge, int, error) {
	lower := strings.ToLower(q)

	countQuery := `
		MATCH (s:Concept)-[r:RELATES_TO]->(t:Concept)
		WHERE toLower(s.id) CONTAINS $q
		   OR toLower(t.id) CONTAINS $q
		   OR toLower(r.relation) CONTAINS $q
		RETURN count(r) AS total
	`
	countResult, err := r.neo.Query(ctx, countQuery, map[string]any{"q": lower})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count edge search results: %w", err)
	}
	total := int(recordCount(countResult))

	query := fmt.Sprintf(`
		MATCH (s:Concept)-[r:RELATES_TO]->(t:Concept)
		WHERE toLower(s.id) CONTAINS $q
		   OR toLower(t.id) CONTAINS $q
		   OR toLower(r.relation) CONTAINS $q
		RETURN %s
		ORDER BY s.id, t.id
		SKIP $offset LIMIT $limit
	`, edgeSelectFields)
	records, err := r.neo.Query(ctx, query, map[string]any{
		"q":      lower,
		"offset": offset,
		"limit":  limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search edges: %w", err)
	}

	edges := make([]model.Edge, 0, len(records))
	for _, rec := range records {
		edges = append(edges, recordToEdge(rec))
	}
	return edges, total, nil
}

// Get returns the edge identified by the (source, target) pair.
// It returns (nil, nil) when the edge does not exist.
func (r *EdgeRepository) Get(ctx context.Context, source, target string) (*model.Edge, error) {
	query := fmt.Sprintf(`
		MATCH (s:Concept {id: $source})-[r:RELATES_TO]->(t:Concept {id: $target})
		RETURN %s
	`, edgeSelectFields)
	records, err := r.neo.Query(ctx, query, map[string]any{
		"source": source,
		"target": target,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get edge: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	edge := recordToEdge(records[0])
	return &edge, nil
}

// Exists reports whether the edge identified by (source, target) is present.
func (r *EdgeRepository) Exists(ctx context.Context, source, target string) (bool, error) {
	query := `
		MATCH (s:Concept {id: $source})-[r:RELATES_TO]->(t:Concept {id: $target})
		RETURN count(r) AS count
	`
	records, err := r.neo.Query(ctx, query, map[string]any{
		"source": source,
		"target": target,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check edge existence: %w", err)
	}
	return recordCount(records) > 0, nil
}

// ListByNodeIDs returns edges where either source or target is in the
// given set of node IDs. This pushes the filtering into Neo4j instead
// of loading all edges and filtering in memory.
func (r *EdgeRepository) ListByNodeIDs(ctx context.Context, nodeIDs []string) ([]model.Edge, error) {
	query := fmt.Sprintf(`
		MATCH (s:Concept)-[r:RELATES_TO]->(t:Concept)
		WHERE s.id IN $ids OR t.id IN $ids
		RETURN %s
	`, edgeSelectFields)
	records, err := r.neo.Query(ctx, query, map[string]any{
		"ids": nodeIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list edges by node IDs: %w", err)
	}

	edges := make([]model.Edge, 0, len(records))
	for _, rec := range records {
		edges = append(edges, recordToEdge(rec))
	}
	return edges, nil
}

// EndpointExists reports whether the Concept node with the given id exists.
// Used by the service layer to validate edge endpoints before creating.
func (r *EdgeRepository) EndpointExists(ctx context.Context, id string) (bool, error) {
	query := `MATCH (n:Concept {id: $id}) RETURN count(n) AS count`
	records, err := r.neo.Query(ctx, query, map[string]any{"id": id})
	if err != nil {
		return false, fmt.Errorf("failed to check node existence: %w", err)
	}
	return recordCount(records) > 0, nil
}

// Create inserts a new RELATES_TO relationship. The caller must have
// already validated that the endpoints exist and that the edge does not
// already exist; Create will return a unique-constraint error from the
// driver if a duplicate is attempted. A uniqueness constraint on the
// pair is not enforced at the schema level because Neo4j does not
// natively support composite uniqueness on relationships; duplicates are
// rejected by the service layer via Exists.
func (r *EdgeRepository) Create(ctx context.Context, req model.EdgeCreateRequest) error {
	cypher := `
		MATCH (s:Concept {id: $source}), (t:Concept {id: $target})
		CREATE (s)-[r:RELATES_TO {
			weight: $weight,
			relation: $relation
		}]->(t)
	`
	return r.neo.Execute(ctx, cypher, map[string]any{
		"source":   req.Source,
		"target":   req.Target,
		"weight":   req.Weight,
		"relation": req.Relation,
	})
}

// Update applies the given partial update to an existing edge and returns
// the post-update state. The (source, target) pair cannot be changed; if
// no fields are supplied, the existing edge is returned unchanged.
func (r *EdgeRepository) Update(ctx context.Context, source, target string, req model.EdgeUpdateRequest) (*model.Edge, error) {
	setClauses := []string{}
	params := map[string]any{
		"source": source,
		"target": target,
	}
	if req.Weight != nil {
		setClauses = append(setClauses, "r.weight = $weight")
		params["weight"] = *req.Weight
	}
	if req.Relation != nil {
		setClauses = append(setClauses, "r.relation = $relation")
		params["relation"] = *req.Relation
	}

	var cypher string
	if len(setClauses) > 0 {
		cypher = fmt.Sprintf(`
			MATCH (s:Concept {id: $source})-[r:RELATES_TO]->(t:Concept {id: $target})
			SET %s
			WITH s, r, t
			RETURN %s
		`, strings.Join(setClauses, ", "), edgeSelectFields)
	} else {
		cypher = fmt.Sprintf(`
			MATCH (s:Concept {id: $source})-[r:RELATES_TO]->(t:Concept {id: $target})
			RETURN %s
		`, edgeSelectFields)
	}

	records, err := r.neo.QueryWrite(ctx, cypher, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update edge: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	edge := recordToEdge(records[0])
	return &edge, nil
}

// Delete removes the edge identified by (source, target).
// It returns (false, nil) when the edge did not exist.
func (r *EdgeRepository) Delete(ctx context.Context, source, target string) (bool, error) {
	cypher := `
		MATCH (s:Concept {id: $source})-[r:RELATES_TO]->(t:Concept {id: $target})
		DELETE r
		RETURN count(r) AS deleted
	`
	records, err := r.neo.QueryWrite(ctx, cypher, map[string]any{
		"source": source,
		"target": target,
	})
	if err != nil {
		return false, fmt.Errorf("failed to delete edge: %w", err)
	}
	return recordCount(records) > 0, nil
}

func recordToEdge(rec *neo4j.Record) model.Edge {
	return model.Edge{
		Source:   valueOrEmpty[string](rec, "source"),
		Target:   valueOrEmpty[string](rec, "target"),
		Weight:   valueOrDefault(rec, "weight", 0.0),
		Relation: valueOrEmpty[string](rec, "relation"),
	}
}
