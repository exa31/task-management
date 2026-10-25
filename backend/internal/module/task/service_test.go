package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"golang/internal/module/reminder"
	"golang/pkg/notification"
)

type mockTransactor struct{}

func (m *mockTransactor) WithTransaction(fn func(tx *sqlx.Tx) error) error {
	return fn(nil)
}

func (m *mockTransactor) WithTransactionCtx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	return fn(nil)
}

func (m *mockTransactor) WithReadOnlyTransactionCtx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	return fn(nil)
}

type mockRepository struct {
	createFunc      func(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error)
	findByIDFunc    func(ctx context.Context, userID string, id int64) (*TaskResponse, error)
	findAllFunc     func(ctx context.Context, userID string, filter TaskFilter) ([]TaskResponse, error)
	updateFunc      func(ctx context.Context, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error)
	deleteFunc      func(ctx context.Context, userID string, id int64) error
	logProgressFunc func(ctx context.Context, userID string, id int64, req LogProgressRequest) (*TaskResponse, error)
	toggleFocusFunc func(ctx context.Context, userID string, id int64) (*TaskResponse, error)

	createSubtaskFunc func(ctx context.Context, userID string, taskID int64, req CreateSubtaskRequest) (*SubtaskResponse, error)
	updateSubtaskFunc func(ctx context.Context, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error)
	deleteSubtaskFunc func(ctx context.Context, userID string, taskID int64, subtaskID int64) error
	countSubtasksFunc func(ctx context.Context, taskID int64) (int, int, error)

	findByIDSystemFunc     func(ctx context.Context, id int64) (*TaskResponse, string, string, int, bool, error)
	markDueWarningSentFunc func(ctx context.Context, id int64) error
	getUserThresholdFunc   func(ctx context.Context, userID string) (int, string, error)

	getWidgetBundleFunc             func(ctx context.Context, userID string) (*WidgetBundleResponse, error)
	getWorkspaceStatsFunc           func(ctx context.Context, userID string) (*WorkspaceStatsResponse, error)
	updateRecurrenceGeneratedAtFunc func(ctx context.Context, id int64, t time.Time) error

	createActivityFunc       func(ctx context.Context, taskID int64, userID string, activityType string, description string, changes []byte) (*TaskActivityResponse, error)
	findActivitiesFunc       func(ctx context.Context, userID string, taskID int64) ([]TaskActivityResponse, error)
	findActivitiesByUserFunc func(ctx context.Context, userID string, from, to *time.Time) ([]TaskActivityResponse, error)

	findRecurringByIDFunc            func(ctx context.Context, userID string, id int64) (*RecurringTaskResponse, error)
	findRecurringByIDSystemFunc      func(ctx context.Context, id int64) (*RecurringTaskResponse, string, error)
	findAllRecurringFunc             func(ctx context.Context, userID string) ([]RecurringTaskResponse, error)
	createRecurringFunc              func(ctx context.Context, userID string, req CreateRecurringTaskRequest) (*RecurringTaskResponse, error)
	updateRecurringFunc              func(ctx context.Context, userID string, id int64, req UpdateRecurringTaskRequest) (*RecurringTaskResponse, error)
	deleteRecurringFunc              func(ctx context.Context, userID string, id int64) error
	toggleActiveRecurringFunc        func(ctx context.Context, userID string, id int64) (*RecurringTaskResponse, error)
	findTodayInstanceFunc            func(ctx context.Context, recurringTaskID int64, todayDate time.Time) (*TaskResponse, error)
	createInstanceFromRecurringFunc   func(ctx context.Context, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error)
	updateRecurringLastGeneratedFunc  func(ctx context.Context, id int64, t time.Time) error
	createRecurringSubtaskFunc        func(ctx context.Context, recurringTaskID int64, title, description string, orderIndex int) error
	updateRecurringSubtaskByTitleFunc func(ctx context.Context, recurringTaskID int64, oldTitle, newTitle, description string) error
	deleteRecurringSubtaskByTitleFunc func(ctx context.Context, recurringTaskID int64, title string) error
	syncRecurringFromTaskFunc         func(ctx context.Context, userID string, task *TaskResponse) error
	convertTaskToRecurringFunc        func(ctx context.Context, userID string, taskID int64, pattern string, interval int, endDate *time.Time) (*RecurringTaskResponse, error)
}

func (m *mockRepository) Create(ctx context.Context, tx *sqlx.Tx, userID string, req CreateTaskRequest) (*TaskResponse, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *mockRepository) FindByID(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*TaskResponse, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *mockRepository) FindAll(ctx context.Context, tx *sqlx.Tx, userID string, filter TaskFilter) ([]TaskResponse, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, userID, filter)
	}
	return []TaskResponse{}, nil
}

func (m *mockRepository) Update(ctx context.Context, tx *sqlx.Tx, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, userID, id, req)
	}
	return nil, nil
}

func (m *mockRepository) Delete(ctx context.Context, tx *sqlx.Tx, userID string, id int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, userID, id)
	}
	return nil
}

func (m *mockRepository) LogProgress(ctx context.Context, tx *sqlx.Tx, userID string, id int64, req LogProgressRequest) (*TaskResponse, error) {
	if m.logProgressFunc != nil {
		return m.logProgressFunc(ctx, userID, id, req)
	}
	return nil, nil
}

