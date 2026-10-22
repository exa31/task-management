package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"


	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Repository interface {
	Create(ctx context.Context, tx *sqlx.Tx, userID string, req CreateTaskRequest) (*TaskResponse, error)
	FindByID(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*TaskResponse, error)
	FindAll(ctx context.Context, tx *sqlx.Tx, userID string, filter TaskFilter) ([]TaskResponse, error)
	Update(ctx context.Context, tx *sqlx.Tx, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error)
	Delete(ctx context.Context, tx *sqlx.Tx, userID string, id int64) error
	LogProgress(ctx context.Context, tx *sqlx.Tx, userID string, id int64, req LogProgressRequest) (*TaskResponse, error)
	ToggleFocus(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*TaskResponse, error)

	CreateSubtask(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, req CreateSubtaskRequest) (*SubtaskResponse, error)
	UpdateSubtask(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error)
	DeleteSubtask(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, subtaskID int64) error
	CountSubtasks(ctx context.Context, tx *sqlx.Tx, taskID int64) (int, int, error)

	CreateActivity(ctx context.Context, tx *sqlx.Tx, taskID int64, userID string, activityType string, description string, changes []byte) (*TaskActivityResponse, error)
	FindActivitiesByTaskID(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64) ([]TaskActivityResponse, error)
	FindActivitiesByUser(ctx context.Context, tx *sqlx.Tx, userID string, from, to *time.Time, limit ...int) ([]TaskActivityResponse, error)
	FindActivityCountsByUser(ctx context.Context, tx *sqlx.Tx, userID string, from, to *time.Time, tz string) ([]ActivityDateCount, error)

	FindByIDSystem(ctx context.Context, tx *sqlx.Tx, id int64) (*TaskResponse, string, string, int, bool, error)
	MarkDueWarningSent(ctx context.Context, tx *sqlx.Tx, id int64) error
	GetUserThresholdAndTimezone(ctx context.Context, tx *sqlx.Tx, userID string) (int, string, error)

	GetWidgetBundle(ctx context.Context, tx *sqlx.Tx, userID string) (*WidgetBundleResponse, error)
	GetWorkspaceStats(ctx context.Context, tx *sqlx.Tx, userID string) (*WorkspaceStatsResponse, error)
	UpdateRecurrenceGeneratedAt(ctx context.Context, tx *sqlx.Tx, id int64, t time.Time) error

	// Recurring Tasks repository methods
	FindRecurringByID(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*RecurringTaskResponse, error)
	FindRecurringByIDSystem(ctx context.Context, tx *sqlx.Tx, id int64) (*RecurringTaskResponse, string, error)
	FindAllRecurring(ctx context.Context, tx *sqlx.Tx, userID string) ([]RecurringTaskResponse, error)
	CreateRecurring(ctx context.Context, tx *sqlx.Tx, userID string, req CreateRecurringTaskRequest) (*RecurringTaskResponse, error)
	UpdateRecurring(ctx context.Context, tx *sqlx.Tx, userID string, id int64, req UpdateRecurringTaskRequest) (*RecurringTaskResponse, error)
	DeleteRecurring(ctx context.Context, tx *sqlx.Tx, userID string, id int64) error
	ToggleActiveRecurring(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*RecurringTaskResponse, error)
	FindTodayInstance(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64, todayDate time.Time) (*TaskResponse, error)
	CreateInstanceFromRecurring(ctx context.Context, tx *sqlx.Tx, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error)
	UpdateRecurringLastGenerated(ctx context.Context, tx *sqlx.Tx, id int64, t time.Time) error
	CreateRecurringSubtask(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64, title, description string, orderIndex int) error
	UpdateRecurringSubtaskByTitle(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64, oldTitle, newTitle, description string) error
	DeleteRecurringSubtaskByTitle(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64, title string) error
	SyncRecurringFromTask(ctx context.Context, tx *sqlx.Tx, userID string, task *TaskResponse) error
	ConvertTaskToRecurring(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, pattern string, interval int, endDate *time.Time) (*RecurringTaskResponse, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}


type taskEntity struct {
	ID                        int64          `db:"id"`
	UserID                    string         `db:"user_id"`
	Title                     string         `db:"title"`
	Description               string         `db:"description"`
	Status                    string         `db:"status"`
	Priority                  string         `db:"priority"`
	DueDate                   *time.Time     `db:"due_date"`
	DueDateWarningSentAt      *time.Time     `db:"due_date_warning_sent_at"`
	IsFocus                   bool           `db:"is_focus"`
	Tags                      pq.StringArray `db:"tags"`
	ProjectID                 *int64         `db:"project_id"`
	ProjectName               *string        `db:"project_name"`
	ProjectColor              *string        `db:"project_color"`
	LastProgressNote          string         `db:"last_progress_note"`
	LastProgressAt            *time.Time     `db:"last_progress_at"`
	SourceNoteID              *int64         `db:"source_note_id"`
	IsRecurring               bool           `db:"is_recurring"`
	RecurrencePattern         string         `db:"recurrence_pattern"`
	RecurrenceInterval        int            `db:"recurrence_interval"`
	RecurrenceEndDate         *time.Time     `db:"recurrence_end_date"`
	RecurrenceLastGeneratedAt *time.Time     `db:"recurrence_last_generated_at"`
	RecurringTaskID           *int64         `db:"recurring_task_id"`
	CreatedAt                 time.Time      `db:"created_at"`
	UpdatedAt                 time.Time      `db:"updated_at"`
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
		ID:                        e.ID,
		UserID:                    e.UserID,
		Title:                     e.Title,
		Description:               e.Description,
		Status:                    e.Status,
		Priority:                  e.Priority,
		DueDate:                   e.DueDate,
		DueDateWarningSentAt:      e.DueDateWarningSentAt,
		IsFocus:                   e.IsFocus,
		Tags:                      tags,
		ProjectID:                 e.ProjectID,
		ProjectName:               e.ProjectName,
		ProjectColor:              e.ProjectColor,
		LastProgressNote:          e.LastProgressNote,
		LastProgressAt:            e.LastProgressAt,
		SourceNoteID:              e.SourceNoteID,
		Subtasks:                  subtasks,
		SubtasksCount:             totalCount,
		SubtasksCompleted:         completedCount,
		ProgressPercent:           percent,
		IsRecurring:               e.IsRecurring || e.RecurringTaskID != nil,
		RecurrencePattern:         e.RecurrencePattern,
		RecurrenceInterval:        e.RecurrenceInterval,
		RecurrenceEndDate:         e.RecurrenceEndDate,
		RecurrenceLastGeneratedAt: e.RecurrenceLastGeneratedAt,
		RecurringTaskID:           e.RecurringTaskID,
		CreatedAt:                 e.CreatedAt,
		UpdatedAt:                 e.UpdatedAt,
	}
}

