-- 000017: Revert rename of organizations to workspaces
-- Revert all table and column renames from workspaces back to organizations

-- Revert invitations
ALTER TABLE invitations RENAME COLUMN workspace_id TO organization_id;

-- Revert junction table
ALTER TABLE workspace_members RENAME COLUMN workspace_id TO organization_id;
ALTER TABLE workspace_members RENAME TO organization_users;

-- Revert foreign key columns in child tables
ALTER TABLE tasks RENAME COLUMN workspace_id TO organization_id;
ALTER TABLE invoices RENAME COLUMN workspace_id TO organization_id;
ALTER TABLE clients RENAME COLUMN workspace_id TO organization_id;
ALTER TABLE projects RENAME COLUMN workspace_id TO organization_id;

-- Rename main table back
ALTER TABLE workspaces RENAME TO organizations;