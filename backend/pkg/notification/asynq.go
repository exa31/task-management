package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypeHTTPPost         = "http:post"
	TypeWebhookSend      = "webhook:send"
	TypeReminderDispatch = "reminder:dispatch"
	QueueDefault         = "default"
)

type HTTPTaskPayload struct {
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       any               `json:"body"`
	Payload    any               `json:"payload,omitempty"`
	Data       any               `json:"data,omitempty"`
	ReminderID int64             `json:"reminder_id,omitempty"`
}

type ReminderScheduler interface {
	ScheduleReminder(ctx context.Context, reminderID int64, remindAt time.Time) error
	CancelReminder(ctx context.Context, reminderID int64) error
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
		"X-Webhook-Secret": s.webhookSecret,
	}

	body := map[string]any{
		"reminder_id": reminderID,
		"id":          reminderID,
	}

	payloadObj := HTTPTaskPayload{
		URL:        triggerURL,
		Method:     "POST",
		Headers:    headers,
		Body:       body,
		Payload:    body,
		Data:       body,
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
