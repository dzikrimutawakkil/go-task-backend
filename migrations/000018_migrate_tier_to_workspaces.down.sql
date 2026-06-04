-- 000018: Revert tier migration from workspaces back to users
-- Re-add tier columns to users and copy data back from workspace owner

-- Re-add tier columns to users
ALTER TABLE users
  ADD COLUMN tier VARCHAR(20) NOT NULL DEFAULT 'free',
  ADD COLUMN tier_expires_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN tier_activated_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN tier_activated_by BIGINT REFERENCES workspaces(id);

-- Migrate data back: copy tier from workspace owner back to users
-- Update the user who is the owner of each workspace
UPDATE users u
SET
  tier = w.tier,
  tier_expires_at = w.tier_expires_at,
  tier_activated_at = w.tier_activated_at,
  tier_activated_by = w.tier_activated_by
FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE wm.user_id = u.id
  AND wm.role = 'owner';

-- Remove tier columns from workspaces
ALTER TABLE workspaces
  DROP COLUMN IF EXISTS tier,
  DROP COLUMN IF EXISTS tier_expires_at,
  DROP COLUMN IF EXISTS tier_activated_at,
  DROP COLUMN IF EXISTS tier_activated_by;