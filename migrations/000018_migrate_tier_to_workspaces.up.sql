-- 000018: Migrate tier from users table to workspaces table
-- Tier is now per-workspace, not per-user

-- Add tier columns to workspaces
ALTER TABLE workspaces
  ADD COLUMN tier VARCHAR(20) NOT NULL DEFAULT 'free',
  ADD COLUMN tier_expires_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN tier_activated_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN tier_activated_by BIGINT REFERENCES users(id);

-- Migrate data: copy tier from owner user to workspace
-- We join workspace_members to find the owner of each workspace
UPDATE workspaces w
SET
  tier = u.tier,
  tier_expires_at = u.tier_expires_at,
  tier_activated_at = u.tier_activated_at,
  tier_activated_by = u.tier_activated_by
FROM workspace_members wm
JOIN users u ON u.id = wm.user_id
WHERE wm.workspace_id = w.id
  AND wm.role = 'owner';

-- Remove tier columns from users (after data successfully migrated)
ALTER TABLE users
  DROP COLUMN IF EXISTS tier,
  DROP COLUMN IF EXISTS tier_expires_at,
  DROP COLUMN IF EXISTS tier_activated_at,
  DROP COLUMN IF EXISTS tier_activated_by;