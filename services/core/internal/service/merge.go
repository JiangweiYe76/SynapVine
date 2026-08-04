// Package service implements business logic. This file hosts the
// MergeService, which writes an approved extraction result (nodes and
// edges produced by the discovery LLM pipeline) into the Neo4j graph.
//
// Merge semantics (see .trae/documents/review-merge-to-neo4j.md):
//   - Nodes are deduplicated by case-insensitive name. An existing node
//     is reused as-is; none of its fields are overwritten.
//   - Newly created nodes carry source='extraction' and source_paper_id
//     so their provenance is traceable.
//   - Relevance (0-10) is scaled to influence_score (0-1) for new nodes.
//   - Edges are created idempotently via MERGE; existing edges are left
//     untouched (no field overwrite).
//   - Edges whose endpoint names cannot be resolved (LLM hallucination)
//     and self-loops are skipped with a warning.
//   - The whole operation runs in a single Neo4j transaction so a
//     failure rolls back every change.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"core/internal/db"
	"core/internal/model"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// MergeService writes approved extraction results into Neo4j.
type MergeService struct {
	neo *db.Neo4j
}

// NewMergeService creates a MergeService backed by the given Neo4j client.
func NewMergeService(neo *db.Neo4j) *MergeService {
	return &MergeService{neo: neo}
}

