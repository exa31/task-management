package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypeHTTPPost         = "task:http_post"
	TypeLegacyHTTPPost   = "http:post"
	TypeWebhookSend      = "webhook:send"
	TypeReminderDispatch = "reminder:dispatch"
	QueueDefault         = "default"
)

type WorkerAsynqHTTPPayload struct {
	URL        string            `json:"url"`
	Method     string            `json:"method,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body"`
	ReminderID int64             `json:"reminder_id,omitempty"`
	TaskID     int64             `json:"task_id,omitempty"`
}

type ReminderScheduler interface {
	ScheduleReminder(ctx context.Context, reminderID int64, remindAt time.Time) error
	CancelReminder(ctx context.Context, reminderID int64) error
}

type TaskDueScheduler interface {
	ScheduleTaskDueAlert(ctx context.Context, taskID int64, remindAt time.Time) error
	CancelTaskDueAlert(ctx context.Context, taskID int64) error
}

type TaskRecurrenceScheduler interface {
	ScheduleTaskRecurrence(ctx context.Context, taskID int64, processAt time.Time) error
	CancelTaskRecurrence(ctx context.Context, taskID int64) error
}

type AsynqScheduler struct {
	client        *asynq.Client
	inspector     *asynq.Inspector
	baseURL       string
	webhookSecret string
	taskType      string
	queue         string
	logger        *slog.Logger
}

func NewAsynqScheduler(redisOpt asynq.RedisClientOpt, baseURL, webhookSecret string, logger *slog.Logger) *AsynqScheduler {
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	taskType := TypeHTTPPost
	queue := QueueDefault

	client := asynq.NewClient(redisOpt)
	inspector := asynq.NewInspector(redisOpt)
	return &AsynqScheduler{
		client:        client,
		inspector:     inspector,
		baseURL:       baseURL,
		webhookSecret: webhookSecret,
		taskType:      taskType,
		queue:         queue,
		logger:        logger,
	}
}

func (s *AsynqScheduler) SetTaskType(t string) {
	if t != "" {
		s.taskType = t
	}
}

func (s *AsynqScheduler) SetQueue(q string) {
	if q != "" {
		s.queue = q
	}
}

func (s *AsynqScheduler) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *AsynqScheduler) ScheduleReminder(ctx context.Context, reminderID int64, remindAt time.Time) error {
	triggerURL := fmt.Sprintf("%s/api/v1/reminders/trigger", s.baseURL)
	headers := map[string]string{
		"Content-Type":     "application/json",
		"X-Internal-Token": s.webhookSecret,
		"X-Webhook-Secret": s.webhookSecret,
	}

	bodyJSON := fmt.Sprintf(`{"reminder_id":%d,"id":%d}`, reminderID, reminderID)

	payloadObj := WorkerAsynqHTTPPayload{
		URL:        triggerURL,
		Method:     "POST",
		Headers:    headers,
		Body:       bodyJSON,
		ReminderID: reminderID,
	}

	payload, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to serialize http task payload: %w", err)
	}

	taskID := fmt.Sprintf("reminder:%d", reminderID)

	// Clean up existing task if re-scheduling/snoozing
	s.deleteTasksByPrefix(taskID)

	task := asynq.NewTask(
		s.taskType,
		payload,
		asynq.TaskID(taskID),
		asynq.MaxRetry(3),
	)

	info, err := s.client.EnqueueContext(ctx, task, asynq.ProcessAt(remindAt), asynq.Queue(s.queue))
	if err != nil && errors.Is(err, asynq.ErrTaskIDConflict) {
		fallbackID := fmt.Sprintf("reminder:%d_%d", reminderID, time.Now().UnixNano())
		fallbackTask := asynq.NewTask(
			s.taskType,
			payload,
			asynq.TaskID(fallbackID),
			asynq.MaxRetry(3),
		)
		info, err = s.client.EnqueueContext(ctx, fallbackTask, asynq.ProcessAt(remindAt), asynq.Queue(s.queue))
	}
	if err != nil {
		s.logger.Error("failed to enqueue http task to asynq",
			slog.Int64("reminder_id", reminderID),
			slog.Time("remind_at", remindAt),
			slog.String("task_type", s.taskType),
			slog.String("queue", s.queue),
			slog.String("target_url", triggerURL),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to enqueue reminder task: %w", err)
	}

	s.logger.Info("reminder http task successfully scheduled in asynq",
		slog.Int64("reminder_id", reminderID),
		slog.String("task_id", info.ID),
		slog.String("task_type", s.taskType),
		slog.String("queue", s.queue),
		slog.String("target_url", triggerURL),
		slog.Time("process_at", remindAt),
	)
	return nil
}

