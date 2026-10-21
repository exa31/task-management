package project

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

type mockRepository struct {
	projects map[int64]*ProjectResponse
	nextID   int64
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		projects: make(map[int64]*ProjectResponse),
		nextID:   1,
	}
}

func (m *mockRepository) Create(ctx context.Context, tx *sqlx.Tx, userID string, req *CreateProjectRequest) (*ProjectResponse, error) {
	id := m.nextID
	m.nextID++
	color := req.Color
	if color == "" {
		color = "indigo"
	}
	p := &ProjectResponse{
		ID:          id,
		UserID:      userID,
		Name:        req.Name,
		Color:       color,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.projects[id] = p
	return p, nil
}

func (m *mockRepository) FindAll(ctx context.Context, tx *sqlx.Tx, userID string) ([]ProjectResponse, error) {
	var list []ProjectResponse
	for _, p := range m.projects {
		if p.UserID == userID {
			list = append(list, *p)
		}
	}
	return list, nil
}

func (m *mockRepository) FindByID(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*ProjectResponse, error) {
	p, ok := m.projects[id]
	if !ok || p.UserID != userID {
		return nil, nil
	}
	return p, nil
}

func (m *mockRepository) Update(ctx context.Context, tx *sqlx.Tx, userID string, id int64, req *UpdateProjectRequest) (*ProjectResponse, error) {
	p, ok := m.projects[id]
	if !ok || p.UserID != userID {
		return nil, nil
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Color != nil {
		p.Color = *req.Color
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	p.UpdatedAt = time.Now()
	return p, nil
}

func (m *mockRepository) Delete(ctx context.Context, tx *sqlx.Tx, userID string, id int64) error {
	p, ok := m.projects[id]
	if !ok || p.UserID != userID {
		return sql.ErrNoRows
	}
	delete(m.projects, id)
	return nil
}

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

func TestProjectService(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo, &mockTransactor{})
	ctx := context.Background()
	userID := "user-123"

	t.Run("Create and GetByID", func(t *testing.T) {
		created, err := svc.Create(ctx, userID, &CreateProjectRequest{
			Name:        "Test Project",
			Color:       "emerald",
			Description: "Desc",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.Name != "Test Project" || created.Color != "emerald" {
			t.Errorf("created project mismatch: %+v", created)
		}

		fetched, err := svc.GetByID(ctx, userID, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fetched.ID != created.ID {
			t.Errorf("expected ID %d, got %d", created.ID, fetched.ID)
		}
	})

	t.Run("List and Update", func(t *testing.T) {
		list, err := svc.List(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 project, got %d", len(list))
		}

		newName := "Updated Name"
		updated, err := svc.Update(ctx, userID, list[0].ID, &UpdateProjectRequest{
			Name: &newName,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != newName {
			t.Errorf("expected updated name %s, got %s", newName, updated.Name)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		list, _ := svc.List(ctx, userID)
		id := list[0].ID

		err := svc.Delete(ctx, userID, id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = svc.GetByID(ctx, userID, id)
		if err == nil {
			t.Error("expected not found error after delete, got nil")
		}
	})
}
