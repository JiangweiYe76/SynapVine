package service

import (
	"context"
	"fmt"

	"core/internal/model"
	"core/internal/repository"
)

// CommunityService provides business logic for community operations.
type CommunityService struct {
	repo *repository.CommunityRepository
}

// NewCommunityService creates a new CommunityService.
func NewCommunityService(repo *repository.CommunityRepository) *CommunityService {
	return &CommunityService{repo: repo}
}

// List returns all communities.
func (s *CommunityService) List(ctx context.Context) (*model.CommunitiesListResponse, error) {
	communities, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return &model.CommunitiesListResponse{Communities: communities}, nil
}

// Get returns a community by ID.
func (s *CommunityService) Get(ctx context.Context, id int) (*model.Community, error) {
	return s.repo.Get(ctx, id)
}

// Create creates a new community.
func (s *CommunityService) Create(ctx context.Context, req model.CommunityCreateRequest) (*model.Community, error) {
	exists, err := s.repo.Exists(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("community already exists")
	}
	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, req.ID)
}

// Update updates an existing community.
func (s *CommunityService) Update(ctx context.Context, id int, req model.CommunityUpdateRequest) (*model.Community, error) {
	return s.repo.Update(ctx, id, req)
}

// Delete deletes a community by ID.
func (s *CommunityService) Delete(ctx context.Context, id int) error {
	_, err := s.repo.Delete(ctx, id)
	return err
}