func (s *AsynqScheduler) deleteTasksByPrefix(prefix string) {
	if s.inspector == nil {
		return
	}
	_ = s.inspector.DeleteTask(s.queue, prefix)

	// Clean scheduled tasks matching prefix (including fallback IDs e.g. prefix_timestamp)
	if scheduled, err := s.inspector.ListScheduledTasks(s.queue); err == nil {
		for _, t := range scheduled {
			if t.ID == prefix || strings.HasPrefix(t.ID, prefix+":") || strings.HasPrefix(t.ID, prefix+"_") {
				_ = s.inspector.DeleteTask(s.queue, t.ID)
			}
		}
	}

	// Clean retry tasks matching prefix
	if retries, err := s.inspector.ListRetryTasks(s.queue); err == nil {
		for _, t := range retries {
			if t.ID == prefix || strings.HasPrefix(t.ID, prefix+":") || strings.HasPrefix(t.ID, prefix+"_") {
				_ = s.inspector.DeleteTask(s.queue, t.ID)
			}
		}
	}

	// Clean pending tasks matching prefix
	if pending, err := s.inspector.ListPendingTasks(s.queue); err == nil {
		for _, t := range pending {
			if t.ID == prefix || strings.HasPrefix(t.ID, prefix+":") || strings.HasPrefix(t.ID, prefix+"_") {
				_ = s.inspector.DeleteTask(s.queue, t.ID)
			}
		}
	}

	// Clean completed tasks matching prefix
	if completed, err := s.inspector.ListCompletedTasks(s.queue); err == nil {
		for _, t := range completed {
			if t.ID == prefix || strings.HasPrefix(t.ID, prefix+":") || strings.HasPrefix(t.ID, prefix+"_") {
				_ = s.inspector.DeleteTask(s.queue, t.ID)
			}
		}
	}

	// Clean archived tasks matching prefix
	if archived, err := s.inspector.ListArchivedTasks(s.queue); err == nil {
		for _, t := range archived {
			if t.ID == prefix || strings.HasPrefix(t.ID, prefix+":") || strings.HasPrefix(t.ID, prefix+"_") {
				_ = s.inspector.DeleteTask(s.queue, t.ID)
			}
		}
	}
}

func (s *AsynqScheduler) CancelReminder(ctx context.Context, reminderID int64) error {
	prefix := fmt.Sprintf("reminder:%d", reminderID)
	s.deleteTasksByPrefix(prefix)
	s.logger.Debug("asynq tasks deleted for reminder", slog.Int64("reminder_id", reminderID))
	return nil
}

func (s *AsynqScheduler) ScheduleTaskDueAlert(ctx context.Context, taskID int64, remindAt time.Time) error {
	triggerURL := fmt.Sprintf("%s/api/v1/tasks/due-warning/trigger", s.baseURL)
	headers := map[string]string{
		"Content-Type":     "application/json",
		"X-Internal-Token": s.webhookSecret,
		"X-Webhook-Secret": s.webhookSecret,
	}

	bodyJSON := fmt.Sprintf(`{"task_id":%d,"id":%d}`, taskID, taskID)

	payloadObj := WorkerAsynqHTTPPayload{
		URL:     triggerURL,
		Method:  "POST",
		Headers: headers,
		Body:    bodyJSON,
		TaskID:  taskID,
	}

	payload, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to serialize task due alert payload: %w", err)
	}

	// Clean up existing tasks if re-scheduling
	taskIDStr := fmt.Sprintf("task_due:%d", taskID)
	s.deleteTasksByPrefix(taskIDStr)

	task := asynq.NewTask(
		s.taskType,
		payload,
		asynq.TaskID(taskIDStr),
		asynq.MaxRetry(3),
	)

	info, err := s.client.EnqueueContext(ctx, task, asynq.ProcessAt(remindAt), asynq.Queue(s.queue))
	if err != nil && errors.Is(err, asynq.ErrTaskIDConflict) {
		fallbackID := fmt.Sprintf("task_due:%d_%d", taskID, time.Now().UnixNano())
		fallbackTask := asynq.NewTask(
			s.taskType,
			payload,
			asynq.TaskID(fallbackID),
			asynq.MaxRetry(3),
		)
		info, err = s.client.EnqueueContext(ctx, fallbackTask, asynq.ProcessAt(remindAt), asynq.Queue(s.queue))
	}
	if err != nil {
		s.logger.Error("failed to enqueue task due alert to asynq",
			slog.Int64("task_id", taskID),
			slog.Time("remind_at", remindAt),
			slog.String("task_type", s.taskType),
			slog.String("queue", s.queue),
			slog.String("target_url", triggerURL),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to enqueue task due alert: %w", err)
	}

	s.logger.Info("task due alert successfully scheduled in asynq",
		slog.Int64("task_id", taskID),
		slog.String("asynq_id", info.ID),
		slog.String("task_type", s.taskType),
		slog.String("queue", s.queue),
		slog.String("target_url", triggerURL),
		slog.Time("process_at", remindAt),
	)
	return nil
}