// Merge writes the given extracted nodes and edges into the Neo4j graph
// inside a single transaction. paperID is recorded on every newly
// created node for provenance. The operation is idempotent: re-running
// it with the same input reuses existing nodes and skips existing edges.
func (s *MergeService) Merge(ctx context.Context, paperID string, nodes []model.ExtractedNode, edges []model.ExtractedEdge) (*model.MergeResult, error) {
	result := &model.MergeResult{}
	nameToID := make(map[string]string) // extracted name -> Neo4j node id

	// Collect unique, trimmed node names for the existing-node lookup.
	nameSet := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		nm := strings.TrimSpace(n.Name)
		if nm != "" {
			nameSet[nm] = true
		}
	}
	names := make([]string, 0, len(nameSet))
	for nm := range nameSet {
		names = append(names, nm)
	}

	err := s.neo.ExecuteInTx(ctx, func(tx neo4j.ManagedTransaction) error {
		// Step 1: find existing nodes by case-insensitive name match.
		// ORDER BY created_at so the first row per name is the earliest
		// created node (deterministic reuse when duplicates exist).
		if len(names) > 0 {
			recs, err := runCollect(ctx, tx, `
				UNWIND $names AS name
				MATCH (n:Concept)
				WHERE toLower(n.name) = toLower(name)
				RETURN name AS q, n.id AS id, n.created_at AS c
				ORDER BY c`, map[string]any{"names": names})
			if err != nil {
				return fmt.Errorf("match existing nodes: %w", err)
			}
			for _, rec := range recs {
				q := recordString(rec, "q")
				if q == "" {
					continue
				}
				if _, ok := nameToID[q]; !ok { // first record wins (earliest)
					nameToID[q] = recordString(rec, "id")
				}
			}
		}
		result.ReusedNodes = len(nameToID)

		// Step 2: create the nodes that were not found, batched in one
		// UNWIND. Relevance (0-10) is clamped and scaled to influence_score
		// (0-1). Duplicate names within the payload are skipped via seen.
		seen := make(map[string]bool, len(nodes))
		var newNodes []map[string]any
		for _, n := range nodes {
			nm := strings.TrimSpace(n.Name)
			if nm == "" || seen[nm] {
				continue
			}
			seen[nm] = true
			if _, ok := nameToID[nm]; ok {
				continue // already exists, will be reused
			}
			score := n.Relevance / 10.0
			if score < 0 {
				score = 0
			} else if score > 1 {
				score = 1
			}
			newNodes = append(newNodes, map[string]any{
				"name":            nm,
				"description":     n.Description,
				"influence_score": score,
			})
		}
		if len(newNodes) > 0 {
			recs, err := runCollect(ctx, tx, `
				UNWIND $new_nodes AS node
				CREATE (n:Concept {
					id: randomUUID(),
					name: node.name,
					description: node.description,
					influence_score: node.influence_score,
					first_appeared: '',
					milestones: [],
					source: 'extraction',
					source_paper_id: $paper_id,
					status: 'active',
					created_at: datetime()
				})
				RETURN n.name AS name, n.id AS id`, map[string]any{
				"new_nodes": newNodes,
				"paper_id":  paperID,
			})
			if err != nil {
				return fmt.Errorf("create extracted nodes: %w", err)
			}
			for _, rec := range recs {
				nameToID[recordString(rec, "name")] = recordString(rec, "id")
			}
			result.CreatedNodes = len(recs)
		}

		// Step 3: resolve edges. source/target are node names from the LLM;
		// map them to the ids resolved above. Skip edges whose endpoints
		// cannot be resolved (hallucinated names) and self-loops.
		var resolved []map[string]any
		for _, e := range edges {
			sName := strings.TrimSpace(e.Source)
			tName := strings.TrimSpace(e.Target)
			sid, sok := nameToID[sName]
			tid, tok := nameToID[tName]
			if !sok || !tok {
				result.SkippedEdges++
				slog.Warn("merge_edge_unresolved_endpoint",
					slog.String("paper_id", paperID),
					slog.String("source", sName),
					slog.String("target", tName))
				continue
			}
			if sid == tid {
				result.SkippedEdges++
				slog.Warn("merge_edge_self_loop",
					slog.String("paper_id", paperID),
					slog.String("name", sName))
				continue
			}
			w := e.Weight
			if w < 0 {
				w = 0
			} else if w > 1 {
				w = 1
			}
			resolved = append(resolved, map[string]any{
				"sid":      sid,
				"tid":      tid,
				"weight":   w,
				"relation": e.Relation,
			})
		}

		if len(resolved) > 0 {
			// 3a: find which edges already exist so we can report the
			// number of truly new edges (MERGE itself is idempotent, but
			// it does not tell us whether it created or matched).
			recs, err := runCollect(ctx, tx, `
				UNWIND $edges AS edge
				MATCH (s:Concept {id: edge.sid})-[r:RELATES_TO]->(t:Concept {id: edge.tid})
				RETURN edge.sid AS s, edge.tid AS t`, map[string]any{"edges": resolved})
			if err != nil {
				return fmt.Errorf("match existing edges: %w", err)
			}
			existing := make(map[string]bool, len(recs))
			for _, rec := range recs {
				existing[recordString(rec, "s")+"|"+recordString(rec, "t")] = true
			}

			// 3b: idempotently create the relationships. ON CREATE SET
			// only fires for newly created edges; existing edges keep
			// their current weight/relation (no overwrite).
			if _, err := tx.Run(ctx, `
				UNWIND $edges AS edge
				MATCH (s:Concept {id: edge.sid}), (t:Concept {id: edge.tid})
				MERGE (s)-[r:RELATES_TO]->(t)
				ON CREATE SET r.weight = edge.weight, r.relation = edge.relation`,
				map[string]any{"edges": resolved}); err != nil {
				return fmt.Errorf("merge edges: %w", err)
			}
			result.CreatedEdges = len(resolved) - len(existing)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("merge extracted data: %w", err)
	}

	slog.Info("merge_completed",
		slog.String("paper_id", paperID),
		slog.Int("created_nodes", result.CreatedNodes),
		slog.Int("reused_nodes", result.ReusedNodes),
		slog.Int("created_edges", result.CreatedEdges),
		slog.Int("skipped_edges", result.SkippedEdges))
	return result, nil
}

// runCollect runs a query within a managed transaction and collects all
// returned records. It is a small wrapper around tx.Run + Result.Collect
// to keep the call sites readable.
func runCollect(ctx context.Context, tx neo4j.ManagedTransaction, cypher string, params map[string]any) ([]*neo4j.Record, error) {
	res, err := tx.Run(ctx, cypher, params)
	if err != nil {
		return nil, err
	}
	return res.Collect(ctx)
}

// recordString safely extracts a string column from a Neo4j record,
// returning "" when the column is missing or not a string.
func recordString(rec *neo4j.Record, key string) string {
	v, ok := rec.Get(key)
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
