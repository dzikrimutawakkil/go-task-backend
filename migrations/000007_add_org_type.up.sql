-- Add organization type for personal vs team workspaces
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS org_type VARCHAR(20) NOT NULL DEFAULT 'personal';