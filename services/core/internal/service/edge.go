package service

import (
	"context"
	"errors"

	"core/internal/model"
	"core/internal/repository"
)

// Sentinel errors returned by EdgeService so handlers can map them to
// HTTP status codes. ErrNodeNotFound is shared with NodeService.
var (
	ErrEdgeNotFound  = errors.New("edge not found")
	ErrEdgeExists    = errors.New("edge already exists")
	ErrInvalidWeight = errors.New("weight must be in [0, 1]")
	ErrSameEndpoints = errors.New("source and target must differ")
)

// EdgeService provides business logic for RELATES_TO edge operations.
type EdgeService struct {
	repo *repository.EdgeRepository
}

// NewEdgeService creates a new EdgeService.
//
// The service validates that both endpoint nodes exist and that the
// (source, target) pair is well-formed before any write.
func NewEdgeService(repo *repository.EdgeRepository) *EdgeService {
	return &EdgeService{repo: repo}
}

// List returns paginated edges. When search is non-empty the result is
// restricted to edges whose source, target, or relation label contains
// the query string (case-insensitive).
func (s *EdgeService) List(ctx context.Context, offset, limit int, search string) (*model.EdgesListResponse, error) {
	var edges []model.Edge
	var total int
	var err error
	if search != "" {
		edges, total, err = s.repo.Search(ctx, search, offset, limit)
	} else {
		edges, total, err = s.repo.List(ctx, offset, limit)
	}
	if err != nil {
		return nil, err
	}
	return &model.EdgesListResponse{
		Edges: edges,
		Pagination: model.Pagination{
			Offset:  offset,
			Limit:   limit,
			Total:   total,
			HasMore: offset+limit < total,
		},
	}, nil
}

// ListByNodeIDs returns all edges connected to any of the given node IDs.
// Filtering is done in Neo4j, not in memory.
func (s *EdgeService) ListByNodeIDs(ctx context.Context, nodeIDs []string) ([]model.Edge, error) {
	return s.repo.ListByNodeIDs(ctx, nodeIDs)
}

// Get returns the edge identified by (source, target). It returns
// (nil, nil) when the edge does not exist so handlers can map that to 404.
func (s *EdgeService) Get(ctx context.Context, source, target string) (*model.Edge, error) {
	return s.repo.Get(ctx, source, target)
}

// Create inserts a new edge. It validates that:
//   - source and target are non-empty and distinct
//   - weight is in [0, 1]
//   - both endpoint nodes exist
//   - the edge does not already exist
func (s *EdgeService) Create(ctx context.Context, req model.EdgeCreateRequest) (*model.Edge, error) {
	if req.Source == "" || req.Target == "" {
		return nil, ErrNodeNotFound
	}
	if req.Source == req.Target {
		return nil, ErrSameEndpoints
	}
	if err := validateWeight(req.Weight); err != nil {
		return nil, err
	}
	if err := s.endpointExists(ctx, req.Source); err != nil {
		return nil, err
	}
	if err := s.endpointExists(ctx, req.Target); err != nil {
		return nil, err
	}
	exists, err := s.repo.Exists(ctx, req.Source, req.Target)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEdgeExists
	}
	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, req.Source, req.Target)
}

// Update applies a partial update to an existing edge. The (source, target)
// pair cannot be changed. Weight (when supplied) is validated to be in
// [0, 1]. Returns (nil, nil) when the edge does not exist.
func (s *EdgeService) Update(ctx context.Context, source, target string, req model.EdgeUpdateRequest) (*model.Edge, error) {
	if source == "" || target == "" {
		return nil, ErrEdgeNotFound
	}
	if req.Weight != nil {
		if err := validateWeight(*req.Weight); err != nil {
			return nil, err
		}
	}
	exists, err := s.repo.Exists(ctx, source, target)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	return s.repo.Update(ctx, source, target, req)
}

// Delete removes an edge. Returns ErrEdgeNotFound when the edge does
// not exist so handlers can return 404 instead of 204.
func (s *EdgeService) Delete(ctx context.Context, source, target string) error {
	exists, err := s.repo.Exists(ctx, source, target)
	if err != nil {
		return err
	}
	if !exists {
		return ErrEdgeNotFound
	}
	_, err = s.repo.Delete(ctx, source, target)
	return err
}

// endpointExists verifies that a Concept node with the given id exists.
func (s *EdgeService) endpointExists(ctx context.Context, id string) error {
	exists, err := s.repo.EndpointExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNodeNotFound
	}
	return nil
}

// validateWeight ensures the weight is within the closed interval [0, 1].
func validateWeight(w float64) error {
	if w < 0 || w > 1 {
		return ErrInvalidWeight
	}
	return nil
}
