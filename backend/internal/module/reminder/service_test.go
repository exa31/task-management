package reminder

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"golang/pkg/notification"
)

type mockRepository struct {
	createFunc              func(ctx context.Context, userID string, req CreateReminderRequest) (*ReminderResponse, error)
	findByIDFunc            func(ctx context.Context, userID string, id int64) (*ReminderResponse, error)
	findAllFunc             func(ctx context.Context, userID string, filter ReminderFilter) ([]ReminderResponse, int, error)
	updateFunc              func(ctx context.Context, userID string, id int64, req UpdateReminderRequest) (*ReminderResponse, error)
	deleteFunc              func(ctx context.Context, userID string, id int64) error
	snoozeFunc              func(ctx context.Context, userID string, id int64, newRemindAt time.Time) (*ReminderResponse, error)
	markDismissedFunc       func(ctx context.Context, userID string, id int64) (*ReminderResponse, error)
	markSentFunc            func(ctx context.Context, id int64) error
	findByIDSystemFunc      func(ctx context.Context, id int64) (*ReminderResponse, string, error)
	registerDeviceTokenFunc func(ctx context.Context, userID, token, platform string) error
	getUserDeviceTokensFunc func(ctx context.Context, userID string) ([]string, error)
}

func (m *mockRepository) Create(ctx context.Context, userID string, req CreateReminderRequest) (*ReminderResponse, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *mockRepository) FindByID(ctx context.Context, userID string, id int64) (*ReminderResponse, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *mockRepository) FindAll(ctx context.Context, userID string, filter ReminderFilter) ([]ReminderResponse, int, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, userID, filter)
	}
	return nil, 0, nil
}

func (m *mockRepository) Update(ctx context.Context, userID string, id int64, req UpdateReminderRequest) (*ReminderResponse, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, userID, id, req)
	}
	return nil, nil
}

func (m *mockRepository) Delete(ctx context.Context, userID string, id int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, userID, id)
	}
	return nil
}

func (m *mockRepository) Snooze(ctx context.Context, userID string, id int64, newRemindAt time.Time) (*ReminderResponse, error) {
	if m.snoozeFunc != nil {
		return m.snoozeFunc(ctx, userID, id, newRemindAt)
	}
	return nil, nil
}

func (m *mockRepository) MarkDismissed(ctx context.Context, userID string, id int64) (*ReminderResponse, error) {
	if m.markDismissedFunc != nil {
		return m.markDismissedFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *mockRepository) MarkSent(ctx context.Context, id int64) error {
	if m.markSentFunc != nil {
		return m.markSentFunc(ctx, id)
	}
	return nil
}

func (m *mockRepository) FindByIDSystem(ctx context.Context, id int64) (*ReminderResponse, string, error) {
	if m.findByIDSystemFunc != nil {
		return m.findByIDSystemFunc(ctx, id)
	}
	return nil, "", nil
}

func (m *mockRepository) RegisterDeviceToken(ctx context.Context, userID, token, platform string) error {
	if m.registerDeviceTokenFunc != nil {
		return m.registerDeviceTokenFunc(ctx, userID, token, platform)
	}
	return nil
}

func (m *mockRepository) GetUserDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	if m.getUserDeviceTokensFunc != nil {
		return m.getUserDeviceTokensFunc(ctx, userID)
	}
	return []string{}, nil
}

type mockScheduler struct {
	scheduledReminders map[int64]time.Time
	cancelledReminders []int64
}

func newMockScheduler() *mockScheduler {
	return &mockScheduler{
		scheduledReminders: make(map[int64]time.Time),
		cancelledReminders: []int64{},
	}
}

func (m *mockScheduler) ScheduleReminder(ctx context.Context, reminderID int64, remindAt time.Time) error {
	m.scheduledReminders[reminderID] = remindAt
	return nil
}

func (m *mockScheduler) CancelReminder(ctx context.Context, reminderID int64) error {
	m.cancelledReminders = append(m.cancelledReminders, reminderID)
	delete(m.scheduledReminders, reminderID)
	return nil
}

