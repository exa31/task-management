-- Add recurrence columns to tasks
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS is_recurring BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS recurrence_pattern VARCHAR(30) NOT NULL DEFAULT '';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS recurrence_interval INT NOT NULL DEFAULT 1;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS recurrence_end_date TIMESTAMP WITH TIME ZONE;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS recurrence_last_generated_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_tasks_user_recurring ON tasks(user_id, is_recurring);

-- Add due_date to subtasks
ALTER TABLE subtasks ADD COLUMN IF NOT EXISTS due_date TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_subtasks_due_date ON subtasks(task_id, due_date);

-- Add reminder_sound to users for custom audio/mood tones
ALTER TABLE users ADD COLUMN IF NOT EXISTS reminder_sound VARCHAR(100) NOT NULL DEFAULT 'default';