type recurringTaskEntity struct {
	ID                 int64      `db:"id"`
	UserID             string     `db:"user_id"`
	ProjectID          *int64     `db:"project_id"`
	ProjectName        *string    `db:"project_name"`
	ProjectColor       *string    `db:"project_color"`
	Title              string     `db:"title"`
	Description        string     `db:"description"`
	Priority           string     `db:"priority"`
	RecurrencePattern  string     `db:"recurrence_pattern"`
	RecurrenceInterval int        `db:"recurrence_interval"`
	DueTime            *string    `db:"due_time"`
	RecurrenceEndDate  *time.Time `db:"recurrence_end_date"`
	IsActive           bool       `db:"is_active"`
	LastGeneratedAt    *time.Time `db:"last_generated_at"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
}

func (e *recurringTaskEntity) toResponse(subtasks []RecurringSubtaskResponse) RecurringTaskResponse {
	if subtasks == nil {
		subtasks = []RecurringSubtaskResponse{}
	}
	return RecurringTaskResponse{
		ID:                 e.ID,
		UserID:             e.UserID,
		ProjectID:          e.ProjectID,
		ProjectName:        e.ProjectName,
		ProjectColor:       e.ProjectColor,
		Title:              e.Title,
		Description:        e.Description,
		Priority:           e.Priority,
		RecurrencePattern:  e.RecurrencePattern,
		RecurrenceInterval: e.RecurrenceInterval,
		DueTime:            e.DueTime,
		RecurrenceEndDate:  e.RecurrenceEndDate,
		IsActive:           e.IsActive,
		LastGeneratedAt:    e.LastGeneratedAt,
		Subtasks:           subtasks,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
}

func (r *repository) fetchSubtasks(ctx context.Context, tx *sqlx.Tx, taskID int64) ([]SubtaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		SELECT id, task_id, title, COALESCE(description, '') AS description, is_completed, order_index, due_date, created_at, updated_at
		FROM subtasks
		WHERE task_id = $1
		ORDER BY order_index ASC, id ASC
	`
	var items []SubtaskResponse
	err := tx.SelectContext(ctx, &items, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subtasks: %w", err)
	}

	if items == nil {
		items = []SubtaskResponse{}
	}
	return items, nil
}