func (s *AsynqScheduler) CancelTaskDueAlert(ctx context.Context, taskID int64) error {
	prefix := fmt.Sprintf("task_due:%d", taskID)
	s.deleteTasksByPrefix(prefix)
	s.logger.Debug("asynq tasks deleted for task due alert", slog.Int64("task_id", taskID))
	return nil
}

func (s *AsynqScheduler) ScheduleTaskRecurrence(ctx context.Context, taskID int64, processAt time.Time) error {
	triggerURL := fmt.Sprintf("%s/api/v1/tasks/recurrence/trigger", s.baseURL)
	headers := map[string]string{
		"Content-Type":     "application/json",
		"X-Internal-Token": s.webhookSecret,
		"X-Webhook-Secret": s.webhookSecret,
	}

	bodyJSON := fmt.Sprintf(`{"task_id":%d,"id":%d}`, taskID, taskID)

	payloadObj := WorkerAsynqHTTPPayload{
		URL:     triggerURL,
		Method:  "POST",
		Headers: headers,
		Body:    bodyJSON,
		TaskID:  taskID,
	}

	payload, err := json.Marshal(payloadObj)
	if err != nil {
		return fmt.Errorf("failed to serialize task recurrence payload: %w", err)
	}

	// Delete any existing scheduled tasks for this taskID
	prefix := fmt.Sprintf("task_recurrence:%d", taskID)
	s.deleteTasksByPrefix(prefix)

	taskIDStr := fmt.Sprintf("task_recurrence:%d", taskID)
	task := asynq.NewTask(
		s.taskType,
		payload,
		asynq.TaskID(taskIDStr),
		asynq.MaxRetry(3),
	)

	info, err := s.client.EnqueueContext(ctx, task, asynq.ProcessAt(processAt), asynq.Queue(s.queue))
	if err != nil && errors.Is(err, asynq.ErrTaskIDConflict) {
		// When the task is currently active in the worker executing this recurrence trigger,
		// Asynq returns ErrTaskIDConflict. We schedule an asynchronous re-enqueue after
		// the worker finishes the active execution, keeping the exact clean taskIDStr without timestamp suffixes.
		go func() {
			time.Sleep(1 * time.Second)
			s.deleteTasksByPrefix(taskIDStr)
			cleanTask := asynq.NewTask(
				s.taskType,
				payload,
				asynq.TaskID(taskIDStr),
				asynq.MaxRetry(3),
			)
			if _, retryErr := s.client.Enqueue(cleanTask, asynq.ProcessAt(processAt), asynq.Queue(s.queue)); retryErr != nil {
				if s.logger != nil {
					s.logger.Warn("failed to re-enqueue recurrence task with clean ID",
						slog.Int64("task_id", taskID),
						slog.String("asynq_id", taskIDStr),
						slog.String("error", retryErr.Error()),
					)
				}
			} else if s.logger != nil {
				s.logger.Info("recurrence successfully rescheduled with clean ID",
					slog.Int64("task_id", taskID),
					slog.String("asynq_id", taskIDStr),
					slog.Time("process_at", processAt),
				)
			}
		}()
		return nil
	}

	if err != nil {
		s.logger.Error("failed to enqueue task recurrence to asynq",
			slog.Int64("task_id", taskID),
			slog.Time("process_at", processAt),
			slog.String("task_type", s.taskType),
			slog.String("queue", s.queue),
			slog.String("target_url", triggerURL),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to enqueue task recurrence: %w", err)
	}

	s.logger.Info("task recurrence successfully scheduled in asynq",
		slog.Int64("task_id", taskID),
		slog.String("asynq_id", info.ID),
		slog.String("target_url", triggerURL),
		slog.Time("process_at", processAt),
	)
	return nil
}

func (s *AsynqScheduler) CancelTaskRecurrence(ctx context.Context, taskID int64) error {
	prefix := fmt.Sprintf("task_recurrence:%d", taskID)
	s.deleteTasksByPrefix(prefix)
	s.logger.Debug("asynq tasks deleted for task recurrence", slog.Int64("task_id", taskID))
	return nil
}
