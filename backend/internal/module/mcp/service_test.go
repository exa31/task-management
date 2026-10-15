package mcp

import (
	"context"
	"testing"
	"time"

	"golang/internal/module/note"
	"golang/internal/module/project"
	"golang/internal/module/reminder"
	"golang/internal/module/task"

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

type mockTokenRepository struct {
	createTokenFunc   func(ctx context.Context, userID string, name string) (*GenerateTokenResponse, error)
	listTokensFunc    func(ctx context.Context, userID string) ([]APITokenResponse, error)
	deleteTokenFunc   func(ctx context.Context, userID string, tokenID int64) error
	validateTokenFunc func(ctx context.Context, rawToken string) (string, error)
}

func (m *mockTokenRepository) CreateToken(ctx context.Context, userID string, name string) (*GenerateTokenResponse, error) {
	if m.createTokenFunc != nil {
		return m.createTokenFunc(ctx, userID, name)
	}
	return nil, nil
}

func (m *mockTokenRepository) ListTokens(ctx context.Context, userID string) ([]APITokenResponse, error) {
	if m.listTokensFunc != nil {
		return m.listTokensFunc(ctx, userID)
	}
	return []APITokenResponse{}, nil
}

func (m *mockTokenRepository) DeleteToken(ctx context.Context, userID string, tokenID int64) error {
	if m.deleteTokenFunc != nil {
		return m.deleteTokenFunc(ctx, userID, tokenID)
	}
	return nil
}

func (m *mockTokenRepository) ValidateToken(ctx context.Context, rawToken string) (string, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(ctx, rawToken)
	}
	return "user-123", nil
}

type dummyNoteService struct {
	note.Service
}

type dummyTaskService struct {
	task.Service
	createFunc        func(ctx context.Context, userID string, req task.CreateTaskRequest) (*task.TaskResponse, error)
	updateFunc        func(ctx context.Context, userID string, taskID int64, req task.UpdateTaskRequest) (*task.TaskResponse, error)
	createSubtaskFunc func(ctx context.Context, userID string, taskID int64, req task.CreateSubtaskRequest) (*task.SubtaskResponse, error)
	updateSubtaskFunc func(ctx context.Context, userID string, taskID int64, subtaskID int64, req task.UpdateSubtaskRequest) (*task.SubtaskResponse, error)
	deleteSubtaskFunc func(ctx context.Context, userID string, taskID int64, subtaskID int64) error
}

func (d *dummyTaskService) Create(ctx context.Context, userID string, req task.CreateTaskRequest) (*task.TaskResponse, error) {
	if d.createFunc != nil {
		return d.createFunc(ctx, userID, req)
	}
	return &task.TaskResponse{ID: 10, Title: req.Title}, nil
}

func (d *dummyTaskService) Update(ctx context.Context, userID string, taskID int64, req task.UpdateTaskRequest) (*task.TaskResponse, error) {
	if d.updateFunc != nil {
		return d.updateFunc(ctx, userID, taskID, req)
	}
	title := "Updated"
	if req.Title != nil {
		title = *req.Title
	}
	return &task.TaskResponse{ID: taskID, Title: title}, nil
}

func (d *dummyTaskService) CreateSubtask(ctx context.Context, userID string, taskID int64, req task.CreateSubtaskRequest) (*task.SubtaskResponse, error) {
	if d.createSubtaskFunc != nil {
		return d.createSubtaskFunc(ctx, userID, taskID, req)
	}
	return &task.SubtaskResponse{ID: 101, TaskID: taskID, Title: req.Title, Description: req.Description}, nil
}

func (d *dummyTaskService) UpdateSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64, req task.UpdateSubtaskRequest) (*task.SubtaskResponse, error) {
	if d.updateSubtaskFunc != nil {
		return d.updateSubtaskFunc(ctx, userID, taskID, subtaskID, req)
	}
	title := "Subtask"
	if req.Title != nil {
		title = *req.Title
	}
	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}
	isComp := false
	if req.IsCompleted != nil {
		isComp = *req.IsCompleted
	}
	return &task.SubtaskResponse{ID: subtaskID, TaskID: taskID, Title: title, Description: desc, IsCompleted: isComp}, nil
}

func (d *dummyTaskService) DeleteSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64) error {
	if d.deleteSubtaskFunc != nil {
		return d.deleteSubtaskFunc(ctx, userID, taskID, subtaskID)
	}
	return nil
}

