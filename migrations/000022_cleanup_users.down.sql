-- Revert 000022: Restore tier columns to users
-- Note: This is for reference only - these columns should NOT be restored

-- This migration is not easily reversible as data was already migrated to workspaces
-- If you need to restore, you would need to:
-- 1. Recreate the columns
-- 2. Copy data back from workspaces table

-- For safety, this down migration does nothing
SELECT 'No rollback needed - tier is now per-workspace' AS note;
