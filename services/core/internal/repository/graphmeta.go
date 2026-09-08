package repository

import (
	"context"
	"fmt"
	"time"

	"core/internal/db"
)

// GraphMetaRepository reads and writes graph-level metadata stored on a
// singleton Neo4j node (:GraphMeta {id: 'graph'}). The metadata lives next
// to the graph data itself so it survives restarts and is updated in the
// same database that the graph mutations target.
type GraphMetaRepository struct {
	neo *db.Neo4j
}

// NewGraphMetaRepository creates a new GraphMetaRepository.
func NewGraphMetaRepository(neo *db.Neo4j) *GraphMetaRepository {
	return &GraphMetaRepository{neo: neo}
}

// Touch updates the graph last_updated timestamp to now, creating the
// metadata node on first use. Callers invoke this after successful graph
// mutations (extraction merge, community re-detection) so clients can
// display when the graph was last changed.
func (r *GraphMetaRepository) Touch(ctx context.Context) error {
	err := r.neo.Execute(ctx, `
		MERGE (m:GraphMeta {id: 'graph'})
		ON CREATE SET m.created_at = datetime()
		SET m.last_updated = datetime()`, nil)
	if err != nil {
		return fmt.Errorf("touch graph meta: %w", err)
	}
	return nil
}

// LastUpdated returns the graph last_updated timestamp. ok is false when
// the metadata node does not exist yet (no mutation has ever run).
func (r *GraphMetaRepository) LastUpdated(ctx context.Context) (time.Time, bool, error) {
	recs, err := r.neo.Query(ctx, `
		MATCH (m:GraphMeta {id: 'graph'})
		RETURN m.last_updated AS ts`, nil)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("read graph meta: %w", err)
	}
	if len(recs) == 0 {
		return time.Time{}, false, nil
	}
	raw, ok := recs[0].Get("ts")
	if !ok || raw == nil {
		return time.Time{}, false, nil
	}
	ts, ok := raw.(time.Time)
	if !ok {
		return time.Time{}, false, fmt.Errorf("graph meta last_updated has unexpected type %T", raw)
	}
	return ts, true, nil
}