func (r *repository) Create(ctx context.Context, tx *sqlx.Tx, userID string, req CreateTaskRequest) (*TaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
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
	interval := req.RecurrenceInterval
	if interval <= 0 {
		interval = 1
	}

	var entity taskEntity
	subtasks := make([]SubtaskResponse, 0, len(req.Subtasks))

	query := `
		INSERT INTO tasks (
			user_id, title, description, status, priority, due_date, is_focus, tags, project_id, source_note_id,
			is_recurring, recurrence_pattern, recurrence_interval, recurrence_end_date, recurring_task_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW())
		RETURNING id, user_id, title, description, status, priority, due_date, due_date_warning_sent_at, is_focus, tags,
			project_id, last_progress_note, last_progress_at, source_note_id,
			is_recurring, recurrence_pattern, recurrence_interval, recurrence_end_date, recurrence_last_generated_at,
			recurring_task_id, created_at, updated_at
	`
	err := tx.QueryRowxContext(ctx, query, userID, req.Title, req.Description, status, priority, req.DueDate, req.IsFocus, pq.Array(tags), req.ProjectID, req.SourceNoteID, req.IsRecurring, req.RecurrencePattern, interval, req.RecurrenceEndDate, req.RecurringTaskID).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	if len(req.Subtasks) > 0 {
		var validTitles []string
		for _, stTitle := range req.Subtasks {
			if strings.TrimSpace(stTitle) != "" {
				validTitles = append(validTitles, stTitle)
			}
		}
		if len(validTitles) > 0 {
			var valueStrings []string
			var valueArgs []interface{}
			argIdx := 1
			for i, stTitle := range validTitles {
				valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, '', FALSE, $%d, NOW(), NOW())", argIdx, argIdx+1, argIdx+2))
				valueArgs = append(valueArgs, entity.ID, stTitle, i)
				argIdx += 3
			}
			batchQuery := fmt.Sprintf(`
				INSERT INTO subtasks (task_id, title, description, is_completed, order_index, created_at, updated_at)
				VALUES %s
				RETURNING id, task_id, title, COALESCE(description, '') AS description, is_completed, order_index, due_date, created_at, updated_at
			`, strings.Join(valueStrings, ", "))

			rows, err := tx.QueryxContext(ctx, batchQuery, valueArgs...)
			if err != nil {
				return nil, fmt.Errorf("failed to batch create subtasks: %w", err)
			}
			defer rows.Close()
			for rows.Next() {
				var st SubtaskResponse
				if err := rows.StructScan(&st); err != nil {
					return nil, fmt.Errorf("failed to scan created subtask: %w", err)
				}
				subtasks = append(subtasks, st)
			}
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("error reading batch created subtasks: %w", err)
			}
		}
	}

	// Fetch with project name & color if project_id is present
	if entity.ProjectID != nil {
		_ = tx.QueryRowxContext(ctx, "SELECT name, color FROM projects WHERE id = $1", *entity.ProjectID).Scan(&entity.ProjectName, &entity.ProjectColor)
	}

	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) FindByID(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*TaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		SELECT 
			t.id, t.user_id, t.title, t.description, t.status, t.priority, t.due_date, t.due_date_warning_sent_at, t.is_focus, t.tags,
			t.project_id, p.name as project_name, p.color as project_color,
			t.last_progress_note, t.last_progress_at, t.source_note_id,
			t.is_recurring, t.recurrence_pattern, t.recurrence_interval, t.recurrence_end_date, t.recurrence_last_generated_at,
			t.recurring_task_id, t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1 AND t.user_id = $2
	`
	var entity taskEntity
	err := tx.GetContext(ctx, &entity, query, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	subtasks, err := r.fetchSubtasks(ctx, tx, entity.ID)
	if err != nil {
		return nil, err
	}

	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) FindAll(ctx context.Context, tx *sqlx.Tx, userID string, filter TaskFilter) ([]TaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		SELECT 
			t.id, t.user_id, t.title, t.description, t.status, t.priority, t.due_date, t.due_date_warning_sent_at, t.is_focus, t.tags,
			t.project_id, p.name as project_name, p.color as project_color,
			t.last_progress_note, t.last_progress_at, t.source_note_id,
			t.is_recurring, t.recurrence_pattern, t.recurrence_interval, t.recurrence_end_date, t.recurrence_last_generated_at,
			t.recurring_task_id, t.created_at, t.updated_at
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

	if filter.From != nil {
		query += fmt.Sprintf(" AND t.due_date >= $%d", argIdx)
		args = append(args, *filter.From)
		argIdx++
	}

	if filter.To != nil {
		query += fmt.Sprintf(" AND t.due_date <= $%d", argIdx)
		args = append(args, *filter.To)
		argIdx++
	}

	offset := filter.Offset
	if offset == 0 && filter.Cursor != nil && *filter.Cursor > 0 {
		offset = int(*filter.Cursor)
	}

	query += " ORDER BY (CASE WHEN t.status = 'done' THEN 1 ELSE 0 END) ASC, (CASE WHEN t.is_focus AND t.status != 'done' THEN 1 ELSE 0 END) DESC, CASE t.status WHEN 'in_progress' THEN 1 WHEN 'todo' THEN 2 WHEN 'review' THEN 3 WHEN 'backlog' THEN 4 ELSE 5 END, CASE t.priority WHEN 'urgent' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END, t.updated_at DESC, t.id DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, offset)
	}

	var entities []taskEntity
	err := tx.SelectContext(ctx, &entities, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	responses := make([]TaskResponse, 0, len(entities))
	for i := range entities {
		subtasks, _ := r.fetchSubtasks(ctx, tx, entities[i].ID)
		responses = append(responses, entities[i].toResponse(subtasks))
	}
	return responses, nil
}

func (r *repository) Update(ctx context.Context, tx *sqlx.Tx, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	existing, err := r.FindByID(ctx, tx, userID, id)
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
	if req.ClearDescription {
		desc = ""
	} else if req.Description != nil {
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

	isRecurring := existing.IsRecurring
	if req.IsRecurring != nil {
		isRecurring = *req.IsRecurring
	}
	recurrencePattern := existing.RecurrencePattern
	if req.RecurrencePattern != nil {
		recurrencePattern = *req.RecurrencePattern
	}
	recurrenceInterval := existing.RecurrenceInterval
	if req.RecurrenceInterval != nil {
		recurrenceInterval = *req.RecurrenceInterval
	}
	if recurrenceInterval <= 0 {
		recurrenceInterval = 1
	}
	recurrenceEndDate := existing.RecurrenceEndDate
	if req.ClearRecurrenceEndDate {
		recurrenceEndDate = nil
	} else if req.HasRecurrenceEndDate {
		recurrenceEndDate = req.RecurrenceEndDate
	} else if req.RecurrenceEndDate != nil {
		recurrenceEndDate = req.RecurrenceEndDate
	}
	recurringTaskID := existing.RecurringTaskID
	if req.RecurringTaskID != nil {
		recurringTaskID = req.RecurringTaskID
	}

	dueDate := existing.DueDate
	var dueDateWarningSentAt = existing.DueDateWarningSentAt
	if req.ClearDueDate {
		dueDate = nil
		dueDateWarningSentAt = nil
	} else if req.HasDueDate {
		dueDate = req.DueDate
		if (existing.DueDate == nil && req.DueDate != nil) ||
			(existing.DueDate != nil && req.DueDate == nil) ||
			(existing.DueDate != nil && req.DueDate != nil && !existing.DueDate.Equal(*req.DueDate)) {
			dueDateWarningSentAt = nil
		}
	} else if req.DueDate != nil {
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
	// When task is marked done, automatically clear focus unless explicitly specified
	if status == "done" && req.IsFocus == nil {
		isFocus = false
	}

	tags := existing.Tags
	if req.Tags != nil {
		tags = *req.Tags
	}

	projectID := existing.ProjectID
	if req.ClearProjectID {
		projectID = nil
	} else if req.HasProjectID {
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
		SET title = $1, description = $2, status = $3, priority = $4, due_date = $5, due_date_warning_sent_at = $6,
			is_focus = $7, tags = $8, project_id = $9, last_progress_note = $10, last_progress_at = $11,
			is_recurring = $12, recurrence_pattern = $13, recurrence_interval = $14, recurrence_end_date = $15,
			recurring_task_id = $16, updated_at = NOW()
		WHERE id = $17 AND user_id = $18
		RETURNING id, user_id, title, description, status, priority, due_date, due_date_warning_sent_at, is_focus, tags,
			project_id, last_progress_note, last_progress_at, source_note_id,
			is_recurring, recurrence_pattern, recurrence_interval, recurrence_end_date, recurrence_last_generated_at,
			recurring_task_id, created_at, updated_at
	`
	var entity taskEntity
	err = tx.QueryRowxContext(ctx, query, title, desc, status, priority, dueDate, dueDateWarningSentAt, isFocus, pq.Array(tags), projectID, lastProgressNote, lastProgressAt, isRecurring, recurrencePattern, recurrenceInterval, recurrenceEndDate, recurringTaskID, id, userID).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	if entity.ProjectID != nil {
		_ = tx.QueryRowxContext(ctx, "SELECT name, color FROM projects WHERE id = $1", *entity.ProjectID).Scan(&entity.ProjectName, &entity.ProjectColor)
	}

	if entity.Status == "done" {
		// Automatically mark all subtasks of completed task as completed
		if _, err := tx.ExecContext(ctx, `UPDATE subtasks SET is_completed = true, updated_at = NOW() WHERE task_id = $1`, entity.ID); err != nil {
			return nil, fmt.Errorf("failed to complete subtasks: %w", err)
		}
	}

	subtasks, _ := r.fetchSubtasks(ctx, tx, entity.ID)
	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) Delete(ctx context.Context, tx *sqlx.Tx, userID string, id int64) error {
	if tx == nil {
		return errors.New("transaction is required")
	}
	query := `DELETE FROM tasks WHERE id = $1 AND user_id = $2`
	_, err := tx.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (r *repository) LogProgress(ctx context.Context, tx *sqlx.Tx, userID string, id int64, req LogProgressRequest) (*TaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	statusUpdate := ""
	args := []interface{}{req.ProgressNote, id, userID}
	if req.Status != nil && *req.Status != "" {
		if *req.Status == "done" {
			statusUpdate = ", status = $4, is_focus = false"
		} else {
			statusUpdate = ", status = $4"
		}
		args = append(args, *req.Status)
	}

	query := fmt.Sprintf(`
		UPDATE tasks
		SET last_progress_note = $1, last_progress_at = NOW(), updated_at = NOW()%s
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, title, description, status, priority, due_date, due_date_warning_sent_at, is_focus, tags,
			project_id, last_progress_note, last_progress_at, source_note_id,
			is_recurring, recurrence_pattern, recurrence_interval, recurrence_end_date, recurrence_last_generated_at,
			recurring_task_id, created_at, updated_at
	`, statusUpdate)

	var entity taskEntity

	err := tx.QueryRowxContext(ctx, query, args...).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to log progress: %w", err)
	}

	if len(req.CompletedSubtaskIDs) > 0 {
		_, err := tx.ExecContext(ctx, `
			UPDATE subtasks SET is_completed = TRUE, updated_at = NOW()
			WHERE task_id = $1 AND id = ANY($2)
		`, id, pq.Array(req.CompletedSubtaskIDs))
		if err != nil {
			return nil, fmt.Errorf("failed to mark completed subtasks: %w", err)
		}
	}

	if entity.ProjectID != nil {
		_ = tx.QueryRowxContext(ctx, "SELECT name, color FROM projects WHERE id = $1", *entity.ProjectID).Scan(&entity.ProjectName, &entity.ProjectColor)
	}

	if entity.Status == "done" {
		if _, err := tx.ExecContext(ctx, `UPDATE subtasks SET is_completed = true, updated_at = NOW() WHERE task_id = $1`, entity.ID); err != nil {
			return nil, fmt.Errorf("failed to complete subtasks: %w", err)
		}
	}

	subtasks, _ := r.fetchSubtasks(ctx, tx, entity.ID)
	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) ToggleFocus(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*TaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		UPDATE tasks
		SET is_focus = NOT is_focus, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, title, description, status, priority, due_date, due_date_warning_sent_at, is_focus, tags,
			project_id, last_progress_note, last_progress_at, source_note_id,
			is_recurring, recurrence_pattern, recurrence_interval, recurrence_end_date, recurrence_last_generated_at,
			recurring_task_id, created_at, updated_at
	`
	var entity taskEntity
	err := tx.QueryRowxContext(ctx, query, id, userID).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle task focus: %w", err)
	}

	if entity.ProjectID != nil {
		_ = tx.QueryRowxContext(ctx, "SELECT name, color FROM projects WHERE id = $1", *entity.ProjectID).Scan(&entity.ProjectName, &entity.ProjectColor)
	}

	subtasks, _ := r.fetchSubtasks(ctx, tx, entity.ID)
	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) CreateSubtask(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, req CreateSubtaskRequest) (*SubtaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	// Verify task ownership
	task, err := r.FindByID(ctx, tx, userID, taskID)
	if err != nil || task == nil {
		return nil, errors.New("task not found or unauthorized")
	}

	query := `
		INSERT INTO subtasks (task_id, title, description, is_completed, order_index, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, FALSE, (SELECT COALESCE(MAX(order_index) + 1, 0) FROM subtasks WHERE task_id = $1), $4, NOW(), NOW())
		RETURNING id, task_id, title, COALESCE(description, '') AS description, is_completed, order_index, due_date, created_at, updated_at
	`
	var st SubtaskResponse
	err = tx.QueryRowxContext(ctx, query, taskID, req.Title, req.Description, req.DueDate).StructScan(&st)
	if err != nil {
		return nil, fmt.Errorf("failed to create subtask: %w", err)
	}
	return &st, nil
}

func (r *repository) UpdateSubtask(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, subtaskID int64, req UpdateSubtaskRequest) (*SubtaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	// Verify task ownership
	task, err := r.FindByID(ctx, tx, userID, taskID)
	if err != nil || task == nil {
		return nil, errors.New("task not found or unauthorized")
	}

	var existing SubtaskResponse
	err = tx.GetContext(ctx, &existing, `SELECT id, task_id, title, COALESCE(description, '') AS description, is_completed, order_index, due_date, created_at, updated_at FROM subtasks WHERE id = $1 AND task_id = $2`, subtaskID, taskID)
	if err != nil {
		return nil, fmt.Errorf("subtask not found: %w", err)
	}

	title := existing.Title
	if req.Title != nil {
		title = *req.Title
	}
	description := existing.Description
	if req.ClearDescription {
		description = ""
	} else if req.HasDescription {
		if req.Description != nil {
			description = *req.Description
		} else {
			description = ""
		}
	} else if req.Description != nil {
		description = *req.Description
	}
	isCompleted := existing.IsCompleted
	if req.IsCompleted != nil {
		isCompleted = *req.IsCompleted
	}
	dueDate := existing.DueDate
	if req.ClearDueDate {
		dueDate = nil
	} else if req.HasDueDate {
		dueDate = req.DueDate
	} else if req.DueDate != nil {
		dueDate = req.DueDate
	}

	query := `
		UPDATE subtasks
		SET title = $1, description = $2, is_completed = $3, due_date = $4, updated_at = NOW()
		WHERE id = $5 AND task_id = $6
		RETURNING id, task_id, title, COALESCE(description, '') AS description, is_completed, order_index, due_date, created_at, updated_at
	`
	var updated SubtaskResponse
	err = tx.QueryRowxContext(ctx, query, title, description, isCompleted, dueDate, subtaskID, taskID).StructScan(&updated)
	if err != nil {
		return nil, fmt.Errorf("failed to update subtask: %w", err)
	}

	return &updated, nil
}

func (r *repository) DeleteSubtask(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, subtaskID int64) error {
	if tx == nil {
		return errors.New("transaction is required")
	}
	task, err := r.FindByID(ctx, tx, userID, taskID)
	if err != nil || task == nil {
		return errors.New("task not found or unauthorized")
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM subtasks WHERE id = $1 AND task_id = $2`, subtaskID, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete subtask: %w", err)
	}
	return nil
}

