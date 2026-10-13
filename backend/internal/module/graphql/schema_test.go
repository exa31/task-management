package graphqlmodule

import (
	"context"
	"testing"
	"time"

	notemodule "golang/internal/module/note"
	projectmodule "golang/internal/module/project"
	remindermodule "golang/internal/module/reminder"
	taskmodule "golang/internal/module/task"
	usermodule "golang/internal/module/user"

	"github.com/graphql-go/graphql"
)

type dummyUserRepo struct{ usermodule.UserRepository }
type dummyProjectRepo struct{ projectmodule.Repository }
type dummyNoteRepo struct{ notemodule.Repository }

type dummyTaskRepo struct {
	taskmodule.Repository
}

func (d *dummyTaskRepo) FindAll(ctx context.Context, userID string, filter taskmodule.TaskFilter) ([]taskmodule.TaskResponse, error) {
	return []taskmodule.TaskResponse{
		{ID: 101, UserID: userID, Title: "Test Task 1", Status: "todo"},
	}, nil
}

func (d *dummyTaskRepo) FindActivitiesByUser(ctx context.Context, userID string, from, to *time.Time) ([]taskmodule.TaskActivityResponse, error) {
	return []taskmodule.TaskActivityResponse{
		{ID: 201, TaskID: 101, TaskTitle: "Test Task 1", UserID: userID, ActivityType: "progress", Description: "Finished 50%"},
	}, nil
}

type dummyReminderRepo struct {
	remindermodule.Repository
}

func (d *dummyReminderRepo) FindAll(ctx context.Context, userID string, filter remindermodule.ReminderFilter) ([]remindermodule.ReminderResponse, int, error) {
	tID := int64(101)
	tTitle := "Test Task 1"
	return []remindermodule.ReminderResponse{
		{ID: 301, UserID: userID, Title: "Reminder 1", Status: "pending", RemindAt: time.Now(), TaskID: &tID, TaskTitle: &tTitle},
	}, 1, nil
}

func TestBuildSchema_CalendarQuery(t *testing.T) {
	cfg := SchemaConfig{
		UserRepo:     &dummyUserRepo{},
		ProjectRepo:  &dummyProjectRepo{},
		TaskRepo:     &dummyTaskRepo{},
		NoteRepo:     &dummyNoteRepo{},
		ReminderRepo: &dummyReminderRepo{},
	}

	schema, err := BuildSchema(cfg)
	if err != nil {
		t.Fatalf("BuildSchema failed: %v", err)
	}

	query := `
		query GetCalendarData($from: String, $to: String) {
			tasks {
				id
				title
				status
			}
			reminders {
				id
				title
				status
				task_title
			}
			activities(from: $from, to: $to) {
				id
				task_id
				task_title
				activity_type
				description
			}
		}
	`

	ctx := context.WithValue(context.Background(), "user_id", "user-uuid-calendar")
	params := graphql.Params{
		Schema:         schema,
		RequestString:  query,
		VariableValues: map[string]interface{}{"from": "2026-10-01T00:00:00Z", "to": "2026-10-31T23:59:59Z"},
		Context:        ctx,
	}

	result := graphql.Do(params)
	if len(result.Errors) > 0 {
		t.Fatalf("graphql.Do returned errors: %v", result.Errors)
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("invalid data format: %v", result.Data)
	}

	tasks, ok := data["tasks"].([]interface{})
	if !ok || len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %v", tasks)
	}

	reminders, ok := data["reminders"].([]interface{})
	if !ok || len(reminders) != 1 {
		t.Fatalf("expected 1 reminder, got %v", reminders)
	}

	activities, ok := data["activities"].([]interface{})
	if !ok || len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %v", activities)
	}
}
