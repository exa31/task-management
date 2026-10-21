package project

import (
	"context"
	"database/sql"
	"errors"

	"golang/database"
	customErrors "golang/pkg/errors"

	"github.com/jmoiron/sqlx"
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
	db   database.Transactor
}

func NewService(repo Repository, db database.Transactor) Service {
	return &service{repo: repo, db: db}
}

func (s *service) Create(ctx context.Context, userID string, req *CreateProjectRequest) (*ProjectResponse, error) {
	var res *ProjectResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		var err error
		res, err = s.repo.Create(ctx, tx, userID, req)
		return err
	})
	return res, err
}

func (s *service) List(ctx context.Context, userID string) ([]ProjectResponse, error) {
	var res []ProjectResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		var err error
		res, err = s.repo.FindAll(ctx, tx, userID)
		return err
	})
	return res, err
}

func (s *service) GetByID(ctx context.Context, userID string, id int64) (*ProjectResponse, error) {
	var p *ProjectResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		var err error
		p, err = s.repo.FindByID(ctx, tx, userID, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, customErrors.ProjectNotFound()
	}
	return p, nil
}

func (s *service) Update(ctx context.Context, userID string, id int64, req *UpdateProjectRequest) (*ProjectResponse, error) {
	var p *ProjectResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		var err error
		p, err = s.repo.Update(ctx, tx, userID, id, req)
		return err
	})
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, customErrors.ProjectNotFound()
	}
	return p, nil
}

func (s *service) Delete(ctx context.Context, userID string, id int64) error {
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Delete(ctx, tx, userID, id)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customErrors.ProjectNotFound()
		}
		return err
	}
	return nil
}
