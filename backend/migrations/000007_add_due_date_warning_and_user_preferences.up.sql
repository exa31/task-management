-- Add user notification preferences and timezone
ALTER TABLE users ADD COLUMN IF NOT EXISTS timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta';
ALTER TABLE users ADD COLUMN IF NOT EXISTS due_date_warning_threshold INT NOT NULL DEFAULT 1440;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE;

-- Add due_date_warning_sent_at to tasks to track alert dispatch status
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS due_date_warning_sent_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_tasks_due_date_warning ON tasks(user_id, due_date, due_date_warning_sent_at);
