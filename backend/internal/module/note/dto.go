package note

import "time"

// CreateNoteRequest payload for creating an idea note.
type CreateNoteRequest struct {
	Title    string   `json:"title" validate:"required,min=1,max=255"`
	Content  string   `json:"content"`
	Type     string   `json:"type" validate:"omitempty,oneof=actionable knowledge"`
	Tags     []string `json:"tags"`
	Color    string   `json:"color"`
	IsPinned bool     `json:"is_pinned"`
}

// UpdateNoteRequest payload for updating a note.
type UpdateNoteRequest struct {
	Title           *string   `json:"title" validate:"omitempty,min=1,max=255"`
	Content         *string   `json:"content"`
	Type            *string   `json:"type" validate:"omitempty,oneof=actionable knowledge"`
	Tags            *[]string `json:"tags"`
	Color           *string   `json:"color"`
	IsPinned        *bool     `json:"is_pinned"`
	ConvertedTaskID *int64    `json:"converted_task_id"`
}

// ConvertToTaskRequest payload for converting an idea into a task.
type ConvertToTaskRequest struct {
	Priority  string     `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	DueDate   *time.Time `json:"due_date"`
	Subtasks  []string   `json:"subtasks"`
	ProjectID *int64     `json:"project_id"`
}

// NoteFilter parameters for filtering notes.
type NoteFilter struct {
	Type   string `json:"type" query:"type"`
	Tag    string `json:"tag" query:"tag"`
	Search string `json:"search" query:"search"`
	Limit  int    `json:"limit" query:"limit" validate:"omitempty,min=1,max=100"`
	Offset int    `json:"offset" query:"offset" validate:"omitempty,min=0"`
}

// NoteResponse represents a single note.
type NoteResponse struct {
	ID              int64     `json:"id" db:"id"`
	UserID          string    `json:"user_id" db:"user_id"`
	Title           string    `json:"title" db:"title"`
	Content         string    `json:"content" db:"content"`
	Type            string    `json:"type" db:"type"`
	Tags            []string  `json:"tags"`
	Color           string    `json:"color" db:"color"`
	IsPinned        bool      `json:"is_pinned" db:"is_pinned"`
	ConvertedTaskID *int64    `json:"converted_task_id" db:"converted_task_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