func (r *repository) CountSubtasks(ctx context.Context, tx *sqlx.Tx, taskID int64) (int, int, error) {
	if tx == nil {
		return 0, 0, errors.New("transaction is required")
	}
	type sc struct {
		Total     int `db:"total"`
		Completed int `db:"completed"`
	}
	var res sc
	err := tx.GetContext(ctx, &res, `
		SELECT COUNT(*) as total, COALESCE(COUNT(*) FILTER (WHERE is_completed = true), 0) as completed
		FROM subtasks WHERE task_id = $1
	`, taskID)
	return res.Total, res.Completed, err
}

func (r *repository) FindByIDSystem(ctx context.Context, tx *sqlx.Tx, id int64) (*TaskResponse, string, string, int, bool, error) {
	if tx == nil {
		return nil, "", "", 0, false, errors.New("transaction is required")
	}
	query := `
		SELECT 
			t.id, t.user_id, t.title, t.description, t.status, t.priority, t.due_date, t.due_date_warning_sent_at, t.is_focus, t.tags,
			t.project_id, p.name as project_name, p.color as project_color,
			t.last_progress_note, t.last_progress_at, t.source_note_id,
			t.is_recurring, t.recurrence_pattern, t.recurrence_interval, t.recurrence_end_date, t.recurrence_last_generated_at,
			t.recurring_task_id, t.created_at, t.updated_at,
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
	err := tx.GetContext(ctx, &se, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", "", 0, false, nil
		}
		return nil, "", "", 0, false, fmt.Errorf("failed to find task system: %w", err)
	}

	subtasks, _ := r.fetchSubtasks(ctx, tx, se.ID)
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

func (r *repository) GetWidgetBundle(ctx context.Context, tx *sqlx.Tx, userID string) (*WidgetBundleResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	allTasks, err := r.FindAll(ctx, tx, userID, TaskFilter{
		Limit: 20,
	})
	if err != nil {
		return nil, err
	}
	var activeFocus []TaskResponse
	seen := make(map[int64]bool)
	// 1. Prioritize Focus & Recurring tasks first
	for _, t := range allTasks {
		if t.Status != "done" && (t.IsFocus || t.IsRecurring || t.RecurringTaskID != nil) {
			activeFocus = append(activeFocus, t)
			seen[t.ID] = true
		}
	}
	// 2. Append remaining active tasks (in_progress, todo, review, backlog)
	for _, t := range allTasks {
		if t.Status != "done" && !seen[t.ID] {
			activeFocus = append(activeFocus, t)
			seen[t.ID] = true
		}
	}

	upcomingQuery := `
		SELECT 
			t.id, t.user_id, t.title, t.description, t.status, t.priority, t.due_date, t.due_date_warning_sent_at, t.is_focus, t.tags,
			t.project_id, p.name as project_name, p.color as project_color,
			t.last_progress_note, t.last_progress_at, t.source_note_id,
			t.is_recurring, t.recurrence_pattern, t.recurrence_interval, t.recurrence_end_date, t.recurrence_last_generated_at,
			t.recurring_task_id, t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN projects p ON p.id = t.project_id
		WHERE t.user_id = $1 AND t.status != 'done' AND t.due_date IS NOT NULL
		ORDER BY t.due_date ASC
		LIMIT 10
	`
	var deadlineEntities []taskEntity
	_ = tx.SelectContext(ctx, &deadlineEntities, upcomingQuery, userID)
	upcoming := make([]TaskResponse, len(deadlineEntities))
	for i, de := range deadlineEntities {
		st, _ := r.fetchSubtasks(ctx, tx, de.ID)
		upcoming[i] = de.toResponse(st)
	}

	var completedToday, totalToday int
	_ = tx.GetContext(ctx, &completedToday, `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND status = 'done' AND updated_at >= CURRENT_DATE`, userID)
	_ = tx.GetContext(ctx, &totalToday, `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND (created_at >= CURRENT_DATE OR (due_date IS NOT NULL AND due_date::date = CURRENT_DATE))`, userID)

	return &WidgetBundleResponse{
		FocusTasks:        activeFocus,
		UpcomingDeadlines: upcoming,
		CompletedToday:    completedToday,
		TotalToday:        totalToday,
	}, nil
}

func (r *repository) UpdateRecurrenceGeneratedAt(ctx context.Context, tx *sqlx.Tx, id int64, t time.Time) error {
	if tx == nil {
		return errors.New("transaction is required")
	}
	query := `UPDATE tasks SET recurrence_last_generated_at = $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, t, id)
	return err
}

func boolPtr(b bool) *bool {
	return &b
}