func (m *mockRepository) ToggleFocus(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*TaskResponse, error) {
	if m.toggleFocusFunc != nil {
		return m.toggleFocusFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *mockRepository) CreateSubtask(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, req CreateSubtaskRequest) (*SubtaskResponse, error) {
	if m.createSubtaskFunc != nil {
		return m.createSubtaskFunc(ctx, userID, taskID, req)
	}
	return nil, nil
}

func (m *mockRepository) GetWidgetBundle(ctx context.Context, tx *sqlx.Tx, userID string) (*WidgetBundleResponse, error) {
	if m.getWidgetBundleFunc != nil {
		return m.getWidgetBundleFunc(ctx, userID)
	}
	return &WidgetBundleResponse{}, nil
}

func (m *mockRepository) GetWorkspaceStats(ctx context.Context, tx *sqlx.Tx, userID string) (*WorkspaceStatsResponse, error) {
	if m.getWorkspaceStatsFunc != nil {
		return m.getWorkspaceStatsFunc(ctx, userID)
	}
	return &WorkspaceStatsResponse{}, nil
}

func (m *mockRepository) UpdateRecurrenceGeneratedAt(ctx context.Context, tx *sqlx.Tx, id int64, t time.Time) error {
	if m.updateRecurrenceGeneratedAtFunc != nil {
		return m.updateRecurrenceGeneratedAtFunc(ctx, id, t)
	}
	return nil
}

func (m *mockRepository) UpdateSubtask(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error) {
	if m.updateSubtaskFunc != nil {
		return m.updateSubtaskFunc(ctx, userID, taskID, subtaskID, req)
	}
	return nil, nil
}

func (m *mockRepository) DeleteSubtask(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, subtaskID int64) error {
	if m.deleteSubtaskFunc != nil {
		return m.deleteSubtaskFunc(ctx, userID, taskID, subtaskID)
	}
	return nil
}

func (m *mockRepository) CountSubtasks(ctx context.Context, tx *sqlx.Tx, taskID int64) (int, int, error) {
	if m.countSubtasksFunc != nil {
		return m.countSubtasksFunc(ctx, taskID)
	}
	return 0, 0, nil
}

func (m *mockRepository) FindByIDSystem(ctx context.Context, tx *sqlx.Tx, id int64) (*TaskResponse, string, string, int, bool, error) {
	if m.findByIDSystemFunc != nil {
		return m.findByIDSystemFunc(ctx, id)
	}
	return nil, "", "", 0, false, nil
}

func (m *mockRepository) MarkDueWarningSent(ctx context.Context, tx *sqlx.Tx, id int64) error {
	if m.markDueWarningSentFunc != nil {
		return m.markDueWarningSentFunc(ctx, id)
	}
	return nil
}

func (m *mockRepository) GetUserThresholdAndTimezone(ctx context.Context, tx *sqlx.Tx, userID string) (int, string, error) {
	if m.getUserThresholdFunc != nil {
		return m.getUserThresholdFunc(ctx, userID)
	}
	return 1440, "Asia/Jakarta", nil
}

func (m *mockRepository) CreateActivity(ctx context.Context, tx *sqlx.Tx, taskID int64, userID string, activityType string, description string, changes []byte) (*TaskActivityResponse, error) {
	if m.createActivityFunc != nil {
		return m.createActivityFunc(ctx, taskID, userID, activityType, description, changes)
	}
	return &TaskActivityResponse{
		ID:           1,
		TaskID:       taskID,
		UserID:       userID,
		ActivityType: activityType,
		Description:  description,
		CreatedAt:    time.Now(),
	}, nil
}

func (m *mockRepository) FindActivitiesByTaskID(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64) ([]TaskActivityResponse, error) {
	if m.findActivitiesFunc != nil {
		return m.findActivitiesFunc(ctx, userID, taskID)
	}
	return []TaskActivityResponse{}, nil
}

func (m *mockRepository) FindActivitiesByUser(ctx context.Context, tx *sqlx.Tx, userID string, from, to *time.Time, limit ...int) ([]TaskActivityResponse, error) {
	if m.findActivitiesByUserFunc != nil {
		return m.findActivitiesByUserFunc(ctx, userID, from, to)
	}
	return []TaskActivityResponse{}, nil
}

func (m *mockRepository) FindActivityCountsByUser(ctx context.Context, tx *sqlx.Tx, userID string, from, to *time.Time, tz string) ([]ActivityDateCount, error) {
	return []ActivityDateCount{}, nil
}

func (m *mockRepository) FindRecurringByID(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*RecurringTaskResponse, error) {
	if m.findRecurringByIDFunc != nil {
		return m.findRecurringByIDFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *mockRepository) FindRecurringByIDSystem(ctx context.Context, tx *sqlx.Tx, id int64) (*RecurringTaskResponse, string, error) {
	if m.findRecurringByIDSystemFunc != nil {
		return m.findRecurringByIDSystemFunc(ctx, id)
	}
	return nil, "Asia/Jakarta", nil
}

func (m *mockRepository) FindAllRecurring(ctx context.Context, tx *sqlx.Tx, userID string) ([]RecurringTaskResponse, error) {
	if m.findAllRecurringFunc != nil {
		return m.findAllRecurringFunc(ctx, userID)
	}
	return []RecurringTaskResponse{}, nil
}

func (m *mockRepository) CreateRecurring(ctx context.Context, tx *sqlx.Tx, userID string, req CreateRecurringTaskRequest) (*RecurringTaskResponse, error) {
	if m.createRecurringFunc != nil {
		return m.createRecurringFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *mockRepository) UpdateRecurring(ctx context.Context, tx *sqlx.Tx, userID string, id int64, req UpdateRecurringTaskRequest) (*RecurringTaskResponse, error) {
	if m.updateRecurringFunc != nil {
		return m.updateRecurringFunc(ctx, userID, id, req)
	}
	return nil, nil
}

func (m *mockRepository) DeleteRecurring(ctx context.Context, tx *sqlx.Tx, userID string, id int64) error {
	if m.deleteRecurringFunc != nil {
		return m.deleteRecurringFunc(ctx, userID, id)
	}
	return nil
}

func (m *mockRepository) ToggleActiveRecurring(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*RecurringTaskResponse, error) {
	if m.toggleActiveRecurringFunc != nil {
		return m.toggleActiveRecurringFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *mockRepository) FindTodayInstance(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64, todayDate time.Time) (*TaskResponse, error) {
	if m.findTodayInstanceFunc != nil {
		return m.findTodayInstanceFunc(ctx, recurringTaskID, todayDate)
	}
	return nil, nil
}

func (m *mockRepository) CreateInstanceFromRecurring(ctx context.Context, tx *sqlx.Tx, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error) {
	if m.createInstanceFromRecurringFunc != nil {
		return m.createInstanceFromRecurringFunc(ctx, master, dueDate)
	}
	return nil, nil
}

func (m *mockRepository) UpdateRecurringLastGenerated(ctx context.Context, tx *sqlx.Tx, id int64, t time.Time) error {
	if m.updateRecurringLastGeneratedFunc != nil {
		return m.updateRecurringLastGeneratedFunc(ctx, id, t)
	}
	return nil
}

func (m *mockRepository) CreateRecurringSubtask(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64, title, description string, orderIndex int) error {
	if m.createRecurringSubtaskFunc != nil {
		return m.createRecurringSubtaskFunc(ctx, recurringTaskID, title, description, orderIndex)
	}
	return nil
}

func (m *mockRepository) UpdateRecurringSubtaskByTitle(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64, oldTitle, newTitle, description string) error {
	if m.updateRecurringSubtaskByTitleFunc != nil {
		return m.updateRecurringSubtaskByTitleFunc(ctx, recurringTaskID, oldTitle, newTitle, description)
	}
	return nil
}

func (m *mockRepository) DeleteRecurringSubtaskByTitle(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64, title string) error {
	if m.deleteRecurringSubtaskByTitleFunc != nil {
		return m.deleteRecurringSubtaskByTitleFunc(ctx, recurringTaskID, title)
	}
	return nil
}

func (m *mockRepository) SyncRecurringFromTask(ctx context.Context, tx *sqlx.Tx, userID string, task *TaskResponse) error {
	if m.syncRecurringFromTaskFunc != nil {
		return m.syncRecurringFromTaskFunc(ctx, userID, task)
	}
	return nil
}

func (m *mockRepository) ConvertTaskToRecurring(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, pattern string, interval int, endDate *time.Time) (*RecurringTaskResponse, error) {
	if m.convertTaskToRecurringFunc != nil {
		return m.convertTaskToRecurringFunc(ctx, userID, taskID, pattern, interval, endDate)
	}
	return nil, nil
}

func TestTaskService_Create(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	mockRepo := &mockRepository{
		createFunc: func(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error) {
			return &TaskResponse{
				ID:       1,
				UserID:   userID,
				Title:    req.Title,
				Priority: req.Priority,
				Status:   req.Status,
			}, nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})
	res, err := svc.Create(ctx, testUserID, CreateTaskRequest{
		Title:    "Ship Feature",
		Priority: "urgent",
		Status:   "todo",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 1 || res.Title != "Ship Feature" {
		t.Errorf("unexpected task result: %+v", res)
	}
}

func TestTaskService_LogProgress(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{
				ID:     id,
				UserID: userID,
				Title:  "Building Backend Feature",
			}, nil
		},
		logProgressFunc: func(ctx context.Context, userID string, id int64, req LogProgressRequest) (*TaskResponse, error) {
			now := time.Now()
			return &TaskResponse{
				ID:               id,
				UserID:           userID,
				Title:            "Building Backend Feature",
				LastProgressNote: req.ProgressNote,
				LastProgressAt:   &now,
			}, nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})
	res, err := svc.LogProgress(ctx, testUserID, 1, LogProgressRequest{
		ProgressNote: "API auth and database migrations complete",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.LastProgressNote != "API auth and database migrations complete" {
		t.Errorf("expected last progress note to match, got %s", res.LastProgressNote)
	}
}

func TestTaskService_ToggleFocus(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{ID: id, UserID: userID, IsFocus: false}, nil
		},
		toggleFocusFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{ID: id, UserID: userID, IsFocus: true}, nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})
	res, err := svc.ToggleFocus(ctx, testUserID, 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.IsFocus {
		t.Errorf("expected is_focus to be true")
	}
}

func TestTaskService_GetByID(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{ID: id, UserID: userID, Title: "Test Task"}, nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})
	res, err := svc.GetByID(ctx, testUserID, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 1 || res.Title != "Test Task" {
		t.Errorf("unexpected task result: %+v", res)
	}
}

func TestTaskService_FindAll(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	mockRepo := &mockRepository{
		findAllFunc: func(ctx context.Context, userID string, filter TaskFilter) ([]TaskResponse, error) {
			return []TaskResponse{
				{ID: 1, UserID: userID, Title: "Task 1"},
				{ID: 2, UserID: userID, Title: "Task 2"},
			}, nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})
	res, err := svc.FindAll(ctx, testUserID, TaskFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(res))
	}
}

func TestTaskService_ProcessDueWarning(t *testing.T) {
	ctx := context.Background()
	dueDate := time.Now().Add(2 * time.Hour)
	markedSent := false

	mockRepo := &mockRepository{
		findByIDSystemFunc: func(ctx context.Context, id int64) (*TaskResponse, string, string, int, bool, error) {
			return &TaskResponse{
				ID:       id,
				UserID:   "user-1",
				Title:    "Urgent Bugfix",
				Status:   "todo",
				Priority: "urgent",
				DueDate:  &dueDate,
			}, "dev@example.com", "Asia/Jakarta", 1440, true, nil
		},
		markDueWarningSentFunc: func(ctx context.Context, id int64) error {
			markedSent = true
			return nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})
	err := svc.ProcessDueWarning(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !markedSent {
		t.Errorf("expected markDueWarningSent to be called")
	}
}

func TestTaskService_Activities(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"
	taskID := int64(42)

	activityCreated := false
	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{
				ID:     id,
				UserID: userID,
				Title:  "Test Task",
			}, nil
		},
		createActivityFunc: func(ctx context.Context, tid int64, uid string, actType string, desc string, changes []byte) (*TaskActivityResponse, error) {
			activityCreated = true
			return &TaskActivityResponse{
				ID:           101,
				TaskID:       tid,
				UserID:       uid,
				ActivityType: actType,
				Description:  desc,
				CreatedAt:    time.Now(),
			}, nil
		},
		findActivitiesFunc: func(ctx context.Context, uid string, tid int64) ([]TaskActivityResponse, error) {
			return []TaskActivityResponse{
				{
					ID:           101,
					TaskID:       tid,
					UserID:       uid,
					ActivityType: "progress_note",
					Description:  "Selesai implementasi endpoint",
					CreatedAt:    time.Now(),
				},
			}, nil
		},
		logProgressFunc: func(ctx context.Context, userID string, id int64, req LogProgressRequest) (*TaskResponse, error) {
			return &TaskResponse{ID: id, UserID: userID, LastProgressNote: req.ProgressNote}, nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})

	// 1. CreateActivity
	act, err := svc.CreateActivity(ctx, testUserID, taskID, CreateTaskActivityRequest{
		Description: "Selesai implementasi endpoint",
	})
	if err != nil {
		t.Fatalf("unexpected error creating activity: %v", err)
	}
	if act == nil || !activityCreated {
		t.Fatalf("expected activity to be created")
	}

	// 2. GetActivities
	activities, err := svc.GetActivities(ctx, testUserID, taskID)
	if err != nil {
		t.Fatalf("unexpected error getting activities: %v", err)
	}
	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}
	if activities[0].Description != "Selesai implementasi endpoint" {
		t.Fatalf("unexpected activity description: %s", activities[0].Description)
	}
}

