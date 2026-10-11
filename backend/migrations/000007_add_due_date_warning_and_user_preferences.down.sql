DROP INDEX IF EXISTS idx_tasks_due_date_warning;
ALTER TABLE tasks DROP COLUMN IF EXISTS due_date_warning_sent_at;

ALTER TABLE users DROP COLUMN IF EXISTS email_notifications_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS due_date_warning_threshold;
ALTER TABLE users DROP COLUMN IF EXISTS timezone;
