-- 000019: Add client_id to projects table
-- Projects can now optionally be linked to a client

ALTER TABLE projects
  ADD COLUMN client_id BIGINT REFERENCES clients(id) ON DELETE SET NULL;