func TestTaskService_SubtaskDueDateValidation(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"
	taskDueDate := time.Now().Add(24 * time.Hour)
	validSubtaskDueDate := taskDueDate.Add(-2 * time.Hour)
	invalidSubtaskDueDate := taskDueDate.Add(2 * time.Hour)

	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{
				ID:      id,
				UserID:  userID,
				Title:   "Parent Task",
				DueDate: &taskDueDate,
				Subtasks: []SubtaskResponse{
					{
						ID:      10,
						TaskID:  id,
						Title:   "Existing Subtask",
						DueDate: &validSubtaskDueDate,
					},
				},
			}, nil
		},
		createSubtaskFunc: func(ctx context.Context, userID string, taskID int64, req CreateSubtaskRequest) (*SubtaskResponse, error) {
			return &SubtaskResponse{
				ID:      11,
				TaskID:  taskID,
				Title:   req.Title,
				DueDate: req.DueDate,
			}, nil
		},
		updateSubtaskFunc: func(ctx context.Context, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error) {
			return &SubtaskResponse{
				ID:      subtaskID,
				TaskID:  taskID,
				Title:   "Updated Subtask",
				DueDate: req.DueDate,
			}, nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})

	// 1. Create subtask exceeding parent due date should fail
	_, err := svc.CreateSubtask(ctx, testUserID, 1, CreateSubtaskRequest{
		Title:   "Too late subtask",
		DueDate: &invalidSubtaskDueDate,
	})
	if err == nil {
		t.Fatalf("expected error when subtask due date exceeds parent due date, got nil")
	}

	// 2. Create subtask within parent due date should succeed
	sub, err := svc.CreateSubtask(ctx, testUserID, 1, CreateSubtaskRequest{
		Title:   "Valid subtask",
		DueDate: &validSubtaskDueDate,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub == nil || sub.Title != "Valid subtask" {
		t.Fatalf("unexpected subtask result: %+v", sub)
	}

	// 3. Update subtask exceeding parent due date should fail
	_, err = svc.UpdateSubtask(ctx, testUserID, 1, 10, UpdateSubtaskRequest{
		DueDate: &invalidSubtaskDueDate,
	})
	if err == nil {
		t.Fatalf("expected error when updating subtask due date to exceed parent due date")
	}

	// 4. Update parent task due date earlier than existing subtask should fail
	earlierParentDueDate := validSubtaskDueDate.Add(-1 * time.Hour)
	_, err = svc.Update(ctx, testUserID, 1, UpdateTaskRequest{
		DueDate: &earlierParentDueDate,
	})
	if err == nil {
		t.Fatalf("expected error when setting task due date earlier than existing subtask due date")
	}
}

func TestTaskService_GetWidgetBundle(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	mockRepo := &mockRepository{
		getWidgetBundleFunc: func(ctx context.Context, userID string) (*WidgetBundleResponse, error) {
			return &WidgetBundleResponse{
				FocusTasks: []TaskResponse{
					{ID: 1, Title: "Focus 1", IsFocus: true},
				},
				UpcomingDeadlines: []TaskResponse{
					{ID: 2, Title: "Deadline 1"},
				},
				CompletedToday: 3,
				TotalToday:     5,
			}, nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})
	bundle, err := svc.GetWidgetBundle(ctx, testUserID)
	if err != nil {
		t.Fatalf("unexpected error getting widget bundle: %v", err)
	}
	if bundle == nil {
		t.Fatalf("expected non-nil widget bundle")
	}
	if len(bundle.FocusTasks) != 1 || len(bundle.UpcomingDeadlines) != 1 {
		t.Fatalf("unexpected counts in bundle: %+v", bundle)
	}
	if bundle.CompletedToday != 3 || bundle.TotalToday != 5 {
		t.Fatalf("unexpected today counts: %d / %d", bundle.CompletedToday, bundle.TotalToday)
	}
}

func TestTaskService_GetWorkspaceStats(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	mockRepo := &mockRepository{
		getWorkspaceStatsFunc: func(ctx context.Context, userID string) (*WorkspaceStatsResponse, error) {
			return &WorkspaceStatsResponse{
				TasksInProgress:  4,
				TasksTodo:        7,
				TasksTotal:       15,
				TotalProjects:    3,
				PendingReminders: 2,
				TotalNotes:       6,
			}, nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})
	stats, err := svc.GetWorkspaceStats(ctx, testUserID)
	if err != nil {
		t.Fatalf("unexpected error getting workspace stats: %v", err)
	}
	if stats == nil {
		t.Fatalf("expected non-nil workspace stats")
	}
	if stats.TasksInProgress != 4 || stats.TotalProjects != 3 || stats.PendingReminders != 2 || stats.TotalNotes != 6 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

type mockTaskScheduler struct {
	dueScheduled        map[int64]time.Time
	recurrenceScheduled map[int64]time.Time
	dueCancelled        map[int64]bool
	recurrenceCancelled map[int64]bool
	remindersCancelled  map[int64]bool
}

func newMockTaskScheduler() *mockTaskScheduler {
	return &mockTaskScheduler{
		dueScheduled:        make(map[int64]time.Time),
		recurrenceScheduled: make(map[int64]time.Time),
		dueCancelled:        make(map[int64]bool),
		recurrenceCancelled: make(map[int64]bool),
		remindersCancelled:  make(map[int64]bool),
	}
}

func (m *mockTaskScheduler) ScheduleReminder(ctx context.Context, reminderID int64, remindAt time.Time) error {
	return nil
}

func (m *mockTaskScheduler) CancelReminder(ctx context.Context, reminderID int64) error {
	m.remindersCancelled[reminderID] = true
	return nil
}

func (m *mockTaskScheduler) ScheduleTaskDueAlert(ctx context.Context, taskID int64, remindAt time.Time) error {
	m.dueScheduled[taskID] = remindAt
	return nil
}

func (m *mockTaskScheduler) CancelTaskDueAlert(ctx context.Context, taskID int64) error {
	delete(m.dueScheduled, taskID)
	m.dueCancelled[taskID] = true
	return nil
}

func (m *mockTaskScheduler) ScheduleTaskRecurrence(ctx context.Context, taskID int64, processAt time.Time) error {
	m.recurrenceScheduled[taskID] = processAt
	return nil
}

func (m *mockTaskScheduler) CancelTaskRecurrence(ctx context.Context, taskID int64) error {
	delete(m.recurrenceScheduled, taskID)
	m.recurrenceCancelled[taskID] = true
	return nil
}

func TestTaskService_Create_WithSchedulerInTransaction(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"
	due := time.Now().Add(2 * time.Hour)

	mockRepo := &mockRepository{
		createFunc: func(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error) {
			return &TaskResponse{
				ID:          100,
				UserID:      userID,
				Title:       req.Title,
				Priority:    req.Priority,
				Status:      req.Status,
				DueDate:     req.DueDate,
				IsRecurring: req.IsRecurring,
			}, nil
		},
		createRecurringFunc: func(ctx context.Context, userID string, req CreateRecurringTaskRequest) (*RecurringTaskResponse, error) {
			return &RecurringTaskResponse{
				ID:       100,
				UserID:   userID,
				Title:    req.Title,
				IsActive: true,
			}, nil
		},
		getUserThresholdFunc: func(ctx context.Context, userID string) (int, string, error) {
			return 30, "Asia/Jakarta", nil
		},
	}

	sched := newMockTaskScheduler()
	svc := NewService(mockRepo, &mockTransactor{})
	svc.SetNotificationDependencies(sched, nil, nil, nil, nil, nil)

	res, err := svc.Create(ctx, testUserID, CreateTaskRequest{
		Title:       "Recurring Task With Due Date",
		Priority:    "high",
		Status:      "todo",
		DueDate:     &due,
		IsRecurring: true,
	})

	if err != nil {
		t.Fatalf("unexpected error creating task: %v", err)
	}
	if res.ID != 100 {
		t.Fatalf("unexpected id: %d", res.ID)
	}

	if _, ok := sched.dueScheduled[100]; !ok {
		t.Errorf("expected task 100 to be scheduled for due alert")
	}
	if _, ok := sched.recurrenceScheduled[100]; !ok {
		t.Errorf("expected task 100 to be scheduled for recurrence")
	}
}

func TestTaskService_ProcessTaskRecurrence_AtomicScheduling(t *testing.T) {
	ctx := context.Background()
	masterID := int64(200)
	var timestampUpdated bool

	mockRepo := &mockRepository{
		findRecurringByIDSystemFunc: func(ctx context.Context, id int64) (*RecurringTaskResponse, string, error) {
			return &RecurringTaskResponse{
				ID:                 masterID,
				UserID:             "user-1",
				Title:              "Daily Recurring Standup",
				IsActive:           true,
				RecurrencePattern:  "daily",
				RecurrenceInterval: 1,
			}, "Asia/Jakarta", nil
		},
		findTodayInstanceFunc: func(ctx context.Context, recurringTaskID int64, todayDate time.Time) (*TaskResponse, error) {
			return nil, nil // No task for today yet
		},
		createInstanceFromRecurringFunc: func(ctx context.Context, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error) {
			return &TaskResponse{
				ID:              501,
				UserID:          master.UserID,
				Title:           master.Title,
				RecurringTaskID: &master.ID,
			}, nil
		},
		updateRecurringLastGeneratedFunc: func(ctx context.Context, id int64, tm time.Time) error {
			timestampUpdated = true
			return nil
		},
	}

	sched := newMockTaskScheduler()
	svc := NewService(mockRepo, &mockTransactor{})
	svc.SetNotificationDependencies(sched, nil, nil, nil, nil, nil)

	err := svc.ProcessTaskRecurrence(ctx, masterID)
	if err != nil {
		t.Fatalf("unexpected error in ProcessTaskRecurrence: %v", err)
	}

	if !timestampUpdated {
		t.Errorf("expected recurrence generated timestamp to be updated")
	}
	if _, ok := sched.recurrenceScheduled[masterID]; !ok {
		t.Errorf("expected master %d to have its next recurrence scheduled", masterID)
	}
}

func TestTaskService_ProcessTaskRecurrence_Idempotency(t *testing.T) {
	ctx := context.Background()
	masterID := int64(300)
	instanceCreated := false

	mockRepo := &mockRepository{
		findRecurringByIDSystemFunc: func(ctx context.Context, id int64) (*RecurringTaskResponse, string, error) {
			return &RecurringTaskResponse{
				ID:                 masterID,
				UserID:             "user-1",
				Title:              "Daily Sholat",
				IsActive:           true,
				RecurrencePattern:  "daily",
				RecurrenceInterval: 1,
			}, "Asia/Jakarta", nil
		},
		findTodayInstanceFunc: func(ctx context.Context, recurringTaskID int64, todayDate time.Time) (*TaskResponse, error) {
			// Today's instance already exists
			return &TaskResponse{
				ID:              128,
				UserID:          "user-1",
				Title:           "Daily Sholat",
				RecurringTaskID: &masterID,
			}, nil
		},
		createInstanceFromRecurringFunc: func(ctx context.Context, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error) {
			instanceCreated = true
			return nil, nil
		},
	}

	sched := newMockTaskScheduler()
	svc := NewService(mockRepo, &mockTransactor{})
	svc.SetNotificationDependencies(sched, nil, nil, nil, nil, nil)

	err := svc.ProcessTaskRecurrence(ctx, masterID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if instanceCreated {
		t.Errorf("expected no new instance to be created when today instance already exists")
	}
	if _, ok := sched.recurrenceScheduled[masterID]; !ok {
		t.Errorf("expected master %d to still have its next recurrence scheduled", masterID)
	}
}

func TestTaskService_AutoCompletionAndUnfocus(t *testing.T) {
	ctx := context.Background()
	taskID := int64(42)
	userID := "user-123"

	t.Run("all subtasks completed triggers parent task done and unfocus", func(t *testing.T) {
		var parentUpdatedWithDone bool
		var parentUpdatedWithUnfocus bool

		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, uID string, id int64) (*TaskResponse, error) {
				return &TaskResponse{
					ID:      taskID,
					UserID:  userID,
					Title:   "Master Task",
					Status:  "in_progress",
					IsFocus: true,
				}, nil
			},
			updateSubtaskFunc: func(ctx context.Context, uID string, tID int64, sID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error) {
				return &SubtaskResponse{
					ID:          sID,
					TaskID:      tID,
					Title:       "Subtask 2",
					IsCompleted: true,
				}, nil
			},
			createActivityFunc: func(ctx context.Context, tID int64, uID string, actType string, desc string, ch []byte) (*TaskActivityResponse, error) {
				return &TaskActivityResponse{ID: 1}, nil
			},
			countSubtasksFunc: func(ctx context.Context, tID int64) (int, int, error) {
				return 2, 2, nil // 2 of 2 completed
			},
			updateFunc: func(ctx context.Context, uID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
				if req.Status != nil && *req.Status == "done" {
					parentUpdatedWithDone = true
				}
				if req.IsFocus != nil && !*req.IsFocus {
					parentUpdatedWithUnfocus = true
				}
				return &TaskResponse{
					ID:      id,
					Status:  "done",
					IsFocus: false,
				}, nil
			},
			getUserThresholdFunc: func(ctx context.Context, uID string) (int, string, error) {
				return 30, "Asia/Jakarta", nil
			},
		}

		svc := NewService(mockRepo, &mockTransactor{})
		isComp := true
		res, err := svc.UpdateSubtask(ctx, userID, taskID, 2, UpdateSubtaskRequest{
			IsCompleted: &isComp,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || !res.IsCompleted {
			t.Errorf("expected subtask to be completed")
		}
		if !parentUpdatedWithDone {
			t.Errorf("expected parent task to be automatically updated with status 'done'")
		}
		if !parentUpdatedWithUnfocus {
			t.Errorf("expected parent task to be automatically unfocused")
		}
	})

	t.Run("subtask unchecked on done task reverts parent task to in_progress", func(t *testing.T) {
		var parentRevertedToInProgress bool

		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, uID string, id int64) (*TaskResponse, error) {
				return &TaskResponse{
					ID:     taskID,
					UserID: userID,
					Title:  "Master Task",
					Status: "done",
				}, nil
			},
			updateSubtaskFunc: func(ctx context.Context, uID string, tID int64, sID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error) {
				return &SubtaskResponse{
					ID:          sID,
					TaskID:      tID,
					Title:       "Subtask 1",
					IsCompleted: false,
				}, nil
			},
			createActivityFunc: func(ctx context.Context, tID int64, uID string, actType string, desc string, ch []byte) (*TaskActivityResponse, error) {
				return &TaskActivityResponse{ID: 2}, nil
			},
			countSubtasksFunc: func(ctx context.Context, tID int64) (int, int, error) {
				return 2, 1, nil // 1 of 2 completed
			},
			updateFunc: func(ctx context.Context, uID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
				if req.Status != nil && *req.Status == "in_progress" {
					parentRevertedToInProgress = true
				}
				return &TaskResponse{
					ID:     id,
					Status: "in_progress",
				}, nil
			},
			getUserThresholdFunc: func(ctx context.Context, uID string) (int, string, error) {
				return 30, "Asia/Jakarta", nil
			},
		}

		svc := NewService(mockRepo, &mockTransactor{})
		isComp := false
		_, err := svc.UpdateSubtask(ctx, userID, taskID, 1, UpdateSubtaskRequest{
			IsCompleted: &isComp,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !parentRevertedToInProgress {
			t.Errorf("expected parent task status to revert to 'in_progress'")
		}
	})
}

type mockTaskPushSender struct {
	sentCh chan map[string]string
}

func (m *mockTaskPushSender) SendPushNotification(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	if m.sentCh != nil {
		m.sentCh <- data
	}
	return nil
}

func (m *mockTaskPushSender) SendPushNotificationWithResult(ctx context.Context, tokens []string, title, body string, data map[string]string) (*notification.PushResult, error) {
	if m.sentCh != nil {
		m.sentCh <- data
	}
	return &notification.PushResult{SuccessCount: len(tokens), TotalTokens: len(tokens)}, nil
}

type mockTaskReminderRepo struct {
	reminder.Repository
	tokens    []string
	reminders []reminder.ReminderResponse
}

func (m *mockTaskReminderRepo) GetUserDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	return m.tokens, nil
}

func (m *mockTaskReminderRepo) DeleteDeviceTokens(ctx context.Context, tokens []string) error {
	return nil
}

func (m *mockTaskReminderRepo) FindAll(ctx context.Context, userID string, filter reminder.ReminderFilter) ([]reminder.ReminderResponse, int, error) {
	if filter.TaskID != nil {
		var matched []reminder.ReminderResponse
		for _, r := range m.reminders {
			if r.TaskID != nil && *r.TaskID == *filter.TaskID {
				matched = append(matched, r)
			}
		}
		return matched, len(matched), nil
	}
	return m.reminders, len(m.reminders), nil
}

func TestTaskService_DeadlineChangePushNotification(t *testing.T) {
	ctx := context.Background()
	userID := "user-push-test"
	taskID := int64(101)

	now := time.Now()
	newDue := now.Add(24 * time.Hour)

	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, uID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{
				ID:       id,
				UserID:   uID,
				Title:    "Push Deadline Task",
				Status:   "todo",
				Priority: "medium",
				DueDate:  nil,
			}, nil
		},
		updateFunc: func(ctx context.Context, uID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
			return &TaskResponse{
				ID:       id,
				UserID:   uID,
				Title:    "Push Deadline Task",
				Status:   "todo",
				Priority: "medium",
				DueDate:  req.DueDate,
			}, nil
		},
		createActivityFunc: func(ctx context.Context, tID int64, uID string, actType string, desc string, ch []byte) (*TaskActivityResponse, error) {
			return &TaskActivityResponse{ID: 1}, nil
		},
		getUserThresholdFunc: func(ctx context.Context, uID string) (int, string, error) {
			return 15, "Asia/Jakarta", nil
		},
	}

	pushCh := make(chan map[string]string, 1)
	pushSender := &mockTaskPushSender{sentCh: pushCh}
	reminderRepo := &mockTaskReminderRepo{tokens: []string{"fcm-device-token-abc"}}

	svc := NewService(mockRepo, &mockTransactor{})
	svc.SetNotificationDependencies(nil, nil, pushSender, reminderRepo, nil, nil)

	_, err := svc.Update(ctx, userID, taskID, UpdateTaskRequest{
		DueDate: &newDue,
	})
	if err != nil {
		t.Fatalf("unexpected error updating task: %v", err)
	}

	select {
	case payload := <-pushCh:
		if payload["type"] != "task_deadline_updated" {
			t.Errorf("expected type 'task_deadline_updated', got '%s'", payload["type"])
		}
		if payload["task_id"] != "101" {
			t.Errorf("expected task_id '101', got '%s'", payload["task_id"])
		}
	case <-time.After(2 * time.Second):
		t.Errorf("timed out waiting for push notification to be dispatched")
	}
}

