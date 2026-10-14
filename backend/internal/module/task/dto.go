package task

import (
	"encoding/json"
	"time"
)

// SubtaskResponse represents a checklist item in a task.
type SubtaskResponse struct {
	ID          int64      `json:"id" db:"id"`
	TaskID      int64      `json:"task_id" db:"task_id"`
	Title       string     `json:"title" db:"title"`
	IsCompleted bool       `json:"is_completed" db:"is_completed"`
	OrderIndex  int        `json:"order_index" db:"order_index"`
	DueDate     *time.Time `json:"due_date" db:"due_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateTaskRequest payload for creating a new task.
type CreateTaskRequest struct {
	Title              string     `json:"title" validate:"required,min=1,max=255"`
	Description        string     `json:"description"`
	Status             string     `json:"status" validate:"omitempty,oneof=backlog todo in_progress review done"`
	Priority           string     `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	DueDate            *time.Time `json:"due_date"`
	IsFocus            bool       `json:"is_focus"`
	Tags               []string   `json:"tags"`
	ProjectID          *int64     `json:"project_id"`
	SourceNoteID       *int64     `json:"source_note_id"`
	Subtasks           []string   `json:"subtasks"`
	IsRecurring        bool       `json:"is_recurring"`
	RecurrencePattern  string     `json:"recurrence_pattern" validate:"omitempty,oneof=daily weekdays weekly monthly"`
	RecurrenceInterval int        `json:"recurrence_interval" validate:"omitempty,min=1,max=365"`
	RecurrenceEndDate  *time.Time `json:"recurrence_end_date"`
}