func (r *repository) MarkDueWarningSent(ctx context.Context, tx *sqlx.Tx, id int64) error {
	if tx == nil {
		return errors.New("transaction is required")
	}
	query := `UPDATE tasks SET due_date_warning_sent_at = NOW() WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id)
	return err
}

func (r *repository) GetUserThresholdAndTimezone(ctx context.Context, tx *sqlx.Tx, userID string) (int, string, error) {
	if tx == nil {
		return 0, "", errors.New("transaction is required")
	}
	query := `SELECT due_date_warning_threshold, timezone FROM users WHERE id = $1`
	var row struct {
		Threshold int    `db:"due_date_warning_threshold"`
		Timezone  string `db:"timezone"`
	}
	err := tx.GetContext(ctx, &row, query, userID)
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

type taskActivityEntity struct {
	ID           int64     `db:"id"`
	TaskID       int64     `db:"task_id"`
	UserID       string    `db:"user_id"`
	ActivityType string    `db:"activity_type"`
	Description  string    `db:"description"`
	Changes      []byte    `db:"changes"`
	CreatedAt    time.Time `db:"created_at"`
}

func (e *taskActivityEntity) toResponse() TaskActivityResponse {
	raw := json.RawMessage(e.Changes)
	if len(raw) == 0 {
		raw = json.RawMessage("{}")
	}
	return TaskActivityResponse{
		ID:           e.ID,
		TaskID:       e.TaskID,
		UserID:       e.UserID,
		ActivityType: e.ActivityType,
		Description:  e.Description,
		Changes:      raw,
		CreatedAt:    e.CreatedAt,
	}
}

func (r *repository) CreateActivity(ctx context.Context, tx *sqlx.Tx, taskID int64, userID string, activityType string, description string, changes []byte) (*TaskActivityResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	if len(changes) == 0 {
		changes = []byte("{}")
	}
	query := `
		INSERT INTO task_activities (task_id, user_id, activity_type, description, changes, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, task_id, user_id, activity_type, description, changes, created_at
	`
	var entity taskActivityEntity
	err := tx.QueryRowxContext(ctx, query, taskID, userID, activityType, description, changes).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create task activity: %w", err)
	}
	res := entity.toResponse()
	return &res, nil
}

func (r *repository) FindActivitiesByTaskID(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64) ([]TaskActivityResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		SELECT ta.id, ta.task_id, ta.user_id, ta.activity_type, ta.description, ta.changes, ta.created_at
		FROM task_activities ta
		INNER JOIN tasks t ON t.id = ta.task_id
		WHERE ta.task_id = $1 AND t.user_id = $2
		ORDER BY ta.created_at DESC, ta.id DESC
	`
	var rows []taskActivityEntity
	err := tx.SelectContext(ctx, &rows, query, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list task activities: %w", err)
	}
	res := make([]TaskActivityResponse, len(rows))
	for i, row := range rows {
		res[i] = row.toResponse()
	}
	return res, nil
}

func (r *repository) FindActivitiesByUser(ctx context.Context, tx *sqlx.Tx, userID string, from, to *time.Time, limit ...int) ([]TaskActivityResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		SELECT ta.id, ta.task_id, COALESCE(t.title, '') AS task_title, ta.user_id, ta.activity_type, ta.description, ta.changes, ta.created_at
		FROM task_activities ta
		LEFT JOIN tasks t ON t.id = ta.task_id
		WHERE ta.user_id = $1
	`
	args := []interface{}{userID}
	argIdx := 2

	if from != nil {
		query += fmt.Sprintf(" AND ta.created_at >= $%d", argIdx)
		args = append(args, *from)
		argIdx++
	}
	if to != nil {
		query += fmt.Sprintf(" AND ta.created_at <= $%d", argIdx)
		args = append(args, *to)
		argIdx++
	}

	query += " ORDER BY ta.created_at DESC, ta.id DESC"
	if len(limit) > 0 && limit[0] > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit[0])
	}

	type activityWithTitleEntity struct {
		ID           int64     `db:"id"`
		TaskID       int64     `db:"task_id"`
		TaskTitle    string    `db:"task_title"`
		UserID       string    `db:"user_id"`
		ActivityType string    `db:"activity_type"`
		Description  string    `db:"description"`
		Changes      []byte    `db:"changes"`
		CreatedAt    time.Time `db:"created_at"`
	}

	var rows []activityWithTitleEntity
	err := tx.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list user activities: %w", err)
	}

	res := make([]TaskActivityResponse, len(rows))
	for i, row := range rows {
		raw := json.RawMessage(row.Changes)
		if len(raw) == 0 {
			raw = json.RawMessage("{}")
		}
		res[i] = TaskActivityResponse{
			ID:           row.ID,
			TaskID:       row.TaskID,
			TaskTitle:    row.TaskTitle,
			UserID:       row.UserID,
			ActivityType: row.ActivityType,
			Description:  row.Description,
			Changes:      raw,
			CreatedAt:    row.CreatedAt,
		}
	}
	return res, nil
}

func (r *repository) FindActivityCountsByUser(ctx context.Context, tx *sqlx.Tx, userID string, from, to *time.Time, tz string) ([]ActivityDateCount, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	query := `
		SELECT TO_CHAR(ta.created_at AT TIME ZONE $2, 'YYYY-MM-DD') AS date, COUNT(*)::int AS count
		FROM task_activities ta
		WHERE ta.user_id = $1
	`
	args := []interface{}{userID, tz}
	argIdx := 3

	if from != nil {
		query += fmt.Sprintf(" AND ta.created_at >= $%d", argIdx)
		args = append(args, *from)
		argIdx++
	}
	if to != nil {
		query += fmt.Sprintf(" AND ta.created_at <= $%d", argIdx)
		args = append(args, *to)
		argIdx++
	}

	query += " GROUP BY date ORDER BY date ASC"

	var rows []ActivityDateCount
	err := tx.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		fallbackQuery := `
			SELECT TO_CHAR(ta.created_at, 'YYYY-MM-DD') AS date, COUNT(*)::int AS count
			FROM task_activities ta
			WHERE ta.user_id = $1
		`
		argsFb := []interface{}{userID}
		argIdxFb := 2
		if from != nil {
			fallbackQuery += fmt.Sprintf(" AND ta.created_at >= $%d", argIdxFb)
			argsFb = append(argsFb, *from)
			argIdxFb++
		}
		if to != nil {
			fallbackQuery += fmt.Sprintf(" AND ta.created_at <= $%d", argIdxFb)
			argsFb = append(argsFb, *to)
			argIdxFb++
		}
		fallbackQuery += " GROUP BY date ORDER BY date ASC"
		errFb := tx.SelectContext(ctx, &rows, fallbackQuery, argsFb...)
		if errFb != nil {
			return nil, fmt.Errorf("failed to count user activities: %w", errFb)
		}
	}
	return rows, nil
}

func (r *repository) GetWorkspaceStats(ctx context.Context, tx *sqlx.Tx, userID string) (*WorkspaceStatsResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		SELECT
			COALESCE((SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND status = 'in_progress'), 0) AS tasks_in_progress,
			COALESCE((SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND status = 'todo'), 0) AS tasks_todo,
			COALESCE((SELECT COUNT(*) FROM tasks WHERE user_id = $1), 0) AS tasks_total,
			COALESCE((SELECT COUNT(*) FROM projects WHERE user_id = $1), 0) AS total_projects,
			COALESCE((SELECT COUNT(*) FROM reminders WHERE user_id = $1 AND status = 'pending'), 0) AS pending_reminders,
			COALESCE((SELECT COUNT(*) FROM notes WHERE user_id = $1), 0) AS total_notes
	`
	var stats WorkspaceStatsResponse
	err := tx.GetContext(ctx, &stats, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workspace stats: %w", err)
	}
	return &stats, nil
}

// --- Recurring Tasks Repository Implementation ---

func (r *repository) fetchRecurringSubtasks(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64) ([]RecurringSubtaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		SELECT id, recurring_task_id, title, COALESCE(description, '') AS description, order_index, created_at
		FROM recurring_subtasks
		WHERE recurring_task_id = $1
		ORDER BY order_index ASC, id ASC
	`
	var items []RecurringSubtaskResponse
	err := tx.SelectContext(ctx, &items, query, recurringTaskID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []RecurringSubtaskResponse{}
	}
	return items, nil
}

func (r *repository) FindRecurringByID(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*RecurringTaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		SELECT 
			rt.id, rt.user_id, rt.project_id, p.name as project_name, p.color as project_color,
			rt.title, rt.description, rt.priority, rt.recurrence_pattern, rt.recurrence_interval,
			rt.due_time, rt.recurrence_end_date, rt.is_active, rt.last_generated_at,
			rt.created_at, rt.updated_at
		FROM recurring_tasks rt
		LEFT JOIN projects p ON p.id = rt.project_id
		WHERE rt.id = $1 AND rt.user_id = $2
	`
	var entity recurringTaskEntity
	err := tx.GetContext(ctx, &entity, query, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find recurring task: %w", err)
	}
	subtasks, _ := r.fetchRecurringSubtasks(ctx, tx, entity.ID)
	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) FindRecurringByIDSystem(ctx context.Context, tx *sqlx.Tx, id int64) (*RecurringTaskResponse, string, error) {
	if tx == nil {
		return nil, "", errors.New("transaction is required")
	}
	query := `
		SELECT 
			rt.id, rt.user_id, rt.project_id, p.name as project_name, p.color as project_color,
			rt.title, rt.description, rt.priority, rt.recurrence_pattern, rt.recurrence_interval,
			rt.due_time, rt.recurrence_end_date, rt.is_active, rt.last_generated_at,
			rt.created_at, rt.updated_at,
			u.timezone
		FROM recurring_tasks rt
		JOIN users u ON u.id = rt.user_id
		LEFT JOIN projects p ON p.id = rt.project_id
		WHERE rt.id = $1
	`
	type sysRecEntity struct {
		recurringTaskEntity
		Timezone string `db:"timezone"`
	}
	var se sysRecEntity
	err := tx.GetContext(ctx, &se, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("failed to find recurring task system: %w", err)
	}

	subtasks, _ := r.fetchRecurringSubtasks(ctx, tx, se.ID)
	res := se.recurringTaskEntity.toResponse(subtasks)
	tz := se.Timezone
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	return &res, tz, nil
}

func (r *repository) FindAllRecurring(ctx context.Context, tx *sqlx.Tx, userID string) ([]RecurringTaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		SELECT 
			rt.id, rt.user_id, rt.project_id, p.name as project_name, p.color as project_color,
			rt.title, rt.description, rt.priority, rt.recurrence_pattern, rt.recurrence_interval,
			rt.due_time, rt.recurrence_end_date, rt.is_active, rt.last_generated_at,
			rt.created_at, rt.updated_at
		FROM recurring_tasks rt
		LEFT JOIN projects p ON p.id = rt.project_id
		WHERE rt.user_id = $1
		ORDER BY rt.id ASC
	`
	var entities []recurringTaskEntity
	err := tx.SelectContext(ctx, &entities, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find all recurring tasks: %w", err)
	}
	res := make([]RecurringTaskResponse, len(entities))
	for i, e := range entities {
		subtasks, _ := r.fetchRecurringSubtasks(ctx, tx, e.ID)
		res[i] = e.toResponse(subtasks)
	}
	return res, nil
}

func (r *repository) CreateRecurring(ctx context.Context, tx *sqlx.Tx, userID string, req CreateRecurringTaskRequest) (*RecurringTaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	interval := req.RecurrenceInterval
	if interval <= 0 {
		interval = 1
	}
	pattern := req.RecurrencePattern
	if pattern == "" {
		pattern = "daily"
	}
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	query := `
		INSERT INTO recurring_tasks (
			user_id, project_id, title, description, priority,
			recurrence_pattern, recurrence_interval, due_time, recurrence_end_date,
			is_active, last_generated_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, TRUE, NULL, NOW(), NOW())
		RETURNING id, user_id, project_id, title, description, priority,
			recurrence_pattern, recurrence_interval, due_time, recurrence_end_date,
			is_active, last_generated_at, created_at, updated_at
	`
	var entity recurringTaskEntity
	err := tx.QueryRowxContext(ctx, query, userID, req.ProjectID, req.Title, req.Description, priority, pattern, interval, req.DueTime, req.RecurrenceEndDate).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create recurring task: %w", err)
	}

	if len(req.Subtasks) > 0 {
		var validItems []CreateRecurringSubtaskItem
		for _, st := range req.Subtasks {
			if strings.TrimSpace(st.Title) != "" {
				validItems = append(validItems, st)
			}
		}
		if len(validItems) > 0 {
			var valueStrings []string
			var valueArgs []interface{}
			argIdx := 1
			for i, st := range validItems {
				orderIdx := st.OrderIndex
				if orderIdx <= 0 {
					orderIdx = i
				}
				valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, NOW(), NOW())", argIdx, argIdx+1, argIdx+2, argIdx+3))
				valueArgs = append(valueArgs, entity.ID, st.Title, st.Description, orderIdx)
				argIdx += 4
			}
			batchQuery := fmt.Sprintf(`
				INSERT INTO recurring_subtasks (recurring_task_id, title, description, order_index, created_at, updated_at)
				VALUES %s
			`, strings.Join(valueStrings, ", "))
			if _, err := tx.ExecContext(ctx, batchQuery, valueArgs...); err != nil {
				return nil, fmt.Errorf("failed to batch insert recurring subtasks: %w", err)
			}
		}
	}

	subtasks, _ := r.fetchRecurringSubtasks(ctx, tx, entity.ID)
	res := entity.toResponse(subtasks)
	if entity.ProjectID != nil {
		_ = tx.QueryRowxContext(ctx, "SELECT name, color FROM projects WHERE id = $1", *entity.ProjectID).Scan(&res.ProjectName, &res.ProjectColor)
	}
	return &res, nil
}

func (r *repository) UpdateRecurring(ctx context.Context, tx *sqlx.Tx, userID string, id int64, req UpdateRecurringTaskRequest) (*RecurringTaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	existing, err := r.FindRecurringByID(ctx, tx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("recurring task not found")
	}

	title := existing.Title
	if req.Title != nil {
		title = *req.Title
	}
	desc := existing.Description
	if req.ClearDescription {
		desc = ""
	} else if req.Description != nil {
		desc = *req.Description
	}
	priority := existing.Priority
	if req.Priority != nil {
		priority = *req.Priority
	}
	pattern := existing.RecurrencePattern
	if req.RecurrencePattern != nil {
		pattern = *req.RecurrencePattern
	}
	interval := existing.RecurrenceInterval
	if req.RecurrenceInterval != nil {
		interval = *req.RecurrenceInterval
	}
	if interval <= 0 {
		interval = 1
	}
	dueTime := existing.DueTime
	if req.ClearDueTime {
		dueTime = nil
	} else if req.HasDueTime {
		dueTime = req.DueTime
	} else if req.DueTime != nil {
		dueTime = req.DueTime
	}
	endDate := existing.RecurrenceEndDate
	if req.ClearRecurrenceEndDate {
		endDate = nil
	} else if req.HasRecurrenceEndDate {
		endDate = req.RecurrenceEndDate
	} else if req.RecurrenceEndDate != nil {
		endDate = req.RecurrenceEndDate
	}
	isActive := existing.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	projectID := existing.ProjectID
	if req.ClearProjectID {
		projectID = nil
	} else if req.HasProjectID {
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

	query := `
		UPDATE recurring_tasks
		SET title = $1, description = $2, priority = $3, project_id = $4,
			recurrence_pattern = $5, recurrence_interval = $6, due_time = $7,
			recurrence_end_date = $8, is_active = $9, updated_at = NOW()
		WHERE id = $10 AND user_id = $11
		RETURNING id, user_id, project_id, title, description, priority,
			recurrence_pattern, recurrence_interval, due_time, recurrence_end_date,
			is_active, last_generated_at, created_at, updated_at
	`
	var entity recurringTaskEntity
	err = tx.QueryRowxContext(ctx, query, title, desc, priority, projectID, pattern, interval, dueTime, endDate, isActive, id, userID).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to update recurring task: %w", err)
	}

	if req.Subtasks != nil {
		if _, err := tx.ExecContext(ctx, "DELETE FROM recurring_subtasks WHERE recurring_task_id = $1", id); err != nil {
			return nil, fmt.Errorf("failed to delete existing recurring subtasks: %w", err)
		}
		var validItems []CreateRecurringSubtaskItem
		for _, st := range *req.Subtasks {
			if strings.TrimSpace(st.Title) != "" {
				validItems = append(validItems, st)
			}
		}
		if len(validItems) > 0 {
			var valueStrings []string
			var valueArgs []interface{}
			argIdx := 1
			for i, st := range validItems {
				orderIdx := st.OrderIndex
				if orderIdx <= 0 {
					orderIdx = i
				}
				valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, NOW(), NOW())", argIdx, argIdx+1, argIdx+2, argIdx+3))
				valueArgs = append(valueArgs, id, st.Title, st.Description, orderIdx)
				argIdx += 4
			}
			batchQuery := fmt.Sprintf(`
				INSERT INTO recurring_subtasks (recurring_task_id, title, description, order_index, created_at, updated_at)
				VALUES %s
			`, strings.Join(valueStrings, ", "))
			if _, err := tx.ExecContext(ctx, batchQuery, valueArgs...); err != nil {
				return nil, fmt.Errorf("failed to batch insert recurring subtasks: %w", err)
			}
		}
	}

	subtasks, _ := r.fetchRecurringSubtasks(ctx, tx, entity.ID)
	res := entity.toResponse(subtasks)
	if entity.ProjectID != nil {
		_ = tx.QueryRowxContext(ctx, "SELECT name, color FROM projects WHERE id = $1", *entity.ProjectID).Scan(&res.ProjectName, &res.ProjectColor)
	}
	return &res, nil
}

func (r *repository) DeleteRecurring(ctx context.Context, tx *sqlx.Tx, userID string, id int64) error {
	if tx == nil {
		return errors.New("transaction is required")
	}
	query := `DELETE FROM recurring_tasks WHERE id = $1 AND user_id = $2`
	_, err := tx.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete recurring task: %w", err)
	}
	return nil
}

func (r *repository) ToggleActiveRecurring(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*RecurringTaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		UPDATE recurring_tasks
		SET is_active = NOT is_active, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, project_id, title, description, priority,
			recurrence_pattern, recurrence_interval, due_time, recurrence_end_date,
			is_active, last_generated_at, created_at, updated_at
	`
	var entity recurringTaskEntity
	err := tx.QueryRowxContext(ctx, query, id, userID).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle recurring task active: %w", err)
	}
	subtasks, _ := r.fetchRecurringSubtasks(ctx, tx, entity.ID)
	res := entity.toResponse(subtasks)
	if entity.ProjectID != nil {
		_ = tx.QueryRowxContext(ctx, "SELECT name, color FROM projects WHERE id = $1", *entity.ProjectID).Scan(&res.ProjectName, &res.ProjectColor)
	}
	return &res, nil
}

