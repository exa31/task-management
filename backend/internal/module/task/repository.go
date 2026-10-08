package task

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
	Create(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error)
	FindByID(ctx context.Context, userID string, id int64) (*TaskResponse, error)
	FindAll(ctx context.Context, userID string, filter TaskFilter) ([]TaskResponse, error)
	Update(ctx context.Context, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error)
	Delete(ctx context.Context, userID string, id int64) error
	LogProgress(ctx context.Context, userID string, id int64, req LogProgressRequest) (*TaskResponse, error)
	ToggleFocus(ctx context.Context, userID string, id int64) (*TaskResponse, error)

	CreateSubtask(ctx context.Context, userID string, taskID int64, title string) (*SubtaskResponse, error)
	UpdateSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error)
	DeleteSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64) error
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

type taskEntity struct {
	ID               int64          `db:"id"`
	UserID           string         `db:"user_id"`
	Title            string         `db:"title"`
	Description      string         `db:"description"`
	Status           string         `db:"status"`
	Priority         string         `db:"priority"`
	DueDate          *time.Time     `db:"due_date"`
	IsFocus          bool           `db:"is_focus"`
	Tags             pq.StringArray `db:"tags"`
	LastProgressNote string         `db:"last_progress_note"`
	LastProgressAt   *time.Time     `db:"last_progress_at"`
	SourceNoteID     *int64         `db:"source_note_id"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
}

func (e *taskEntity) toResponse(subtasks []SubtaskResponse) TaskResponse {
	tags := []string(e.Tags)
	if tags == nil {
		tags = []string{}
	}
	if subtasks == nil {
		subtasks = []SubtaskResponse{}
	}

	completedCount := 0
	for _, st := range subtasks {
		if st.IsCompleted {
			completedCount++
		}
	}
	totalCount := len(subtasks)
	percent := 0
	if totalCount > 0 {
		percent = (completedCount * 100) / totalCount
	}

	return TaskResponse{
		ID:                e.ID,
		UserID:            e.UserID,
		Title:             e.Title,
		Description:       e.Description,
		Status:            e.Status,
		Priority:          e.Priority,
		DueDate:           e.DueDate,
		IsFocus:           e.IsFocus,
		Tags:              tags,
		LastProgressNote:  e.LastProgressNote,
		LastProgressAt:    e.LastProgressAt,
		SourceNoteID:      e.SourceNoteID,
		Subtasks:          subtasks,
		SubtasksCount:     totalCount,
		SubtasksCompleted: completedCount,
		ProgressPercent:   percent,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
}

func (r *repository) fetchSubtasks(ctx context.Context, taskID int64) ([]SubtaskResponse, error) {
	query := `
		SELECT id, task_id, title, is_completed, order_index, created_at, updated_at
		FROM subtasks
		WHERE task_id = $1
		ORDER BY order_index ASC, id ASC
	`
	var items []SubtaskResponse
	err := r.exec(ctx).SelectContext(ctx, &items, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subtasks: %w", err)
	}

	if items == nil {
		items = []SubtaskResponse{}
	}
	return items, nil
}

func (r *repository) Create(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error) {
	status := req.Status
	if status == "" {
		status = "todo"
	}
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	var entity taskEntity
	subtasks := make([]SubtaskResponse, 0, len(req.Subtasks))

	query := `
		INSERT INTO tasks (user_id, title, description, status, priority, due_date, is_focus, tags, source_note_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, user_id, title, description, status, priority, due_date, is_focus, tags, last_progress_note, last_progress_at, source_note_id, created_at, updated_at
	`
	err := r.exec(ctx).QueryRowxContext(ctx, query, userID, req.Title, req.Description, status, priority, req.DueDate, req.IsFocus, pq.Array(tags), req.SourceNoteID).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	for i, stTitle := range req.Subtasks {
		if stTitle == "" {
			continue
		}
		subQuery := `
			INSERT INTO subtasks (task_id, title, is_completed, order_index, created_at, updated_at)
			VALUES ($1, $2, FALSE, $3, NOW(), NOW())
			RETURNING id, task_id, title, is_completed, order_index, created_at, updated_at
		`
		var st SubtaskResponse
		if err := r.exec(ctx).QueryRowxContext(ctx, subQuery, entity.ID, stTitle, i).StructScan(&st); err != nil {
			return nil, fmt.Errorf("failed to create subtask: %w", err)
		}
		subtasks = append(subtasks, st)
	}

	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) FindByID(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
	query := `
		SELECT id, user_id, title, description, status, priority, due_date, is_focus, tags, last_progress_note, last_progress_at, source_note_id, created_at, updated_at
		FROM tasks
		WHERE id = $1 AND user_id = $2
	`
	var entity taskEntity
	err := r.exec(ctx).GetContext(ctx, &entity, query, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	subtasks, err := r.fetchSubtasks(ctx, entity.ID)
	if err != nil {
		return nil, err
	}

	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) FindAll(ctx context.Context, userID string, filter TaskFilter) ([]TaskResponse, error) {
	query := `
		SELECT id, user_id, title, description, status, priority, due_date, is_focus, tags, last_progress_note, last_progress_at, source_note_id, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argIdx := 2

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.Priority != "" {
		query += fmt.Sprintf(" AND priority = $%d", argIdx)
		args = append(args, filter.Priority)
		argIdx++
	}

	if filter.IsFocus != nil {
		query += fmt.Sprintf(" AND is_focus = $%d", argIdx)
		args = append(args, *filter.IsFocus)
		argIdx++
	}

	if filter.Tag != "" {
		query += fmt.Sprintf(" AND $%d = ANY(tags)", argIdx)
		args = append(args, filter.Tag)
		argIdx++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	if filter.Cursor != nil && *filter.Cursor > 0 {
		query += fmt.Sprintf(" AND id < $%d", argIdx)
		args = append(args, *filter.Cursor)
		argIdx++
	}

	if filter.Paginate == "cursor" || filter.Cursor != nil {
		query += " ORDER BY id DESC"
	} else {
		query += " ORDER BY is_focus DESC, CASE status WHEN 'in_progress' THEN 1 WHEN 'todo' THEN 2 WHEN 'backlog' THEN 3 WHEN 'review' THEN 4 ELSE 5 END, updated_at DESC, id DESC"
	}

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 && filter.Cursor == nil {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
	}

	var entities []taskEntity
	err := r.exec(ctx).SelectContext(ctx, &entities, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	responses := make([]TaskResponse, 0, len(entities))
	for i := range entities {
		subtasks, _ := r.fetchSubtasks(ctx, entities[i].ID)
		responses = append(responses, entities[i].toResponse(subtasks))
	}
	return responses, nil
}

func (r *repository) Update(ctx context.Context, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
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

	desc := existing.Description
	if req.Description != nil {
		desc = *req.Description
	}

	status := existing.Status
	if req.Status != nil {
		status = *req.Status
	}

	priority := existing.Priority
	if req.Priority != nil {
		priority = *req.Priority
	}

	dueDate := existing.DueDate
	if req.DueDate != nil {
		dueDate = req.DueDate
	}

	isFocus := existing.IsFocus
	if req.IsFocus != nil {
		isFocus = *req.IsFocus
	}

	tags := existing.Tags
	if req.Tags != nil {
		tags = *req.Tags
	}

	lastProgressNote := existing.LastProgressNote
	var lastProgressAt = existing.LastProgressAt
	if req.LastProgressNote != nil {
		lastProgressNote = *req.LastProgressNote
		now := time.Now()
		lastProgressAt = &now
	}

	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, priority = $4, due_date = $5, is_focus = $6, tags = $7, last_progress_note = $8, last_progress_at = $9, updated_at = NOW()
		WHERE id = $10 AND user_id = $11
		RETURNING id, user_id, title, description, status, priority, due_date, is_focus, tags, last_progress_note, last_progress_at, source_note_id, created_at, updated_at
	`
	var entity taskEntity
	err = r.exec(ctx).QueryRowxContext(ctx, query, title, desc, status, priority, dueDate, isFocus, pq.Array(tags), lastProgressNote, lastProgressAt, id, userID).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	subtasks, _ := r.fetchSubtasks(ctx, entity.ID)
	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) Delete(ctx context.Context, userID string, id int64) error {
	query := `DELETE FROM tasks WHERE id = $1 AND user_id = $2`
	_, err := r.exec(ctx).ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (r *repository) LogProgress(ctx context.Context, userID string, id int64, req LogProgressRequest) (*TaskResponse, error) {
	statusUpdate := ""
	args := []interface{}{req.ProgressNote, id, userID}
	if req.Status != nil && *req.Status != "" {
		statusUpdate = ", status = $4"
		args = append(args, *req.Status)
	}

	query := fmt.Sprintf(`
		UPDATE tasks
		SET last_progress_note = $1, last_progress_at = NOW(), updated_at = NOW()%s
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, title, description, status, priority, due_date, is_focus, tags, last_progress_note, last_progress_at, source_note_id, created_at, updated_at
	`, statusUpdate)

	var entity taskEntity

	err := r.exec(ctx).QueryRowxContext(ctx, query, args...).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to log progress: %w", err)
	}

	if len(req.CompletedSubtaskIDs) > 0 {
		_, err := r.exec(ctx).ExecContext(ctx, `
			UPDATE subtasks SET is_completed = TRUE, updated_at = NOW()
			WHERE task_id = $1 AND id = ANY($2)
		`, id, pq.Array(req.CompletedSubtaskIDs))
		if err != nil {
			return nil, fmt.Errorf("failed to mark completed subtasks: %w", err)
		}
	}

	subtasks, _ := r.fetchSubtasks(ctx, entity.ID)
	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) ToggleFocus(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
	query := `
		UPDATE tasks
		SET is_focus = NOT is_focus, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, title, description, status, priority, due_date, is_focus, tags, last_progress_note, last_progress_at, source_note_id, created_at, updated_at
	`
	var entity taskEntity
	err := r.exec(ctx).QueryRowxContext(ctx, query, id, userID).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle task focus: %w", err)
	}

	subtasks, _ := r.fetchSubtasks(ctx, entity.ID)
	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) CreateSubtask(ctx context.Context, userID string, taskID int64, title string) (*SubtaskResponse, error) {
	// Verify task ownership
	task, err := r.FindByID(ctx, userID, taskID)
	if err != nil || task == nil {
		return nil, errors.New("task not found or unauthorized")
	}

	query := `
		INSERT INTO subtasks (task_id, title, is_completed, order_index, created_at, updated_at)
		VALUES ($1, $2, FALSE, (SELECT COALESCE(MAX(order_index) + 1, 0) FROM subtasks WHERE task_id = $1), NOW(), NOW())
		RETURNING id, task_id, title, is_completed, order_index, created_at, updated_at
	`
	var st SubtaskResponse
	err = r.exec(ctx).QueryRowxContext(ctx, query, taskID, title).StructScan(&st)
	if err != nil {
		return nil, fmt.Errorf("failed to create subtask: %w", err)
	}
	return &st, nil
}

func (r *repository) UpdateSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error) {
	// Verify task ownership
	task, err := r.FindByID(ctx, userID, taskID)
	if err != nil || task == nil {
		return nil, errors.New("task not found or unauthorized")
	}

	var existing SubtaskResponse
	err = r.exec(ctx).GetContext(ctx, &existing, `SELECT id, task_id, title, is_completed, order_index, created_at, updated_at FROM subtasks WHERE id = $1 AND task_id = $2`, subtaskID, taskID)
	if err != nil {
		return nil, fmt.Errorf("subtask not found: %w", err)
	}

	title := existing.Title
	if req.Title != nil {
		title = *req.Title
	}
	isCompleted := existing.IsCompleted
	if req.IsCompleted != nil {
		isCompleted = *req.IsCompleted
	}

	query := `
		UPDATE subtasks
		SET title = $1, is_completed = $2, updated_at = NOW()
		WHERE id = $3 AND task_id = $4
		RETURNING id, task_id, title, is_completed, order_index, created_at, updated_at
	`
	var updated SubtaskResponse
	err = r.exec(ctx).QueryRowxContext(ctx, query, title, isCompleted, subtaskID, taskID).StructScan(&updated)
	if err != nil {
		return nil, fmt.Errorf("failed to update subtask: %w", err)
	}

	return &updated, nil
}

func (r *repository) DeleteSubtask(ctx context.Context, userID string, taskID int64, subtaskID int64) error {
	task, err := r.FindByID(ctx, userID, taskID)
	if err != nil || task == nil {
		return errors.New("task not found or unauthorized")
	}

	_, err = r.exec(ctx).ExecContext(ctx, `DELETE FROM subtasks WHERE id = $1 AND task_id = $2`, subtaskID, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete subtask: %w", err)
	}
	return nil
}
