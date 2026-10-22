package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var flexibleDateFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

func parseFlexibleTime(val interface{}) (*time.Time, bool, error) {
	if val == nil {
		return nil, true, nil
	}
	str, ok := val.(string)
	if !ok {
		return nil, false, errors.New("invalid date type")
	}
	str = strings.TrimSpace(str)
	if str == "" || str == "null" {
		return nil, true, nil
	}
	for _, layout := range flexibleDateFormats {
		if t, err := time.Parse(layout, str); err == nil {
			return &t, true, nil
		}
	}
	return nil, false, fmt.Errorf("cannot parse %q as date/time", str)
}

// SubtaskResponse represents a checklist item in a task.
type SubtaskResponse struct {
	ID          int64      `json:"id" db:"id"`
	TaskID      int64      `json:"task_id" db:"task_id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
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
	RecurringTaskID    *int64     `json:"recurring_task_id"`
}

func (c *CreateTaskRequest) UnmarshalJSON(data []byte) error {
	type Alias struct {
		Title              string   `json:"title" validate:"required,min=1,max=255"`
		Description        string   `json:"description"`
		Status             string   `json:"status" validate:"omitempty,oneof=backlog todo in_progress review done"`
		Priority           string   `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
		IsFocus            bool     `json:"is_focus"`
		Tags               []string `json:"tags"`
		ProjectID          *int64   `json:"project_id"`
		SourceNoteID       *int64   `json:"source_note_id"`
		Subtasks           []string `json:"subtasks"`
		IsRecurring        bool     `json:"is_recurring"`
		RecurrencePattern  string   `json:"recurrence_pattern" validate:"omitempty,oneof=daily weekdays weekly monthly"`
		RecurrenceInterval int      `json:"recurrence_interval" validate:"omitempty,min=1,max=365"`
		RecurringTaskID    *int64   `json:"recurring_task_id"`
	}
	var aux Alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.Title = aux.Title
	c.Description = aux.Description
	c.Status = aux.Status
	c.Priority = aux.Priority
	c.IsFocus = aux.IsFocus
	c.Tags = aux.Tags
	c.ProjectID = aux.ProjectID
	c.SourceNoteID = aux.SourceNoteID
	c.Subtasks = aux.Subtasks
	c.IsRecurring = aux.IsRecurring
	c.RecurrencePattern = aux.RecurrencePattern
	c.RecurrenceInterval = aux.RecurrenceInterval
	c.RecurringTaskID = aux.RecurringTaskID

	if c.ProjectID != nil && *c.ProjectID <= 0 {
		c.ProjectID = nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err == nil {
		if val, exists := raw["due_date"]; exists {
			t, _, err := parseFlexibleTime(val)
			if err != nil {
				return err
			}
			c.DueDate = t
		}
		if val, exists := raw["recurrence_end_date"]; exists {
			t, _, err := parseFlexibleTime(val)
			if err != nil {
				return err
			}
			c.RecurrenceEndDate = t
		}
	}
	return nil
}

