DROP INDEX IF EXISTS idx_tasks_unique_daily_recurring;
DROP INDEX IF EXISTS idx_tasks_recurring_task_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS recurring_task_id;
DROP TABLE IF EXISTS recurring_subtasks;
DROP TABLE IF EXISTS recurring_tasks;
