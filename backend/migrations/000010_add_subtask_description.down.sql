-- Remove description column from subtasks
ALTER TABLE subtasks DROP COLUMN IF EXISTS description;
