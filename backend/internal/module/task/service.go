package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang/database"
	"golang/internal/module/note"
	"golang/internal/module/reminder"
	"golang/pkg/errors"
	"golang/pkg/notification"
	"golang/pkg/realtime"

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

	// SetNotificationDependencies injects optional alerting services (Asynq, email, push, SSE, logger)
	SetNotificationDependencies(
		scheduler notification.TaskDueScheduler,
		emailSender notification.EmailSender,
		pushSender notification.PushSender,
		reminderRepo reminder.Repository,
		hub *realtime.Hub,
		logger *slog.Logger,
	)

	// ProcessDueWarning is invoked by external background Asynq worker when task due alert triggers
	ProcessDueWarning(ctx context.Context, taskID int64) error
}

type service struct {
	repo         Repository
	db           database.Transactor
	scheduler    notification.TaskDueScheduler
	emailSender  notification.EmailSender
	pushSender   notification.PushSender
	reminderRepo reminder.Repository
	hub          *realtime.Hub
	logger       *slog.Logger
}

func NewService(repo Repository, db database.Transactor) Service {
	return &service{repo: repo, db: db}
}

func (s *service) SetNotificationDependencies(
	scheduler notification.TaskDueScheduler,
	emailSender notification.EmailSender,
	pushSender notification.PushSender,
	reminderRepo reminder.Repository,
	hub *realtime.Hub,
	logger *slog.Logger,
) {
	s.scheduler = scheduler
	s.emailSender = emailSender
	s.pushSender = pushSender
	s.reminderRepo = reminderRepo
	s.hub = hub
	s.logger = logger
}

func (s *service) Create(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error) {
	var res *TaskResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var err error
		res, err = s.repo.Create(txCtx, userID, req)
		return err
	})
	if err != nil {
		return nil, err
	}

	// Schedule Task Due Alert if due date is provided
	if res != nil && res.DueDate != nil && s.scheduler != nil && res.Status != "done" {
		thresholdMinutes, _, _ := s.repo.GetUserThresholdAndTimezone(ctx, userID)
		warningTime := res.DueDate.Add(-time.Duration(thresholdMinutes) * time.Minute)
		now := time.Now()
		if warningTime.After(now) {
			_ = s.scheduler.ScheduleTaskDueAlert(ctx, res.ID, warningTime)
		} else if res.DueDate.After(now) {
			_ = s.scheduler.ScheduleTaskDueAlert(ctx, res.ID, now.Add(5*time.Second))
		}
	}

	return res, nil
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
	if err != nil {
		return nil, err
	}

	if res != nil && s.scheduler != nil {
		if res.Status == "done" {
			_ = s.scheduler.CancelTaskDueAlert(ctx, res.ID)
		} else if req.DueDate != nil {
			if res.DueDate == nil {
				_ = s.scheduler.CancelTaskDueAlert(ctx, res.ID)
			} else {
				thresholdMinutes, _, _ := s.repo.GetUserThresholdAndTimezone(ctx, userID)
				warningTime := res.DueDate.Add(-time.Duration(thresholdMinutes) * time.Minute)
				now := time.Now()
				if warningTime.After(now) {
					_ = s.scheduler.ScheduleTaskDueAlert(ctx, res.ID, warningTime)
				} else if res.DueDate.After(now) {
					_ = s.scheduler.ScheduleTaskDueAlert(ctx, res.ID, now.Add(5*time.Second))
				} else {
					_ = s.scheduler.CancelTaskDueAlert(ctx, res.ID)
				}
			}
		}
	}

	return res, nil
}

func (s *service) Delete(ctx context.Context, userID string, id int64) error {
	if s.scheduler != nil {
		_ = s.scheduler.CancelTaskDueAlert(ctx, id)
	}
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
		var err2 error
		res, err2 = s.repo.CreateSubtask(txCtx, userID, taskID, req.Title)
		return err2
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
		var err2 error
		res, err2 = s.repo.UpdateSubtask(txCtx, userID, taskID, subtaskID, req)
		return err2
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

func (s *service) ProcessDueWarning(ctx context.Context, taskID int64) error {
	task, userEmail, userTimezone, _, emailEnabled, err := s.repo.FindByIDSystem(ctx, taskID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to find task system for due warning", "task_id", taskID, "error", err)
		}
		return err
	}
	if task == nil {
		if s.logger != nil {
			s.logger.Info("task not found for due warning, skipping", "task_id", taskID)
		}
		return nil
	}

	if task.Status == "done" {
		if s.logger != nil {
			s.logger.Info("task is already done, skipping due warning", "task_id", taskID)
		}
		return nil
	}

	if task.DueDateWarningSentAt != nil {
		if s.logger != nil {
			s.logger.Info("due warning already sent for task, skipping", "task_id", taskID)
		}
		return nil
	}

	// 1. In-App Realtime SSE
	if s.hub != nil {
		s.hub.Broadcast(&realtime.Message{
			Type:   "task:due_warning",
			UserID: task.UserID,
			Data: map[string]any{
				"task_id":  task.ID,
				"title":    task.Title,
				"due_date": task.DueDate,
				"priority": task.Priority,
			},
		})
		if s.logger != nil {
			s.logger.Info("realtime task due warning broadcast sent", "task_id", taskID, "user_id", task.UserID)
		}
	}

	// 2. Email Notification (via Resend)
	if s.emailSender != nil && emailEnabled && userEmail != "" && task.DueDate != nil {
		dueStr := task.DueDate.Format("02 Jan 2006 15:04")
		if loc, locErr := time.LoadLocation(userTimezone); locErr == nil {
			dueStr = task.DueDate.In(loc).Format("02 Jan 2006 15:04") + " (" + userTimezone + ")"
		}
		subject := fmt.Sprintf("Deadline Reminder: %s", task.Title)
		desc := fmt.Sprintf("This task is due on %s. %s", dueStr, task.Description)
		if err := s.emailSender.SendReminderEmail(ctx, userEmail, subject, desc, &task.Title); err != nil {
			if s.logger != nil {
				s.logger.Warn("resend task due email failed", "task_id", taskID, "email", userEmail, "error", err)
			}
		}
	}

	// 3. Push Notification (via FCM)
	if s.pushSender != nil && s.reminderRepo != nil {
		tokens, err := s.reminderRepo.GetUserDeviceTokens(ctx, task.UserID)
		if err == nil && len(tokens) > 0 {
			dueStr := ""
			if task.DueDate != nil {
				dueStr = task.DueDate.Format(time.RFC3339)
			}
			fcmData := map[string]string{
				"task_id":  fmt.Sprintf("%d", task.ID),
				"title":    task.Title,
				"due_date": dueStr,
				"type":     "task:due_warning",
			}
			body := fmt.Sprintf("Task '%s' is approaching its deadline.", task.Title)
			if err := s.pushSender.SendPushNotification(ctx, tokens, "Task Due Soon", body, fcmData); err != nil {
				if s.logger != nil {
					s.logger.Warn("fcm push task due warning failed", "task_id", taskID, "error", err)
				}
			}
		}
	}

	// 4. Mark status as warning sent
	if err := s.repo.MarkDueWarningSent(ctx, taskID); err != nil {
		if s.logger != nil {
			s.logger.Error("failed to mark due warning sent", "task_id", taskID, "error", err)
		}
		return err
	}

	if s.logger != nil {
		s.logger.Info("task due warning processed successfully", "task_id", taskID)
	}
	return nil
}
