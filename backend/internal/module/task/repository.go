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

	FindByIDSystem(ctx context.Context, id int64) (*TaskResponse, string, string, int, bool, error)
	MarkDueWarningSent(ctx context.Context, id int64) error
	GetUserThresholdAndTimezone(ctx context.Context, userID string) (int, string, error)
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
	DueDate              *time.Time     `db:"due_date"`
	DueDateWarningSentAt *time.Time     `db:"due_date_warning_sent_at"`
	IsFocus              bool           `db:"is_focus"`
	Tags                 pq.StringArray `db:"tags"`
	ProjectID            *int64         `db:"project_id"`
	ProjectName          *string        `db:"project_name"`
	ProjectColor         *string        `db:"project_color"`
	LastProgressNote     string         `db:"last_progress_note"`
	LastProgressAt       *time.Time     `db:"last_progress_at"`
	SourceNoteID         *int64         `db:"source_note_id"`
	CreatedAt            time.Time      `db:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at"`
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
		Status:               e.Status,
		Priority:             e.Priority,
		DueDate:              e.DueDate,
		DueDateWarningSentAt: e.DueDateWarningSentAt,
		IsFocus:              e.IsFocus,
		Tags:                 tags,
		ProjectID:         e.ProjectID,
		ProjectName:       e.ProjectName,
		ProjectColor:      e.ProjectColor,
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
		INSERT INTO tasks (user_id, title, description, status, priority, due_date, is_focus, tags, project_id, source_note_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, user_id, title, description, status, priority, due_date, due_date_warning_sent_at, is_focus, tags, project_id, last_progress_note, last_progress_at, source_note_id, created_at, updated_at
	`
	err := r.exec(ctx).QueryRowxContext(ctx, query, userID, req.Title, req.Description, status, priority, req.DueDate, req.IsFocus, pq.Array(tags), req.ProjectID, req.SourceNoteID).StructScan(&entity)
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

	// Fetch with project name & color if project_id is present
	if entity.ProjectID != nil {
		_ = r.exec(ctx).QueryRowxContext(ctx, "SELECT name, color FROM projects WHERE id = $1", *entity.ProjectID).Scan(&entity.ProjectName, &entity.ProjectColor)
	}

	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) FindByID(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.title, t.description, t.status, t.priority, t.due_date, t.due_date_warning_sent_at, t.is_focus, t.tags,
			t.project_id, p.name as project_name, p.color as project_color,
			t.last_progress_note, t.last_progress_at, t.source_note_id, t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1 AND t.user_id = $2
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
		SELECT 
			t.id, t.user_id, t.title, t.description, t.status, t.priority, t.due_date, t.due_date_warning_sent_at, t.is_focus, t.tags,
			t.project_id, p.name as project_name, p.color as project_color,
			t.last_progress_note, t.last_progress_at, t.source_note_id, t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN projects p ON p.id = t.project_id
		WHERE t.user_id = $1
	`
	args := []interface{}{userID}
	argIdx := 2

	if filter.ProjectID != nil {
		query += fmt.Sprintf(" AND t.project_id = $%d", argIdx)
		args = append(args, *filter.ProjectID)
		argIdx++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND t.status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.Priority != "" {
		query += fmt.Sprintf(" AND t.priority = $%d", argIdx)
		args = append(args, filter.Priority)
		argIdx++
	}

	if filter.IsFocus != nil {
		query += fmt.Sprintf(" AND t.is_focus = $%d", argIdx)
		args = append(args, *filter.IsFocus)
		argIdx++
	}

	if filter.Tag != "" {
		query += fmt.Sprintf(" AND $%d = ANY(t.tags)", argIdx)
		args = append(args, filter.Tag)
		argIdx++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (t.title ILIKE $%d OR t.description ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	if filter.Cursor != nil && *filter.Cursor > 0 {
		query += fmt.Sprintf(" AND t.id < $%d", argIdx)
		args = append(args, *filter.Cursor)
		argIdx++
	}

	if filter.Paginate == "cursor" || filter.Cursor != nil {
		query += " ORDER BY t.id DESC"
	} else {
		query += " ORDER BY t.is_focus DESC, CASE t.status WHEN 'in_progress' THEN 1 WHEN 'todo' THEN 2 WHEN 'backlog' THEN 3 WHEN 'review' THEN 4 ELSE 5 END, t.updated_at DESC, t.id DESC"
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
	var dueDateWarningSentAt = existing.DueDateWarningSentAt
	if req.DueDate != nil {
		dueDate = req.DueDate
		if (existing.DueDate == nil && req.DueDate != nil) ||
			(existing.DueDate != nil && req.DueDate == nil) ||
			(existing.DueDate != nil && req.DueDate != nil && !existing.DueDate.Equal(*req.DueDate)) {
			dueDateWarningSentAt = nil
		}
	}

	isFocus := existing.IsFocus
	if req.IsFocus != nil {
		isFocus = *req.IsFocus
	}

	tags := existing.Tags
	if req.Tags != nil {
		tags = *req.Tags
	}

	projectID := existing.ProjectID
	if req.HasProjectID {
		projectID = req.ProjectID
		if projectID != nil && *projectID <= 0 {
			projectID = nil
		}
	} else if req.ProjectID != nil {
		projectID = req.ProjectID
		if *projectID <= 0 {
			projectID = nil
		}
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
		SET title = $1, description = $2, status = $3, priority = $4, due_date = $5, due_date_warning_sent_at = $6, is_focus = $7, tags = $8, project_id = $9, last_progress_note = $10, last_progress_at = $11, updated_at = NOW()
		WHERE id = $12 AND user_id = $13
		RETURNING id, user_id, title, description, status, priority, due_date, due_date_warning_sent_at, is_focus, tags, project_id, last_progress_note, last_progress_at, source_note_id, created_at, updated_at
	`
	var entity taskEntity
	err = r.exec(ctx).QueryRowxContext(ctx, query, title, desc, status, priority, dueDate, dueDateWarningSentAt, isFocus, pq.Array(tags), projectID, lastProgressNote, lastProgressAt, id, userID).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	if entity.ProjectID != nil {
		_ = r.exec(ctx).QueryRowxContext(ctx, "SELECT name, color FROM projects WHERE id = $1", *entity.ProjectID).Scan(&entity.ProjectName, &entity.ProjectColor)
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

func (r *repository) FindByIDSystem(ctx context.Context, id int64) (*TaskResponse, string, string, int, bool, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.title, t.description, t.status, t.priority, t.due_date, t.due_date_warning_sent_at, t.is_focus, t.tags,
			t.project_id, p.name as project_name, p.color as project_color,
			t.last_progress_note, t.last_progress_at, t.source_note_id, t.created_at, t.updated_at,
			u.email, u.timezone, u.due_date_warning_threshold, u.email_notifications_enabled
		FROM tasks t
		JOIN users u ON u.id = t.user_id
		LEFT JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1
	`
	type systemEntity struct {
		taskEntity
		Email                     string `db:"email"`
		Timezone                  string `db:"timezone"`
		DueDateWarningThreshold   int    `db:"due_date_warning_threshold"`
		EmailNotificationsEnabled bool   `db:"email_notifications_enabled"`
	}
	var se systemEntity
	err := r.exec(ctx).GetContext(ctx, &se, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", "", 0, false, nil
		}
		return nil, "", "", 0, false, fmt.Errorf("failed to find task system: %w", err)
	}

	subtasks, _ := r.fetchSubtasks(ctx, se.ID)
	res := se.taskEntity.toResponse(subtasks)
	tz := se.Timezone
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	threshold := se.DueDateWarningThreshold
	if threshold <= 0 {
		threshold = 1440
	}
	return &res, se.Email, tz, threshold, se.EmailNotificationsEnabled, nil
}

func (r *repository) MarkDueWarningSent(ctx context.Context, id int64) error {
	query := `UPDATE tasks SET due_date_warning_sent_at = NOW() WHERE id = $1`
	_, err := r.exec(ctx).ExecContext(ctx, query, id)
	return err
}

func (r *repository) GetUserThresholdAndTimezone(ctx context.Context, userID string) (int, string, error) {
	query := `SELECT due_date_warning_threshold, timezone FROM users WHERE id = $1`
	var row struct {
		Threshold int    `db:"due_date_warning_threshold"`
		Timezone  string `db:"timezone"`
	}
	err := r.exec(ctx).GetContext(ctx, &row, query, userID)
	if err != nil {
		return 1440, "Asia/Jakarta", nil
	}
	if row.Threshold <= 0 {
		row.Threshold = 1440
	}
	if row.Timezone == "" {
		row.Timezone = "Asia/Jakarta"
	}
	return row.Threshold, row.Timezone, nil
}

