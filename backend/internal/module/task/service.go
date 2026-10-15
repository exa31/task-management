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

	// Widget & Recurrence methods
	GetWidgetBundle(ctx context.Context, userID string) (*WidgetBundleResponse, error)
	GetWorkspaceStats(ctx context.Context, userID string) (*WorkspaceStatsResponse, error)
	ProcessTaskRecurrence(ctx context.Context, taskID int64) error
}

type service struct {
	repo                Repository
	db                  database.Transactor
	scheduler           notification.TaskDueScheduler
	recurrenceScheduler notification.TaskRecurrenceScheduler
	emailSender         notification.EmailSender
	pushSender          notification.PushSender
	reminderRepo        reminder.Repository
	hub                 *realtime.Hub
	logger              *slog.Logger
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
	if rs, ok := scheduler.(notification.TaskRecurrenceScheduler); ok {
		s.recurrenceScheduler = rs
	}
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

			thresholdMinutes, userTimezone, tzErr := s.repo.GetUserThresholdAndTimezone(txCtx, userID)
			if tzErr != nil && s.logger != nil {
				s.logger.WarnContext(txCtx, "failed to get user threshold and timezone", "error", tzErr, "user_id", userID)
			}

			// Schedule Task Due Alert if due date is provided
			if res.DueDate != nil && s.scheduler != nil && res.Status != "done" {
				warningTime := res.DueDate.Add(-time.Duration(thresholdMinutes) * time.Minute)
				now := time.Now()
				if warningTime.After(now) {
					if err := s.scheduler.ScheduleTaskDueAlert(txCtx, res.ID, warningTime); err != nil && s.logger != nil {
						s.logger.WarnContext(txCtx, "failed to schedule task due alert", "error", err, "task_id", res.ID)
					}
				} else if res.DueDate.After(now) {
					if err := s.scheduler.ScheduleTaskDueAlert(txCtx, res.ID, now.Add(5*time.Second)); err != nil && s.logger != nil {
						s.logger.WarnContext(txCtx, "failed to schedule task due alert", "error", err, "task_id", res.ID)
					}
				}
			}

			// Schedule Task Recurrence if recurring
			if res.IsRecurring && s.recurrenceScheduler != nil {
				nextMidnight := getNextMidnightInUserTimezone(userTimezone)
				if err := s.recurrenceScheduler.ScheduleTaskRecurrence(txCtx, res.ID, nextMidnight); err != nil && s.logger != nil {
					s.logger.WarnContext(txCtx, "failed to schedule task recurrence", "error", err, "task_id", res.ID)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
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
	var userTimezone string
	var dueChanged bool
	var dueRemoved bool
	var urgentChanged bool
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}

		if req.DueDate != nil {
			for _, st := range entity.Subtasks {
				if st.DueDate != nil && st.DueDate.After(*req.DueDate) {
					return errors.BadRequest("Task due date cannot be earlier than any existing subtask due date")
				}
			}
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
				if res.Priority == "urgent" {
					urgentChanged = true
				}
			}
			if (entity.DueDate == nil && res.DueDate != nil) || (entity.DueDate != nil && res.DueDate == nil) || (entity.DueDate != nil && res.DueDate != nil && !entity.DueDate.Equal(*res.DueDate)) {
				dueChanged = true
				if res.DueDate == nil {
					dueRemoved = true
				}
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
				if _, err := s.repo.CreateActivity(txCtx, res.ID, userID, "task_updated", desc, changesJSON); err != nil {
					return err
				}
			}

			thresholdMinutes, tz, tzErr := s.repo.GetUserThresholdAndTimezone(txCtx, userID)
			if tzErr != nil && s.logger != nil {
				s.logger.WarnContext(txCtx, "failed to get user threshold and timezone", "error", tzErr, "user_id", userID)
			}
			userTimezone = tz

			if s.scheduler != nil {
				if res.Status == "done" {
					if err := s.scheduler.CancelTaskDueAlert(txCtx, res.ID); err != nil && s.logger != nil {
						s.logger.WarnContext(txCtx, "failed to cancel task due alert", "error", err, "task_id", res.ID)
					}
				} else if req.DueDate != nil {
					if res.DueDate == nil {
						if err := s.scheduler.CancelTaskDueAlert(txCtx, res.ID); err != nil && s.logger != nil {
							s.logger.WarnContext(txCtx, "failed to cancel task due alert", "error", err, "task_id", res.ID)
						}
					} else {
						warningTime := res.DueDate.Add(-time.Duration(thresholdMinutes) * time.Minute)
						now := time.Now()
						if warningTime.After(now) {
							if err := s.scheduler.ScheduleTaskDueAlert(txCtx, res.ID, warningTime); err != nil && s.logger != nil {
								s.logger.WarnContext(txCtx, "failed to schedule task due alert", "error", err, "task_id", res.ID)
							}
						} else if res.DueDate.After(now) {
							if err := s.scheduler.ScheduleTaskDueAlert(txCtx, res.ID, now.Add(5*time.Second)); err != nil && s.logger != nil {
								s.logger.WarnContext(txCtx, "failed to schedule task due alert", "error", err, "task_id", res.ID)
							}
						} else {
							if err := s.scheduler.CancelTaskDueAlert(txCtx, res.ID); err != nil && s.logger != nil {
								s.logger.WarnContext(txCtx, "failed to cancel task due alert", "error", err, "task_id", res.ID)
							}
						}
					}
				}
			}

			if s.recurrenceScheduler != nil {
				if req.IsRecurring != nil {
					if !*req.IsRecurring {
						if err := s.recurrenceScheduler.CancelTaskRecurrence(txCtx, res.ID); err != nil && s.logger != nil {
							s.logger.WarnContext(txCtx, "failed to cancel task recurrence", "error", err, "task_id", res.ID)
						}
					} else {
						nextMidnight := getNextMidnightInUserTimezone(userTimezone)
						if err := s.recurrenceScheduler.ScheduleTaskRecurrence(txCtx, res.ID, nextMidnight); err != nil && s.logger != nil {
							s.logger.WarnContext(txCtx, "failed to schedule task recurrence", "error", err, "task_id", res.ID)
						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if res != nil {
		if dueChanged || urgentChanged {
			s.notifyTaskDeadlineChanged(res.ID, res.Title, res.DueDate, urgentChanged, dueRemoved, userID)
		}
	}

	return res, nil
}

func (s *service) Delete(ctx context.Context, userID string, id int64) error {
	var deletedEntity *TaskResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(txCtx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		deletedEntity = entity
		if err := s.repo.Delete(txCtx, userID, id); err != nil {
			return err
		}

		if s.scheduler != nil {
			if err := s.scheduler.CancelTaskDueAlert(txCtx, id); err != nil && s.logger != nil {
				s.logger.WarnContext(txCtx, "failed to cancel task due alert on delete", "error", err, "task_id", id)
			}
		}
		if s.recurrenceScheduler != nil {
			if err := s.recurrenceScheduler.CancelTaskRecurrence(txCtx, id); err != nil && s.logger != nil {
				s.logger.WarnContext(txCtx, "failed to cancel task recurrence on delete", "error", err, "task_id", id)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if deletedEntity != nil && deletedEntity.DueDate != nil {
		s.notifyTaskDeadlineChanged(id, deletedEntity.Title, nil, false, true, userID)
	}

	return nil
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
			if _, err := s.repo.CreateActivity(txCtx, res.ID, userID, "progress_logged", req.ProgressNote, changesJSON); err != nil {
				return err
			}
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
			if _, err := s.repo.CreateActivity(txCtx, res.ID, userID, "focus_toggled", desc, changesJSON); err != nil {
				return err
			}
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
		if req.DueDate != nil && entity.DueDate != nil && req.DueDate.After(*entity.DueDate) {
			return errors.BadRequest("Subtask due date cannot exceed task due date")
		}
		var err2 error
		res, err2 = s.repo.CreateSubtask(txCtx, userID, taskID, req)
		if err2 != nil {
			return err2
		}
		if res != nil {
			desc := fmt.Sprintf("Added subtask: %s", req.Title)
			changesJSON, _ := json.Marshal(map[string]any{"subtask_id": res.ID, "title": req.Title})
			if _, err := s.repo.CreateActivity(txCtx, taskID, userID, "subtask_created", desc, changesJSON); err != nil {
				return err
			}
			if entity.Status == "done" {
				progressStatus := "in_progress"
				if _, err := s.Update(txCtx, userID, taskID, UpdateTaskRequest{
					Status: &progressStatus,
				}); err != nil {
					return err
				}
			}
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
		if req.DueDate != nil && entity.DueDate != nil && req.DueDate.After(*entity.DueDate) {
			return errors.BadRequest("Subtask due date cannot exceed task due date")
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
			if _, err := s.repo.CreateActivity(txCtx, taskID, userID, "subtask_toggled", desc, changesJSON); err != nil {
				return err
			}

			if req.IsCompleted != nil {
				total, completed, countErr := s.repo.CountSubtasks(txCtx, taskID)
				if countErr == nil && total > 0 {
					if completed == total && entity.Status != "done" {
						doneStatus := "done"
						noFocus := false
						if _, err := s.Update(txCtx, userID, taskID, UpdateTaskRequest{
							Status:  &doneStatus,
							IsFocus: &noFocus,
						}); err != nil {
							return err
						}
					} else if completed < total && entity.Status == "done" {
						progressStatus := "in_progress"
						if _, err := s.Update(txCtx, userID, taskID, UpdateTaskRequest{
							Status: &progressStatus,
						}); err != nil {
							return err
						}
					}
				}
			}
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
		if _, err := s.repo.CreateActivity(txCtx, taskID, userID, "subtask_deleted", desc, changesJSON); err != nil {
			return err
		}

		total, completed, countErr := s.repo.CountSubtasks(txCtx, taskID)
		if countErr == nil && total > 0 && completed == total && entity.Status != "done" {
			doneStatus := "done"
			noFocus := false
			if _, err := s.Update(txCtx, userID, taskID, UpdateTaskRequest{
				Status:  &doneStatus,
				IsFocus: &noFocus,
			}); err != nil {
				return err
			}
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
		_, err = s.repo.LogProgress(txCtx, userID, taskID, LogProgressRequest{
			ProgressNote: req.Description,
		})
		return err
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
		ProjectID:    req.ProjectID,
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

	// 4. Mark status as warning sent in transaction
	err = s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		return s.repo.MarkDueWarningSent(txCtx, taskID)
	})
	if err != nil {
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

func (s *service) GetWidgetBundle(ctx context.Context, userID string) (*WidgetBundleResponse, error) {
	var bundle *WidgetBundleResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var err error
		bundle, err = s.repo.GetWidgetBundle(txCtx, userID)
		return err
	})
	return bundle, err
}

func (s *service) GetWorkspaceStats(ctx context.Context, userID string) (*WorkspaceStatsResponse, error) {
	var stats *WorkspaceStatsResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var err error
		stats, err = s.repo.GetWorkspaceStats(txCtx, userID)
		return err
	})
	return stats, err
}

func (s *service) ProcessTaskRecurrence(ctx context.Context, taskID int64) error {
	task, _, userTimezone, _, _, err := s.repo.FindByIDSystem(ctx, taskID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to find task system for recurrence", "task_id", taskID, "error", err)
		}
		return err
	}
	if task == nil || !task.IsRecurring {
		if s.logger != nil {
			s.logger.Info("task is not recurring or not found, skipping recurrence", "task_id", taskID)
		}
		return nil
	}

	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		loc = time.Local
	}
	nowInLoc := time.Now().In(loc)

	// Avoid duplicate generation within the same local calendar day
	if task.RecurrenceLastGeneratedAt != nil {
		lastInLoc := task.RecurrenceLastGeneratedAt.In(loc)
		if lastInLoc.Year() == nowInLoc.Year() && lastInLoc.YearDay() == nowInLoc.YearDay() {
			if s.logger != nil {
				s.logger.Info("recurrence already generated today, ensuring next schedule is set", "task_id", taskID)
			}
			if s.recurrenceScheduler != nil {
				nextMidnight := getNextMidnightInUserTimezone(userTimezone)
				_ = s.recurrenceScheduler.ScheduleTaskRecurrence(ctx, taskID, nextMidnight)
			}
			return nil
		}
	}

	// Roll over task for next cycle (creates subtasks, updates timestamp, and schedules next cycle in a single transaction)
	if _, err := s.cloneAndRollOverRecurringTask(ctx, task, userTimezone); err != nil {
		if s.logger != nil {
			s.logger.Error("failed to rollover recurring task in ProcessTaskRecurrence", "task_id", taskID, "error", err)
		}
		return err
	}

	if s.logger != nil {
		s.logger.Info("task recurrence processed successfully", "task_id", taskID)
	}
	return nil
}

func (s *service) cloneAndRollOverRecurringTask(ctx context.Context, task *TaskResponse, userTimezone string) (*TaskResponse, error) {
	if task.RecurrenceEndDate != nil && time.Now().After(*task.RecurrenceEndDate) {
		return nil, nil
	}

	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		loc = time.Local
	}

	var newDueDate *time.Time
	if task.DueDate != nil {
		next := computeNextDueDate(*task.DueDate, task.RecurrencePattern, task.RecurrenceInterval, loc)
		if task.RecurrenceEndDate != nil && next.After(*task.RecurrenceEndDate) {
			return nil, nil
		}
		newDueDate = &next
	}

	var newSubtasks []CreateSubtaskRequest
	for _, st := range task.Subtasks {
		var newSubDue *time.Time
		if st.DueDate != nil && task.DueDate != nil && newDueDate != nil {
			offset := task.DueDate.Sub(*st.DueDate)
			calc := newDueDate.Add(-offset)
			newSubDue = &calc
		} else if st.DueDate != nil {
			calc := computeNextDueDate(*st.DueDate, task.RecurrencePattern, task.RecurrenceInterval, loc)
			newSubDue = &calc
		}
		newSubtasks = append(newSubtasks, CreateSubtaskRequest{
			Title:   st.Title,
			DueDate: newSubDue,
		})
	}

	newTaskReq := CreateTaskRequest{
		Title:              task.Title,
		Description:        task.Description,
		Status:             "todo",
		Priority:           task.Priority,
		DueDate:            newDueDate,
		ProjectID:          task.ProjectID,
		Tags:               task.Tags,
		IsRecurring:        true,
		RecurrencePattern:  task.RecurrencePattern,
		RecurrenceInterval: task.RecurrenceInterval,
		RecurrenceEndDate:  task.RecurrenceEndDate,
	}

	var created *TaskResponse
	err = s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var createErr error
		created, createErr = s.repo.Create(txCtx, task.UserID, newTaskReq)
		if createErr != nil {
			return createErr
		}

		if created != nil {
			changes, _ := json.Marshal(map[string]any{
				"title":    created.Title,
				"status":   created.Status,
				"priority": created.Priority,
			})
			if _, err := s.repo.CreateActivity(txCtx, created.ID, task.UserID, "task_created", "Recurring task created", changes); err != nil {
				return err
			}

			for _, subReq := range newSubtasks {
				if _, err := s.repo.CreateSubtask(txCtx, task.UserID, created.ID, subReq); err != nil {
					return err
				}
			}

			if created.DueDate != nil && s.scheduler != nil && created.Status != "done" {
				thresholdMinutes, _, tzErr := s.repo.GetUserThresholdAndTimezone(txCtx, task.UserID)
				if tzErr != nil && s.logger != nil {
					s.logger.WarnContext(txCtx, "failed to get user threshold and timezone", "error", tzErr, "user_id", task.UserID)
				}
				warningTime := created.DueDate.Add(-time.Duration(thresholdMinutes) * time.Minute)
				now := time.Now()
				if warningTime.After(now) {
					if err := s.scheduler.ScheduleTaskDueAlert(txCtx, created.ID, warningTime); err != nil && s.logger != nil {
						s.logger.WarnContext(txCtx, "failed to schedule task due alert for new recurring task", "error", err, "task_id", created.ID)
					}
				} else if created.DueDate.After(now) {
					if err := s.scheduler.ScheduleTaskDueAlert(txCtx, created.ID, now.Add(5*time.Second)); err != nil && s.logger != nil {
						s.logger.WarnContext(txCtx, "failed to schedule task due alert for new recurring task", "error", err, "task_id", created.ID)
					}
				}
			}
		}

		now := time.Now()
		if err := s.repo.UpdateRecurrenceGeneratedAt(txCtx, task.ID, now); err != nil {
			return err
		}

		if s.recurrenceScheduler != nil {
			nextMidnight := getNextMidnightInUserTimezone(userTimezone)
			if err := s.recurrenceScheduler.ScheduleTaskRecurrence(txCtx, task.ID, nextMidnight); err != nil {
				return fmt.Errorf("failed to schedule next recurrence: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to rollover recurring task in transaction", "task_id", task.ID, "error", err)
		}
		return nil, err
	}

	if s.hub != nil && created != nil {
		s.hub.Broadcast(&realtime.Message{
			Type:   "task:recurrence_created",
			UserID: task.UserID,
			Data: map[string]any{
				"old_task_id": task.ID,
				"new_task_id": created.ID,
				"title":       created.Title,
			},
		})
	}

	return created, nil
}

func computeNextDueDate(currentDueDate time.Time, pattern string, interval int, loc *time.Location) time.Time {
	if interval <= 0 {
		interval = 1
	}
	t := currentDueDate.In(loc)
	switch strings.ToLower(pattern) {
	case "daily":
		return t.AddDate(0, 0, interval)
	case "weekdays":
		next := t.AddDate(0, 0, 1)
		for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
			next = next.AddDate(0, 0, 1)
		}
		return next
	case "weekly":
		return t.AddDate(0, 0, interval*7)
	case "monthly":
		return t.AddDate(0, interval, 0)
	default:
		return t.AddDate(0, 0, interval)
	}
}

func getNextMidnightInUserTimezone(userTimezone string) time.Time {
	if userTimezone == "" {
		userTimezone = "Asia/Jakarta"
	}
	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		loc = time.Local
	}
	nowInLoc := time.Now().In(loc)
	nextMidnight := time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day()+1, 0, 0, 0, 0, loc).UTC()
	return nextMidnight
}

func (s *service) notifyTaskDeadlineChanged(taskID int64, title string, dueDate *time.Time, isUrgent bool, isRemoved bool, userID string) {
	if s.pushSender == nil || s.reminderRepo == nil {
		return
	}
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		tokens, err := s.reminderRepo.GetUserDeviceTokens(bgCtx, userID)
		if err != nil || len(tokens) == 0 {
			return
		}

		var body string
		if isRemoved {
			body = fmt.Sprintf("Batas waktu '%s' telah dihapus", title)
		} else if dueDate != nil {
			body = fmt.Sprintf("Batas waktu '%s' diatur ke %s", title, dueDate.Format("02 Jan 15:04"))
		} else if isUrgent {
			body = fmt.Sprintf("Tugas '%s' ditandai sebagai mendesak", title)
		} else {
			body = fmt.Sprintf("Tugas '%s' telah diperbarui", title)
		}

		pushData := map[string]string{
			"type":    "task_deadline_updated",
			"task_id": fmt.Sprintf("%d", taskID),
			"title":   title,
		}
		if err := s.pushSender.SendPushNotification(bgCtx, tokens, "Pembaruan Batas Waktu", body, pushData); err != nil && s.logger != nil {
			s.logger.WarnContext(bgCtx, "failed to dispatch task deadline push notification", "task_id", taskID, "error", err)
		}
	}()
}
