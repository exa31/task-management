package project

import (
	"context"
	"database/sql"
	"errors"

	customErrors "golang/pkg/errors"
)

type Service interface {
	Create(ctx context.Context, userID string, req *CreateProjectRequest) (*ProjectResponse, error)
	List(ctx context.Context, userID string) ([]ProjectResponse, error)
	GetByID(ctx context.Context, userID string, id int64) (*ProjectResponse, error)
	Update(ctx context.Context, userID string, id int64, req *UpdateProjectRequest) (*ProjectResponse, error)
	Delete(ctx context.Context, userID string, id int64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, userID string, req *CreateProjectRequest) (*ProjectResponse, error) {
	return s.repo.Create(ctx, userID, req)
}

func (s *service) List(ctx context.Context, userID string) ([]ProjectResponse, error) {
	return s.repo.FindAll(ctx, userID)
}

func (s *service) GetByID(ctx context.Context, userID string, id int64) (*ProjectResponse, error) {
	p, err := s.repo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, customErrors.ProjectNotFound()
	}
	return p, nil
}

func (s *service) Update(ctx context.Context, userID string, id int64, req *UpdateProjectRequest) (*ProjectResponse, error) {
	p, err := s.repo.Update(ctx, userID, id, req)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, customErrors.ProjectNotFound()
	}
	return p, nil
}

func (s *service) Delete(ctx context.Context, userID string, id int64) error {
	err := s.repo.Delete(ctx, userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customErrors.ProjectNotFound()
		}
		return err
	}
	return nil
}
