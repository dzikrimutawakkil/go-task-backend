-- Revert: rename workspace_type back to org_type
ALTER TABLE workspaces RENAME COLUMN workspace_type TO org_type;