func (r *repository) FindTodayInstance(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64, todayDate time.Time) (*TaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		SELECT 
			t.id, t.user_id, t.title, t.description, t.status, t.priority, t.due_date, t.due_date_warning_sent_at, t.is_focus, t.tags,
			t.project_id, p.name as project_name, p.color as project_color,
			t.last_progress_note, t.last_progress_at, t.source_note_id,
			t.is_recurring, t.recurrence_pattern, t.recurrence_interval, t.recurrence_end_date, t.recurrence_last_generated_at,
			t.recurring_task_id, t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN projects p ON p.id = t.project_id
		WHERE t.recurring_task_id = $1 AND ((t.due_date AT TIME ZONE 'UTC')::date) = (($2::timestamptz AT TIME ZONE 'UTC')::date)
		LIMIT 1
	`
	var entity taskEntity
	err := tx.GetContext(ctx, &entity, query, recurringTaskID, todayDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find today's instance: %w", err)
	}
	subtasks, _ := r.fetchSubtasks(ctx, tx, entity.ID)
	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) CreateInstanceFromRecurring(ctx context.Context, tx *sqlx.Tx, master *RecurringTaskResponse, dueDate time.Time) (*TaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	query := `
		INSERT INTO tasks (
			user_id, title, description, status, priority, due_date, is_focus, tags, project_id,
			is_recurring, recurrence_pattern, recurrence_interval, recurrence_end_date, recurring_task_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, 'todo', $4, $5, FALSE, '{}', $6, TRUE, $7, $8, $9, $10, NOW(), NOW())
		ON CONFLICT (recurring_task_id, ((due_date AT TIME ZONE 'UTC')::date)) WHERE recurring_task_id IS NOT NULL AND due_date IS NOT NULL
		DO NOTHING
		RETURNING id, user_id, title, description, status, priority, due_date, due_date_warning_sent_at, is_focus, tags,
			project_id, last_progress_note, last_progress_at, source_note_id,
			is_recurring, recurrence_pattern, recurrence_interval, recurrence_end_date, recurrence_last_generated_at,
			recurring_task_id, created_at, updated_at
	`
	var entity taskEntity
	err := tx.QueryRowxContext(ctx, query,
		master.UserID, master.Title, master.Description, master.Priority, dueDate, master.ProjectID,
		master.RecurrencePattern, master.RecurrenceInterval, master.RecurrenceEndDate, master.ID,
	).StructScan(&entity)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return r.FindTodayInstance(ctx, tx, master.ID, dueDate)
		}
		return nil, fmt.Errorf("failed to create instance from recurring task: %w", err)
	}

	if len(master.Subtasks) > 0 {
		var validItems []RecurringSubtaskResponse
		for _, st := range master.Subtasks {
			if strings.TrimSpace(st.Title) != "" {
				validItems = append(validItems, st)
			}
		}
		if len(validItems) > 0 {
			var valueStrings []string
			var valueArgs []interface{}
			argIdx := 1
			for _, st := range validItems {
				valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, FALSE, $%d, NOW(), NOW())", argIdx, argIdx+1, argIdx+2, argIdx+3))
				valueArgs = append(valueArgs, entity.ID, st.Title, st.Description, st.OrderIndex)
				argIdx += 4
			}
			batchQuery := fmt.Sprintf(`
				INSERT INTO subtasks (task_id, title, description, is_completed, order_index, created_at, updated_at)
				VALUES %s
			`, strings.Join(valueStrings, ", "))
			if _, err := tx.ExecContext(ctx, batchQuery, valueArgs...); err != nil {
				return nil, fmt.Errorf("failed to batch insert subtasks for recurring instance: %w", err)
			}
		}
	}

	if entity.ProjectID != nil {
		_ = tx.QueryRowxContext(ctx, "SELECT name, color FROM projects WHERE id = $1", *entity.ProjectID).Scan(&entity.ProjectName, &entity.ProjectColor)
	}

	subtasks, _ := r.fetchSubtasks(ctx, tx, entity.ID)
	res := entity.toResponse(subtasks)
	return &res, nil
}

func (r *repository) UpdateRecurringLastGenerated(ctx context.Context, tx *sqlx.Tx, id int64, t time.Time) error {
	if tx == nil {
		return errors.New("transaction is required")
	}
	query := `UPDATE recurring_tasks SET last_generated_at = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, t, id)
	return err
}

