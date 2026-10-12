package reminder

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang/pkg/notification"
	"golang/pkg/realtime"
)

type Service interface {
	Create(ctx context.Context, userID string, req CreateReminderRequest) (*ReminderResponse, error)
	GetByID(ctx context.Context, userID string, id int64) (*ReminderResponse, error)
	List(ctx context.Context, userID string, filter ReminderFilter) ([]ReminderResponse, int, error)
	Update(ctx context.Context, userID string, id int64, req UpdateReminderRequest) (*ReminderResponse, error)
	Delete(ctx context.Context, userID string, id int64) error
	Snooze(ctx context.Context, userID string, id int64, req SnoozeReminderRequest) (*ReminderResponse, error)
	Dismiss(ctx context.Context, userID string, id int64) (*ReminderResponse, error)
	RegisterDeviceToken(ctx context.Context, userID string, req RegisterDeviceTokenRequest) error
	TestFCM(ctx context.Context, callingUserID string, req TestFCMRequest, isAdminOrDev bool) (*TestFCMResponse, error)

	// ProcessReminder is invoked by background Asynq worker when a reminder is due
	ProcessReminder(ctx context.Context, reminderID int64) error
}

type service struct {
	repo        Repository
	scheduler   notification.ReminderScheduler
	emailSender notification.EmailSender
	pushSender  notification.PushSender
	hub         *realtime.Hub
	logger      *slog.Logger
}

func NewService(
	repo Repository,
	scheduler notification.ReminderScheduler,
	emailSender notification.EmailSender,
	pushSender notification.PushSender,
	hub *realtime.Hub,
	logger *slog.Logger,
) Service {
	return &service{
		repo:        repo,
		scheduler:   scheduler,
		emailSender: emailSender,
		pushSender:  pushSender,
		hub:         hub,
		logger:      logger,
	}
}