func TestTaskService_Delete_CleansUpAsynqJobs(t *testing.T) {
	ctx := context.Background()
	userID := "user-del-test"
	taskID := int64(201)
	now := time.Now()
	due := now.Add(4 * time.Hour)
	isRecurring := true

	taskDeleted := false
	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, uID string, id int64) (*TaskResponse, error) {
			if id == taskID {
				return &TaskResponse{
					ID:          taskID,
					UserID:      uID,
					Title:       "Task to Delete",
					Status:      "in_progress",
					Priority:    "high",
					DueDate:     &due,
					IsRecurring: isRecurring,
				}, nil
			}
			return nil, errors.New("not found")
		},
		deleteFunc: func(ctx context.Context, uID string, id int64) error {
			if id == taskID {
				taskDeleted = true
				return nil
			}
			return errors.New("not found")
		},
	}

	rem1ID := int64(501)
	rem2ID := int64(502)
	reminderRepo := &mockTaskReminderRepo{
		reminders: []reminder.ReminderResponse{
			{ID: rem1ID, UserID: userID, TaskID: &taskID, Title: "Reminder 1"},
			{ID: rem2ID, UserID: userID, TaskID: &taskID, Title: "Reminder 2"},
		},
	}

	sched := newMockTaskScheduler()
	sched.dueScheduled[taskID] = due
	sched.recurrenceScheduled[taskID] = now.Add(24 * time.Hour)

	svc := NewService(mockRepo, &mockTransactor{})
	svc.SetNotificationDependencies(sched, nil, nil, reminderRepo, nil, nil)

	err := svc.Delete(ctx, userID, taskID)
	if err != nil {
		t.Fatalf("unexpected error deleting task: %v", err)
	}

	if !taskDeleted {
		t.Errorf("expected task to be deleted in repo")
	}

	// Verify task due alert was cancelled
	if !sched.dueCancelled[taskID] {
		t.Errorf("expected task due alert to be cancelled for task %d", taskID)
	}

	// Verify task recurrence was cancelled
	if !sched.recurrenceCancelled[taskID] {
		t.Errorf("expected task recurrence to be cancelled for task %d", taskID)
	}

	// Verify associated reminders were cancelled in Asynq
	if !sched.remindersCancelled[rem1ID] {
		t.Errorf("expected reminder %d to be cancelled in Asynq", rem1ID)
	}
	if !sched.remindersCancelled[rem2ID] {
		t.Errorf("expected reminder %d to be cancelled in Asynq", rem2ID)
	}
}

func TestTaskService_LogProgress_Done_CancelsDueAlert(t *testing.T) {
	ctx := context.Background()
	userID := "user-progress-test"
	taskID := int64(301)
	due := time.Now().Add(2 * time.Hour)

	statusDone := "done"
	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, uID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{
				ID:       taskID,
				UserID:   uID,
				Title:    "Task in progress",
				Status:   "in_progress",
				Priority: "medium",
				DueDate:  &due,
			}, nil
		},
		logProgressFunc: func(ctx context.Context, uID string, id int64, req LogProgressRequest) (*TaskResponse, error) {
			return &TaskResponse{
				ID:       taskID,
				UserID:   uID,
				Title:    "Task in progress",
				Status:   *req.Status,
				Priority: "medium",
				DueDate:  &due,
			}, nil
		},
		createActivityFunc: func(ctx context.Context, tID int64, uID string, actType string, desc string, ch []byte) (*TaskActivityResponse, error) {
			return &TaskActivityResponse{ID: 1}, nil
		},
	}

	sched := newMockTaskScheduler()
	sched.dueScheduled[taskID] = due

	svc := NewService(mockRepo, &mockTransactor{})
	svc.SetNotificationDependencies(sched, nil, nil, nil, nil, nil)

	_, err := svc.LogProgress(ctx, userID, taskID, LogProgressRequest{
		ProgressNote: "Done with task!",
		Status:       &statusDone,
	})
	if err != nil {
		t.Fatalf("unexpected error logging progress: %v", err)
	}

	if !sched.dueCancelled[taskID] {
		t.Errorf("expected task due alert to be cancelled when task is marked done")
	}
}

