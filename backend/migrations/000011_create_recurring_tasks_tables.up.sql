-- Create recurring_tasks table (Master Recurring Template)
CREATE TABLE IF NOT EXISTS recurring_tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id BIGINT REFERENCES projects(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    recurrence_pattern VARCHAR(30) NOT NULL DEFAULT 'daily',
    recurrence_interval INT NOT NULL DEFAULT 1,
    due_time TIME,
    recurrence_end_date TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_generated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recurring_tasks_user_active ON recurring_tasks(user_id, is_active);
CREATE INDEX IF NOT EXISTS idx_recurring_tasks_project ON recurring_tasks(project_id);

-- Create recurring_subtasks table (Subtask Templates for Recurring Tasks)
CREATE TABLE IF NOT EXISTS recurring_subtasks (
    id BIGSERIAL PRIMARY KEY,
    recurring_task_id BIGINT NOT NULL REFERENCES recurring_tasks(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    order_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recurring_subtasks_task_order ON recurring_subtasks(recurring_task_id, order_index);

-- Add recurring_task_id reference to tasks table
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS recurring_task_id BIGINT REFERENCES recurring_tasks(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_recurring_task_id ON tasks(recurring_task_id);

-- Enforce database-level idempotency: only one task per recurring master per due date
CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_unique_daily_recurring 
ON tasks(recurring_task_id, ((due_date AT TIME ZONE 'UTC')::date)) 
WHERE recurring_task_id IS NOT NULL AND due_date IS NOT NULL;
