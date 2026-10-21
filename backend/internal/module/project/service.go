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

func NewService(repo Repository, db ...database.Transactor) Service {
	var transactor database.Transactor
	if len(db) > 0 {
		transactor = db[0]
	}
	return &service{repo: repo, db: transactor}
}

func (s *service) Create(ctx context.Context, userID string, req *CreateProjectRequest) (*ProjectResponse, error) {
	if s.db != nil {
		var res *ProjectResponse
		err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
			var err error
			res, err = s.repo.Create(ctx, tx, userID, req)
			return err
		})
		return res, err
	}
	return s.repo.Create(ctx, nil, userID, req)
}

func (s *service) List(ctx context.Context, userID string) ([]ProjectResponse, error) {
	if s.db != nil {
		var res []ProjectResponse
		err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
			var err error
			res, err = s.repo.FindAll(ctx, tx, userID)
			return err
		})
		return res, err
	}
	return s.repo.FindAll(ctx, nil, userID)
}

func (s *service) GetByID(ctx context.Context, userID string, id int64) (*ProjectResponse, error) {
	var p *ProjectResponse
	var err error
	if s.db != nil {
		err = s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
			var err2 error
			p, err2 = s.repo.FindByID(ctx, tx, userID, id)
			return err2
		})
	} else {
		p, err = s.repo.FindByID(ctx, nil, userID, id)
	}
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
	var err error
	if s.db != nil {
		err = s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
			var err2 error
			p, err2 = s.repo.Update(ctx, tx, userID, id, req)
			return err2
		})
	} else {
		p, err = s.repo.Update(ctx, nil, userID, id, req)
	}
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, customErrors.ProjectNotFound()
	}
	return p, nil
}

func (s *service) Delete(ctx context.Context, userID string, id int64) error {
	var err error
	if s.db != nil {
		err = s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
			return s.repo.Delete(ctx, tx, userID, id)
		})
	} else {
		err = s.repo.Delete(ctx, nil, userID, id)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customErrors.ProjectNotFound()
		}
		return err
	}
	return nil
}
