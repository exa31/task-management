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

	// Recurring Tasks methods
	CreateRecurringTask(ctx context.Context, userID string, req CreateRecurringTaskRequest) (*RecurringTaskResponse, error)
	FindAllRecurringTasks(ctx context.Context, userID string) ([]RecurringTaskResponse, error)
	GetRecurringTaskByID(ctx context.Context, userID string, id int64) (*RecurringTaskResponse, error)
	UpdateRecurringTask(ctx context.Context, userID string, id int64, req UpdateRecurringTaskRequest) (*RecurringTaskResponse, error)
	DeleteRecurringTask(ctx context.Context, userID string, id int64) error
	ToggleActiveRecurringTask(ctx context.Context, userID string, id int64) (*RecurringTaskResponse, error)
}

type service struct {
	repo                Repository
	db                  database.Transactor
	scheduler           notification.TaskDueScheduler
	recurrenceScheduler notification.TaskRecurrenceScheduler
	reminderScheduler   notification.ReminderScheduler
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
	if rems, ok := scheduler.(notification.ReminderScheduler); ok {
		s.reminderScheduler = rems
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
		ctx := database.ContextWithTx(ctx, tx)

		thresholdMinutes, userTimezone, tzErr := s.repo.GetUserThresholdAndTimezone(ctx, tx, userID)
		if tzErr != nil && s.logger != nil {
			s.logger.WarnContext(ctx, "failed to get user threshold and timezone", "error", tzErr, "user_id", userID)
		}

		if req.IsRecurring && req.RecurringTaskID == nil {
			subItems := make([]CreateRecurringSubtaskItem, len(req.Subtasks))
			for i, s := range req.Subtasks {
				subItems[i] = CreateRecurringSubtaskItem{
					Title:      s,
					OrderIndex: i,
				}
			}
			var dueTime *string
			if req.DueDate != nil {
				loc, err := time.LoadLocation(userTimezone)
				if err != nil {
					loc = time.Local
				}
				formatted := req.DueDate.In(loc).Format("15:04:05")
				dueTime = &formatted
			}
			recReq := CreateRecurringTaskRequest{
				Title:              req.Title,
				Description:        req.Description,
				Priority:           req.Priority,
				ProjectID:          req.ProjectID,
				RecurrencePattern:  req.RecurrencePattern,
				RecurrenceInterval: req.RecurrenceInterval,
				RecurrenceEndDate:  req.RecurrenceEndDate,
				DueTime:            dueTime,
				Subtasks:           subItems,
			}
			recMaster, recErr := s.repo.CreateRecurring(ctx, tx, userID, recReq)
			if recErr != nil {
				return recErr
			}
			req.RecurringTaskID = &recMaster.ID

			if s.recurrenceScheduler != nil {
				nextMidnight := getNextMidnightInUserTimezone(userTimezone)
				if err := s.recurrenceScheduler.ScheduleTaskRecurrence(ctx, recMaster.ID, nextMidnight); err != nil && s.logger != nil {
					s.logger.WarnContext(ctx, "failed to schedule task recurrence for master", "error", err, "recurring_id", recMaster.ID)
				}
			}
		}

		var err error
		res, err = s.repo.Create(ctx, tx, userID, req)
		if err != nil {
			return err
		}
		if res != nil {
			changes, _ := json.Marshal(map[string]any{
				"title":    res.Title,
				"status":   res.Status,
				"priority": res.Priority,
			})
			if _, err := s.repo.CreateActivity(ctx, tx, res.ID, userID, "task_created", "Task created", changes); err != nil {
				return err
			}

			// Schedule Task Due Alert if due date is provided
			if res.DueDate != nil && s.scheduler != nil && res.Status != "done" {
				warningTime := res.DueDate.Add(-time.Duration(thresholdMinutes) * time.Minute)
				now := time.Now()
				if warningTime.After(now) {
					if err := s.scheduler.ScheduleTaskDueAlert(ctx, res.ID, warningTime); err != nil && s.logger != nil {
						s.logger.WarnContext(ctx, "failed to schedule task due alert", "error", err, "task_id", res.ID)
					}
				} else if res.DueDate.After(now) {
					if err := s.scheduler.ScheduleTaskDueAlert(ctx, res.ID, now.Add(5*time.Second)); err != nil && s.logger != nil {
						s.logger.WarnContext(ctx, "failed to schedule task due alert", "error", err, "task_id", res.ID)
					}
				}
			}

			// Fallback scheduling if recurring without master
			if res.IsRecurring && res.RecurringTaskID == nil && s.recurrenceScheduler != nil {
				nextMidnight := getNextMidnightInUserTimezone(userTimezone)
				if err := s.recurrenceScheduler.ScheduleTaskRecurrence(ctx, res.ID, nextMidnight); err != nil && s.logger != nil {
					s.logger.WarnContext(ctx, "failed to schedule task recurrence", "error", err, "task_id", res.ID)
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
		ctx := database.ContextWithTx(ctx, tx)
		var err error
		entity, err = s.repo.FindByID(ctx, tx, userID, id)
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
		ctx := database.ContextWithTx(ctx, tx)
		var err error
		res, err = s.repo.FindAll(ctx, tx, userID, filter)
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
		ctx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(ctx, tx, userID, id)
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

		res, err = s.repo.Update(ctx, tx, userID, id, req)
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
				if _, err := s.repo.CreateActivity(ctx, tx, res.ID, userID, "task_updated", desc, changesJSON); err != nil {
					return err
				}
			}

			thresholdMinutes, tz, tzErr := s.repo.GetUserThresholdAndTimezone(ctx, tx, userID)
			if tzErr != nil && s.logger != nil {
				s.logger.WarnContext(ctx, "failed to get user threshold and timezone", "error", tzErr, "user_id", userID)
			}
			userTimezone = tz

			if s.scheduler != nil {
				if res.Status == "done" || res.DueDate == nil {
					if err := s.scheduler.CancelTaskDueAlert(ctx, res.ID); err != nil && s.logger != nil {
						s.logger.WarnContext(ctx, "failed to cancel task due alert", "error", err, "task_id", res.ID)
					}
				} else if req.HasDueDate || req.DueDate != nil {
					warningTime := res.DueDate.Add(-time.Duration(thresholdMinutes) * time.Minute)
					now := time.Now()
					if warningTime.After(now) {
						if err := s.scheduler.ScheduleTaskDueAlert(ctx, res.ID, warningTime); err != nil && s.logger != nil {
							s.logger.WarnContext(ctx, "failed to schedule task due alert", "error", err, "task_id", res.ID)
						}
					} else if res.DueDate.After(now) {
						if err := s.scheduler.ScheduleTaskDueAlert(ctx, res.ID, now.Add(5*time.Second)); err != nil && s.logger != nil {
							s.logger.WarnContext(ctx, "failed to schedule task due alert", "error", err, "task_id", res.ID)
						}
					} else {
						if err := s.scheduler.CancelTaskDueAlert(ctx, res.ID); err != nil && s.logger != nil {
							s.logger.WarnContext(ctx, "failed to cancel task due alert", "error", err, "task_id", res.ID)
						}
					}
				}
			}

			// Recurrence handling on Update
			if req.IsRecurring != nil {
				if !*req.IsRecurring {
					// User turned off recurrence
					if entity.RecurringTaskID != nil {
						if _, err := tx.ExecContext(ctx, "UPDATE recurring_tasks SET is_active = FALSE, updated_at = NOW() WHERE id = $1 AND user_id = $2", *entity.RecurringTaskID, userID); err != nil {
							return fmt.Errorf("failed to deactivate recurring master: %w", err)
						}
						if s.recurrenceScheduler != nil {
							_ = s.recurrenceScheduler.CancelTaskRecurrence(ctx, *entity.RecurringTaskID)
						}
						if _, err := tx.ExecContext(ctx, "UPDATE tasks SET recurring_task_id = NULL WHERE id = $1 AND user_id = $2", res.ID, userID); err != nil {
							return fmt.Errorf("failed to unlink task from recurring master: %w", err)
						}
						res.RecurringTaskID = nil
					}
				} else if entity.RecurringTaskID == nil {
					// User turned on recurrence for previously non-recurring task
					pattern := "daily"
					if req.RecurrencePattern != nil && *req.RecurrencePattern != "" {
						pattern = *req.RecurrencePattern
					}
					interval := 1
					if req.RecurrenceInterval != nil && *req.RecurrenceInterval > 0 {
						interval = *req.RecurrenceInterval
					}
					recMaster, convertErr := s.repo.ConvertTaskToRecurring(ctx, tx, userID, res.ID, pattern, interval, req.RecurrenceEndDate)
					if convertErr != nil {
						return fmt.Errorf("failed to convert task to recurring master: %w", convertErr)
					}
					if recMaster != nil {
						res.RecurringTaskID = &recMaster.ID
						res.IsRecurring = true
						if s.recurrenceScheduler != nil {
							nextMidnight := getNextMidnightInUserTimezone(userTimezone)
							_ = s.recurrenceScheduler.ScheduleTaskRecurrence(ctx, recMaster.ID, nextMidnight)
						}
					}
				}
			} else if res.RecurringTaskID != nil && res.IsRecurring {
				// Existing recurring task updated: sync fields to recurring_tasks template
				if err := s.repo.SyncRecurringFromTask(ctx, tx, userID, res); err != nil {
					return fmt.Errorf("failed to sync recurring task master: %w", err)
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
		ctx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(ctx, tx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		deletedEntity = entity

		// 1. Cancel Asynq jobs for any reminders associated with this task before cascading delete
		if s.reminderRepo != nil && s.reminderScheduler != nil {
			rems, _, remErr := s.reminderRepo.FindAll(ctx, userID, reminder.ReminderFilter{
				TaskID: &id,
				Status: "all",
				Limit:  100,
			})
			if remErr == nil {
				for _, rem := range rems {
					if cErr := s.reminderScheduler.CancelReminder(ctx, rem.ID); cErr != nil && s.logger != nil {
						s.logger.WarnContext(ctx, "failed to cancel reminder asynq job on task delete", "error", cErr, "reminder_id", rem.ID)
					}
				}
			} else if s.logger != nil {
				s.logger.WarnContext(ctx, "failed to query reminders for task on delete", "error", remErr, "task_id", id)
			}
		}

		if err := s.repo.Delete(ctx, tx, userID, id); err != nil {
			return err
		}

		// 2. Cancel Task Due Alert in Asynq
		if s.scheduler != nil {
			if err := s.scheduler.CancelTaskDueAlert(ctx, id); err != nil && s.logger != nil {
				s.logger.WarnContext(ctx, "failed to cancel task due alert on delete", "error", err, "task_id", id)
			}
		}
		// 3. Cancel Task Recurrence in Asynq
		if s.recurrenceScheduler != nil {
			if err := s.recurrenceScheduler.CancelTaskRecurrence(ctx, id); err != nil && s.logger != nil {
				s.logger.WarnContext(ctx, "failed to cancel task recurrence on delete", "error", err, "task_id", id)
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
		ctx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(ctx, tx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		res, err = s.repo.LogProgress(ctx, tx, userID, id, req)
		if err != nil {
			return err
		}
		if res != nil {
			if req.Status != nil && *req.Status == "done" && s.scheduler != nil {
				if err := s.scheduler.CancelTaskDueAlert(ctx, id); err != nil && s.logger != nil {
					s.logger.WarnContext(ctx, "failed to cancel task due alert on progress done", "error", err, "task_id", id)
				}
			}

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
			if _, err := s.repo.CreateActivity(ctx, tx, res.ID, userID, "progress_logged", req.ProgressNote, changesJSON); err != nil {
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
		ctx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(ctx, tx, userID, id)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		res, err = s.repo.ToggleFocus(ctx, tx, userID, id)
		if err != nil {
			return err
		}
		if res != nil {
			desc := "Removed from focus"
			if res.IsFocus {
				desc = "Added to focus horizon"
			}
			changesJSON, _ := json.Marshal(map[string]any{"is_focus": res.IsFocus})
			if _, err := s.repo.CreateActivity(ctx, tx, res.ID, userID, "focus_toggled", desc, changesJSON); err != nil {
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
		ctx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(ctx, tx, userID, taskID)
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
		res, err2 = s.repo.CreateSubtask(ctx, tx, userID, taskID, req)
		if err2 != nil {
			return err2
		}
		if res != nil {
			// Sync to recurring_subtasks template if this task belongs to a recurring series
			if entity.RecurringTaskID != nil {
				if err := s.repo.CreateRecurringSubtask(ctx, tx, *entity.RecurringTaskID, req.Title, req.Description, res.OrderIndex); err != nil {
					return fmt.Errorf("failed to sync recurring subtask: %w", err)
				}
			}
			desc := fmt.Sprintf("Added subtask: %s", req.Title)
			changesJSON, _ := json.Marshal(map[string]any{"subtask_id": res.ID, "title": req.Title})
			if _, err := s.repo.CreateActivity(ctx, tx, taskID, userID, "subtask_created", desc, changesJSON); err != nil {
				return err
			}
			if entity.Status == "done" {
				progressStatus := "in_progress"
				if err := s.updateTaskStatusTx(ctx, tx, userID, taskID, progressStatus, nil); err != nil {
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

func (s *service) updateTaskStatusTx(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, newStatus string, isFocus *bool) error {
	req := UpdateTaskRequest{
		Status:  &newStatus,
		IsFocus: isFocus,
	}
	res, err := s.repo.Update(ctx, tx, userID, taskID, req)
	if err != nil {
		return err
	}
	if res != nil {
		desc := fmt.Sprintf("Status changed to '%s' (auto-updated by subtasks)", newStatus)
		changesJSON, _ := json.Marshal(map[string]any{"status": newStatus})
		_, _ = s.repo.CreateActivity(ctx, tx, taskID, userID, "task_updated", desc, changesJSON)
		if s.scheduler != nil && newStatus == "done" {
			_ = s.scheduler.CancelTaskDueAlert(ctx, taskID)
		}
	}
	return nil
}

func (s *service) UpdateSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error) {
	var res *SubtaskResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(ctx, tx, userID, taskID)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		if req.DueDate != nil && entity.DueDate != nil && req.DueDate.After(*entity.DueDate) {
			return errors.BadRequest("Subtask due date cannot exceed task due date")
		}
		var oldTitle string
		for _, st := range entity.Subtasks {
			if st.ID == subtaskID {
				oldTitle = st.Title
				break
			}
		}
		var err2 error
		res, err2 = s.repo.UpdateSubtask(ctx, tx, userID, taskID, subtaskID, req)
		if err2 != nil {
			return err2
		}
		if res != nil {
			// Sync to recurring_subtasks template if this task belongs to a recurring series
			if entity.RecurringTaskID != nil && (req.Title != nil || req.Description != nil || req.HasDescription) {
				if err := s.repo.UpdateRecurringSubtaskByTitle(ctx, tx, *entity.RecurringTaskID, oldTitle, res.Title, res.Description); err != nil {
					return fmt.Errorf("failed to sync recurring subtask update: %w", err)
				}
			}
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
			if _, err := s.repo.CreateActivity(ctx, tx, taskID, userID, "subtask_toggled", desc, changesJSON); err != nil {
				return err
			}

			if req.IsCompleted != nil {
				total, completed, countErr := s.repo.CountSubtasks(ctx, tx, taskID)
				if countErr == nil && total > 0 {
					if completed == total && entity.Status != "done" {
						doneStatus := "done"
						noFocus := false
						if err := s.updateTaskStatusTx(ctx, tx, userID, taskID, doneStatus, &noFocus); err != nil {
							return err
						}
					} else if completed < total && entity.Status == "done" {
						progressStatus := "in_progress"
						if err := s.updateTaskStatusTx(ctx, tx, userID, taskID, progressStatus, nil); err != nil {
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
		ctx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(ctx, tx, userID, taskID)
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
		if err := s.repo.DeleteSubtask(ctx, tx, userID, taskID, subtaskID); err != nil {
			return err
		}
		// Sync delete to recurring_subtasks template if this task belongs to a recurring series
		if entity.RecurringTaskID != nil && subtaskTitle != "" {
			if err := s.repo.DeleteRecurringSubtaskByTitle(ctx, tx, *entity.RecurringTaskID, subtaskTitle); err != nil {
				return fmt.Errorf("failed to sync recurring subtask deletion: %w", err)
			}
		}
		desc := fmt.Sprintf("Deleted subtask: %s", subtaskTitle)
		changesJSON, _ := json.Marshal(map[string]any{"subtask_id": subtaskID, "title": subtaskTitle})
		if _, err := s.repo.CreateActivity(ctx, tx, taskID, userID, "subtask_deleted", desc, changesJSON); err != nil {
			return err
		}

		total, completed, countErr := s.repo.CountSubtasks(ctx, tx, taskID)
		if countErr == nil && total > 0 && completed == total && entity.Status != "done" {
			doneStatus := "done"
			noFocus := false
			if err := s.updateTaskStatusTx(ctx, tx, userID, taskID, doneStatus, &noFocus); err != nil {
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
		ctx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(ctx, tx, userID, taskID)
		if err != nil {
			return err
		}
		if entity == nil {
			return errors.TaskNotFound()
		}
		res, err = s.repo.FindActivitiesByTaskID(ctx, tx, userID, taskID)
		return err
	})
	return res, err
}

func (s *service) CreateActivity(ctx context.Context, userID string, taskID int64, req CreateTaskActivityRequest) (*TaskActivityResponse, error) {
	var res *TaskActivityResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		entity, err := s.repo.FindByID(ctx, tx, userID, taskID)
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
		res, err = s.repo.CreateActivity(ctx, tx, taskID, userID, actType, req.Description, changesJSON)
		if err != nil {
			return err
		}
		_, err = s.repo.LogProgress(ctx, tx, userID, taskID, LogProgressRequest{
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
	var (
		task         *TaskResponse
		userEmail    string
		userTimezone string
		emailEnabled bool
	)
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		var err error
		task, userEmail, userTimezone, _, emailEnabled, err = s.repo.FindByIDSystem(ctx, tx, taskID)
		return err
	})
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
			if res, err := s.pushSender.SendPushNotificationWithResult(ctx, tokens, "Task Due Soon", body, fcmData); err != nil {
				if s.logger != nil {
					s.logger.Warn("fcm push task due warning failed", "task_id", taskID, "error", err)
				}
			} else if res != nil && len(res.InvalidTokens) > 0 {
				_ = s.reminderRepo.DeleteDeviceTokens(ctx, res.InvalidTokens)
			}
		}
	}

		// 4. Mark status as warning sent in transaction
	err = s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		return s.repo.MarkDueWarningSent(ctx, tx, taskID)
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
		ctx := database.ContextWithTx(ctx, tx)
		var err error
		bundle, err = s.repo.GetWidgetBundle(ctx, tx, userID)
		return err
	})
	return bundle, err
}

func (s *service) GetWorkspaceStats(ctx context.Context, userID string) (*WorkspaceStatsResponse, error) {
	var stats *WorkspaceStatsResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		var err error
		stats, err = s.repo.GetWorkspaceStats(ctx, tx, userID)
		return err
	})
	return stats, err
}

func (s *service) ProcessTaskRecurrence(ctx context.Context, targetID int64) error {
	var (
		master       *RecurringTaskResponse
		userTimezone string
	)
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		var err error
		master, userTimezone, err = s.repo.FindRecurringByIDSystem(ctx, tx, targetID)
		if err != nil {
			return err
		}
		// Fallback check: could targetID be a legacy tasks.id?
		if master == nil {
			legacyTask, _, _, _, _, legacyErr := s.repo.FindByIDSystem(ctx, tx, targetID)
			if legacyErr == nil && legacyTask != nil && legacyTask.RecurringTaskID != nil {
				master, userTimezone, err = s.repo.FindRecurringByIDSystem(ctx, tx, *legacyTask.RecurringTaskID)
				return err
			}
		}
		return nil
	})
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to find recurring task for recurrence", "recurring_id", targetID, "error", err)
		}
		return err
	}

	if master == nil {
		if s.logger != nil {
			s.logger.Info("recurring task not found or retired, skipping", "id", targetID)
		}
		return nil
	}

	if !master.IsActive {
		if s.logger != nil {
			s.logger.Info("recurring task is not active, skipping", "recurring_id", targetID)
		}
		return nil
	}

	if master.RecurrenceEndDate != nil && time.Now().After(*master.RecurrenceEndDate) {
		if s.logger != nil {
			s.logger.Info("recurring task has reached end date, deactivating", "recurring_id", targetID)
		}
		_ = s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
			ctx := database.ContextWithTx(ctx, tx)
			_, err := s.repo.ToggleActiveRecurring(ctx, tx, master.UserID, master.ID)
			return err
		})
		if s.recurrenceScheduler != nil {
			_ = s.recurrenceScheduler.CancelTaskRecurrence(ctx, master.ID)
		}
		return nil
	}

	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		loc = time.Local
	}
	nowInLoc := time.Now().In(loc)
	todayDueDate := time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), 23, 59, 59, 0, loc)
	if master.DueTime != nil && *master.DueTime != "" {
		if t, parseErr := time.Parse("15:04:05", *master.DueTime); parseErr == nil {
			todayDueDate = time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
		} else if t, parseErr2 := time.Parse("15:04", *master.DueTime); parseErr2 == nil {
			todayDueDate = time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), t.Hour(), t.Minute(), 0, 0, loc)
		}
	}

	// 1. Idempotency Check: if today's instance already exists, skip creation
	var existing *TaskResponse
	err = s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		var err error
		existing, err = s.repo.FindTodayInstance(ctx, tx, master.ID, todayDueDate)
		return err
	})
	if err != nil {
		return err
	}
	if existing != nil {
		if s.logger != nil {
			s.logger.Info("recurring task instance already exists today, ensuring next midnight schedule is set", "recurring_id", master.ID)
		}
		if s.recurrenceScheduler != nil {
			nextMidnight := getNextMidnightInUserTimezone(userTimezone)
			_ = s.recurrenceScheduler.ScheduleTaskRecurrence(ctx, master.ID, nextMidnight)
		}
		return nil
	}

	// 2. Create today's instance in transaction
	var created *TaskResponse
	err = s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		var createErr error
		created, createErr = s.repo.CreateInstanceFromRecurring(ctx, tx, master, todayDueDate)
		if createErr != nil {
			return createErr
		}

		if created != nil {
			changes, _ := json.Marshal(map[string]any{
				"title":             created.Title,
				"status":            created.Status,
				"priority":          created.Priority,
				"recurring_task_id": master.ID,
			})
			if _, err := s.repo.CreateActivity(ctx, tx, created.ID, master.UserID, "task_created", "Recurring task generated", changes); err != nil {
				return err
			}
		}

		if err := s.repo.UpdateRecurringLastGenerated(ctx, tx, master.ID, time.Now()); err != nil {
			return err
		}

		if s.recurrenceScheduler != nil {
			nextMidnight := getNextMidnightInUserTimezone(userTimezone)
			if err := s.recurrenceScheduler.ScheduleTaskRecurrence(ctx, master.ID, nextMidnight); err != nil {
				return fmt.Errorf("failed to schedule next recurrence: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to create recurring task instance in transaction", "recurring_id", master.ID, "error", err)
		}
		return err
	}

	if s.hub != nil && created != nil {
		s.hub.Broadcast(&realtime.Message{
			Type:   "task:recurrence_created",
			UserID: master.UserID,
			Data: map[string]any{
				"recurring_task_id": master.ID,
				"new_task_id":       created.ID,
				"title":             created.Title,
			},
		})
	}

	if s.logger != nil {
		s.logger.Info("recurring task processed successfully", "recurring_id", master.ID, "task_id", created.ID)
	}
	return nil
}

func (s *service) CreateRecurringTask(ctx context.Context, userID string, req CreateRecurringTaskRequest) (*RecurringTaskResponse, error) {
	var res *RecurringTaskResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		var err error
		res, err = s.repo.CreateRecurring(ctx, tx, userID, req)
		if err != nil {
			return err
		}

		_, userTimezone, tzErr := s.repo.GetUserThresholdAndTimezone(ctx, tx, userID)
		if tzErr != nil && s.logger != nil {
			s.logger.WarnContext(ctx, "failed to get user timezone", "error", tzErr, "user_id", userID)
		}

		// Generate today's initial instance if not existing
		loc, err := time.LoadLocation(userTimezone)
		if err != nil {
			loc = time.Local
		}
		nowInLoc := time.Now().In(loc)
		todayDueDate := time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), 23, 59, 59, 0, loc)
		if res.DueTime != nil && *res.DueTime != "" {
			if t, parseErr := time.Parse("15:04:05", *res.DueTime); parseErr == nil {
				todayDueDate = time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
			} else if t, parseErr2 := time.Parse("15:04", *res.DueTime); parseErr2 == nil {
				todayDueDate = time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), t.Hour(), t.Minute(), 0, 0, loc)
			}
		}

		createdInstance, err := s.repo.CreateInstanceFromRecurring(ctx, tx, res, todayDueDate)
		if err != nil {
			return err
		}
		if createdInstance != nil {
			if err := s.repo.UpdateRecurringLastGenerated(ctx, tx, res.ID, time.Now()); err != nil {
				return err
			}
			changes, _ := json.Marshal(map[string]any{
				"title":             createdInstance.Title,
				"status":            createdInstance.Status,
				"priority":          createdInstance.Priority,
				"recurring_task_id": res.ID,
			})
			if _, err := s.repo.CreateActivity(ctx, tx, createdInstance.ID, userID, "task_created", "Initial recurring task instance generated", changes); err != nil {
				return err
			}
		}

		// Schedule next rollover at midnight
		if s.recurrenceScheduler != nil {
			nextMidnight := getNextMidnightInUserTimezone(userTimezone)
			if err := s.recurrenceScheduler.ScheduleTaskRecurrence(ctx, res.ID, nextMidnight); err != nil && s.logger != nil {
				s.logger.WarnContext(ctx, "failed to schedule recurring task recurrence in asynq", "error", err, "recurring_id", res.ID)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *service) FindAllRecurringTasks(ctx context.Context, userID string) ([]RecurringTaskResponse, error) {
	var list []RecurringTaskResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		var err error
		list, err = s.repo.FindAllRecurring(ctx, tx, userID)
		return err
	})
	return list, err
}