func (s *service) Create(ctx context.Context, userID string, req CreateReminderRequest) (*ReminderResponse, error) {
	res, err := s.repo.Create(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	// Schedule in Asynq
	if s.scheduler != nil {
		scheduleTime := res.RemindAt
		if scheduleTime.Before(time.Now()) {
			scheduleTime = time.Now().Add(5 * time.Second)
		}
		if err := s.scheduler.ScheduleReminder(ctx, res.ID, scheduleTime); err != nil {
			s.logger.Warn("failed to schedule reminder in asynq",
				slog.Int64("reminder_id", res.ID),
				slog.String("error", err.Error()),
			)
		}
	}

	return res, nil
}

func (s *service) GetByID(ctx context.Context, userID string, id int64) (*ReminderResponse, error) {
	return s.repo.FindByID(ctx, userID, id)
}

func (s *service) List(ctx context.Context, userID string, filter ReminderFilter) ([]ReminderResponse, int, error) {
	return s.repo.FindAll(ctx, userID, filter)
}

func (s *service) Update(ctx context.Context, userID string, id int64, req UpdateReminderRequest) (*ReminderResponse, error) {
	res, err := s.repo.Update(ctx, userID, id, req)
	if err != nil {
		return nil, err
	}

	if req.RemindAt != nil && s.scheduler != nil && res.Status == "pending" {
		scheduleTime := *req.RemindAt
		if scheduleTime.Before(time.Now()) {
			scheduleTime = time.Now().Add(5 * time.Second)
		}
		_ = s.scheduler.ScheduleReminder(ctx, res.ID, scheduleTime)
	}

	return res, nil
}

func (s *service) Delete(ctx context.Context, userID string, id int64) error {
	if s.scheduler != nil {
		_ = s.scheduler.CancelReminder(ctx, id)
	}
	return s.repo.Delete(ctx, userID, id)
}

func (s *service) Snooze(ctx context.Context, userID string, id int64, req SnoozeReminderRequest) (*ReminderResponse, error) {
	newRemindAt := time.Now().Add(time.Duration(req.Minutes) * time.Minute)
	res, err := s.repo.Snooze(ctx, userID, id, newRemindAt)
	if err != nil {
		return nil, err
	}

	if s.scheduler != nil {
		_ = s.scheduler.ScheduleReminder(ctx, res.ID, newRemindAt)
	}

	return res, nil
}

func (s *service) Dismiss(ctx context.Context, userID string, id int64) (*ReminderResponse, error) {
	if s.scheduler != nil {
		_ = s.scheduler.CancelReminder(ctx, id)
	}
	return s.repo.MarkDismissed(ctx, userID, id)
}

func (s *service) RegisterDeviceToken(ctx context.Context, userID string, req RegisterDeviceTokenRequest) error {
	return s.repo.RegisterDeviceToken(ctx, userID, req.Token, req.Platform)
}

func (s *service) ProcessReminder(ctx context.Context, reminderID int64) error {
	rem, userEmail, err := s.repo.FindByIDSystem(ctx, reminderID)
	if err != nil {
		s.logger.Error("reminder processor failed to fetch reminder",
			slog.Int64("reminder_id", reminderID),
			slog.String("error", err.Error()),
		)
		return err
	}

	if rem.Status != "pending" {
		s.logger.Info("reminder is not pending, skipping dispatch",
			slog.Int64("reminder_id", reminderID),
			slog.String("status", rem.Status),
		)
		return nil
	}

	// 1. In-App Realtime alert
	if s.hub != nil {
		s.hub.Broadcast(&realtime.Message{
			Type:   "reminder:triggered",
			UserID: rem.UserID,
			Data:   rem,
		})
		s.logger.Info("realtime reminder broadcast sent", slog.Int64("reminder_id", reminderID), slog.String("user_id", rem.UserID))
	}

	// 2. Email notification via Resend
	if s.emailSender != nil && userEmail != "" {
		if err := s.emailSender.SendReminderEmail(ctx, userEmail, rem.Title, rem.Description, rem.TaskTitle); err != nil {
			s.logger.Warn("resend email reminder failed",
				slog.Int64("reminder_id", reminderID),
				slog.String("email", userEmail),
				slog.String("error", err.Error()),
			)
		}
	}

	// 3. Push notification via FCM
	if s.pushSender != nil {
		tokens, err := s.repo.GetUserDeviceTokens(ctx, rem.UserID)
		if err == nil && len(tokens) > 0 {
			data := map[string]string{
				"reminder_id": fmt.Sprintf("%d", rem.ID),
				"title":       rem.Title,
			}
			if rem.TaskID != nil {
				data["task_id"] = fmt.Sprintf("%d", *rem.TaskID)
			}
			body := rem.Description
			if body == "" && rem.TaskTitle != nil {
				body = fmt.Sprintf("Task: %s", *rem.TaskTitle)
			}
			if err := s.pushSender.SendPushNotification(ctx, tokens, rem.Title, body, data); err != nil {
				s.logger.Warn("fcm push reminder failed",
					slog.Int64("reminder_id", reminderID),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	// Mark status as sent
	if err := s.repo.MarkSent(ctx, reminderID); err != nil {
		s.logger.Error("failed to mark reminder as sent",
			slog.Int64("reminder_id", reminderID),
			slog.String("error", err.Error()),
		)
		return err
	}

	s.logger.Info("reminder successfully processed and marked sent", slog.Int64("reminder_id", reminderID))
	return nil
}

func (s *service) TestFCM(ctx context.Context, callingUserID string, req TestFCMRequest, isAdminOrDev bool) (*TestFCMResponse, error) {
	targetUserID := callingUserID
	if req.TargetUserID != "" && req.TargetUserID != callingUserID {
		if !isAdminOrDev {
			return nil, errors.New("forbidden: cannot send test push notifications to other users without authorization")
		}
		targetUserID = req.TargetUserID
	}

	var tokens []string
	if req.Token != "" {
		if !isAdminOrDev && req.TargetUserID != "" && req.TargetUserID != callingUserID {
			return nil, errors.New("forbidden: custom token testing restricted")
		}
		tokens = []string{req.Token}
	} else {
		var err error
		tokens, err = s.repo.GetUserDeviceTokens(ctx, targetUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch device tokens: %w", err)
		}
	}

	title := req.Title
	if title == "" {
		title = "Task Hub Test Notification"
	}

	body := req.Body
	if body == "" {
		body = "Push notification delivered successfully from Task Hub backend!"
	}

	data := req.Data
	if data == nil {
		data = make(map[string]string)
	}
	if _, ok := data["type"]; !ok {
		data["type"] = "test_fcm"
	}
	if _, ok := data["user_id"]; !ok {
		data["user_id"] = targetUserID
	}
	if _, ok := data["timestamp"]; !ok {
		data["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	}

	// Mask tokens for secure response preview
	maskedTokens := make([]string, len(tokens))
	for i, t := range tokens {
		if len(t) > 16 {
			maskedTokens[i] = fmt.Sprintf("%s...%s", t[:8], t[len(t)-6:])
		} else {
			maskedTokens[i] = t
		}
	}

	if len(tokens) == 0 {
		return &TestFCMResponse{
			Success:      false,
			Message:      fmt.Sprintf("No registered device tokens found for user %s. Register device token via POST /reminders/device-token first.", targetUserID),
			TargetUserID: targetUserID,
			TokensCount:  0,
			Tokens:       maskedTokens,
			Title:        title,
			Body:         body,
			Data:         data,
		}, nil
	}

	if s.pushSender == nil {
		return &TestFCMResponse{
			Success:      false,
			Message:      "FCM push sender is not configured on backend server (check FCM_SERVICE_ACCOUNT_KEY or FCM_CREDENTIALS_BASE64)",
			TargetUserID: targetUserID,
			TokensCount:  len(tokens),
			Tokens:       maskedTokens,
			Title:        title,
			Body:         body,
			Data:         data,
		}, nil
	}

	pushResult, err := s.pushSender.SendPushNotificationWithResult(ctx, tokens, title, body, data)
	if err != nil {
		return &TestFCMResponse{
			Success:      false,
			Message:      fmt.Sprintf("FCM dispatch error: %v", err),
			TargetUserID: targetUserID,
			TokensCount:  len(tokens),
			Tokens:       maskedTokens,
			Title:        title,
			Body:         body,
			Data:         data,
		}, nil
	}

	success := pushResult.SuccessCount > 0
	msg := fmt.Sprintf("FCM push dispatched to %d device(s): %d successful, %d failed", len(tokens), pushResult.SuccessCount, pushResult.FailureCount)
	if pushResult.FailureCount > 0 && len(pushResult.Errors) > 0 {
		msg = fmt.Sprintf("%s. Failures: %s", msg, strings.Join(pushResult.Errors, "; "))
	}

	return &TestFCMResponse{
		Success:      success,
		Message:      msg,
		TargetUserID: targetUserID,
		TokensCount:  len(tokens),
		SuccessCount: pushResult.SuccessCount,
		FailureCount: pushResult.FailureCount,
		Tokens:       maskedTokens,
		Title:        title,
		Body:         body,
		Data:         data,
	}, nil
}
