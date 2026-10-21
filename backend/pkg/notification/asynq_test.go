package notification

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

func TestAsynqScheduler_ConfigAndPayload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	redisOpt := asynq.RedisClientOpt{Addr: "127.0.0.1:6379"}

	scheduler := NewAsynqScheduler(redisOpt, "http://localhost:3000/", "test-secret", logger)
	defer scheduler.Close()

	if scheduler.baseURL != "http://localhost:3000" {
		t.Errorf("expected trimmed baseURL 'http://localhost:3000', got '%s'", scheduler.baseURL)
	}
	if scheduler.queue != QueueDefault {
		t.Errorf("expected default queue '%s', got '%s'", QueueDefault, scheduler.queue)
	}

	scheduler.SetQueue("critical")
	if scheduler.queue != "critical" {
		t.Errorf("expected queue 'critical', got '%s'", scheduler.queue)
	}

	scheduler.SetTaskType("custom:type")
	if scheduler.taskType != "custom:type" {
		t.Errorf("expected taskType 'custom:type', got '%s'", scheduler.taskType)
	}

	// Payload serialization verification
	payload := WorkerAsynqHTTPPayload{
		URL:     "http://localhost:3000/trigger",
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    `{"task_id":1}`,
		TaskID:  1,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	var parsed WorkerAsynqHTTPPayload
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if parsed.TaskID != 1 || parsed.URL != payload.URL {
		t.Errorf("payload mismatch: %+v", parsed)
	}
}

func getTestRedisOpt() *asynq.RedisClientOpt {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:63791" // Default local tunnel if active
	}
	pass := os.Getenv("TEST_REDIS_PASSWORD")
	if pass == "" {
		pass = "ekasyafrino"
	}
	return &asynq.RedisClientOpt{
		Addr:     addr,
		Password: pass,
	}
}

func TestAsynqScheduler_Integration(t *testing.T) {
	redisOpt := getTestRedisOpt()
	inspector := asynq.NewInspector(*redisOpt)
	defer inspector.Close()

	// Quick connectivity check
	if _, err := inspector.Queues(); err != nil {
		t.Skipf("Skipping integration test; Redis not reachable at %s: %v", redisOpt.Addr, err)
		return
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	scheduler := NewAsynqScheduler(*redisOpt, "http://localhost:3000", "test-secret", logger)
	defer scheduler.Close()

	ctx := context.Background()
	testTaskID := int64(77777)
	processAt := time.Now().Add(12 * time.Hour)

	t.Run("ScheduleTaskRecurrence and Cancel", func(t *testing.T) {
		// Clean start
		_ = scheduler.CancelTaskRecurrence(ctx, testTaskID)

		err := scheduler.ScheduleTaskRecurrence(ctx, testTaskID, processAt)
		if err != nil {
			t.Fatalf("ScheduleTaskRecurrence failed: %v", err)
		}

		expectedID := "task_recurrence:77777"
		info, err := inspector.GetTaskInfo("default", expectedID)
		if err != nil {
			t.Fatalf("could not find scheduled task %s: %v", expectedID, err)
		}
		if info.State != asynq.TaskStateScheduled {
			t.Errorf("expected state 'scheduled', got '%s'", info.State)
		}

		// Re-schedule (simulating update) - should clean previous and remain clean ID
		err = scheduler.ScheduleTaskRecurrence(ctx, testTaskID, processAt.Add(1*time.Hour))
		if err != nil {
			t.Fatalf("re-scheduling failed: %v", err)
		}

		infoAfter, err := inspector.GetTaskInfo("default", expectedID)
		if err != nil {
			t.Fatalf("task lost after re-scheduling: %v", err)
		}
		if infoAfter.ID != expectedID {
			t.Errorf("expected clean task ID %s, got %s", expectedID, infoAfter.ID)
		}

		// Cancel
		err = scheduler.CancelTaskRecurrence(ctx, testTaskID)
		if err != nil {
			t.Fatalf("CancelTaskRecurrence failed: %v", err)
		}

		_, err = inspector.GetTaskInfo("default", expectedID)
		if err == nil {
			t.Errorf("task %s should have been deleted after cancel", expectedID)
		}
	})

	t.Run("ScheduleTaskDueAlert and Cancel", func(t *testing.T) {
		_ = scheduler.CancelTaskDueAlert(ctx, testTaskID)

		err := scheduler.ScheduleTaskDueAlert(ctx, testTaskID, processAt)
		if err != nil {
			t.Fatalf("ScheduleTaskDueAlert failed: %v", err)
		}

		expectedID := "task_due:77777"
		info, err := inspector.GetTaskInfo("default", expectedID)
		if err != nil {
			t.Fatalf("could not find task %s: %v", expectedID, err)
		}
		if info.State != asynq.TaskStateScheduled {
			t.Errorf("expected state 'scheduled', got '%s'", info.State)
		}

		_ = scheduler.CancelTaskDueAlert(ctx, testTaskID)
		_, err = inspector.GetTaskInfo("default", expectedID)
		if err == nil {
			t.Errorf("task %s should have been deleted", expectedID)
		}
	})

	t.Run("ScheduleReminder and Cancel", func(t *testing.T) {
		_ = scheduler.CancelReminder(ctx, testTaskID)

		err := scheduler.ScheduleReminder(ctx, testTaskID, processAt)
		if err != nil {
			t.Fatalf("ScheduleReminder failed: %v", err)
		}

		expectedID := "reminder:77777"
		info, err := inspector.GetTaskInfo("default", expectedID)
		if err != nil {
			t.Fatalf("could not find task %s: %v", expectedID, err)
		}
		if info.State != asynq.TaskStateScheduled {
			t.Errorf("expected state 'scheduled', got '%s'", info.State)
		}

		_ = scheduler.CancelReminder(ctx, testTaskID)
		_, err = inspector.GetTaskInfo("default", expectedID)
		if err == nil {
			t.Errorf("task %s should have been deleted", expectedID)
		}
	})
}