func (s *service) GetRecurringTaskByID(ctx context.Context, userID string, id int64) (*RecurringTaskResponse, error) {
	var res *RecurringTaskResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		var err error
		res, err = s.repo.FindRecurringByID(ctx, tx, userID, id)
		if err != nil {
			return err
		}
		if res == nil {
			return errors.TaskNotFound()
		}
		return nil
	})
	return res, err
}

func (s *service) UpdateRecurringTask(ctx context.Context, userID string, id int64, req UpdateRecurringTaskRequest) (*RecurringTaskResponse, error) {
	var res *RecurringTaskResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		var err error
		res, err = s.repo.UpdateRecurring(ctx, tx, userID, id, req)
		if err != nil {
			return err
		}
		if res != nil && s.recurrenceScheduler != nil {
			if !res.IsActive {
				_ = s.recurrenceScheduler.CancelTaskRecurrence(ctx, res.ID)
			} else {
				_, tz, _ := s.repo.GetUserThresholdAndTimezone(ctx, tx, userID)
				nextMidnight := getNextMidnightInUserTimezone(tz)
				_ = s.recurrenceScheduler.ScheduleTaskRecurrence(ctx, res.ID, nextMidnight)
			}
		}
		return nil
	})
	return res, err
}

func (s *service) DeleteRecurringTask(ctx context.Context, userID string, id int64) error {
	return s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		if err := s.repo.DeleteRecurring(ctx, tx, userID, id); err != nil {
			return err
		}
		if s.recurrenceScheduler != nil {
			_ = s.recurrenceScheduler.CancelTaskRecurrence(ctx, id)
		}
		return nil
	})
}

