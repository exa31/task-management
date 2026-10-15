-- Add description column to subtasks
ALTER TABLE subtasks ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