func (r *repository) CreateRecurringSubtask(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64, title, description string, orderIndex int) error {
	if tx == nil {
		return errors.New("transaction is required")
	}
	if title == "" {
		return nil
	}
	query := `
		INSERT INTO recurring_subtasks (recurring_task_id, title, description, order_index, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`
	_, err := tx.ExecContext(ctx, query, recurringTaskID, title, description, orderIndex)
	return err
}

func (r *repository) UpdateRecurringSubtaskByTitle(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64, oldTitle, newTitle, description string) error {
	if tx == nil {
		return errors.New("transaction is required")
	}
	if oldTitle == "" && newTitle == "" {
		return nil
	}
	targetTitle := oldTitle
	if targetTitle == "" {
		targetTitle = newTitle
	}
	finalTitle := newTitle
	if finalTitle == "" {
		finalTitle = oldTitle
	}
	query := `
		UPDATE recurring_subtasks
		SET title = $1, description = $2, updated_at = NOW()
		WHERE recurring_task_id = $3 AND title = $4
	`
	_, err := tx.ExecContext(ctx, query, finalTitle, description, recurringTaskID, targetTitle)
	return err
}

func (r *repository) DeleteRecurringSubtaskByTitle(ctx context.Context, tx *sqlx.Tx, recurringTaskID int64, title string) error {
	if tx == nil {
		return errors.New("transaction is required")
	}
	if title == "" {
		return nil
	}
	query := `
		DELETE FROM recurring_subtasks
		WHERE recurring_task_id = $1 AND title = $2
	`
	_, err := tx.ExecContext(ctx, query, recurringTaskID, title)
	return err
}

