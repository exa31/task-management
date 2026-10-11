package task

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
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

	createSubtaskFunc func(ctx context.Context, userID string, taskID int64, title string) (*SubtaskResponse, error)
	updateSubtaskFunc func(ctx context.Context, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error)
	deleteSubtaskFunc func(ctx context.Context, userID string, taskID int64, subtaskID int64) error

	findByIDSystemFunc     func(ctx context.Context, id int64) (*TaskResponse, string, string, int, bool, error)
	markDueWarningSentFunc func(ctx context.Context, id int64) error
	getUserThresholdFunc   func(ctx context.Context, userID string) (int, string, error)
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

func (m *mockRepository) CreateSubtask(ctx context.Context, userID string, taskID int64, title string) (*SubtaskResponse, error) {
	if m.createSubtaskFunc != nil {
		return m.createSubtaskFunc(ctx, userID, taskID, title)
	}
	return nil, nil
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