func TestTaskService_RecurringSubtasksSync(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"
	masterID := int64(10)
	taskID := int64(100)

	var (
		createdRecSubtaskTitle string
		updatedRecSubtaskTitle string
		deletedRecSubtaskTitle string
		syncedTask             bool
	)

	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{
				ID:              taskID,
				UserID:          userID,
				Title:           "Daily Task",
				RecurringTaskID: &masterID,
				IsRecurring:     true,
				Subtasks: []SubtaskResponse{
					{
						ID:     1,
						TaskID: taskID,
						Title:  "Old Subtask",
					},
				},
			}, nil
		},
		createSubtaskFunc: func(ctx context.Context, userID string, taskID int64, req CreateSubtaskRequest) (*SubtaskResponse, error) {
			return &SubtaskResponse{
				ID:         2,
				TaskID:     taskID,
				Title:      req.Title,
				OrderIndex: 1,
			}, nil
		},
		updateSubtaskFunc: func(ctx context.Context, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error) {
			title := "Updated Subtask"
			if req.Title != nil {
				title = *req.Title
			}
			return &SubtaskResponse{
				ID:     subtaskID,
				TaskID: taskID,
				Title:  title,
			}, nil
		},
		deleteSubtaskFunc: func(ctx context.Context, userID string, taskID int64, subtaskID int64) error {
			return nil
		},
		updateFunc: func(ctx context.Context, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
			return &TaskResponse{
				ID:              id,
				UserID:          userID,
				Title:           "Updated Daily Task",
				RecurringTaskID: &masterID,
				IsRecurring:     true,
			}, nil
		},
		createRecurringSubtaskFunc: func(ctx context.Context, recID int64, title, description string, orderIndex int) error {
			createdRecSubtaskTitle = title
			return nil
		},
		updateRecurringSubtaskByTitleFunc: func(ctx context.Context, recID int64, oldTitle, newTitle, description string) error {
			updatedRecSubtaskTitle = newTitle
			return nil
		},
		deleteRecurringSubtaskByTitleFunc: func(ctx context.Context, recID int64, title string) error {
			deletedRecSubtaskTitle = title
			return nil
		},
		syncRecurringFromTaskFunc: func(ctx context.Context, userID string, task *TaskResponse) error {
			syncedTask = true
			return nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})

	// 1. CreateSubtask on recurring task instance
	_, err := svc.CreateSubtask(ctx, testUserID, taskID, CreateSubtaskRequest{Title: "New Subtask"})
	if err != nil {
		t.Fatalf("CreateSubtask failed: %v", err)
	}
	if createdRecSubtaskTitle != "New Subtask" {
		t.Errorf("expected recurring subtask 'New Subtask', got '%s'", createdRecSubtaskTitle)
	}

	// 2. UpdateSubtask on recurring task instance
	newTitle := "Renamed Subtask"
	_, err = svc.UpdateSubtask(ctx, testUserID, taskID, 1, UpdateSubtaskRequest{Title: &newTitle})
	if err != nil {
		t.Fatalf("UpdateSubtask failed: %v", err)
	}
	if updatedRecSubtaskTitle != "Renamed Subtask" {
		t.Errorf("expected updated recurring subtask 'Renamed Subtask', got '%s'", updatedRecSubtaskTitle)
	}

	// 3. DeleteSubtask on recurring task instance
	err = svc.DeleteSubtask(ctx, testUserID, taskID, 1)
	if err != nil {
		t.Fatalf("DeleteSubtask failed: %v", err)
	}
	if deletedRecSubtaskTitle != "Old Subtask" {
		t.Errorf("expected deleted recurring subtask 'Old Subtask', got '%s'", deletedRecSubtaskTitle)
	}

	// 4. Update on recurring task instance
	updatedTitle := "Updated Daily Task"
	_, err = svc.Update(ctx, testUserID, taskID, UpdateTaskRequest{Title: &updatedTitle})
	if err != nil {
		t.Fatalf("Update task failed: %v", err)
	}
	if !syncedTask {
		t.Errorf("expected recurring task master to be synced")
	}
}

func TestTaskService_RecurringSubtasksSync_Errors(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-rec-err"
	taskID := int64(102)
	masterID := int64(502)

	expectedErr := errors.New("db error")

	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{
				ID:              taskID,
				UserID:          testUserID,
				Title:           "Daily Task",
				RecurringTaskID: &masterID,
				IsRecurring:     true,
				Subtasks: []SubtaskResponse{
					{ID: 1, TaskID: taskID, Title: "Existing Subtask"},
				},
			}, nil
		},
		createSubtaskFunc: func(ctx context.Context, userID string, taskID int64, req CreateSubtaskRequest) (*SubtaskResponse, error) {
			return &SubtaskResponse{ID: 2, TaskID: taskID, Title: req.Title}, nil
		},
		updateSubtaskFunc: func(ctx context.Context, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error) {
			return &SubtaskResponse{ID: subtaskID, TaskID: taskID, Title: "Renamed"}, nil
		},
		deleteSubtaskFunc: func(ctx context.Context, userID string, taskID int64, subtaskID int64) error {
			return nil
		},
		createRecurringSubtaskFunc: func(ctx context.Context, recID int64, title, description string, orderIndex int) error {
			return expectedErr
		},
		updateRecurringSubtaskByTitleFunc: func(ctx context.Context, recID int64, oldTitle, newTitle, description string) error {
			return expectedErr
		},
		deleteRecurringSubtaskByTitleFunc: func(ctx context.Context, recID int64, title string) error {
			return expectedErr
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})

	// 1. CreateSubtask error
	_, err := svc.CreateSubtask(ctx, testUserID, taskID, CreateSubtaskRequest{Title: "Fail Subtask"})
	if err == nil {
		t.Fatal("expected error on CreateSubtask when recurring subtask creation fails, got nil")
	}

	// 2. UpdateSubtask error
	newTitle := "Fail Rename"
	_, err = svc.UpdateSubtask(ctx, testUserID, taskID, 1, UpdateSubtaskRequest{Title: &newTitle})
	if err == nil {
		t.Fatal("expected error on UpdateSubtask when recurring subtask update fails, got nil")
	}

	// 3. DeleteSubtask error
	err = svc.DeleteSubtask(ctx, testUserID, taskID, 1)
	if err == nil {
		t.Fatal("expected error on DeleteSubtask when recurring subtask deletion fails, got nil")
	}
}

func TestTaskService_Update_ClearDueDate_CancelsAlert(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"
	taskID := int64(100)
	initialDue := time.Now().Add(2 * time.Hour)

	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{
				ID:       taskID,
				UserID:   userID,
				Title:    "Existing Task",
				DueDate:  &initialDue,
				Status:   "todo",
				Priority: "normal",
			}, nil
		},
		updateFunc: func(ctx context.Context, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
			var updatedDue *time.Time
			if !req.ClearDueDate {
				updatedDue = req.DueDate
			}
			return &TaskResponse{
				ID:       taskID,
				UserID:   userID,
				Title:    "Existing Task",
				DueDate:  updatedDue,
				Status:   "todo",
				Priority: "normal",
			}, nil
		},
		getUserThresholdFunc: func(ctx context.Context, userID string) (int, string, error) {
			return 30, "Asia/Jakarta", nil
		},
	}

	scheduler := newMockTaskScheduler()
	scheduler.dueScheduled[taskID] = initialDue.Add(-30 * time.Minute)

	svc := NewService(mockRepo, &mockTransactor{})
	svc.SetNotificationDependencies(scheduler, nil, nil, nil, nil, nil)

	res, err := svc.Update(ctx, testUserID, taskID, UpdateTaskRequest{
		ClearDueDate: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DueDate != nil {
		t.Fatalf("expected due date to be nil, got %v", res.DueDate)
	}
	if !scheduler.dueCancelled[taskID] {
		t.Fatalf("expected scheduler alert to be cancelled for task %d", taskID)
	}
}

func TestTaskService_Update_ConvertToRecurring_PreservesDueDate(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"
	taskID := int64(101)
	initialDue := time.Now().Add(2 * time.Hour)
	recMasterID := int64(999)

	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{
				ID:          taskID,
				UserID:      userID,
				Title:       "Existing Task",
				DueDate:     &initialDue,
				Status:      "todo",
				Priority:    "normal",
				IsRecurring: false,
			}, nil
		},
		updateFunc: func(ctx context.Context, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
			isRec := true
			return &TaskResponse{
				ID:          taskID,
				UserID:      userID,
				Title:       "Existing Task",
				DueDate:     &initialDue, // preserved!
				Status:      "todo",
				Priority:    "normal",
				IsRecurring: isRec,
			}, nil
		},
		convertTaskToRecurringFunc: func(ctx context.Context, userID string, tid int64, pattern string, interval int, endDate *time.Time) (*RecurringTaskResponse, error) {
			return &RecurringTaskResponse{
				ID:       recMasterID,
				UserID:   userID,
				Title:    "Existing Task",
				IsActive: true,
			}, nil
		},
		getUserThresholdFunc: func(ctx context.Context, userID string) (int, string, error) {
			return 30, "Asia/Jakarta", nil
		},
	}

	scheduler := newMockTaskScheduler()
	scheduler.dueScheduled[taskID] = initialDue.Add(-30 * time.Minute)

	isRecurring := true
	svc := NewService(mockRepo, &mockTransactor{})
	svc.SetNotificationDependencies(scheduler, nil, nil, nil, nil, nil)

	res, err := svc.Update(ctx, testUserID, taskID, UpdateTaskRequest{
		IsRecurring: &isRecurring,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsRecurring {
		t.Fatal("expected task to be recurring")
	}
	if res.DueDate == nil {
		t.Fatal("expected due date to be preserved, got nil")
	}
	if scheduler.dueCancelled[taskID] {
		t.Fatalf("scheduler alert should not be cancelled when due date is preserved")
	}
}

func TestRecurrencePattern_IsOccurrenceDay(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}

	// 2026-10-19 is Monday, 2026-10-23 is Friday, 2026-10-24 is Saturday, 2026-10-25 is Sunday
	monday := time.Date(2026, 10, 19, 10, 0, 0, 0, loc)
	tuesday := time.Date(2026, 10, 20, 10, 0, 0, 0, loc)
	wednesday := time.Date(2026, 10, 21, 10, 0, 0, 0, loc)
	thursday := time.Date(2026, 10, 22, 10, 0, 0, 0, loc)
	friday := time.Date(2026, 10, 23, 10, 0, 0, 0, loc)
	saturday := time.Date(2026, 10, 24, 10, 0, 0, 0, loc)
	sunday := time.Date(2026, 10, 25, 10, 0, 0, 0, loc)

	// Weekdays pattern
	t.Run("weekdays pattern skips Saturday and Sunday", func(t *testing.T) {
		if !isOccurrenceDay("weekdays", 1, monday, monday, loc) {
			t.Errorf("expected Monday to be occurrence day for weekdays")
		}
		if !isOccurrenceDay("weekdays", 1, tuesday, monday, loc) {
			t.Errorf("expected Tuesday to be occurrence day for weekdays")
		}
		if !isOccurrenceDay("weekdays", 1, wednesday, monday, loc) {
			t.Errorf("expected Wednesday to be occurrence day for weekdays")
		}
		if !isOccurrenceDay("weekdays", 1, thursday, monday, loc) {
			t.Errorf("expected Thursday to be occurrence day for weekdays")
		}
		if !isOccurrenceDay("weekdays", 1, friday, monday, loc) {
			t.Errorf("expected Friday to be occurrence day for weekdays")
		}
		if isOccurrenceDay("weekdays", 1, saturday, monday, loc) {
			t.Errorf("expected Saturday to NOT be occurrence day for weekdays")
		}
		if isOccurrenceDay("weekdays", 1, sunday, monday, loc) {
			t.Errorf("expected Sunday to NOT be occurrence day for weekdays")
		}
	})

	// Daily pattern
	t.Run("daily pattern with interval", func(t *testing.T) {
		if !isOccurrenceDay("daily", 1, saturday, monday, loc) {
			t.Errorf("expected Saturday to be occurrence day for daily interval 1")
		}
		// interval 2: Monday (0 days) = true, Tuesday (1 day) = false, Wednesday (2 days) = true
		if !isOccurrenceDay("daily", 2, monday, monday, loc) {
			t.Errorf("expected Monday to be occurrence day for daily interval 2")
		}
		if isOccurrenceDay("daily", 2, tuesday, monday, loc) {
			t.Errorf("expected Tuesday to NOT be occurrence day for daily interval 2")
		}
		if !isOccurrenceDay("daily", 2, wednesday, monday, loc) {
			t.Errorf("expected Wednesday to be occurrence day for daily interval 2")
		}
	})

	// Weekly pattern
	t.Run("weekly pattern matches weekday", func(t *testing.T) {
		nextMonday := monday.AddDate(0, 0, 7)
		if !isOccurrenceDay("weekly", 1, nextMonday, monday, loc) {
			t.Errorf("expected next Monday to be occurrence day for weekly")
		}
		if isOccurrenceDay("weekly", 1, tuesday, monday, loc) {
			t.Errorf("expected Tuesday to NOT be occurrence day for weekly anchored on Monday")
		}
	})
}

