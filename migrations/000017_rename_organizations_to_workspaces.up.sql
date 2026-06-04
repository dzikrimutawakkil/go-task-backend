-- 000017: Rename organizations table to workspaces and update all related columns
-- This migration renames the organizations table and all related foreign keys

-- Rename main table
ALTER TABLE organizations RENAME TO workspaces;

-- Rename foreign key columns in child tables
ALTER TABLE projects RENAME COLUMN organization_id TO workspace_id;
ALTER TABLE clients RENAME COLUMN organization_id TO workspace_id;
ALTER TABLE invoices RENAME COLUMN organization_id TO workspace_id;
ALTER TABLE tasks RENAME COLUMN organization_id TO workspace_id;

-- Rename junction/join table
ALTER TABLE organization_users RENAME TO workspace_members;
ALTER TABLE workspace_members RENAME COLUMN organization_id TO workspace_id;

-- Rename invitations table column
ALTER TABLE invitations RENAME COLUMN organization_id TO workspace_id;