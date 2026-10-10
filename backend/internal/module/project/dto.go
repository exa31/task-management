package project

import "time"

// CreateProjectRequest payload for creating a project/module.
type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Color       string `json:"color" validate:"omitempty,max=30"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// UpdateProjectRequest payload for modifying a project/module.
type UpdateProjectRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Color       *string `json:"color" validate:"omitempty,max=30"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// ProjectResponse represents a single project entity with summary stats.
type ProjectResponse struct {
	ID          int64     `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Color       string    `json:"color" db:"color"`
	Description string    `json:"description" db:"description"`
	TaskCount   int       `json:"task_count" db:"task_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