func TestRecurrencePattern_GetNextOccurrenceMidnight(t *testing.T) {
	tz := "Asia/Jakarta"
	loc, _ := time.LoadLocation(tz)

	friday := time.Date(2026, 10, 23, 10, 0, 0, 0, loc)
	saturday := time.Date(2026, 10, 24, 10, 0, 0, 0, loc)
	sunday := time.Date(2026, 10, 25, 10, 0, 0, 0, loc)
	monday := time.Date(2026, 10, 26, 10, 0, 0, 0, loc)
	thursday := time.Date(2026, 10, 22, 10, 0, 0, 0, loc)

	expectedMondayMidnight := time.Date(2026, 10, 26, 0, 0, 0, 0, loc).UTC()
	expectedFridayMidnight := time.Date(2026, 10, 23, 0, 0, 0, 0, loc).UTC()
	expectedTuesdayMidnight := time.Date(2026, 10, 27, 0, 0, 0, 0, loc).UTC()

	// Friday -> Monday midnight
	nextFromFri := getNextOccurrenceMidnight("weekdays", 1, friday, tz, friday)
	if !nextFromFri.Equal(expectedMondayMidnight) {
		t.Errorf("expected Friday next occurrence to be Monday midnight %v, got %v", expectedMondayMidnight, nextFromFri)
	}

	// Saturday -> Monday midnight
	nextFromSat := getNextOccurrenceMidnight("weekdays", 1, saturday, tz, friday)
	if !nextFromSat.Equal(expectedMondayMidnight) {
		t.Errorf("expected Saturday next occurrence to be Monday midnight %v, got %v", expectedMondayMidnight, nextFromSat)
	}

	// Sunday -> Monday midnight
	nextFromSun := getNextOccurrenceMidnight("weekdays", 1, sunday, tz, friday)
	if !nextFromSun.Equal(expectedMondayMidnight) {
		t.Errorf("expected Sunday next occurrence to be Monday midnight %v, got %v", expectedMondayMidnight, nextFromSun)
	}

	// Monday -> Tuesday midnight
	nextFromMon := getNextOccurrenceMidnight("weekdays", 1, monday, tz, friday)
	if !nextFromMon.Equal(expectedTuesdayMidnight) {
		t.Errorf("expected Monday next occurrence to be Tuesday midnight %v, got %v", expectedTuesdayMidnight, nextFromMon)
	}

	// Thursday -> Friday midnight
	nextFromThu := getNextOccurrenceMidnight("weekdays", 1, thursday, tz, friday)
	if !nextFromThu.Equal(expectedFridayMidnight) {
		t.Errorf("expected Thursday next occurrence to be Friday midnight %v, got %v", expectedFridayMidnight, nextFromThu)
	}
}

func TestTaskService_ProcessTaskRecurrence_WeekdaysWeekendSkipped(t *testing.T) {
	ctx := context.Background()
	masterID := int64(888)
	instanceCreated := false

	loc, _ := time.LoadLocation("Asia/Jakarta")
	saturdayTime := time.Date(2026, 10, 24, 0, 0, 2, 0, loc) // Saturday midnight
	expectedMondayMidnight := time.Date(2026, 10, 26, 0, 0, 0, 0, loc).UTC()

	mockRepo := &mockRepository{
		findRecurringByIDSystemFunc: func(ctx context.Context, id int64) (*RecurringTaskResponse, string, error) {
			return &RecurringTaskResponse{
				ID:                 masterID,
				UserID:             "user-1",
				Title:              "Absen",
				IsActive:           true,
				RecurrencePattern:  "weekdays",
				RecurrenceInterval: 1,
				CreatedAt:          time.Date(2026, 10, 21, 10, 0, 0, 0, loc),
			}, "Asia/Jakarta", nil
		},
		findTodayInstanceFunc: func(ctx context.Context, recurringTaskID int64, todayDate time.Time) (*TaskResponse, error) {
			return nil, nil
		},
		createInstanceFromRecurringFunc: func(ctx context.Context, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error) {
			instanceCreated = true
			return &TaskResponse{ID: 999, Title: master.Title}, nil
		},
	}

	sched := newMockTaskScheduler()
	svc := NewService(mockRepo, &mockTransactor{}).(*service)
	svc.SetNotificationDependencies(sched, nil, nil, nil, nil, nil)
	svc.nowFunc = func() time.Time {
		return saturdayTime
	}

	err := svc.ProcessTaskRecurrence(ctx, masterID)
	if err != nil {
		t.Fatalf("unexpected error in ProcessTaskRecurrence: %v", err)
	}

	// Crucial assertion: no instance should ever be created on Saturday for weekdays!
	if instanceCreated {
		t.Errorf("expected NO instance to be created on Saturday for 'weekdays' pattern")
	}

	// Must schedule next recurrence directly to Monday midnight
	scheduledAt, ok := sched.recurrenceScheduled[masterID]
	if !ok {
		t.Fatalf("expected next recurrence to be scheduled for master %d", masterID)
	}
	if !scheduledAt.Equal(expectedMondayMidnight) {
		t.Errorf("expected scheduled time to be Monday midnight %v, got %v", expectedMondayMidnight, scheduledAt)
	}
}

func TestTaskService_ProcessTaskRecurrence_WeekdaysFridayGeneratesAndSchedulesMonday(t *testing.T) {
	ctx := context.Background()
	masterID := int64(889)
	instanceCreated := false
	var createdDueDate time.Time

	loc, _ := time.LoadLocation("Asia/Jakarta")
	fridayTime := time.Date(2026, 10, 23, 0, 0, 2, 0, loc) // Friday midnight
	expectedMondayMidnight := time.Date(2026, 10, 26, 0, 0, 0, 0, loc).UTC()

	mockRepo := &mockRepository{
		findRecurringByIDSystemFunc: func(ctx context.Context, id int64) (*RecurringTaskResponse, string, error) {
			return &RecurringTaskResponse{
				ID:                 masterID,
				UserID:             "user-1",
				Title:              "Absen",
				IsActive:           true,
				RecurrencePattern:  "weekdays",
				RecurrenceInterval: 1,
				CreatedAt:          time.Date(2026, 10, 21, 10, 0, 0, 0, loc),
			}, "Asia/Jakarta", nil
		},
		findTodayInstanceFunc: func(ctx context.Context, recurringTaskID int64, todayDate time.Time) (*TaskResponse, error) {
			return nil, nil
		},
		createInstanceFromRecurringFunc: func(ctx context.Context, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error) {
			instanceCreated = true
			createdDueDate = dueDate
			return &TaskResponse{ID: 1000, Title: master.Title}, nil
		},
		updateRecurringLastGeneratedFunc: func(ctx context.Context, id int64, tm time.Time) error {
			return nil
		},
	}

	sched := newMockTaskScheduler()
	svc := NewService(mockRepo, &mockTransactor{}).(*service)
	svc.SetNotificationDependencies(sched, nil, nil, nil, nil, nil)
	svc.nowFunc = func() time.Time {
		return fridayTime
	}

	err := svc.ProcessTaskRecurrence(ctx, masterID)
	if err != nil {
		t.Fatalf("unexpected error in ProcessTaskRecurrence: %v", err)
	}

	// Instance MUST be created on Friday
	if !instanceCreated {
		t.Errorf("expected instance to be created on Friday for 'weekdays' pattern")
	}
	if createdDueDate.Day() != 23 {
		t.Errorf("expected due date day 23, got %d", createdDueDate.Day())
	}

	// Next occurrence MUST skip Saturday and Sunday and be Monday midnight
	scheduledAt, ok := sched.recurrenceScheduled[masterID]
	if !ok {
		t.Fatalf("expected next recurrence to be scheduled for master %d", masterID)
	}
	if !scheduledAt.Equal(expectedMondayMidnight) {
		t.Errorf("expected scheduled time to be Monday midnight %v, got %v", expectedMondayMidnight, scheduledAt)
	}
}

