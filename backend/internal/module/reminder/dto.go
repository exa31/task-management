package reminder

import "time"

// CreateReminderRequest payload for creating a new reminder.
type CreateReminderRequest struct {
	TaskID      *int64    `json:"task_id"`
	Title       string    `json:"title" validate:"required,min=1,max=255"`
	Description string    `json:"description" validate:"max=1000"`
	RemindAt    time.Time `json:"remind_at" validate:"required"`
}

// UpdateReminderRequest payload for updating an existing reminder.
type UpdateReminderRequest struct {
	TaskID      *int64     `json:"task_id"`
	ClearTaskID bool       `json:"clear_task_id"`
	Title       *string    `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string    `json:"description" validate:"omitempty,max=1000"`
	RemindAt    *time.Time `json:"remind_at"`
	Status      *string    `json:"status" validate:"omitempty,oneof=pending sent dismissed"`
}

// SnoozeReminderRequest payload for postponing a reminder.
type SnoozeReminderRequest struct {
	Minutes int `json:"minutes" validate:"required,min=1,max=43200"`
}

// TriggerReminderRequest payload for external worker HTTP POST trigger.
type TriggerReminderRequest struct {
	ReminderID int64 `json:"reminder_id" query:"reminder_id"`
	ID         int64 `json:"id" query:"id"`
}

func (r *TriggerReminderRequest) TargetID() int64 {
	if r.ReminderID != 0 {
		return r.ReminderID
	}
	return r.ID
}

// RegisterDeviceTokenRequest payload for saving push notification tokens (FCM).
type RegisterDeviceTokenRequest struct {
	Token    string `json:"token" validate:"required,min=10"`
	Platform string `json:"platform" validate:"omitempty,oneof=web android ios"`
}

// ReminderFilter holds query parameters for filtering reminders.
type ReminderFilter struct {
	TaskID *int64 `json:"task_id" query:"task_id"`
	Status string `json:"status" query:"status" validate:"omitempty,oneof=pending sent dismissed all"`
	Limit  int    `json:"limit" query:"limit" validate:"omitempty,min=1,max=100"`
	Offset int    `json:"offset" query:"offset" validate:"omitempty,min=0"`
}

// ReminderResponse represents a reminder in API responses.
type ReminderResponse struct {
	ID          int64      `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	TaskID      *int64     `json:"task_id" db:"task_id"`
	TaskTitle   *string    `json:"task_title" db:"task_title"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	RemindAt    time.Time  `json:"remind_at" db:"remind_at"`
	Status      string     `json:"status" db:"status"`
	SnoozeCount int        `json:"snooze_count" db:"snooze_count"`
	SentAt      *time.Time `json:"sent_at" db:"sent_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// TestFCMRequest payload for testing FCM push notification dispatch.
type TestFCMRequest struct {
	TargetUserID string            `json:"target_user_id" validate:"omitempty"`
	Token        string            `json:"token" validate:"omitempty"`
	Title        string            `json:"title" validate:"omitempty,max=255"`
	Body         string            `json:"body" validate:"omitempty,max=1000"`
	Data         map[string]string `json:"data" validate:"omitempty"`
}

// TestFCMResponse represents the outcome of a test FCM push notification.
type TestFCMResponse struct {
	Success      bool              `json:"success"`
	Message      string            `json:"message"`
	TargetUserID string            `json:"target_user_id"`
	TokensCount  int               `json:"tokens_count"`
	SuccessCount int               `json:"success_count"`
	FailureCount int               `json:"failure_count"`
	Tokens       []string          `json:"tokens_masked"`
	Title        string            `json:"title"`
	Body         string            `json:"body"`
	Data         map[string]string `json:"data,omitempty"`
}
