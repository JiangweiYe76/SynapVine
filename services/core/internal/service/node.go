package service

import (
	"context"
	"fmt"

	"core/internal/model"
	"core/internal/repository"
)

// NodeService provides business logic for node operations.
type NodeService struct {
	repo *repository.NodeRepository
}

// NewNodeService creates a new NodeService.
func NewNodeService(repo *repository.NodeRepository) *NodeService {
	return &NodeService{repo: repo}
}

// List returns paginated nodes.
func (s *NodeService) List(ctx context.Context, offset, limit int, search string) (*model.NodesListResponse, error) {
	var nodes []model.Node
	var total int
	var err error

	if search != "" {
		nodes, total, err = s.repo.Search(ctx, search, offset, limit)
	} else {
		nodes, total, err = s.repo.List(ctx, offset, limit)
	}
	if err != nil {
		return nil, err
	}

	return &model.NodesListResponse{
		Nodes: nodes,
		Pagination: model.Pagination{
			Offset:  offset,
			Limit:   limit,
			Total:   total,
			HasMore: offset+limit < total,
		},
	}, nil
}

// Get returns a node by ID.
func (s *NodeService) Get(ctx context.Context, id string) (*model.Node, error) {
	return s.repo.Get(ctx, id)
}

// Create creates a new node.
func (s *NodeService) Create(ctx context.Context, req model.NodeCreateRequest) (*model.Node, error) {
	exists, err := s.repo.Exists(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("node already exists")
	}
	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, req.ID)
}

// Update updates an existing node.
func (s *NodeService) Update(ctx context.Context, id string, req model.NodeUpdateRequest) (*model.Node, error) {
	return s.repo.Update(ctx, id, req)
}

// Delete deletes a node by ID.
func (s *NodeService) Delete(ctx context.Context, id string) error {
	_, err := s.repo.Delete(ctx, id)
	return err
}