// UpdateTaskRequest payload for updating an existing task.
type UpdateTaskRequest struct {
	Title              *string    `json:"title" validate:"omitempty,min=1,max=255"`
	Description        *string    `json:"description"`
	Status             *string    `json:"status" validate:"omitempty,oneof=backlog todo in_progress review done"`
	Priority           *string    `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	DueDate              *time.Time `json:"due_date"`
	HasDueDate           bool       `json:"-"`
	IsFocus              *bool      `json:"is_focus"`
	Tags                 *[]string  `json:"tags"`
	ProjectID            *int64     `json:"project_id"`
	HasProjectID         bool       `json:"-"`
	LastProgressNote     *string    `json:"last_progress_note"`
	IsRecurring          *bool      `json:"is_recurring"`
	RecurrencePattern    *string    `json:"recurrence_pattern" validate:"omitempty,oneof=daily weekdays weekly monthly"`
	RecurrenceInterval   *int       `json:"recurrence_interval" validate:"omitempty,min=1,max=365"`
	RecurrenceEndDate    *time.Time `json:"recurrence_end_date"`
	HasRecurrenceEndDate bool       `json:"-"`
}

// UnmarshalJSON detects if project_id, due_date, or recurrence_end_date was explicitly provided in JSON (including null).
func (u *UpdateTaskRequest) UnmarshalJSON(data []byte) error {
	type Alias UpdateTaskRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err == nil {
		if _, exists := raw["project_id"]; exists {
			u.HasProjectID = true
		}
		if _, exists := raw["due_date"]; exists {
			u.HasDueDate = true
		}
		if _, exists := raw["recurrence_end_date"]; exists {
			u.HasRecurrenceEndDate = true
		}
	}
	return nil
}

// LogProgressRequest payload for updating context ("inget-inget udah sampe mana").
type LogProgressRequest struct {
	ProgressNote        string  `json:"progress_note" validate:"required"`
	Status              *string `json:"status" validate:"omitempty,oneof=backlog todo in_progress review done"`
	CompletedSubtaskIDs []int64 `json:"completed_subtask_ids"`
}

// CreateSubtaskRequest payload for adding a subtask checklist item.
type CreateSubtaskRequest struct {
	Title   string     `json:"title" validate:"required,min=1,max=255"`
	DueDate *time.Time `json:"due_date"`
}

// UpdateSubtaskRequest payload for modifying a subtask checklist item.
type UpdateSubtaskRequest struct {
	Title       *string    `json:"title"`
	IsCompleted *bool      `json:"is_completed"`
	DueDate     *time.Time `json:"due_date"`
	HasDueDate  bool       `json:"-"`
}

// UnmarshalJSON detects if due_date was explicitly provided in JSON (including null).
func (u *UpdateSubtaskRequest) UnmarshalJSON(data []byte) error {
	type Alias UpdateSubtaskRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err == nil {
		if _, exists := raw["due_date"]; exists {
			u.HasDueDate = true
		}
	}
	return nil
}

// TaskFilter parameters for querying tasks.
type TaskFilter struct {
	Status    string `json:"status" query:"status"`
	Priority  string `json:"priority" query:"priority"`
	IsFocus   *bool  `json:"is_focus" query:"is_focus"`
	Tag       string `json:"tag" query:"tag"`
	Search    string `json:"search" query:"search"`
	ProjectID *int64 `json:"project_id" query:"project_id"`
	Limit     int    `json:"limit" query:"limit" validate:"omitempty,min=1,max=100"`
	Offset    int    `json:"offset" query:"offset" validate:"omitempty,min=0"`
	Cursor    *int64 `json:"cursor" query:"cursor"`
	Paginate  string `json:"paginate" query:"paginate"`
}

// TaskCursorResponse represents a paginated task list response using cursor.
type TaskCursorResponse struct {
	Items      []TaskResponse `json:"items"`
	NextCursor *int64         `json:"next_cursor"`
	HasMore    bool           `json:"has_more"`
}

// TaskResponse represents a detailed task object.
type TaskResponse struct {
	ID                        int64             `json:"id" db:"id"`
	UserID                    string            `json:"user_id" db:"user_id"`
	Title                     string            `json:"title" db:"title"`
	Description               string            `json:"description" db:"description"`
	Status                    string            `json:"status" db:"status"`
	Priority                  string            `json:"priority" db:"priority"`
	DueDate                   *time.Time        `json:"due_date" db:"due_date"`
	DueDateWarningSentAt      *time.Time        `json:"due_date_warning_sent_at" db:"due_date_warning_sent_at"`
	IsFocus                   bool              `json:"is_focus" db:"is_focus"`
	Tags                      []string          `json:"tags"`
	ProjectID                 *int64            `json:"project_id" db:"project_id"`
	ProjectName               *string           `json:"project_name" db:"project_name"`
	ProjectColor              *string           `json:"project_color" db:"project_color"`
	LastProgressNote          string            `json:"last_progress_note" db:"last_progress_note"`
	LastProgressAt            *time.Time        `json:"last_progress_at" db:"last_progress_at"`
	SourceNoteID              *int64            `json:"source_note_id" db:"source_note_id"`
	Subtasks                  []SubtaskResponse `json:"subtasks"`
	SubtasksCount             int               `json:"subtasks_count"`
	SubtasksCompleted         int               `json:"subtasks_completed"`
	ProgressPercent           int               `json:"progress_percent"`
	IsRecurring               bool              `json:"is_recurring" db:"is_recurring"`
	RecurrencePattern         string            `json:"recurrence_pattern" db:"recurrence_pattern"`
	RecurrenceInterval        int               `json:"recurrence_interval" db:"recurrence_interval"`
	RecurrenceEndDate         *time.Time        `json:"recurrence_end_date" db:"recurrence_end_date"`
	RecurrenceLastGeneratedAt *time.Time        `json:"recurrence_last_generated_at" db:"recurrence_last_generated_at"`
	CreatedAt                 time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time         `json:"updated_at" db:"updated_at"`
}

// TriggerRecurrenceRequest payload for Asynq worker rollover trigger.
type TriggerRecurrenceRequest struct {
	TaskID int64 `json:"task_id" query:"task_id"`
	ID     int64 `json:"id" query:"id"`
}

func (r *TriggerRecurrenceRequest) TargetID() int64 {
	if r.TaskID != 0 {
		return r.TaskID
	}
	return r.ID
}

// TriggerDueWarningRequest payload for external worker HTTP POST trigger.
type TriggerDueWarningRequest struct {
	TaskID int64 `json:"task_id" query:"task_id"`
	ID     int64 `json:"id" query:"id"`
}

func (r *TriggerDueWarningRequest) TargetID() int64 {
	if r.TaskID != 0 {
		return r.TaskID
	}
	return r.ID
}

// TaskActivityResponse represents a historical audit log or progress event for a task.
type TaskActivityResponse struct {
	ID           int64           `json:"id" db:"id"`
	TaskID       int64           `json:"task_id" db:"task_id"`
	TaskTitle    string          `json:"task_title,omitempty" db:"task_title"`
	UserID       string          `json:"user_id" db:"user_id"`
	ActivityType string          `json:"activity_type" db:"activity_type"`
	Description  string          `json:"description" db:"description"`
	Changes      json.RawMessage `json:"changes" db:"changes"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

// CreateTaskActivityRequest payload for posting a manual activity note / progress comment.
type CreateTaskActivityRequest struct {
	ActivityType string `json:"activity_type" validate:"omitempty,max=50"`
	Description  string `json:"description" validate:"required,min=1"`
}

// WidgetBundleResponse aggregates data for iOS Home Screen & Lock Screen widgets.
type WidgetBundleResponse struct {
	FocusTasks        []TaskResponse `json:"focus_tasks"`
	UpcomingDeadlines []TaskResponse `json:"upcoming_deadlines"`
	CompletedToday    int            `json:"completed_today"`
	TotalToday        int            `json:"total_today"`
}

// WorkspaceStatsResponse provides lightweight counts for the navigation sidebar and dashboard badges.
type WorkspaceStatsResponse struct {
	TasksInProgress  int `json:"tasks_in_progress" db:"tasks_in_progress"`
	TasksTodo        int `json:"tasks_todo" db:"tasks_todo"`
	TasksTotal       int `json:"tasks_total" db:"tasks_total"`
	TotalProjects    int `json:"total_projects" db:"total_projects"`
	PendingReminders int `json:"pending_reminders" db:"pending_reminders"`
	TotalNotes       int `json:"total_notes" db:"total_notes"`
}

