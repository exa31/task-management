ALTER TABLE users DROP COLUMN IF EXISTS reminder_sound;

DROP INDEX IF EXISTS idx_subtasks_due_date;
ALTER TABLE subtasks DROP COLUMN IF EXISTS due_date;

DROP INDEX IF EXISTS idx_tasks_user_recurring;
ALTER TABLE tasks DROP COLUMN IF EXISTS recurrence_last_generated_at;
ALTER TABLE tasks DROP COLUMN IF EXISTS recurrence_end_date;
ALTER TABLE tasks DROP COLUMN IF EXISTS recurrence_interval;
ALTER TABLE tasks DROP COLUMN IF EXISTS recurrence_pattern;
ALTER TABLE tasks DROP COLUMN IF EXISTS is_recurring;
