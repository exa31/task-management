package note

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
	createFunc        func(ctx context.Context, userID string, req CreateNoteRequest) (*NoteResponse, error)
	findByIDFunc      func(ctx context.Context, userID string, id int64) (*NoteResponse, error)
	findAllFunc       func(ctx context.Context, userID string, filter NoteFilter) ([]NoteResponse, error)
	updateFunc        func(ctx context.Context, userID string, id int64, req UpdateNoteRequest) (*NoteResponse, error)
	deleteFunc        func(ctx context.Context, userID string, id int64) error
	markConvertedFunc func(ctx context.Context, userID string, id int64, taskID int64) error
}

func (m *mockRepository) Create(ctx context.Context, userID string, req CreateNoteRequest) (*NoteResponse, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *mockRepository) FindByID(ctx context.Context, userID string, id int64) (*NoteResponse, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, userID, id)
	}
	return nil, nil
}

func (m *mockRepository) FindAll(ctx context.Context, userID string, filter NoteFilter) ([]NoteResponse, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, userID, filter)
	}
	return []NoteResponse{}, nil
}

func (m *mockRepository) Update(ctx context.Context, userID string, id int64, req UpdateNoteRequest) (*NoteResponse, error) {
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

func (m *mockRepository) MarkConverted(ctx context.Context, userID string, id int64, taskID int64) error {
	if m.markConvertedFunc != nil {
		return m.markConvertedFunc(ctx, userID, id, taskID)
	}
	return nil
}

type mockTaskCreator struct {
	createFromNoteFunc func(ctx context.Context, userID string, noteID int64, title string, content string, req ConvertToTaskRequest) (int64, error)
}

func (m *mockTaskCreator) CreateFromNote(ctx context.Context, userID string, noteID int64, title string, content string, req ConvertToTaskRequest) (int64, error) {
	if m.createFromNoteFunc != nil {
		return m.createFromNoteFunc(ctx, userID, noteID, title, content, req)
	}
	return 100, nil
}

func TestNoteService_Create(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	mockRepo := &mockRepository{
		createFunc: func(ctx context.Context, userID string, req CreateNoteRequest) (*NoteResponse, error) {
			return &NoteResponse{
				ID:      10,
				UserID:  userID,
				Title:   req.Title,
				Content: req.Content,
			}, nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})
	res, err := svc.Create(ctx, testUserID, CreateNoteRequest{
		Title:   "New Note",
		Content: "Some idea",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 10 || res.Title != "New Note" {
		t.Errorf("unexpected result: %+v", res)
	}
}

func TestNoteService_GetByID(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	tests := []struct {
		name      string
		id        int64
		mockRepo  Repository
		expectErr bool
	}{
		{
			name: "Success - Note found",
			id:   1,
			mockRepo: &mockRepository{
				findByIDFunc: func(ctx context.Context, userID string, id int64) (*NoteResponse, error) {
					return &NoteResponse{ID: 1, UserID: userID, Title: "Great Idea", Content: "Details"}, nil
				},
			},
			expectErr: false,
		},
		{
			name: "Error - Note not found",
			id:   999,
			mockRepo: &mockRepository{
				findByIDFunc: func(ctx context.Context, userID string, id int64) (*NoteResponse, error) {
					return nil, nil
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.mockRepo, &mockTransactor{})
			res, err := svc.GetByID(ctx, testUserID, tt.id)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if res == nil || res.ID != tt.id {
					t.Errorf("expected note ID %d, got %v", tt.id, res)
				}
			}
		})
	}
}

func TestNoteService_ConvertToTask(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	mockRepo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*NoteResponse, error) {
			return &NoteResponse{
				ID:        1,
				UserID:    userID,
				Title:     "Convertible Note",
				Content:   "Notes content",
				CreatedAt: time.Now(),
			}, nil
		},
		markConvertedFunc: func(ctx context.Context, userID string, id int64, taskID int64) error {
			return nil
		},
	}

	mockCreator := &mockTaskCreator{
		createFromNoteFunc: func(ctx context.Context, userID string, noteID int64, title string, content string, req ConvertToTaskRequest) (int64, error) {
			return 42, nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})
	svc.SetTaskCreator(mockCreator)

	noteRes, taskID, err := svc.ConvertToTask(ctx, testUserID, 1, ConvertToTaskRequest{
		Priority: "high",
		Subtasks: []string{"Step 1", "Step 2"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if taskID != 42 {
		t.Errorf("expected taskID 42, got %d", taskID)
	}

	if noteRes == nil || noteRes.ConvertedTaskID == nil || *noteRes.ConvertedTaskID != 42 {
		t.Errorf("expected note to have converted_task_id = 42")
	}
}

func TestNoteService_FindAll(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-uuid-123"

	mockRepo := &mockRepository{
		findAllFunc: func(ctx context.Context, userID string, filter NoteFilter) ([]NoteResponse, error) {
			return []NoteResponse{
				{ID: 1, UserID: userID, Title: "Note 1"},
				{ID: 2, UserID: userID, Title: "Note 2"},
			}, nil
		},
	}

	svc := NewService(mockRepo, &mockTransactor{})
	res, err := svc.FindAll(ctx, testUserID, NoteFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(res))
	}
}
