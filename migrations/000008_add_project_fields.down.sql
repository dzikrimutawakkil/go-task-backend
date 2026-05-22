-- 000008_add_project_fields.down.sql
-- Remove status, priority, progress, budget, deadline from projects table

ALTER TABLE projects DROP COLUMN IF EXISTS status;
ALTER TABLE projects DROP COLUMN IF EXISTS priority;
ALTER TABLE projects DROP COLUMN IF EXISTS progress;
ALTER TABLE projects DROP COLUMN IF EXISTS budget;
ALTER TABLE projects DROP COLUMN IF EXISTS deadline;