func TestTaskService_CreateRecurringTask_WeekendSkipsInitialInstance(t *testing.T) {
	ctx := context.Background()
	userID := "user-1"
	loc, _ := time.LoadLocation("Asia/Jakarta")
	saturdayTime := time.Date(2026, 10, 24, 15, 30, 0, 0, loc) // Saturday afternoon
	expectedMondayMidnight := time.Date(2026, 10, 26, 0, 0, 0, 0, loc).UTC()

	instanceCreated := false
	mockRepo := &mockRepository{
		createRecurringFunc: func(ctx context.Context, uid string, req CreateRecurringTaskRequest) (*RecurringTaskResponse, error) {
			return &RecurringTaskResponse{
				ID:                 777,
				UserID:             uid,
				Title:              req.Title,
				RecurrencePattern:  req.RecurrencePattern,
				RecurrenceInterval: req.RecurrenceInterval,
				CreatedAt:          saturdayTime,
			}, nil
		},
		getUserThresholdFunc: func(ctx context.Context, uid string) (int, string, error) {
			return 30, "Asia/Jakarta", nil
		},
		createInstanceFromRecurringFunc: func(ctx context.Context, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error) {
			instanceCreated = true
			return &TaskResponse{ID: 1001, Title: master.Title}, nil
		},
	}

	sched := newMockTaskScheduler()
	svc := NewService(mockRepo, &mockTransactor{}).(*service)
	svc.SetNotificationDependencies(sched, nil, nil, nil, nil, nil)
	svc.nowFunc = func() time.Time {
		return saturdayTime
	}

	res, err := svc.CreateRecurringTask(ctx, userID, CreateRecurringTaskRequest{
		Title:              "Absen",
		RecurrencePattern:  "weekdays",
		RecurrenceInterval: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("expected recurring task response, got nil")
	}

	// Should not generate instance on Saturday
	if instanceCreated {
		t.Errorf("expected NO initial instance created on Saturday for 'weekdays' pattern")
	}

	// Asynq should be scheduled for Monday midnight
	scheduledAt, ok := sched.recurrenceScheduled[res.ID]
	if !ok {
		t.Fatalf("expected recurrence to be scheduled for master %d", res.ID)
	}
	if !scheduledAt.Equal(expectedMondayMidnight) {
		t.Errorf("expected scheduled time to be Monday midnight %v, got %v", expectedMondayMidnight, scheduledAt)
	}
}

func TestRecurrencePattern_MonthlyEdgeCases(t *testing.T) {
	tz := "Asia/Jakarta"
	loc, err := time.LoadLocation(tz)
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}

	t.Run("month end clamping for 31st in non-leap February", func(t *testing.T) {
		// Anchored on Jan 31, 2027 (non-leap year)
		jan31 := time.Date(2027, 1, 31, 10, 0, 0, 0, loc)
		feb27 := time.Date(2027, 2, 27, 10, 0, 0, 0, loc)
		feb28 := time.Date(2027, 2, 28, 10, 0, 0, 0, loc)
		mar30 := time.Date(2027, 3, 30, 10, 0, 0, 0, loc)
		mar31 := time.Date(2027, 3, 31, 10, 0, 0, 0, loc)
		apr29 := time.Date(2027, 4, 29, 10, 0, 0, 0, loc)
		apr30 := time.Date(2027, 4, 30, 10, 0, 0, 0, loc)

		if isOccurrenceDay("monthly", 1, feb27, jan31, loc) {
			t.Errorf("expected Feb 27 to NOT be occurrence day for Jan 31 anchor")
		}
		if !isOccurrenceDay("monthly", 1, feb28, jan31, loc) {
			t.Errorf("expected Feb 28 (last day of Feb 2027) to be occurrence day for Jan 31 anchor")
		}

		if isOccurrenceDay("monthly", 1, mar30, jan31, loc) {
			t.Errorf("expected March 30 to NOT be occurrence day")
		}
		if !isOccurrenceDay("monthly", 1, mar31, jan31, loc) {
			t.Errorf("expected March 31 to be occurrence day")
		}

		if isOccurrenceDay("monthly", 1, apr29, jan31, loc) {
			t.Errorf("expected April 29 to NOT be occurrence day")
		}
		if !isOccurrenceDay("monthly", 1, apr30, jan31, loc) {
			t.Errorf("expected April 30 (last day of April) to be occurrence day for Jan 31 anchor")
		}

		// getNextOccurrenceMidnight from Jan 31 must clamp to Feb 28 midnight
		expectedFeb28Midnight := time.Date(2027, 2, 28, 0, 0, 0, 0, loc).UTC()
		nextMidnight := getNextOccurrenceMidnight("monthly", 1, jan31, tz, jan31)
		if !nextMidnight.Equal(expectedFeb28Midnight) {
			t.Errorf("expected Jan 31 next occurrence to be Feb 28 %v, got %v", expectedFeb28Midnight, nextMidnight)
		}
	})

	t.Run("month end clamping for 31st in leap year February 2028", func(t *testing.T) {
		// Anchored on Jan 31, 2028 (leap year)
		jan31Leap := time.Date(2028, 1, 31, 10, 0, 0, 0, loc)
		feb28Leap := time.Date(2028, 2, 28, 10, 0, 0, 0, loc)
		feb29Leap := time.Date(2028, 2, 29, 10, 0, 0, 0, loc)

		if isOccurrenceDay("monthly", 1, feb28Leap, jan31Leap, loc) {
			t.Errorf("expected Feb 28 in leap year to NOT be occurrence day for Jan 31 anchor (Feb 29 is last day)")
		}
		if !isOccurrenceDay("monthly", 1, feb29Leap, jan31Leap, loc) {
			t.Errorf("expected Feb 29 in leap year to be occurrence day for Jan 31 anchor")
		}

		expectedFeb29Midnight := time.Date(2028, 2, 29, 0, 0, 0, 0, loc).UTC()
		nextMidnight := getNextOccurrenceMidnight("monthly", 1, jan31Leap, tz, jan31Leap)
		if !nextMidnight.Equal(expectedFeb29Midnight) {
			t.Errorf("expected Jan 31 2028 next occurrence to be Feb 29 %v, got %v", expectedFeb29Midnight, nextMidnight)
		}
	})

	t.Run("year boundary rollover December to January", func(t *testing.T) {
		dec31 := time.Date(2026, 12, 31, 14, 0, 0, 0, loc)
		expectedJan31Midnight := time.Date(2027, 1, 31, 0, 0, 0, 0, loc).UTC()

		nextMidnight := getNextOccurrenceMidnight("monthly", 1, dec31, tz, dec31)
		if !nextMidnight.Equal(expectedJan31Midnight) {
			t.Errorf("expected Dec 31 next occurrence to be Jan 31 %v, got %v", expectedJan31Midnight, nextMidnight)
		}
	})

	t.Run("monthly with interval greater than 1", func(t *testing.T) {
		jan15 := time.Date(2026, 1, 15, 9, 0, 0, 0, loc)
		feb15 := time.Date(2026, 2, 15, 9, 0, 0, 0, loc)
		mar15 := time.Date(2026, 3, 15, 9, 0, 0, 0, loc)

		// Interval = 2: Jan 15 (true), Feb 15 (false), Mar 15 (true)
		if !isOccurrenceDay("monthly", 2, jan15, jan15, loc) {
			t.Errorf("expected Jan 15 to be occurrence day for interval 2")
		}
		if isOccurrenceDay("monthly", 2, feb15, jan15, loc) {
			t.Errorf("expected Feb 15 to NOT be occurrence day for interval 2")
		}
		if !isOccurrenceDay("monthly", 2, mar15, jan15, loc) {
			t.Errorf("expected Mar 15 to be occurrence day for interval 2")
		}

		expectedMar15Midnight := time.Date(2026, 3, 15, 0, 0, 0, 0, loc).UTC()
		nextMidnight := getNextOccurrenceMidnight("monthly", 2, jan15, tz, jan15)
		if !nextMidnight.Equal(expectedMar15Midnight) {
			t.Errorf("expected Jan 15 interval 2 next occurrence to be Mar 15 %v, got %v", expectedMar15Midnight, nextMidnight)
		}
	})
}

func TestRecurrencePattern_WeeklyIntervalEdgeCases(t *testing.T) {
	tz := "Asia/Jakarta"
	loc, _ := time.LoadLocation(tz)

	// Wednesday Oct 21, 2026
	wedWeek1 := time.Date(2026, 10, 21, 10, 0, 0, 0, loc)
	wedWeek2 := time.Date(2026, 10, 28, 10, 0, 0, 0, loc)
	wedWeek3 := time.Date(2026, 11, 4, 10, 0, 0, 0, loc)
	monWeek2 := time.Date(2026, 10, 26, 10, 0, 0, 0, loc)

	t.Run("bi-weekly interval 2 occurrences", func(t *testing.T) {
		if !isOccurrenceDay("weekly", 2, wedWeek1, wedWeek1, loc) {
			t.Errorf("expected Week 1 Wednesday to be occurrence day")
		}
		if isOccurrenceDay("weekly", 2, wedWeek2, wedWeek1, loc) {
			t.Errorf("expected Week 2 Wednesday to NOT be occurrence day for interval 2")
		}
		if !isOccurrenceDay("weekly", 2, wedWeek3, wedWeek1, loc) {
			t.Errorf("expected Week 3 Wednesday to be occurrence day for interval 2")
		}
		if isOccurrenceDay("weekly", 2, monWeek2, wedWeek1, loc) {
			t.Errorf("expected Monday to NOT be occurrence day for Wednesday weekly")
		}
	})

	t.Run("bi-weekly next occurrence scheduling", func(t *testing.T) {
		expectedWedWeek3Midnight := time.Date(2026, 11, 4, 0, 0, 0, 0, loc).UTC()

		// Scheduling from Week 1 Wed -> Week 3 Wed
		nextFromWeek1 := getNextOccurrenceMidnight("weekly", 2, wedWeek1, tz, wedWeek1)
		if !nextFromWeek1.Equal(expectedWedWeek3Midnight) {
			t.Errorf("expected Week 1 Wed next occurrence to be Week 3 Wed %v, got %v", expectedWedWeek3Midnight, nextFromWeek1)
		}

		// Scheduling from off-day (Week 2 Monday) -> Week 3 Wed
		nextFromWeek2Mon := getNextOccurrenceMidnight("weekly", 2, monWeek2, tz, wedWeek1)
		if !nextFromWeek2Mon.Equal(expectedWedWeek3Midnight) {
			t.Errorf("expected Week 2 Mon next occurrence to be Week 3 Wed %v, got %v", expectedWedWeek3Midnight, nextFromWeek2Mon)
		}
	})
}

