-- 000020: Rename org_type to workspace_type for consistency
-- All "organization" references have been migrated to "workspace"
ALTER TABLE workspaces RENAME COLUMN org_type TO workspace_type;
