package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
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

	GetActivities(ctx context.Context, userID string, taskID int64) ([]TaskActivityResponse, error)
	CreateActivity(ctx context.Context, userID string, taskID int64, req CreateTaskActivityRequest) (*TaskActivityResponse, error)

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
		if err != nil {
			return err
		}
		if res != nil {
			changes, _ := json.Marshal(map[string]any{
				"title":    res.Title,
				"status":   res.Status,
				"priority": res.Priority,
			})
			_, _ = s.repo.CreateActivity(txCtx, res.ID, userID, "task_created", "Task created", changes)
		}
		return nil
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
		if err != nil {
			return err
		}
		if res != nil {
			var changeDescs []string
			diffs := make(map[string]map[string]any)

			if entity.Title != res.Title {
				diffs["title"] = map[string]any{"old": entity.Title, "new": res.Title}
				changeDescs = append(changeDescs, fmt.Sprintf("Title changed to '%s'", res.Title))
			}
			if entity.Status != res.Status {
				diffs["status"] = map[string]any{"old": entity.Status, "new": res.Status}
				changeDescs = append(changeDescs, fmt.Sprintf("Status changed from '%s' to '%s'", entity.Status, res.Status))
			}
			if entity.Priority != res.Priority {
				diffs["priority"] = map[string]any{"old": entity.Priority, "new": res.Priority}
				changeDescs = append(changeDescs, fmt.Sprintf("Priority changed from '%s' to '%s'", entity.Priority, res.Priority))
			}
			if (entity.DueDate == nil && res.DueDate != nil) || (entity.DueDate != nil && res.DueDate == nil) || (entity.DueDate != nil && res.DueDate != nil && !entity.DueDate.Equal(*res.DueDate)) {
				diffs["due_date"] = map[string]any{"old": entity.DueDate, "new": res.DueDate}
				if res.DueDate != nil {
					changeDescs = append(changeDescs, fmt.Sprintf("Due date set to %s", res.DueDate.Format("02 Jan 2006 15:04")))
				} else {
					changeDescs = append(changeDescs, "Due date removed")
				}
			}
			if (entity.ProjectID == nil && res.ProjectID != nil) || (entity.ProjectID != nil && res.ProjectID == nil) || (entity.ProjectID != nil && res.ProjectID != nil && *entity.ProjectID != *res.ProjectID) {
				diffs["project_id"] = map[string]any{"old": entity.ProjectID, "new": res.ProjectID}
				changeDescs = append(changeDescs, "Project changed")
			}
			if entity.Description != res.Description {
				diffs["description"] = map[string]any{"old": entity.Description, "new": res.Description}
				changeDescs = append(changeDescs, "Description updated")
			}

			if len(changeDescs) > 0 {
				changesJSON, _ := json.Marshal(diffs)
				desc := strings.Join(changeDescs, ", ")
				_, _ = s.repo.CreateActivity(txCtx, res.ID, userID, "task_updated", desc, changesJSON)
			}
		}
		return nil
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
		if err != nil {
			return err
		}
		if res != nil {
			changesMap := map[string]any{
				"progress_note": req.ProgressNote,
			}
			if req.Status != nil && *req.Status != "" {
				changesMap["status"] = *req.Status
			}
			if len(req.CompletedSubtaskIDs) > 0 {
				changesMap["completed_subtask_ids"] = req.CompletedSubtaskIDs
			}
			changesJSON, _ := json.Marshal(changesMap)
			_, _ = s.repo.CreateActivity(txCtx, res.ID, userID, "progress_logged", req.ProgressNote, changesJSON)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if s.hub != nil {
		s.hub.Broadcast(&realtime.Message{
			Type:   "task:activity_created",
			UserID: userID,
			Data: map[string]any{
				"task_id": id,
			},
		})
	}
	return res, nil
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
		if err != nil {
			return err
		}
		if res != nil {
			desc := "Removed from focus"
			if res.IsFocus {
				desc = "Added to focus horizon"
			}
			changesJSON, _ := json.Marshal(map[string]any{"is_focus": res.IsFocus})
			_, _ = s.repo.CreateActivity(txCtx, res.ID, userID, "focus_toggled", desc, changesJSON)
		}
		return nil
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
		if err2 != nil {
			return err2
		}
		if res != nil {
			desc := fmt.Sprintf("Added subtask: %s", req.Title)
			changesJSON, _ := json.Marshal(map[string]any{"subtask_id": res.ID, "title": req.Title})
			_, _ = s.repo.CreateActivity(txCtx, taskID, userID, "subtask_created", desc, changesJSON)
		}
		return nil
	})
	if err == nil && s.hub != nil {
		s.hub.Broadcast(&realtime.Message{
			Type:   "task:activity_created",
			UserID: userID,
			Data: map[string]any{
				"task_id": taskID,
			},
		})
	}
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
		if err2 != nil {
			return err2
		}
		if res != nil {
			var desc string
			if req.IsCompleted != nil {
				if *req.IsCompleted {
					desc = fmt.Sprintf("Completed subtask: %s", res.Title)
				} else {
					desc = fmt.Sprintf("Marked subtask as incomplete: %s", res.Title)
				}
			} else {
				desc = fmt.Sprintf("Updated subtask: %s", res.Title)
			}
			changesJSON, _ := json.Marshal(map[string]any{
				"subtask_id":   subtaskID,
				"title":        res.Title,
				"is_completed": res.IsCompleted,
			})
			_, _ = s.repo.CreateActivity(txCtx, taskID, userID, "subtask_toggled", desc, changesJSON)
		}
		return nil
	})
	if err == nil && s.hub != nil {
		s.hub.Broadcast(&realtime.Message{
			Type:   "task:activity_created",
			UserID: userID,
			Data: map[string]any{
				"task_id": taskID,
			},
		})
	}
	return res, err
}

