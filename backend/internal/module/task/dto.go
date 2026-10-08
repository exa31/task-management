package task

import "time"

// SubtaskResponse represents a checklist item in a task.
type SubtaskResponse struct {
	ID          int64     `json:"id" db:"id"`
	TaskID      int64     `json:"task_id" db:"task_id"`
	Title       string    `json:"title" db:"title"`
	IsCompleted bool      `json:"is_completed" db:"is_completed"`
	OrderIndex  int       `json:"order_index" db:"order_index"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateTaskRequest payload for creating a new task.
type CreateTaskRequest struct {
	Title        string     `json:"title" validate:"required,min=1,max=255"`
	Description  string     `json:"description"`
	Status       string     `json:"status" validate:"omitempty,oneof=backlog todo in_progress review done"`
	Priority     string     `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	DueDate      *time.Time `json:"due_date"`
	IsFocus      bool       `json:"is_focus"`
	Tags         []string   `json:"tags"`
	SourceNoteID *int64     `json:"source_note_id"`
	Subtasks     []string   `json:"subtasks"`
}

// UpdateTaskRequest payload for updating an existing task.
type UpdateTaskRequest struct {
	Title            *string    `json:"title" validate:"omitempty,min=1,max=255"`
	Description      *string    `json:"description"`
	Status           *string    `json:"status" validate:"omitempty,oneof=backlog todo in_progress review done"`
	Priority         *string    `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	DueDate          *time.Time `json:"due_date"`
	IsFocus          *bool      `json:"is_focus"`
	Tags             *[]string  `json:"tags"`
	LastProgressNote *string    `json:"last_progress_note"`
}

// LogProgressRequest payload for updating context ("inget-inget udah sampe mana").
type LogProgressRequest struct {
	ProgressNote        string  `json:"progress_note" validate:"required"`
	Status              *string `json:"status" validate:"omitempty,oneof=backlog todo in_progress review done"`
	CompletedSubtaskIDs []int64 `json:"completed_subtask_ids"`
}

// CreateSubtaskRequest payload for adding a subtask checklist item.
type CreateSubtaskRequest struct {
	Title string `json:"title" validate:"required,min=1,max=255"`
}

// UpdateSubtaskRequest payload for modifying a subtask checklist item.
type UpdateSubtaskRequest struct {
	Title       *string `json:"title"`
	IsCompleted *bool   `json:"is_completed"`
}

// TaskFilter parameters for querying tasks.
type TaskFilter struct {
	Status   string `query:"status"`
	Priority string `query:"priority"`
	IsFocus  *bool  `query:"is_focus"`
	Tag      string `query:"tag"`
	Search   string `query:"search"`
	Limit    int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset   int    `query:"offset" validate:"omitempty,min=0"`
	Cursor   *int64 `query:"cursor"`
	Paginate string `query:"paginate"`
}

// TaskCursorResponse represents a paginated task list response using cursor.
type TaskCursorResponse struct {
	Items      []TaskResponse `json:"items"`
	NextCursor *int64         `json:"next_cursor"`
	HasMore    bool           `json:"has_more"`
}

// TaskResponse represents a detailed task object.
type TaskResponse struct {
	ID                int64             `json:"id" db:"id"`
	UserID            string            `json:"user_id" db:"user_id"`
	Title             string            `json:"title" db:"title"`
	Description       string            `json:"description" db:"description"`
	Status            string            `json:"status" db:"status"`
	Priority          string            `json:"priority" db:"priority"`
	DueDate           *time.Time        `json:"due_date" db:"due_date"`
	IsFocus           bool              `json:"is_focus" db:"is_focus"`
	Tags              []string          `json:"tags"`
	LastProgressNote  string            `json:"last_progress_note" db:"last_progress_note"`
	LastProgressAt    *time.Time        `json:"last_progress_at" db:"last_progress_at"`
	SourceNoteID      *int64            `json:"source_note_id" db:"source_note_id"`
	Subtasks          []SubtaskResponse `json:"subtasks"`
	SubtasksCount     int               `json:"subtasks_count"`
	SubtasksCompleted int               `json:"subtasks_completed"`
	ProgressPercent   int               `json:"progress_percent"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
}