func TestService_CreateAndSnooze(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockRepo := &mockRepository{
		createFunc: func(ctx context.Context, userID string, req CreateReminderRequest) (*ReminderResponse, error) {
			return &ReminderResponse{
				ID:       10,
				UserID:   userID,
				Title:    req.Title,
				RemindAt: req.RemindAt,
				Status:   "pending",
			}, nil
		},
		snoozeFunc: func(ctx context.Context, userID string, id int64, newRemindAt time.Time) (*ReminderResponse, error) {
			return &ReminderResponse{
				ID:          id,
				UserID:      userID,
				Title:       "Test Reminder",
				RemindAt:    newRemindAt,
				Status:      "pending",
				SnoozeCount: 1,
			}, nil
		},
	}

	scheduler := newMockScheduler()
	svc := NewService(mockRepo, scheduler, nil, nil, nil, logger)

	t.Run("Create Reminder schedules in scheduler", func(t *testing.T) {
		remindAt := time.Now().Add(1 * time.Hour)
		resp, err := svc.Create(ctx, "user-1", CreateReminderRequest{
			Title:    "Review PR",
			RemindAt: remindAt,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ID != 10 {
			t.Errorf("expected ID 10, got %d", resp.ID)
		}
		if _, ok := scheduler.scheduledReminders[10]; !ok {
			t.Errorf("expected reminder 10 to be scheduled in scheduler")
		}
	})

	t.Run("Snooze Reminder updates remind_at and reschedules", func(t *testing.T) {
		resp, err := svc.Snooze(ctx, "user-1", 10, SnoozeReminderRequest{Minutes: 15})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.SnoozeCount != 1 {
			t.Errorf("expected snooze count 1, got %d", resp.SnoozeCount)
		}
		if _, ok := scheduler.scheduledReminders[10]; !ok {
			t.Errorf("expected reminder 10 to be rescheduled")
		}
	})
}

func TestService_ProcessReminder(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	sentCalled := false
	mockRepo := &mockRepository{
		findByIDSystemFunc: func(ctx context.Context, id int64) (*ReminderResponse, string, error) {
			if id == 99 {
				return nil, "", errors.New("not found")
			}
			return &ReminderResponse{
				ID:     id,
				UserID: "user-123",
				Title:  "Meeting soon",
				Status: "pending",
			}, "test@example.com", nil
		},
		markSentFunc: func(ctx context.Context, id int64) error {
			sentCalled = true
			return nil
		},
	}

	svc := NewService(mockRepo, nil, nil, nil, nil, logger)

	t.Run("Process pending reminder marks as sent", func(t *testing.T) {
		err := svc.ProcessReminder(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !sentCalled {
			t.Errorf("expected markSent to be called")
		}
	})

	t.Run("Process non-existent reminder returns error", func(t *testing.T) {
		err := svc.ProcessReminder(ctx, 99)
		if err == nil {
			t.Fatalf("expected error for non-existent reminder")
		}
	})
}

type mockTestPushSender struct {
	sendPushNotificationWithResultFunc func(ctx context.Context, tokens []string, title, body string, data map[string]string) (*notification.PushResult, error)
}

func (m *mockTestPushSender) SendPushNotification(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	_, err := m.SendPushNotificationWithResult(ctx, tokens, title, body, data)
	return err
}

func (m *mockTestPushSender) SendPushNotificationWithResult(ctx context.Context, tokens []string, title, body string, data map[string]string) (*notification.PushResult, error) {
	if m.sendPushNotificationWithResultFunc != nil {
		return m.sendPushNotificationWithResultFunc(ctx, tokens, title, body, data)
	}
	return &notification.PushResult{SuccessCount: len(tokens), FailureCount: 0, TotalTokens: len(tokens)}, nil
}

func TestService_TestFCM(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockRepo := &mockRepository{
		getUserDeviceTokensFunc: func(ctx context.Context, userID string) ([]string, error) {
			if userID == "user-with-tokens" {
				return []string{"fcm-token-alpha-1234567890", "fcm-token-beta-1234567890"}, nil
			}
			return []string{}, nil
		},
	}

	mockPush := &mockTestPushSender{}
	svc := NewService(mockRepo, nil, nil, mockPush, nil, logger)

	t.Run("Returns graceful message when target user has no registered tokens", func(t *testing.T) {
		res, err := svc.TestFCM(ctx, "user-no-tokens", TestFCMRequest{}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Success {
			t.Errorf("expected success to be false when 0 tokens found")
		}
		if res.TokensCount != 0 {
			t.Errorf("expected 0 tokens, got %d", res.TokensCount)
		}
	})

	t.Run("Rejects cross-user targeting when not admin/dev", func(t *testing.T) {
		_, err := svc.TestFCM(ctx, "caller-user", TestFCMRequest{TargetUserID: "other-user"}, false)
		if err == nil {
			t.Fatalf("expected forbidden error for unauthorized cross-user targeting")
		}
	})

	t.Run("Allows cross-user targeting when isAdminOrDev is true", func(t *testing.T) {
		res, err := svc.TestFCM(ctx, "caller-user", TestFCMRequest{TargetUserID: "user-with-tokens"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Success {
			t.Errorf("expected push to succeed")
		}
		if res.TokensCount != 2 {
			t.Errorf("expected 2 tokens, got %d", res.TokensCount)
		}
		if res.TargetUserID != "user-with-tokens" {
			t.Errorf("expected target_user_id to be user-with-tokens, got %s", res.TargetUserID)
		}
	})

	t.Run("Dispatches to custom token directly", func(t *testing.T) {
		customToken := "direct-fcm-custom-token-xyz-12345678"
		res, err := svc.TestFCM(ctx, "caller-user", TestFCMRequest{
			Token: customToken,
			Title: "Direct Test",
			Body:  "Direct message",
		}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Success {
			t.Errorf("expected push to succeed for direct token")
		}
		if res.TokensCount != 1 {
			t.Errorf("expected 1 token, got %d", res.TokensCount)
		}
		if res.Title != "Direct Test" {
			t.Errorf("expected title 'Direct Test', got '%s'", res.Title)
		}
	})

	t.Run("Gracefully handles nil PushSender without panicking", func(t *testing.T) {
		svcNoPush := NewService(mockRepo, nil, nil, nil, nil, logger)
		res, err := svcNoPush.TestFCM(ctx, "user-with-tokens", TestFCMRequest{}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Success {
			t.Errorf("expected success to be false when pushSender is nil")
		}
	})
}
