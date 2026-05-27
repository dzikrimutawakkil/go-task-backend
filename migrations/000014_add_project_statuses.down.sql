-- Q19: Rollback project_statuses table
ALTER TABLE projects DROP COLUMN IF EXISTS status_id;
DROP TABLE IF EXISTS project_statuses;