func TestRecurrencePattern_DailyIntervalEdgeCases(t *testing.T) {
	tz := "Asia/Jakarta"
	loc, _ := time.LoadLocation(tz)

	day0 := time.Date(2026, 10, 1, 10, 0, 0, 0, loc)
	day1 := time.Date(2026, 10, 2, 10, 0, 0, 0, loc)
	day2 := time.Date(2026, 10, 3, 10, 0, 0, 0, loc)
	day3 := time.Date(2026, 10, 4, 10, 0, 0, 0, loc)
	day6 := time.Date(2026, 10, 7, 10, 0, 0, 0, loc)

	t.Run("daily interval 3 occurrence check", func(t *testing.T) {
		if !isOccurrenceDay("daily", 3, day0, day0, loc) {
			t.Errorf("expected Day 0 to be occurrence day")
		}
		if isOccurrenceDay("daily", 3, day1, day0, loc) {
			t.Errorf("expected Day 1 to NOT be occurrence day for interval 3")
		}
		if isOccurrenceDay("daily", 3, day2, day0, loc) {
			t.Errorf("expected Day 2 to NOT be occurrence day for interval 3")
		}
		if !isOccurrenceDay("daily", 3, day3, day0, loc) {
			t.Errorf("expected Day 3 to be occurrence day for interval 3")
		}
		if !isOccurrenceDay("daily", 3, day6, day0, loc) {
			t.Errorf("expected Day 6 to be occurrence day for interval 3")
		}
	})

	t.Run("daily interval 3 scheduling", func(t *testing.T) {
		expectedDay3Midnight := time.Date(2026, 10, 4, 0, 0, 0, 0, loc).UTC()
		nextMidnight := getNextOccurrenceMidnight("daily", 3, day0, tz, day0)
		if !nextMidnight.Equal(expectedDay3Midnight) {
			t.Errorf("expected Day 0 next occurrence to be Day 3 %v, got %v", expectedDay3Midnight, nextMidnight)
		}
	})

	t.Run("fallback resilience on zero or negative interval", func(t *testing.T) {
		if !isOccurrenceDay("daily", 0, day1, day0, loc) {
			t.Errorf("expected interval 0 to safely default to 1")
		}
		if !isOccurrenceDay("daily", -5, day1, day0, loc) {
			t.Errorf("expected interval -5 to safely default to 1")
		}

		expectedDay1Midnight := time.Date(2026, 10, 2, 0, 0, 0, 0, loc).UTC()
		nextMidnight := getNextOccurrenceMidnight("daily", 0, day0, tz, day0)
		if !nextMidnight.Equal(expectedDay1Midnight) {
			t.Errorf("expected interval 0 to default to +1 day %v, got %v", expectedDay1Midnight, nextMidnight)
		}
	})
}

func TestTaskService_ProcessTaskRecurrence_EndDateEdgeCases(t *testing.T) {
	ctx := context.Background()
	masterID := int64(777)
	loc, _ := time.LoadLocation("Asia/Jakarta")

	t.Run("deactivates when current time is after RecurrenceEndDate", func(t *testing.T) {
		endDate := time.Date(2026, 10, 20, 23, 59, 59, 0, loc)
		nowTime := time.Date(2026, 10, 21, 0, 0, 2, 0, loc) // 1 day after end date

		toggledDeactivated := false
		cancelledRecurrence := false
		instanceCreated := false

		mockRepo := &mockRepository{
			findRecurringByIDSystemFunc: func(ctx context.Context, id int64) (*RecurringTaskResponse, string, error) {
				return &RecurringTaskResponse{
					ID:                 masterID,
					UserID:             "user-1",
					Title:              "Expiring Task",
					IsActive:           true,
					RecurrencePattern:  "daily",
					RecurrenceInterval: 1,
					RecurrenceEndDate:  &endDate,
					CreatedAt:          time.Date(2026, 10, 1, 0, 0, 0, 0, loc),
				}, "Asia/Jakarta", nil
			},
			toggleActiveRecurringFunc: func(ctx context.Context, uid string, id int64) (*RecurringTaskResponse, error) {
				toggledDeactivated = true
				return &RecurringTaskResponse{ID: id, IsActive: false}, nil
			},
			createInstanceFromRecurringFunc: func(ctx context.Context, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error) {
				instanceCreated = true
				return &TaskResponse{ID: 100}, nil
			},
		}

		sched := newMockTaskScheduler()
		svc := NewService(mockRepo, &mockTransactor{}).(*service)
		svc.SetNotificationDependencies(sched, nil, nil, nil, nil, nil)
		svc.nowFunc = func() time.Time {
			return nowTime
		}

		err := svc.ProcessTaskRecurrence(ctx, masterID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !toggledDeactivated {
			t.Errorf("expected master task to be toggled deactivated")
		}
		if instanceCreated {
			t.Errorf("expected NO task instance to be created when after end date")
		}
		if !sched.recurrenceCancelled[masterID] {
			cancelledRecurrence = true
		}
		if cancelledRecurrence {
			t.Errorf("expected recurrence to be cancelled in scheduler")
		}
	})

	t.Run("creates last instance and does not schedule beyond end date", func(t *testing.T) {
		// End date is today at 23:59:59
		endDate := time.Date(2026, 10, 20, 23, 59, 59, 0, loc)
		nowTime := time.Date(2026, 10, 20, 0, 0, 2, 0, loc) // morning of end date

		instanceCreated := false
		mockRepo := &mockRepository{
			findRecurringByIDSystemFunc: func(ctx context.Context, id int64) (*RecurringTaskResponse, string, error) {
				return &RecurringTaskResponse{
					ID:                 masterID,
					UserID:             "user-1",
					Title:              "Last Day Task",
					IsActive:           true,
					RecurrencePattern:  "daily",
					RecurrenceInterval: 1,
					RecurrenceEndDate:  &endDate,
					CreatedAt:          time.Date(2026, 10, 1, 0, 0, 0, 0, loc),
				}, "Asia/Jakarta", nil
			},
			findTodayInstanceFunc: func(ctx context.Context, recurringTaskID int64, todayDate time.Time) (*TaskResponse, error) {
				return nil, nil
			},
			createInstanceFromRecurringFunc: func(ctx context.Context, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error) {
				instanceCreated = true
				return &TaskResponse{ID: 200, Title: master.Title}, nil
			},
			updateRecurringLastGeneratedFunc: func(ctx context.Context, id int64, tm time.Time) error {
				return nil
			},
		}

		sched := newMockTaskScheduler()
		svc := NewService(mockRepo, &mockTransactor{}).(*service)
		svc.SetNotificationDependencies(sched, nil, nil, nil, nil, nil)
		svc.nowFunc = func() time.Time {
			return nowTime
		}

		err := svc.ProcessTaskRecurrence(ctx, masterID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Instance MUST be created for today
		if !instanceCreated {
			t.Errorf("expected instance to be created on the final valid day")
		}

		// Crucial assertion: Next recurrence (tomorrow) is beyond end date, so should NOT be scheduled!
		if _, ok := sched.recurrenceScheduled[masterID]; ok {
			t.Errorf("expected NO new recurrence scheduled because tomorrow exceeds RecurrenceEndDate")
		}
	})
}

func TestTaskService_UpdateTask_SyncToRecurringMasterEdgeCases(t *testing.T) {
	ctx := context.Background()
	taskID := int64(101)
	masterID := int64(505)
	testUserID := "user-sync-test"
	loc, _ := time.LoadLocation("Asia/Jakarta")

	t.Run("switching recurrence pattern from daily to weekdays reschedules to Monday on weekend", func(t *testing.T) {
		fridayTime := time.Date(2026, 10, 23, 15, 0, 0, 0, loc)
		expectedMondayMidnight := time.Date(2026, 10, 26, 0, 0, 0, 0, loc).UTC()

		synced := false
		newPattern := "weekdays"
		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, uid string, id int64) (*TaskResponse, error) {
				return &TaskResponse{
					ID:                 taskID,
					UserID:             uid,
					Title:              "Standup",
					IsRecurring:        true,
					RecurringTaskID:    &masterID,
					RecurrencePattern:  "daily",
					RecurrenceInterval: 1,
					CreatedAt:          fridayTime,
				}, nil
			},
			updateFunc: func(ctx context.Context, uid string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
				return &TaskResponse{
					ID:                 taskID,
					UserID:             uid,
					Title:              "Standup",
					IsRecurring:        true,
					RecurringTaskID:    &masterID,
					RecurrencePattern:  newPattern,
					RecurrenceInterval: 1,
					CreatedAt:          fridayTime,
				}, nil
			},
			syncRecurringFromTaskFunc: func(ctx context.Context, uid string, task *TaskResponse) error {
				synced = true
				return nil
			},
			getUserThresholdFunc: func(ctx context.Context, uid string) (int, string, error) {
				return 30, "Asia/Jakarta", nil
			},
		}

		sched := newMockTaskScheduler()
		svc := NewService(mockRepo, &mockTransactor{}).(*service)
		svc.SetNotificationDependencies(sched, nil, nil, nil, nil, nil)
		svc.nowFunc = func() time.Time {
			return fridayTime
		}

		res, err := svc.Update(ctx, testUserID, taskID, UpdateTaskRequest{
			RecurrencePattern: &newPattern,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || res.RecurrencePattern != "weekdays" {
			t.Fatalf("expected updated pattern to be weekdays")
		}
		if !synced {
			t.Errorf("expected recurring template to be synced from task")
		}

		// Asynq must be rescheduled to Monday midnight
		scheduledAt, ok := sched.recurrenceScheduled[masterID]
		if !ok {
			t.Fatalf("expected recurrence to be rescheduled for master %d", masterID)
		}
		if !scheduledAt.Equal(expectedMondayMidnight) {
			t.Errorf("expected scheduled time to be Monday midnight %v, got %v", expectedMondayMidnight, scheduledAt)
		}
	})

	t.Run("turning off is_recurring deactivates master and cancels Asynq schedule", func(t *testing.T) {
		isRecurring := false
		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, uid string, id int64) (*TaskResponse, error) {
				return &TaskResponse{
					ID:              taskID,
					UserID:          uid,
					Title:           "Weekly Meeting",
					IsRecurring:     true,
					RecurringTaskID: &masterID,
				}, nil
			},
			updateFunc: func(ctx context.Context, uid string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
				return &TaskResponse{
					ID:              taskID,
					UserID:          uid,
					Title:           "Weekly Meeting",
					IsRecurring:     false,
					RecurringTaskID: nil,
				}, nil
			},
			getUserThresholdFunc: func(ctx context.Context, uid string) (int, string, error) {
				return 30, "Asia/Jakarta", nil
			},
		}

		sched := newMockTaskScheduler()
		svc := NewService(mockRepo, &mockTransactor{}).(*service)
		svc.SetNotificationDependencies(sched, nil, nil, nil, nil, nil)

		res, err := svc.Update(ctx, testUserID, taskID, UpdateTaskRequest{
			IsRecurring: &isRecurring,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsRecurring {
			t.Errorf("expected is_recurring to be false")
		}
		if res.RecurringTaskID != nil {
			t.Errorf("expected recurring_task_id to be nil")
		}
		if !sched.recurrenceCancelled[masterID] {
			t.Errorf("expected Asynq task recurrence to be cancelled for master %d", masterID)
		}
	})
}



