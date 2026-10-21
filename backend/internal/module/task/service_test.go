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
	createInstanceFromRecurringFunc  func(ctx context.Context, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error)
	updateRecurringLastGeneratedFunc func(ctx context.Context, id int64, t time.Time) error
}

func (m *mockRepository) Create(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *mockRepository) FindByID(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *mockRepository) FindAll(ctx context.Context, userID string, filter TaskFilter) ([]TaskResponse, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, userID, filter)
	}
	return []TaskResponse{}, nil
}

func (m *mockRepository) Update(ctx context.Context, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
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

func (m *mockRepository) LogProgress(ctx context.Context, userID string, id int64, req LogProgressRequest) (*TaskResponse, error) {
	if m.logProgressFunc != nil {
		return m.logProgressFunc(ctx, userID, id, req)
	}
	return nil, nil
}

func (m *mockRepository) ToggleFocus(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
	if m.toggleFocusFunc != nil {
		return m.toggleFocusFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *mockRepository) CreateSubtask(ctx context.Context, userID string, taskID int64, req CreateSubtaskRequest) (*SubtaskResponse, error) {
	if m.createSubtaskFunc != nil {
		return m.createSubtaskFunc(ctx, userID, taskID, req)
	}
	return nil, nil
}

func (m *mockRepository) GetWidgetBundle(ctx context.Context, userID string) (*WidgetBundleResponse, error) {
	if m.getWidgetBundleFunc != nil {
		return m.getWidgetBundleFunc(ctx, userID)
	}
	return &WidgetBundleResponse{}, nil
}

func (m *mockRepository) GetWorkspaceStats(ctx context.Context, userID string) (*WorkspaceStatsResponse, error) {
	if m.getWorkspaceStatsFunc != nil {
		return m.getWorkspaceStatsFunc(ctx, userID)
	}
	return &WorkspaceStatsResponse{}, nil
}

func (m *mockRepository) UpdateRecurrenceGeneratedAt(ctx context.Context, id int64, t time.Time) error {
	if m.updateRecurrenceGeneratedAtFunc != nil {
		return m.updateRecurrenceGeneratedAtFunc(ctx, id, t)
	}
	return nil
}

func (m *mockRepository) UpdateSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error) {
	if m.updateSubtaskFunc != nil {
		return m.updateSubtaskFunc(ctx, userID, taskID, subtaskID, req)
	}
	return nil, nil
}

func (m *mockRepository) DeleteSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64) error {
	if m.deleteSubtaskFunc != nil {
		return m.deleteSubtaskFunc(ctx, userID, taskID, subtaskID)
	}
	return nil
}

func (m *mockRepository) CountSubtasks(ctx context.Context, taskID int64) (int, int, error) {
	if m.countSubtasksFunc != nil {
		return m.countSubtasksFunc(ctx, taskID)
	}
	return 0, 0, nil
}

func (m *mockRepository) FindByIDSystem(ctx context.Context, id int64) (*TaskResponse, string, string, int, bool, error) {
	if m.findByIDSystemFunc != nil {
		return m.findByIDSystemFunc(ctx, id)
	}
	return nil, "", "", 0, false, nil
}

func (m *mockRepository) MarkDueWarningSent(ctx context.Context, id int64) error {
	if m.markDueWarningSentFunc != nil {
		return m.markDueWarningSentFunc(ctx, id)
	}
	return nil
}

func (m *mockRepository) GetUserThresholdAndTimezone(ctx context.Context, userID string) (int, string, error) {
	if m.getUserThresholdFunc != nil {
		return m.getUserThresholdFunc(ctx, userID)
	}
	return 1440, "Asia/Jakarta", nil
}

func (m *mockRepository) CreateActivity(ctx context.Context, taskID int64, userID string, activityType string, description string, changes []byte) (*TaskActivityResponse, error) {
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

func (m *mockRepository) FindActivitiesByTaskID(ctx context.Context, userID string, taskID int64) ([]TaskActivityResponse, error) {
	if m.findActivitiesFunc != nil {
		return m.findActivitiesFunc(ctx, userID, taskID)
	}
	return []TaskActivityResponse{}, nil
}

func (m *mockRepository) FindActivitiesByUser(ctx context.Context, userID string, from, to *time.Time, limit ...int) ([]TaskActivityResponse, error) {
	if m.findActivitiesByUserFunc != nil {
		return m.findActivitiesByUserFunc(ctx, userID, from, to)
	}
	return []TaskActivityResponse{}, nil
}

func (m *mockRepository) FindActivityCountsByUser(ctx context.Context, userID string, from, to *time.Time, tz string) ([]ActivityDateCount, error) {
	return []ActivityDateCount{}, nil
}

func (m *mockRepository) FindRecurringByID(ctx context.Context, userID string, id int64) (*RecurringTaskResponse, error) {
	if m.findRecurringByIDFunc != nil {
		return m.findRecurringByIDFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *mockRepository) FindRecurringByIDSystem(ctx context.Context, id int64) (*RecurringTaskResponse, string, error) {
	if m.findRecurringByIDSystemFunc != nil {
		return m.findRecurringByIDSystemFunc(ctx, id)
	}
	return nil, "Asia/Jakarta", nil
}

func (m *mockRepository) FindAllRecurring(ctx context.Context, userID string) ([]RecurringTaskResponse, error) {
	if m.findAllRecurringFunc != nil {
		return m.findAllRecurringFunc(ctx, userID)
	}
	return []RecurringTaskResponse{}, nil
}

func (m *mockRepository) CreateRecurring(ctx context.Context, userID string, req CreateRecurringTaskRequest) (*RecurringTaskResponse, error) {
	if m.createRecurringFunc != nil {
		return m.createRecurringFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *mockRepository) UpdateRecurring(ctx context.Context, userID string, id int64, req UpdateRecurringTaskRequest) (*RecurringTaskResponse, error) {
	if m.updateRecurringFunc != nil {
		return m.updateRecurringFunc(ctx, userID, id, req)
	}
	return nil, nil
}

func (m *mockRepository) DeleteRecurring(ctx context.Context, userID string, id int64) error {
	if m.deleteRecurringFunc != nil {
		return m.deleteRecurringFunc(ctx, userID, id)
	}
	return nil
}

func (m *mockRepository) ToggleActiveRecurring(ctx context.Context, userID string, id int64) (*RecurringTaskResponse, error) {
	if m.toggleActiveRecurringFunc != nil {
		return m.toggleActiveRecurringFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *mockRepository) FindTodayInstance(ctx context.Context, recurringTaskID int64, todayDate time.Time) (*TaskResponse, error) {
	if m.findTodayInstanceFunc != nil {
		return m.findTodayInstanceFunc(ctx, recurringTaskID, todayDate)
	}
	return nil, nil
}

func (m *mockRepository) CreateInstanceFromRecurring(ctx context.Context, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error) {
	if m.createInstanceFromRecurringFunc != nil {
		return m.createInstanceFromRecurringFunc(ctx, master, dueDate)
	}
	return nil, nil
}

func (m *mockRepository) UpdateRecurringLastGenerated(ctx context.Context, id int64, t time.Time) error {
	if m.updateRecurringLastGeneratedFunc != nil {
		return m.updateRecurringLastGeneratedFunc(ctx, id, t)
	}
	return nil
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
