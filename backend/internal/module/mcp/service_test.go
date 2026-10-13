package mcp

import (
	"context"
	"testing"
	"time"

	"golang/internal/module/note"
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
	createFunc func(ctx context.Context, userID string, req task.CreateTaskRequest) (*task.TaskResponse, error)
}

func (d *dummyTaskService) Create(ctx context.Context, userID string, req task.CreateTaskRequest) (*task.TaskResponse, error) {
	if d.createFunc != nil {
		return d.createFunc(ctx, userID, req)
	}
	return &task.TaskResponse{ID: 10, Title: req.Title}, nil
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

	// 3. Tool call - create_task
	callResp := svc.HandleJSONRPC(ctx, testUserID, JSONRPCRequest{
		Method: "tools/call",
		ID:     3,
		Params: map[string]interface{}{
			"name": "create_task",
			"arguments": map[string]interface{}{
				"title": "Build MCP Client",
			},
		},
	})
	if callResp.Error != nil {
		t.Fatalf("tool call failed: %v", callResp.Error)
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