func (s *service) ToggleActiveRecurringTask(ctx context.Context, userID string, id int64) (*RecurringTaskResponse, error) {
	var res *RecurringTaskResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		ctx := database.ContextWithTx(ctx, tx)
		var err error
		res, err = s.repo.ToggleActiveRecurring(ctx, tx, userID, id)
		if err != nil {
			return err
		}
		if res != nil && s.recurrenceScheduler != nil {
			if !res.IsActive {
				_ = s.recurrenceScheduler.CancelTaskRecurrence(ctx, res.ID)
			} else {
				_, tz, _ := s.repo.GetUserThresholdAndTimezone(ctx, tx, userID)
				nextMidnight := getNextMidnightInUserTimezone(tz)
				_ = s.recurrenceScheduler.ScheduleTaskRecurrence(ctx, res.ID, nextMidnight)
			}
		}
		return nil
	})
	return res, err
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
		if res, err := s.pushSender.SendPushNotificationWithResult(bgCtx, tokens, "Pembaruan Batas Waktu", body, pushData); err != nil && s.logger != nil {
			s.logger.WarnContext(bgCtx, "failed to dispatch task deadline push notification", "task_id", taskID, "error", err)
		} else if res != nil && len(res.InvalidTokens) > 0 {
			_ = s.reminderRepo.DeleteDeviceTokens(bgCtx, res.InvalidTokens)
		}
	}()
}