func TestMCPService_CreateToken(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	mockRepo := &mockTokenRepository{
		createTokenFunc: func(ctx context.Context, userID string, name string) (*GenerateTokenResponse, error) {
			return &GenerateTokenResponse{
				ID:          1,
				Name:        name,
				TokenPrefix: "mcp_pat_12345678",
				RawToken:    "mcp_pat_12345678abcdef",
				CreatedAt:   time.Now(),
			}, nil
		},
	}

	svc := NewService(mockRepo, &dummyNoteService{}, &dummyTaskService{}, &mockTransactor{})
	res, err := svc.CreateToken(ctx, testUserID, "Cursor IDE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 1 || res.Name != "Cursor IDE" {
		t.Errorf("unexpected token response: %+v", res)
	}
}

func TestMCPService_ListTokens(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	mockRepo := &mockTokenRepository{
		listTokensFunc: func(ctx context.Context, userID string) ([]APITokenResponse, error) {
			return []APITokenResponse{
				{ID: 1, UserID: userID, Name: "Token 1"},
				{ID: 2, UserID: userID, Name: "Token 2"},
			}, nil
		},
	}

	svc := NewService(mockRepo, &dummyNoteService{}, &dummyTaskService{}, &mockTransactor{})
	tokens, err := svc.ListTokens(ctx, testUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
}

func TestMCPService_DeleteToken(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	deleted := false
	mockRepo := &mockTokenRepository{
		deleteTokenFunc: func(ctx context.Context, userID string, tokenID int64) error {
			deleted = true
			return nil
		},
	}

	svc := NewService(mockRepo, &dummyNoteService{}, &dummyTaskService{}, &mockTransactor{})
	err := svc.DeleteToken(ctx, testUserID, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Errorf("expected delete to have been called")
	}
}

func TestMCPService_ValidateToken(t *testing.T) {
	ctx := context.Background()

	mockRepo := &mockTokenRepository{
		validateTokenFunc: func(ctx context.Context, rawToken string) (string, error) {
			return "user-999", nil
		},
	}

	svc := NewService(mockRepo, &dummyNoteService{}, &dummyTaskService{}, &mockTransactor{})
	uid, err := svc.ValidateToken(ctx, "mcp_pat_secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid != "user-999" {
		t.Errorf("expected user-999, got %s", uid)
	}
}

func TestMCPService_HandleJSONRPC(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	taskSvc := &dummyTaskService{
		createFunc: func(ctx context.Context, userID string, req task.CreateTaskRequest) (*task.TaskResponse, error) {
			return &task.TaskResponse{
				ID:    42,
				Title: req.Title,
			}, nil
		},
	}

	svc := NewService(&mockTokenRepository{}, &dummyNoteService{}, taskSvc, &mockTransactor{})

	// 1. Initialize
	initResp := svc.HandleJSONRPC(ctx, testUserID, JSONRPCRequest{
		Method: "initialize",
		ID:     1,
	})
	if initResp.Error != nil {
		t.Fatalf("initialize failed: %v", initResp.Error)
	}

	// 2. Tools list
	listResp := svc.HandleJSONRPC(ctx, testUserID, JSONRPCRequest{
		Method: "tools/list",
		ID:     2,
	})
	if listResp.Error != nil {
		t.Fatalf("tools/list failed: %v", listResp.Error)
	}

	var capturedCreateReq task.CreateTaskRequest
	taskSvc.createFunc = func(ctx context.Context, userID string, req task.CreateTaskRequest) (*task.TaskResponse, error) {
		capturedCreateReq = req
		return &task.TaskResponse{
			ID:    42,
			Title: req.Title,
		}, nil
	}

	// 3. Tool call - create_task with due date & recurrence
	callResp := svc.HandleJSONRPC(ctx, testUserID, JSONRPCRequest{
		Method: "tools/call",
		ID:     3,
		Params: map[string]interface{}{
			"name": "create_task",
			"arguments": map[string]interface{}{
				"title":               "Build MCP Client",
				"due_date":            "2026-10-15T18:00:00Z",
				"is_recurring":        true,
				"recurrence_pattern":  "weekly",
				"recurrence_interval": 2,
			},
		},
	})
	if callResp.Error != nil {
		t.Fatalf("tool call failed: %v", callResp.Error)
	}
	if capturedCreateReq.DueDate == nil {
		t.Fatalf("expected DueDate to be parsed, got nil")
	}
	if !capturedCreateReq.IsRecurring || capturedCreateReq.RecurrencePattern != "weekly" || capturedCreateReq.RecurrenceInterval != 2 {
		t.Errorf("unexpected recurrence params: %+v", capturedCreateReq)
	}

	// 4. Tool call - update_task with recurrence
	var capturedUpdateReq task.UpdateTaskRequest
	taskSvc.updateFunc = func(ctx context.Context, userID string, taskID int64, req task.UpdateTaskRequest) (*task.TaskResponse, error) {
		capturedUpdateReq = req
		return &task.TaskResponse{ID: taskID, Title: "Updated Task"}, nil
	}
	updateResp := svc.HandleJSONRPC(ctx, testUserID, JSONRPCRequest{
		Method: "tools/call",
		ID:     4,
		Params: map[string]interface{}{
			"name": "update_task",
			"arguments": map[string]interface{}{
				"task_id":            42,
				"due_date":           "2026-10-20T10:00:00Z",
				"is_recurring":       true,
				"recurrence_pattern": "daily",
			},
		},
	})
	if updateResp.Error != nil {
		t.Fatalf("update_task failed: %v", updateResp.Error)
	}
	if capturedUpdateReq.DueDate == nil {
		t.Fatalf("expected DueDate in update, got nil")
	}
	if capturedUpdateReq.IsRecurring == nil || !*capturedUpdateReq.IsRecurring || capturedUpdateReq.RecurrencePattern == nil || *capturedUpdateReq.RecurrencePattern != "daily" {
		t.Errorf("unexpected update recurrence params: %+v", capturedUpdateReq)
	}
}

type dummyReminderService struct {
	reminder.Service
	createFunc func(ctx context.Context, userID string, req reminder.CreateReminderRequest) (*reminder.ReminderResponse, error)
	listFunc   func(ctx context.Context, userID string, filter reminder.ReminderFilter) ([]reminder.ReminderResponse, int, error)
}

func (d *dummyReminderService) Create(ctx context.Context, userID string, req reminder.CreateReminderRequest) (*reminder.ReminderResponse, error) {
	if d.createFunc != nil {
		return d.createFunc(ctx, userID, req)
	}
	return &reminder.ReminderResponse{ID: 101, Title: req.Title, RemindAt: req.RemindAt}, nil
}

func (d *dummyReminderService) List(ctx context.Context, userID string, filter reminder.ReminderFilter) ([]reminder.ReminderResponse, int, error) {
	if d.listFunc != nil {
		return d.listFunc(ctx, userID, filter)
	}
	return []reminder.ReminderResponse{{ID: 101, Title: "Standup"}}, 1, nil
}

func TestMCPService_ReminderTools(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	remSvc := &dummyReminderService{}
	svc := NewService(&mockTokenRepository{}, &dummyNoteService{}, &dummyTaskService{}, &mockTransactor{})
	svc.SetReminderService(remSvc)

	// Test create_reminder
	createResp := svc.HandleJSONRPC(ctx, testUserID, JSONRPCRequest{
		Method: "tools/call",
		ID:     1,
		Params: map[string]interface{}{
			"name": "create_reminder",
			"arguments": map[string]interface{}{
				"title":     "Team Sync",
				"remind_at": "2026-10-15T09:00:00Z",
			},
		},
	})
	if createResp.Error != nil {
		t.Fatalf("create_reminder failed: %v", createResp.Error)
	}

	// Test list_reminders
	listResp := svc.HandleJSONRPC(ctx, testUserID, JSONRPCRequest{
		Method: "tools/call",
		ID:     2,
		Params: map[string]interface{}{
			"name": "list_reminders",
			"arguments": map[string]interface{}{
				"status": "pending",
			},
		},
	})
	if listResp.Error != nil {
		t.Fatalf("list_reminders failed: %v", listResp.Error)
	}
}

type dummyProjectService struct {
	project.Service
	listFunc func(ctx context.Context, userID string) ([]project.ProjectResponse, error)
}

func (d *dummyProjectService) List(ctx context.Context, userID string) ([]project.ProjectResponse, error) {
	if d.listFunc != nil {
		return d.listFunc(ctx, userID)
	}
	return []project.ProjectResponse{
		{ID: 2, Name: "Task-management", Color: "#3b82f6", TaskCount: 5},
	}, nil
}

func TestMCPService_ProjectTools(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	var capturedReq task.CreateTaskRequest
	taskSvc := &dummyTaskService{
		createFunc: func(ctx context.Context, userID string, req task.CreateTaskRequest) (*task.TaskResponse, error) {
			capturedReq = req
			return &task.TaskResponse{
				ID:        55,
				Title:     req.Title,
				ProjectID: req.ProjectID,
			}, nil
		},
	}

	projSvc := &dummyProjectService{}
	svc := NewService(&mockTokenRepository{}, &dummyNoteService{}, taskSvc, &mockTransactor{})
	svc.SetProjectService(projSvc)

	// 1. Test list_projects
	listResp := svc.HandleJSONRPC(ctx, testUserID, JSONRPCRequest{
		Method: "tools/call",
		ID:     1,
		Params: map[string]interface{}{
			"name":      "list_projects",
			"arguments": map[string]interface{}{},
		},
	})
	if listResp.Error != nil {
		t.Fatalf("list_projects failed: %v", listResp.Error)
	}

	// 2. Test create_task with project_id
	createResp := svc.HandleJSONRPC(ctx, testUserID, JSONRPCRequest{
		Method: "tools/call",
		ID:     2,
		Params: map[string]interface{}{
			"name": "create_task",
			"arguments": map[string]interface{}{
				"title":      "MCP with Project Link",
				"project_id": 2,
			},
		},
	})
	if createResp.Error != nil {
		t.Fatalf("create_task with project_id failed: %v", createResp.Error)
	}
	if capturedReq.ProjectID == nil || *capturedReq.ProjectID != 2 {
		t.Fatalf("expected project_id 2 to be linked, got %v", capturedReq.ProjectID)
	}
}

func TestMCPService_SubtaskTools(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	var createdSubtaskReq task.CreateSubtaskRequest
	var updatedSubtaskReq task.UpdateSubtaskRequest
	var deletedSubtaskID int64

	taskSvc := &dummyTaskService{
		createSubtaskFunc: func(ctx context.Context, userID string, taskID int64, req task.CreateSubtaskRequest) (*task.SubtaskResponse, error) {
			createdSubtaskReq = req
			return &task.SubtaskResponse{
				ID:          201,
				TaskID:      taskID,
				Title:       req.Title,
				Description: req.Description,
			}, nil
		},
		updateSubtaskFunc: func(ctx context.Context, userID string, taskID int64, subtaskID int64, req task.UpdateSubtaskRequest) (*task.SubtaskResponse, error) {
			updatedSubtaskReq = req
			return &task.SubtaskResponse{
				ID:          subtaskID,
				TaskID:      taskID,
				Title:       *req.Title,
				Description: *req.Description,
				IsCompleted: *req.IsCompleted,
			}, nil
		},
		deleteSubtaskFunc: func(ctx context.Context, userID string, taskID int64, subtaskID int64) error {
			deletedSubtaskID = subtaskID
			return nil
		},
	}

	svc := NewService(&mockTokenRepository{}, &dummyNoteService{}, taskSvc, &mockTransactor{})

	// 1. Test create_subtask
	createResp := svc.HandleJSONRPC(ctx, testUserID, JSONRPCRequest{
		Method: "tools/call",
		ID:     1,
		Params: map[string]interface{}{
			"name": "create_subtask",
			"arguments": map[string]interface{}{
				"task_id":     10,
				"title":       "New Subtask Title",
				"description": "Subtask detail notes",
				"due_date":    "2026-10-20T10:00:00Z",
			},
		},
	})
	if createResp.Error != nil {
		t.Fatalf("create_subtask failed: %v", createResp.Error)
	}
	if createdSubtaskReq.Title != "New Subtask Title" || createdSubtaskReq.Description != "Subtask detail notes" {
		t.Fatalf("unexpected create_subtask params: %+v", createdSubtaskReq)
	}

	// 2. Test update_subtask
	updateResp := svc.HandleJSONRPC(ctx, testUserID, JSONRPCRequest{
		Method: "tools/call",
		ID:     2,
		Params: map[string]interface{}{
			"name": "update_subtask",
			"arguments": map[string]interface{}{
				"task_id":      10,
				"subtask_id":   201,
				"title":        "Updated Subtask",
				"description":  "Updated notes",
				"is_completed": true,
			},
		},
	})
	if updateResp.Error != nil {
		t.Fatalf("update_subtask failed: %v", updateResp.Error)
	}
	if updatedSubtaskReq.Title == nil || *updatedSubtaskReq.Title != "Updated Subtask" {
		t.Fatalf("unexpected update_subtask title: %+v", updatedSubtaskReq)
	}

	// 3. Test delete_subtask
	delResp := svc.HandleJSONRPC(ctx, testUserID, JSONRPCRequest{
		Method: "tools/call",
		ID:     3,
		Params: map[string]interface{}{
			"name": "delete_subtask",
			"arguments": map[string]interface{}{
				"task_id":    10,
				"subtask_id": 201,
			},
		},
	})
	if delResp.Error != nil {
		t.Fatalf("delete_subtask failed: %v", delResp.Error)
	}
	if deletedSubtaskID != 201 {
		t.Fatalf("expected subtask 201 to be deleted, got %d", deletedSubtaskID)
	}
}
