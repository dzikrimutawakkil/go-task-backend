-- 000019: Revert add client_id to projects
-- Remove the client_id column from projects

ALTER TABLE projects DROP COLUMN IF EXISTS client_id;