// UpdateTaskRequest payload for updating an existing task.
type UpdateTaskRequest struct {
	Title                  *string    `json:"title" validate:"omitempty,min=1,max=255"`
	Description            *string    `json:"description"`
	ClearDescription       bool       `json:"clear_description"`
	HasDescription         bool       `json:"-"`
	Status                 *string    `json:"status" validate:"omitempty,oneof=backlog todo in_progress review done"`
	Priority               *string    `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	DueDate                *time.Time `json:"due_date"`
	ClearDueDate           bool       `json:"clear_due_date"`
	HasDueDate             bool       `json:"-"`
	IsFocus                *bool      `json:"is_focus"`
	Tags                   *[]string  `json:"tags"`
	ProjectID              *int64     `json:"project_id"`
	ClearProjectID         bool       `json:"clear_project_id"`
	HasProjectID           bool       `json:"-"`
	LastProgressNote       *string    `json:"last_progress_note"`
	IsRecurring            *bool      `json:"is_recurring"`
	RecurrencePattern      *string    `json:"recurrence_pattern" validate:"omitempty,oneof=daily weekdays weekly monthly"`
	RecurrenceInterval     *int       `json:"recurrence_interval" validate:"omitempty,min=1,max=365"`
	RecurrenceEndDate      *time.Time `json:"recurrence_end_date"`
	ClearRecurrenceEndDate bool       `json:"clear_recurrence_end_date"`
	HasRecurrenceEndDate   bool       `json:"-"`
	RecurringTaskID        *int64     `json:"recurring_task_id"`
}

// UnmarshalJSON detects if project_id, due_date, recurrence_end_date, or description was provided or explicitly cleared.
func (u *UpdateTaskRequest) UnmarshalJSON(data []byte) error {
	type Alias struct {
		Title                  *string   `json:"title" validate:"omitempty,min=1,max=255"`
		Description            *string   `json:"description"`
		ClearDescription       bool      `json:"clear_description"`
		Status                 *string   `json:"status" validate:"omitempty,oneof=backlog todo in_progress review done"`
		Priority               *string   `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
		ClearDueDate           bool      `json:"clear_due_date"`
		IsFocus                *bool     `json:"is_focus"`
		Tags                   *[]string `json:"tags"`
		ProjectID              *int64    `json:"project_id"`
		ClearProjectID         bool      `json:"clear_project_id"`
		LastProgressNote       *string   `json:"last_progress_note"`
		IsRecurring            *bool     `json:"is_recurring"`
		RecurrencePattern      *string   `json:"recurrence_pattern" validate:"omitempty,oneof=daily weekdays weekly monthly"`
		RecurrenceInterval     *int      `json:"recurrence_interval" validate:"omitempty,min=1,max=365"`
		ClearRecurrenceEndDate bool      `json:"clear_recurrence_end_date"`
		RecurringTaskID        *int64    `json:"recurring_task_id"`
	}
	var aux Alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	u.Title = aux.Title
	u.Description = aux.Description
	u.ClearDescription = aux.ClearDescription
	u.Status = aux.Status
	u.Priority = aux.Priority
	u.ClearDueDate = aux.ClearDueDate
	u.IsFocus = aux.IsFocus
	u.Tags = aux.Tags
	u.ProjectID = aux.ProjectID
	u.ClearProjectID = aux.ClearProjectID
	u.LastProgressNote = aux.LastProgressNote
	u.IsRecurring = aux.IsRecurring
	u.RecurrencePattern = aux.RecurrencePattern
	u.RecurrenceInterval = aux.RecurrenceInterval
	u.ClearRecurrenceEndDate = aux.ClearRecurrenceEndDate
	u.RecurringTaskID = aux.RecurringTaskID

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err == nil {
		if val, exists := raw["project_id"]; exists {
			u.HasProjectID = true
			if val == nil {
				u.ProjectID = nil
				u.ClearProjectID = true
			}
		}
		if u.ClearProjectID || (u.ProjectID != nil && *u.ProjectID <= 0) {
			u.HasProjectID = true
			u.ProjectID = nil
			u.ClearProjectID = true
		}

		if val, exists := raw["due_date"]; exists {
			u.HasDueDate = true
			t, _, err := parseFlexibleTime(val)
			if err != nil {
				return err
			}
			u.DueDate = t
			if t == nil {
				u.ClearDueDate = true
			}
		}
		if u.ClearDueDate {
			u.HasDueDate = true
			u.DueDate = nil
		}

		if val, exists := raw["recurrence_end_date"]; exists {
			u.HasRecurrenceEndDate = true
			t, _, err := parseFlexibleTime(val)
			if err != nil {
				return err
			}
			u.RecurrenceEndDate = t
			if t == nil {
				u.ClearRecurrenceEndDate = true
			}
		}
		if u.ClearRecurrenceEndDate {
			u.HasRecurrenceEndDate = true
			u.RecurrenceEndDate = nil
		}

		if val, exists := raw["description"]; exists {
			u.HasDescription = true
			if val == nil {
				empty := ""
				u.Description = &empty
				u.ClearDescription = true
			} else if str, ok := val.(string); ok && strings.TrimSpace(str) == "" {
				empty := ""
				u.Description = &empty
				u.ClearDescription = true
			}
		}
		if u.ClearDescription {
			u.HasDescription = true
			empty := ""
			u.Description = &empty
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
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

func (c *CreateSubtaskRequest) UnmarshalJSON(data []byte) error {
	type Alias struct {
		Title       string `json:"title" validate:"required,min=1,max=255"`
		Description string `json:"description"`
	}
	var aux Alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.Title = aux.Title
	c.Description = aux.Description

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err == nil {
		if val, exists := raw["due_date"]; exists {
			t, _, err := parseFlexibleTime(val)
			if err != nil {
				return err
			}
			c.DueDate = t
		}
	}
	return nil
}

// UpdateSubtaskRequest payload for modifying a subtask checklist item.
type UpdateSubtaskRequest struct {
	Title            *string    `json:"title"`
	Description      *string    `json:"description"`
	ClearDescription bool       `json:"clear_description"`
	HasDescription   bool       `json:"-"`
	IsCompleted      *bool      `json:"is_completed"`
	DueDate          *time.Time `json:"due_date"`
	ClearDueDate     bool       `json:"clear_due_date"`
	HasDueDate       bool       `json:"-"`
}

// UnmarshalJSON detects if due_date, description, or clear flags were explicitly provided.
func (u *UpdateSubtaskRequest) UnmarshalJSON(data []byte) error {
	type Alias struct {
		Title            *string `json:"title"`
		Description      *string `json:"description"`
		ClearDescription bool    `json:"clear_description"`
		IsCompleted      *bool   `json:"is_completed"`
		ClearDueDate     bool    `json:"clear_due_date"`
	}
	var aux Alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	u.Title = aux.Title
	u.Description = aux.Description
	u.ClearDescription = aux.ClearDescription
	u.IsCompleted = aux.IsCompleted
	u.ClearDueDate = aux.ClearDueDate

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err == nil {
		if val, exists := raw["due_date"]; exists {
			u.HasDueDate = true
			t, _, err := parseFlexibleTime(val)
			if err != nil {
				return err
			}
			u.DueDate = t
			if t == nil {
				u.ClearDueDate = true
			}
		}
		if u.ClearDueDate {
			u.HasDueDate = true
			u.DueDate = nil
		}

		if val, exists := raw["description"]; exists {
			u.HasDescription = true
			if val == nil {
				empty := ""
				u.Description = &empty
				u.ClearDescription = true
			} else if str, ok := val.(string); ok && strings.TrimSpace(str) == "" {
				empty := ""
				u.Description = &empty
				u.ClearDescription = true
			}
		}
		if u.ClearDescription {
			u.HasDescription = true
			empty := ""
			u.Description = &empty
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
	Cursor    *int64     `json:"cursor" query:"cursor"`
	Paginate  string     `json:"paginate" query:"paginate"`
	From      *time.Time `json:"from" query:"from"`
	To        *time.Time `json:"to" query:"to"`
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
	RecurringTaskID           *int64            `json:"recurring_task_id" db:"recurring_task_id"`
	CreatedAt                 time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time         `json:"updated_at" db:"updated_at"`
}

// TriggerRecurrenceRequest payload for Asynq worker rollover trigger.
type TriggerRecurrenceRequest struct {
	RecurringTaskID int64 `json:"recurring_task_id" query:"recurring_task_id"`
	TaskID          int64 `json:"task_id" query:"task_id"`
	ID              int64 `json:"id" query:"id"`
}

func (r *TriggerRecurrenceRequest) TargetID() int64 {
	if r.RecurringTaskID != 0 {
		return r.RecurringTaskID
	}
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

// ActivityDateCount represents an aggregate count of activities for a single day.
type ActivityDateCount struct {
	Date  string `json:"date" db:"date"`
	Count int    `json:"count" db:"count"`
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

// RecurringTaskResponse represents a master recurring rule template.
type RecurringTaskResponse struct {
	ID                 int64                      `json:"id" db:"id"`
	UserID             string                     `json:"user_id" db:"user_id"`
	ProjectID          *int64                     `json:"project_id" db:"project_id"`
	ProjectName        *string                    `json:"project_name" db:"project_name"`
	ProjectColor       *string                    `json:"project_color" db:"project_color"`
	Title              string                     `json:"title" db:"title"`
	Description        string                     `json:"description" db:"description"`
	Priority           string                     `json:"priority" db:"priority"`
	RecurrencePattern  string                     `json:"recurrence_pattern" db:"recurrence_pattern"`
	RecurrenceInterval int                        `json:"recurrence_interval" db:"recurrence_interval"`
	DueTime            *string                    `json:"due_time" db:"due_time"`
	RecurrenceEndDate  *time.Time                 `json:"recurrence_end_date" db:"recurrence_end_date"`
	IsActive           bool                       `json:"is_active" db:"is_active"`
	LastGeneratedAt    *time.Time                 `json:"last_generated_at" db:"last_generated_at"`
	Subtasks           []RecurringSubtaskResponse `json:"subtasks"`
	CreatedAt          time.Time                  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time                  `json:"updated_at" db:"updated_at"`
}

// RecurringSubtaskResponse represents a template subtask within a recurring task.
type RecurringSubtaskResponse struct {
	ID              int64     `json:"id" db:"id"`
	RecurringTaskID int64     `json:"recurring_task_id" db:"recurring_task_id"`
	Title           string    `json:"title" db:"title"`
	Description     string    `json:"description" db:"description"`
	OrderIndex      int       `json:"order_index" db:"order_index"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// CreateRecurringTaskRequest payload for creating a new master recurring task.
type CreateRecurringTaskRequest struct {
	Title              string                       `json:"title" validate:"required,min=1,max=255"`
	Description        string                       `json:"description"`
	Priority           string                       `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	ProjectID          *int64                       `json:"project_id"`
	RecurrencePattern  string                       `json:"recurrence_pattern" validate:"required,oneof=daily weekdays weekly monthly"`
	RecurrenceInterval int                          `json:"recurrence_interval" validate:"omitempty,min=1,max=365"`
	DueTime            *string                      `json:"due_time"`
	RecurrenceEndDate  *time.Time                   `json:"recurrence_end_date"`
	Subtasks           []CreateRecurringSubtaskItem `json:"subtasks"`
}

func (c *CreateRecurringTaskRequest) UnmarshalJSON(data []byte) error {
	type Alias struct {
		Title              string                       `json:"title" validate:"required,min=1,max=255"`
		Description        string                       `json:"description"`
		Priority           string                       `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
		ProjectID          *int64                       `json:"project_id"`
		RecurrencePattern  string                       `json:"recurrence_pattern" validate:"required,oneof=daily weekdays weekly monthly"`
		RecurrenceInterval int                          `json:"recurrence_interval" validate:"omitempty,min=1,max=365"`
		DueTime            *string                      `json:"due_time"`
		Subtasks           []CreateRecurringSubtaskItem `json:"subtasks"`
	}
	var aux Alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.Title = aux.Title
	c.Description = aux.Description
	c.Priority = aux.Priority
	c.ProjectID = aux.ProjectID
	c.RecurrencePattern = aux.RecurrencePattern
	c.RecurrenceInterval = aux.RecurrenceInterval
	c.DueTime = aux.DueTime
	c.Subtasks = aux.Subtasks

	if c.ProjectID != nil && *c.ProjectID <= 0 {
		c.ProjectID = nil
	}
	if c.DueTime != nil && strings.TrimSpace(*c.DueTime) == "" {
		c.DueTime = nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err == nil {
		if val, exists := raw["recurrence_end_date"]; exists {
			t, _, err := parseFlexibleTime(val)
			if err != nil {
				return err
			}
			c.RecurrenceEndDate = t
		}
	}
	return nil
}

// CreateRecurringSubtaskItem describes a subtask in a recurring task creation request.
type CreateRecurringSubtaskItem struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description"`
	OrderIndex  int    `json:"order_index"`
}

// UpdateRecurringTaskRequest payload for updating a master recurring task.
type UpdateRecurringTaskRequest struct {
	Title                  *string                       `json:"title" validate:"omitempty,min=1,max=255"`
	Description            *string                       `json:"description"`
	ClearDescription       bool                          `json:"clear_description"`
	HasDescription         bool                          `json:"-"`
	Priority               *string                       `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	ProjectID              *int64                        `json:"project_id"`
	ClearProjectID         bool                          `json:"clear_project_id"`
	HasProjectID           bool                          `json:"-"`
	RecurrencePattern      *string                       `json:"recurrence_pattern" validate:"omitempty,oneof=daily weekdays weekly monthly"`
	RecurrenceInterval     *int                          `json:"recurrence_interval" validate:"omitempty,min=1,max=365"`
	DueTime                *string                       `json:"due_time"`
	ClearDueTime           bool                          `json:"clear_due_time"`
	HasDueTime             bool                          `json:"-"`
	RecurrenceEndDate      *time.Time                    `json:"recurrence_end_date"`
	ClearRecurrenceEndDate bool                          `json:"clear_recurrence_end_date"`
	HasRecurrenceEndDate   bool                          `json:"-"`
	IsActive               *bool                         `json:"is_active"`
	Subtasks               *[]CreateRecurringSubtaskItem `json:"subtasks"`
}

func (u *UpdateRecurringTaskRequest) UnmarshalJSON(data []byte) error {
	type Alias struct {
		Title                  *string                       `json:"title" validate:"omitempty,min=1,max=255"`
		Description            *string                       `json:"description"`
		ClearDescription       bool                          `json:"clear_description"`
		Priority               *string                       `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
		ProjectID              *int64                        `json:"project_id"`
		ClearProjectID         bool                          `json:"clear_project_id"`
		RecurrencePattern      *string                       `json:"recurrence_pattern" validate:"omitempty,oneof=daily weekdays weekly monthly"`
		RecurrenceInterval     *int                          `json:"recurrence_interval" validate:"omitempty,min=1,max=365"`
		DueTime                *string                       `json:"due_time"`
		ClearDueTime           bool                          `json:"clear_due_time"`
		ClearRecurrenceEndDate bool                          `json:"clear_recurrence_end_date"`
		IsActive               *bool                         `json:"is_active"`
		Subtasks               *[]CreateRecurringSubtaskItem `json:"subtasks"`
	}
	var aux Alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	u.Title = aux.Title
	u.Description = aux.Description
	u.ClearDescription = aux.ClearDescription
	u.Priority = aux.Priority
	u.ProjectID = aux.ProjectID
	u.ClearProjectID = aux.ClearProjectID
	u.RecurrencePattern = aux.RecurrencePattern
	u.RecurrenceInterval = aux.RecurrenceInterval
	u.DueTime = aux.DueTime
	u.ClearDueTime = aux.ClearDueTime
	u.ClearRecurrenceEndDate = aux.ClearRecurrenceEndDate
	u.IsActive = aux.IsActive
	u.Subtasks = aux.Subtasks

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err == nil {
		if val, exists := raw["project_id"]; exists {
			u.HasProjectID = true
			if val == nil {
				u.ProjectID = nil
				u.ClearProjectID = true
			}
		}
		if u.ClearProjectID || (u.ProjectID != nil && *u.ProjectID <= 0) {
			u.HasProjectID = true
			u.ProjectID = nil
			u.ClearProjectID = true
		}

		if val, exists := raw["due_time"]; exists {
			u.HasDueTime = true
			if val == nil {
				u.DueTime = nil
				u.ClearDueTime = true
			} else if s, ok := val.(string); ok && strings.TrimSpace(s) == "" {
				u.DueTime = nil
				u.ClearDueTime = true
			}
		}
		if u.ClearDueTime {
			u.HasDueTime = true
			u.DueTime = nil
		}

		if val, exists := raw["recurrence_end_date"]; exists {
			u.HasRecurrenceEndDate = true
			t, _, err := parseFlexibleTime(val)
			if err != nil {
				return err
			}
			u.RecurrenceEndDate = t
			if t == nil {
				u.ClearRecurrenceEndDate = true
			}
		}
		if u.ClearRecurrenceEndDate {
			u.HasRecurrenceEndDate = true
			u.RecurrenceEndDate = nil
		}

		if val, exists := raw["description"]; exists {
			u.HasDescription = true
			if val == nil {
				empty := ""
				u.Description = &empty
				u.ClearDescription = true
			} else if str, ok := val.(string); ok && strings.TrimSpace(str) == "" {
				empty := ""
				u.Description = &empty
				u.ClearDescription = true
			}
		}
		if u.ClearDescription {
			u.HasDescription = true
			empty := ""
			u.Description = &empty
		}
	}
	return nil
}