func (s *service) DeleteSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64) error {
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, taskID)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		subtaskTitle := fmt.Sprintf("#%d", subtaskID)
		for _, st := range entity.Subtasks {
			if st.ID == subtaskID {
				subtaskTitle = st.Title
				break
			}
		}
		if err := s.repo.DeleteSubtask(txCtx, userID, taskID, subtaskID); err != nil {
			return err
		}
		desc := fmt.Sprintf("Deleted subtask: %s", subtaskTitle)
		changesJSON, _ := json.Marshal(map[string]any{"subtask_id": subtaskID, "title": subtaskTitle})
		_, _ = s.repo.CreateActivity(txCtx, taskID, userID, "subtask_deleted", desc, changesJSON)
		return nil
	})
	if err == nil && s.hub != nil {
		s.hub.Broadcast(&realtime.Message{
			Type:   "task:activity_created",
			UserID: userID,
			Data: map[string]any{
				"task_id": taskID,
			},
		})
	}
	return err
}

func (s *service) GetActivities(ctx context.Context, userID string, taskID int64) ([]TaskActivityResponse, error) {
	var res []TaskActivityResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, taskID)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		res, err = s.repo.FindActivitiesByTaskID(txCtx, userID, taskID)
		return err
	})
	return res, err
}

func (s *service) CreateActivity(ctx context.Context, userID string, taskID int64, req CreateTaskActivityRequest) (*TaskActivityResponse, error) {
	var res *TaskActivityResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, taskID)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		actType := req.ActivityType
		if actType == "" {
			actType = "progress_note"
		}
		changesJSON, _ := json.Marshal(map[string]any{"progress_note": req.Description})
		res, err = s.repo.CreateActivity(txCtx, taskID, userID, actType, req.Description, changesJSON)
		if err != nil {
			return err
		}
		_, _ = s.repo.LogProgress(txCtx, userID, taskID, LogProgressRequest{
			ProgressNote: req.Description,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if s.hub != nil {
		s.hub.Broadcast(&realtime.Message{
			Type:   "task:activity_created",
			UserID: userID,
			Data: map[string]any{
				"task_id": taskID,
			},
		})
	}
	return res, nil
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
