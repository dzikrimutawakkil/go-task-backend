-- 000008_add_project_fields.up.sql
-- Add status, priority, progress, budget, deadline to projects table

ALTER TABLE projects ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'backlog';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS priority VARCHAR(20) DEFAULT 'medium';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS progress INT DEFAULT 0;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS budget DECIMAL(15,2);
ALTER TABLE projects ADD COLUMN IF NOT EXISTS deadline DATE;