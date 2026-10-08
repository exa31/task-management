package task

import (
	"context"

	"golang/database"
	"golang/internal/module/note"
	"golang/pkg/errors"

	"github.com/jmoiron/sqlx"
)

type Service interface {
	Create(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error)
	GetByID(ctx context.Context, userID string, id int64) (*TaskResponse, error)
	FindAll(ctx context.Context, userID string, filter TaskFilter) ([]TaskResponse, error)
	Update(ctx context.Context, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error)
	Delete(ctx context.Context, userID string, id int64) error
	LogProgress(ctx context.Context, userID string, id int64, req LogProgressRequest) (*TaskResponse, error)
	ToggleFocus(ctx context.Context, userID string, id int64) (*TaskResponse, error)

	CreateSubtask(ctx context.Context, userID string, taskID int64, req CreateSubtaskRequest) (*SubtaskResponse, error)
	UpdateSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error)
	DeleteSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64) error

	// CreateFromNote implements note.TaskCreator
	CreateFromNote(ctx context.Context, userID string, noteID int64, title string, content string, req note.ConvertToTaskRequest) (int64, error)
}

type service struct {
	repo Repository
	db   database.Transactor
}

func NewService(repo Repository, db database.Transactor) Service {
	return &service{repo: repo, db: db}
}

func (s *service) Create(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error) {
	var res *TaskResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var err error
		res, err = s.repo.Create(txCtx, userID, req)
		return err
	})
	return res, err
}

func (s *service) GetByID(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
	var entity *TaskResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var err error
		entity, err = s.repo.FindByID(txCtx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *service) FindAll(ctx context.Context, userID string, filter TaskFilter) ([]TaskResponse, error) {
	var res []TaskResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var err error
		res, err = s.repo.FindAll(txCtx, userID, filter)
		return err
	})
	return res, err
}

func (s *service) Update(ctx context.Context, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
	var res *TaskResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
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
			return errors.TaskNotFound()
		}
		return s.repo.Delete(txCtx, userID, id)
	})
}

func (s *service) LogProgress(ctx context.Context, userID string, id int64, req LogProgressRequest) (*TaskResponse, error) {
	var res *TaskResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		res, err = s.repo.LogProgress(txCtx, userID, id, req)
		return err
	})
	return res, err
}

func (s *service) ToggleFocus(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
	var res *TaskResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		res, err = s.repo.ToggleFocus(txCtx, userID, id)
		return err
	})
	return res, err
}

func (s *service) CreateSubtask(ctx context.Context, userID string, taskID int64, req CreateSubtaskRequest) (*SubtaskResponse, error) {
	var res *SubtaskResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, taskID)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		res, err = s.repo.CreateSubtask(txCtx, userID, taskID, req.Title)
		return err
	})
	return res, err
}

func (s *service) UpdateSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error) {
	var res *SubtaskResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, taskID)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		res, err = s.repo.UpdateSubtask(txCtx, userID, taskID, subtaskID, req)
		return err
	})
	return res, err
}

func (s *service) DeleteSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64) error {
	return s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, taskID)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		return s.repo.DeleteSubtask(txCtx, userID, taskID, subtaskID)
	})
}

func (s *service) CreateFromNote(ctx context.Context, userID string, noteID int64, title string, content string, req note.ConvertToTaskRequest) (int64, error) {
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	res, err := s.Create(ctx, userID, CreateTaskRequest{
		Title:        title,
		Description:  content,
		Status:       "todo",
		Priority:     priority,
		DueDate:      req.DueDate,
		SourceNoteID: &noteID,
		Subtasks:     req.Subtasks,
	})
	if err != nil {
		return 0, err
	}
	return res.ID, nil
}
