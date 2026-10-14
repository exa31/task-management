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
	_ = s.inspector.DeleteTask(s.queue, taskID)

	task := asynq.NewTask(
		s.taskType,
		payload,
		asynq.TaskID(taskID),
		asynq.MaxRetry(3),
		asynq.Retention(24*time.Hour),
	)

	info, err := s.client.EnqueueContext(ctx, task, asynq.ProcessAt(remindAt), asynq.Queue(s.queue))
	if err != nil && errors.Is(err, asynq.ErrTaskIDConflict) {
		fallbackID := fmt.Sprintf("reminder:%d_%d", reminderID, time.Now().UnixNano())
		fallbackTask := asynq.NewTask(
			s.taskType,
			payload,
			asynq.TaskID(fallbackID),
			asynq.MaxRetry(3),
			asynq.Retention(24*time.Hour),
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

func (s *AsynqScheduler) CancelReminder(ctx context.Context, reminderID int64) error {
	taskID := fmt.Sprintf("reminder:%d", reminderID)
	err := s.inspector.DeleteTask(s.queue, taskID)
	if err != nil {
		s.logger.Debug("asynq task deletion error (may already be finished or absent)",
			slog.Int64("reminder_id", reminderID),
			slog.String("queue", s.queue),
			slog.String("error", err.Error()),
		)
	}
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

	taskIDStr := fmt.Sprintf("task_due:%d", taskID)

	// Clean up existing task if re-scheduling
	_ = s.inspector.DeleteTask(s.queue, taskIDStr)

	task := asynq.NewTask(
		s.taskType,
		payload,
		asynq.TaskID(taskIDStr),
		asynq.MaxRetry(3),
		asynq.Retention(24*time.Hour),
	)

	info, err := s.client.EnqueueContext(ctx, task, asynq.ProcessAt(remindAt), asynq.Queue(s.queue))
	if err != nil && errors.Is(err, asynq.ErrTaskIDConflict) {
		fallbackID := fmt.Sprintf("task_due:%d_%d", taskID, time.Now().UnixNano())
		fallbackTask := asynq.NewTask(
			s.taskType,
			payload,
			asynq.TaskID(fallbackID),
			asynq.MaxRetry(3),
			asynq.Retention(24*time.Hour),
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
	taskIDStr := fmt.Sprintf("task_due:%d", taskID)
	err := s.inspector.DeleteTask(s.queue, taskIDStr)
	if err != nil {
		s.logger.Debug("asynq task deletion error for task due alert",
			slog.Int64("task_id", taskID),
			slog.String("queue", s.queue),
			slog.String("error", err.Error()),
		)
	}
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

	// Delete any existing scheduled tasks for this taskID in the scheduled set
	prefix := fmt.Sprintf("task_recurrence:%d", taskID)
	if scheduled, lErr := s.inspector.ListScheduledTasks(s.queue); lErr == nil {
		for _, t := range scheduled {
			if t.ID == prefix || strings.HasPrefix(t.ID, prefix+":") {
				_ = s.inspector.DeleteTask(s.queue, t.ID)
			}
		}
	}
	_ = s.inspector.DeleteTask(s.queue, prefix)

	taskIDStr := fmt.Sprintf("task_recurrence:%d:%d", taskID, processAt.Unix())
	task := asynq.NewTask(
		s.taskType,
		payload,
		asynq.TaskID(taskIDStr),
		asynq.MaxRetry(3),
		asynq.Retention(24*time.Hour),
	)

	info, err := s.client.EnqueueContext(ctx, task, asynq.ProcessAt(processAt), asynq.Queue(s.queue))
	if err != nil && errors.Is(err, asynq.ErrTaskIDConflict) {
		fallbackID := fmt.Sprintf("task_recurrence:%d:%d_%d", taskID, processAt.Unix(), time.Now().UnixNano())
		fallbackTask := asynq.NewTask(
			s.taskType,
			payload,
			asynq.TaskID(fallbackID),
			asynq.MaxRetry(3),
			asynq.Retention(24*time.Hour),
		)
		info, err = s.client.EnqueueContext(ctx, fallbackTask, asynq.ProcessAt(processAt), asynq.Queue(s.queue))
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
	if scheduled, lErr := s.inspector.ListScheduledTasks(s.queue); lErr == nil {
		for _, t := range scheduled {
			if t.ID == prefix || strings.HasPrefix(t.ID, prefix+":") {
				_ = s.inspector.DeleteTask(s.queue, t.ID)
			}
		}
	}
	err := s.inspector.DeleteTask(s.queue, prefix)
	if err != nil {
		s.logger.Debug("asynq task deletion error for task recurrence",
			slog.Int64("task_id", taskID),
			slog.String("queue", s.queue),
			slog.String("error", err.Error()),
		)
	}
	return nil
}

