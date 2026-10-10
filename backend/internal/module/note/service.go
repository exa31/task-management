package note

import (
	"context"

	"golang/database"
	"golang/pkg/errors"

	"github.com/jmoiron/sqlx"
)

type TaskCreator interface {
	CreateFromNote(ctx context.Context, userID string, noteID int64, title string, content string, req ConvertToTaskRequest) (int64, error)
}

type Service interface {
	Create(ctx context.Context, userID string, req CreateNoteRequest) (*NoteResponse, error)
	GetByID(ctx context.Context, userID string, id int64) (*NoteResponse, error)
	FindAll(ctx context.Context, userID string, filter NoteFilter) ([]NoteResponse, error)
	Update(ctx context.Context, userID string, id int64, req UpdateNoteRequest) (*NoteResponse, error)
	Delete(ctx context.Context, userID string, id int64) error
	ConvertToTask(ctx context.Context, userID string, id int64, req ConvertToTaskRequest) (*NoteResponse, int64, error)
	SetTaskCreator(creator TaskCreator)
}

type service struct {
	repo        Repository
	taskCreator TaskCreator
	db          database.Transactor
}

func NewService(repo Repository, db database.Transactor) Service {
	return &service{repo: repo, db: db}
}

func (s *service) SetTaskCreator(creator TaskCreator) {
	s.taskCreator = creator
}

func (s *service) Create(ctx context.Context, userID string, req CreateNoteRequest) (*NoteResponse, error) {
	var res *NoteResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var err error
		res, err = s.repo.Create(txCtx, userID, req)
		return err
	})
	return res, err
}

func (s *service) GetByID(ctx context.Context, userID string, id int64) (*NoteResponse, error) {
	var entity *NoteResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var err error
		entity, err = s.repo.FindByID(txCtx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.NoteNotFound()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *service) FindAll(ctx context.Context, userID string, filter NoteFilter) ([]NoteResponse, error) {
	var res []NoteResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var err error
		res, err = s.repo.FindAll(txCtx, userID, filter)
		return err
	})
	return res, err
}

func (s *service) Update(ctx context.Context, userID string, id int64, req UpdateNoteRequest) (*NoteResponse, error) {
	var res *NoteResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.NoteNotFound()
		}
		res, err = s.repo.Update(txCtx, userID, id, req)
		return err
	})
	return res, err
}

func (s *service) Delete(ctx context.Context, userID string, id int64) error {
	return s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.NoteNotFound()
		}
		return s.repo.Delete(txCtx, userID, id)
	})
}

func (s *service) ConvertToTask(ctx context.Context, userID string, id int64, req ConvertToTaskRequest) (*NoteResponse, int64, error) {
	var res *NoteResponse
	var taskID int64

	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.NoteNotFound()
		}

		if entity.Type == "knowledge" {
			return errors.NoteNotActionable()
		}

		if s.taskCreator == nil {
			return errors.InternalError("Task integration not configured", nil)
		}

		var errTask error
		taskID, errTask = s.taskCreator.CreateFromNote(txCtx, userID, id, entity.Title, entity.Content, req)
		if errTask != nil {
			return errTask
		}

		if err := s.repo.MarkConverted(txCtx, userID, id, taskID); err != nil {
			return err
		}

		entity.ConvertedTaskID = &taskID
		res = entity
		return nil
	})

	if err != nil {
		return nil, 0, err
	}
	return res, taskID, nil
}
