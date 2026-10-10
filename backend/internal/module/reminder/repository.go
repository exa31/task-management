package reminder

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang/database"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, userID string, req CreateReminderRequest) (*ReminderResponse, error)
	FindByID(ctx context.Context, userID string, id int64) (*ReminderResponse, error)
	FindAll(ctx context.Context, userID string, filter ReminderFilter) ([]ReminderResponse, int, error)
	Update(ctx context.Context, userID string, id int64, req UpdateReminderRequest) (*ReminderResponse, error)
	Delete(ctx context.Context, userID string, id int64) error
	Snooze(ctx context.Context, userID string, id int64, newRemindAt time.Time) (*ReminderResponse, error)
	MarkDismissed(ctx context.Context, userID string, id int64) (*ReminderResponse, error)
	MarkSent(ctx context.Context, id int64) error
	FindByIDSystem(ctx context.Context, id int64) (*ReminderResponse, string, error) // Returns reminder + user email
	RegisterDeviceToken(ctx context.Context, userID, token, platform string) error
	GetUserDeviceTokens(ctx context.Context, userID string) ([]string, error)
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

type reminderEntity struct {
	ID          int64      `db:"id"`
	UserID      string     `db:"user_id"`
	TaskID      *int64     `db:"task_id"`
	TaskTitle   *string    `db:"task_title"`
	Title       string     `db:"title"`
	Description string     `db:"description"`
	RemindAt    time.Time  `db:"remind_at"`
	Status      string     `db:"status"`
	SnoozeCount int        `db:"snooze_count"`
	SentAt      *time.Time `db:"sent_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

func (e *reminderEntity) toResponse() ReminderResponse {
	return ReminderResponse{
		ID:          e.ID,
		UserID:      e.UserID,
		TaskID:      e.TaskID,
		TaskTitle:   e.TaskTitle,
		Title:       e.Title,
		Description: e.Description,
		RemindAt:    e.RemindAt,
		Status:      e.Status,
		SnoozeCount: e.SnoozeCount,
		SentAt:      e.SentAt,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

const reminderSelectColumns = `
	r.id, r.user_id, r.task_id, t.title AS task_title, r.title, r.description,
	r.remind_at, r.status, r.snooze_count, r.sent_at, r.created_at, r.updated_at
`

func (r *repository) Create(ctx context.Context, userID string, req CreateReminderRequest) (*ReminderResponse, error) {
	query := `
		INSERT INTO reminders (user_id, task_id, title, description, remind_at, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id, user_id, task_id, title, description, remind_at, status, snooze_count, sent_at, created_at, updated_at
	`

	var entity reminderEntity
	err := r.exec(ctx).QueryRowxContext(ctx, query,
		userID,
		req.TaskID,
		req.Title,
		req.Description,
		req.RemindAt,
	).StructScan(&entity)

	if err != nil {
		return nil, fmt.Errorf("failed to insert reminder: %w", err)
	}

	if entity.TaskID != nil {
		var taskTitle string
		_ = r.exec(ctx).GetContext(ctx, &taskTitle, `SELECT title FROM tasks WHERE id = $1`, *entity.TaskID)
		if taskTitle != "" {
			entity.TaskTitle = &taskTitle
		}
	}

	resp := entity.toResponse()
	return &resp, nil
}

func (r *repository) FindByID(ctx context.Context, userID string, id int64) (*ReminderResponse, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM reminders r
		LEFT JOIN tasks t ON r.task_id = t.id
		WHERE r.id = $1 AND r.user_id = $2
	`, reminderSelectColumns)

	var entity reminderEntity
	err := r.exec(ctx).GetContext(ctx, &entity, query, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("reminder not found")
		}
		return nil, fmt.Errorf("failed to query reminder: %w", err)
	}

	resp := entity.toResponse()
	return &resp, nil
}

