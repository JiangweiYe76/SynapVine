package service

import (
	"context"
	"fmt"
	"log/slog"

	"core/internal/community"
	"core/internal/repository"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// CommunityDetectorService orchestrates community detection and writes results back to Neo4j.
type CommunityDetectorService struct {
	nodeRepo *repository.NodeRepository
	commRepo *repository.CommunityRepository
}

// NewCommunityDetectorService creates a new CommunityDetectorService.
func NewCommunityDetectorService(nodeRepo *repository.NodeRepository, commRepo *repository.CommunityRepository) *CommunityDetectorService {
	return &CommunityDetectorService{
		nodeRepo: nodeRepo,
		commRepo: commRepo,
	}
}

// DetectAndStore runs Louvain community detection on the full graph and persists results to Neo4j.
func (s *CommunityDetectorService) DetectAndStore(ctx context.Context) error {
	nodes, edges, err := s.nodeRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load graph: %w", err)
	}

	slog.Info("community_detection_started",
		slog.Int("nodes", len(nodes)),
		slog.Int("edges", len(edges)),
	)

	// Flat communities
	flatCommunities := community.Detect(nodes, edges)
	community.AssignCommunities(nodes, flatCommunities)
	community.ComputeDegrees(&nodes, edges)

	// Hierarchical communities
	config := community.CommunityConfig{
		MaxLevels:        3,
		MinCommunitySize: 3,
	}
	root, maxLevel := community.DetectHierarchical(nodes, edges, config)
	community.AssignHierarchicalCommunities(nodes, root)

	flatCount := len(flatCommunities)
	hierCount := community.CountAllCommunities(root)

	slog.Info("community_detection_completed",
		slog.Int("flat_communities", flatCount),
		slog.Int("hierarchical_communities", hierCount),
		slog.Int("max_level", maxLevel),
	)

	// Flatten hierarchical tree for persistence
	allComms := community.FlattenHierarchicalCommunities(root)

	// Build node-to-community assignments (leaf community ID)
	assignments := make([]struct {
		NodeID      string `json:"node_id"`
		CommunityID string `json:"community_id"`
	}, 0, len(nodes))
	for _, n := range nodes {
		if n.CommunityID == nil {
			continue
		}
		assignments = append(assignments, struct {
			NodeID      string `json:"node_id"`
			CommunityID string `json:"community_id"`
		}{
			NodeID:      n.ID,
			CommunityID: *n.CommunityID,
		})
	}

	// Write back to Neo4j in a single transaction. If any step fails
	// the entire transaction is rolled back and no community data is lost.
	if err := s.commRepo.ExecuteInTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := s.commRepo.ClearAllTx(ctx, tx); err != nil {
			return fmt.Errorf("clear old communities: %w", err)
		}
		if len(allComms) > 0 {
			if err := s.commRepo.CreateBatchTx(ctx, tx, allComms); err != nil {
				return fmt.Errorf("create communities: %w", err)
			}
		}
		if len(assignments) > 0 {
			if err := s.commRepo.AssignNodesBatchTx(ctx, tx, assignments); err != nil {
				return fmt.Errorf("assign nodes: %w", err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to persist communities: %w", err)
	}

	slog.Info("community_detection_persisted",
		slog.Int("communities_created", len(allComms)),
		slog.Int("assignments_created", len(assignments)),
	)

	return nil
}
