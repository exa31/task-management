package note

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang/database"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Repository interface {
	Create(ctx context.Context, userID string, req CreateNoteRequest) (*NoteResponse, error)
	FindByID(ctx context.Context, userID string, id int64) (*NoteResponse, error)
	FindAll(ctx context.Context, userID string, filter NoteFilter) ([]NoteResponse, error)
	Update(ctx context.Context, userID string, id int64, req UpdateNoteRequest) (*NoteResponse, error)
	Delete(ctx context.Context, userID string, id int64) error
	MarkConverted(ctx context.Context, userID string, id int64, taskID int64) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) exec(ctx context.Context) database.QueryExecutor {
	return database.GetExecutor(ctx, r.db)
}

type noteEntity struct {
	ID              int64          `db:"id"`
	UserID          string         `db:"user_id"`
	Title           string         `db:"title"`
	Content         string         `db:"content"`
	Type            string         `db:"type"`
	Tags            pq.StringArray `db:"tags"`
	Color           string         `db:"color"`
	IsPinned        bool           `db:"is_pinned"`
	ConvertedTaskID *int64         `db:"converted_task_id"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
}

func (e *noteEntity) toResponse() NoteResponse {
	tags := []string(e.Tags)
	if tags == nil {
		tags = []string{}
	}
	color := e.Color
	if color == "" {
		color = "default"
	}
	noteType := e.Type
	if noteType == "" {
		noteType = "actionable"
	}
	return NoteResponse{
		ID:              e.ID,
		UserID:          e.UserID,
		Title:           e.Title,
		Content:         e.Content,
		Type:            noteType,
		Tags:            tags,
		Color:           color,
		IsPinned:        e.IsPinned,
		ConvertedTaskID: e.ConvertedTaskID,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

func (r *repository) Create(ctx context.Context, userID string, req CreateNoteRequest) (*NoteResponse, error) {
	query := `
		INSERT INTO notes (user_id, title, content, type, tags, color, is_pinned, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, user_id, title, content, type, tags, color, is_pinned, converted_task_id, created_at, updated_at
	`
	color := req.Color
	if color == "" {
		color = "default"
	}

	noteType := req.Type
	if noteType == "" {
		noteType = "actionable"
	}

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	var entity noteEntity
	err := r.exec(ctx).QueryRowxContext(ctx, query, userID, req.Title, req.Content, noteType, pq.Array(tags), color, req.IsPinned).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	res := entity.toResponse()
	return &res, nil
}

func (r *repository) FindByID(ctx context.Context, userID string, id int64) (*NoteResponse, error) {
	query := `
		SELECT id, user_id, title, content, type, tags, color, is_pinned, converted_task_id, created_at, updated_at
		FROM notes
		WHERE id = $1 AND user_id = $2
	`
	var entity noteEntity
	err := r.exec(ctx).GetContext(ctx, &entity, query, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find note: %w", err)
	}

	res := entity.toResponse()
	return &res, nil
}

func (r *repository) FindAll(ctx context.Context, userID string, filter NoteFilter) ([]NoteResponse, error) {
	query := `
		SELECT id, user_id, title, content, type, tags, color, is_pinned, converted_task_id, created_at, updated_at
		FROM notes
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argIdx := 2

	if filter.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, filter.Type)
		argIdx++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR content ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	if filter.Tag != "" {
		query += fmt.Sprintf(" AND $%d = ANY(tags)", argIdx)
		args = append(args, filter.Tag)
		argIdx++
	}

	query += " ORDER BY is_pinned DESC, updated_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
	}

	var entities []noteEntity
	err := r.exec(ctx).SelectContext(ctx, &entities, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}

	responses := make([]NoteResponse, 0, len(entities))
	for i := range entities {
		responses = append(responses, entities[i].toResponse())
	}
	return responses, nil
}

func (r *repository) Update(ctx context.Context, userID string, id int64, req UpdateNoteRequest) (*NoteResponse, error) {
	existing, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	title := existing.Title
	if req.Title != nil {
		title = *req.Title
	}

	content := existing.Content
	if req.Content != nil {
		content = *req.Content
	}

	noteType := existing.Type
	if req.Type != nil {
		noteType = *req.Type
	}

	tags := existing.Tags
	if req.Tags != nil {
		tags = *req.Tags
	}

	color := existing.Color
	if req.Color != nil {
		color = *req.Color
	}

	isPinned := existing.IsPinned
	if req.IsPinned != nil {
		isPinned = *req.IsPinned
	}

	convertedTaskID := existing.ConvertedTaskID
	if req.ConvertedTaskID != nil {
		convertedTaskID = req.ConvertedTaskID
	}

	query := `
		UPDATE notes
		SET title = $1, content = $2, type = $3, tags = $4, color = $5, is_pinned = $6, converted_task_id = $7, updated_at = NOW()
		WHERE id = $8 AND user_id = $9
		RETURNING id, user_id, title, content, type, tags, color, is_pinned, converted_task_id, created_at, updated_at
	`
	var entity noteEntity
	err = r.exec(ctx).QueryRowxContext(ctx, query, title, content, noteType, pq.Array(tags), color, isPinned, convertedTaskID, id, userID).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	res := entity.toResponse()
	return &res, nil
}

func (r *repository) Delete(ctx context.Context, userID string, id int64) error {
	query := `DELETE FROM notes WHERE id = $1 AND user_id = $2`
	_, err := r.exec(ctx).ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}
	return nil
}

func (r *repository) MarkConverted(ctx context.Context, userID string, id int64, taskID int64) error {
	query := `UPDATE notes SET converted_task_id = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`
	_, err := r.exec(ctx).ExecContext(ctx, query, taskID, id, userID)
	if err != nil {
		return fmt.Errorf("failed to mark note as converted: %w", err)
	}
	return nil
}