func (r *repository) FindAll(ctx context.Context, userID string, filter ReminderFilter) ([]ReminderResponse, int, error) {
	where := "WHERE r.user_id = $1"
	args := []interface{}{userID}
	argIdx := 2

	if filter.TaskID != nil {
		where += fmt.Sprintf(" AND r.task_id = $%d", argIdx)
		args = append(args, *filter.TaskID)
		argIdx++
	}

	if filter.Status != "" && filter.Status != "all" {
		where += fmt.Sprintf(" AND r.status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM reminders r %s", where)
	var total int
	err := r.exec(ctx).GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count reminders: %w", err)
	}

	// Pagination
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM reminders r
		LEFT JOIN tasks t ON r.task_id = t.id
		%s
		ORDER BY r.remind_at ASC
		LIMIT $%d OFFSET $%d
	`, reminderSelectColumns, where, argIdx, argIdx+1)

	args = append(args, limit, offset)

	var entities []reminderEntity
	err = r.exec(ctx).SelectContext(ctx, &entities, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to select reminders: %w", err)
	}

	responses := make([]ReminderResponse, len(entities))
	for i, e := range entities {
		responses[i] = e.toResponse()
	}

	return responses, total, nil
}

func (r *repository) Update(ctx context.Context, userID string, id int64, req UpdateReminderRequest) (*ReminderResponse, error) {
	current, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	taskID := current.TaskID
	if req.TaskID != nil {
		taskID = req.TaskID
	}
	title := current.Title
	if req.Title != nil {
		title = *req.Title
	}
	description := current.Description
	if req.Description != nil {
		description = *req.Description
	}
	remindAt := current.RemindAt
	if req.RemindAt != nil {
		remindAt = *req.RemindAt
	}
	status := current.Status
	if req.Status != nil {
		status = *req.Status
	}

	query := `
		UPDATE reminders
		SET task_id = $1, title = $2, description = $3, remind_at = $4, status = $5, updated_at = NOW()
		WHERE id = $6 AND user_id = $7
		RETURNING id, user_id, task_id, title, description, remind_at, status, snooze_count, sent_at, created_at, updated_at
	`

	var entity reminderEntity
	err = r.exec(ctx).QueryRowxContext(ctx, query,
		taskID, title, description, remindAt, status, id, userID,
	).StructScan(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to update reminder: %w", err)
	}

	if entity.TaskID != nil {
		var taskTitle string
		_ = r.exec(ctx).GetContext(ctx, &taskTitle, `SELECT title FROM tasks WHERE id = $1`, *entity.TaskID)
		if taskTitle != "" {
			entity.TaskTitle = &taskTitle
		}
	}

	resp := entity.toResponse()
	return &resp, nil
}

func (r *repository) Delete(ctx context.Context, userID string, id int64) error {
	query := `DELETE FROM reminders WHERE id = $1 AND user_id = $2`
	res, err := r.exec(ctx).ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return errors.New("reminder not found")
	}
	return nil
}

func (r *repository) Snooze(ctx context.Context, userID string, id int64, newRemindAt time.Time) (*ReminderResponse, error) {
	query := `
		UPDATE reminders
		SET remind_at = $1, status = 'pending', snooze_count = snooze_count + 1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, task_id, title, description, remind_at, status, snooze_count, sent_at, created_at, updated_at
	`

	var entity reminderEntity
	err := r.exec(ctx).QueryRowxContext(ctx, query, newRemindAt, id, userID).StructScan(&entity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("reminder not found")
		}
		return nil, fmt.Errorf("failed to snooze reminder: %w", err)
	}

	if entity.TaskID != nil {
		var taskTitle string
		_ = r.exec(ctx).GetContext(ctx, &taskTitle, `SELECT title FROM tasks WHERE id = $1`, *entity.TaskID)
		if taskTitle != "" {
			entity.TaskTitle = &taskTitle
		}
	}

	resp := entity.toResponse()
	return &resp, nil
}

func (r *repository) MarkDismissed(ctx context.Context, userID string, id int64) (*ReminderResponse, error) {
	query := `
		UPDATE reminders
		SET status = 'dismissed', updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, task_id, title, description, remind_at, status, snooze_count, sent_at, created_at, updated_at
	`

	var entity reminderEntity
	err := r.exec(ctx).QueryRowxContext(ctx, query, id, userID).StructScan(&entity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("reminder not found")
		}
		return nil, fmt.Errorf("failed to dismiss reminder: %w", err)
	}

	resp := entity.toResponse()
	return &resp, nil
}

func (r *repository) MarkSent(ctx context.Context, id int64) error {
	query := `UPDATE reminders SET status = 'sent', sent_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.exec(ctx).ExecContext(ctx, query, id)
	return err
}

func (r *repository) FindByIDSystem(ctx context.Context, id int64) (*ReminderResponse, string, error) {
	query := fmt.Sprintf(`
		SELECT %s, u.email AS user_email
		FROM reminders r
		LEFT JOIN tasks t ON r.task_id = t.id
		JOIN users u ON r.user_id = u.id
		WHERE r.id = $1
	`, reminderSelectColumns)

	type systemEntity struct {
		reminderEntity
		UserEmail string `db:"user_email"`
	}

	var se systemEntity
	err := r.exec(ctx).GetContext(ctx, &se, query, id)
	if err != nil {
		return nil, "", err
	}

	resp := se.reminderEntity.toResponse()
	return &resp, se.UserEmail, nil
}

func (r *repository) RegisterDeviceToken(ctx context.Context, userID, token, platform string) error {
	if platform == "" {
		platform = "web"
	}
	query := `
		INSERT INTO user_device_tokens (user_id, token, platform, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (token) DO UPDATE
		SET user_id = EXCLUDED.user_id, platform = EXCLUDED.platform, updated_at = NOW()
	`
	_, err := r.exec(ctx).ExecContext(ctx, query, userID, token, platform)
	return err
}

func (r *repository) GetUserDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	query := `SELECT token FROM user_device_tokens WHERE user_id = $1`
	var tokens []string
	err := r.exec(ctx).SelectContext(ctx, &tokens, query, userID)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}