func (r *repository) SyncRecurringFromTask(ctx context.Context, tx *sqlx.Tx, userID string, task *TaskResponse) error {
	if tx == nil {
		return errors.New("transaction is required")
	}
	if task == nil || task.RecurringTaskID == nil {
		return nil
	}
	interval := task.RecurrenceInterval
	if interval <= 0 {
		interval = 1
	}
	pattern := task.RecurrencePattern
	if pattern == "" {
		pattern = "daily"
	}

	var dueTime *string
	if task.DueDate != nil {
		_, tz, _ := r.GetUserThresholdAndTimezone(ctx, tx, userID)
		loc, err := time.LoadLocation(tz)
		if err != nil {
			loc = time.Local
		}
		formatted := task.DueDate.In(loc).Format("15:04:05")
		dueTime = &formatted
	}

	query := `
		UPDATE recurring_tasks
		SET title = $1, description = $2, priority = $3, project_id = $4,
		    recurrence_pattern = $5, recurrence_interval = $6, recurrence_end_date = $7,
		    due_time = $8, updated_at = NOW()
		WHERE id = $9 AND user_id = $10
	`
	_, err := tx.ExecContext(ctx, query,
		task.Title, task.Description, task.Priority, task.ProjectID,
		pattern, interval, task.RecurrenceEndDate,
		dueTime,
		*task.RecurringTaskID, userID,
	)
	return err
}

func (r *repository) ConvertTaskToRecurring(ctx context.Context, tx *sqlx.Tx, userID string, taskID int64, pattern string, interval int, endDate *time.Time) (*RecurringTaskResponse, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	task, err := r.FindByID(ctx, tx, userID, taskID)
	if err != nil || task == nil {
		return nil, errors.New("task not found")
	}
	if interval <= 0 {
		interval = 1
	}
	if pattern == "" {
		pattern = "daily"
	}

	var dueTime *string
	if task.DueDate != nil {
		_, tz, _ := r.GetUserThresholdAndTimezone(ctx, tx, userID)
		loc, err := time.LoadLocation(tz)
		if err != nil {
			loc = time.Local
		}
		formatted := task.DueDate.In(loc).Format("15:04:05")
		dueTime = &formatted
	}

	// 1. Create recurring_tasks master
	recQuery := `
		INSERT INTO recurring_tasks (
			user_id, project_id, title, description, priority,
			recurrence_pattern, recurrence_interval, due_time, recurrence_end_date, is_active,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, TRUE, NOW(), NOW())
		RETURNING id, user_id, project_id, title, description, priority,
			recurrence_pattern, recurrence_interval, due_time, recurrence_end_date,
			is_active, last_generated_at, created_at, updated_at
	`
	var entity recurringTaskEntity
	err = tx.QueryRowxContext(ctx, recQuery,
		userID, task.ProjectID, task.Title, task.Description, task.Priority,
		pattern, interval, dueTime, endDate,
	).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create recurring task: %w", err)
	}

	// 2. Copy existing subtasks from tasks -> recurring_subtasks in batch
	if len(task.Subtasks) > 0 {
		var validItems []SubtaskResponse
		for _, st := range task.Subtasks {
			if strings.TrimSpace(st.Title) != "" {
				validItems = append(validItems, st)
			}
		}
		if len(validItems) > 0 {
			var valueStrings []string
			var valueArgs []interface{}
			argIdx := 1
			for i, st := range validItems {
				orderIdx := st.OrderIndex
				if orderIdx <= 0 {
					orderIdx = i
				}
				valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, NOW(), NOW())", argIdx, argIdx+1, argIdx+2, argIdx+3))
				valueArgs = append(valueArgs, entity.ID, st.Title, st.Description, orderIdx)
				argIdx += 4
			}
			batchQuery := fmt.Sprintf(`
				INSERT INTO recurring_subtasks (recurring_task_id, title, description, order_index, created_at, updated_at)
				VALUES %s
			`, strings.Join(valueStrings, ", "))
			if _, err := tx.ExecContext(ctx, batchQuery, valueArgs...); err != nil {
				return nil, fmt.Errorf("failed to batch copy subtasks to recurring master: %w", err)
			}
		}
	}

	// 3. Link task instance to this new master (preserves existing due_date on this task)
	_, err = tx.ExecContext(ctx, `
		UPDATE tasks
		SET is_recurring = TRUE, recurrence_pattern = $1, recurrence_interval = $2,
		    recurrence_end_date = $3, recurring_task_id = $4, updated_at = NOW()
		WHERE id = $5 AND user_id = $6
	`, pattern, interval, endDate, entity.ID, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to link task to recurring master: %w", err)
	}

	subtasks, _ := r.fetchRecurringSubtasks(ctx, tx, entity.ID)
	res := entity.toResponse(subtasks)
	if entity.ProjectID != nil {
		_ = tx.QueryRowxContext(ctx, "SELECT name, color FROM projects WHERE id = $1", *entity.ProjectID).Scan(&res.ProjectName, &res.ProjectColor)
	}
	return &res, nil
}


