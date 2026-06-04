-- 000022: Clean up users table - drop legacy columns
-- These columns were already migrated to workspaces table

-- Drop tier columns from users (already migrated to workspaces in 000018)
ALTER TABLE users DROP COLUMN IF EXISTS tier;
ALTER TABLE users DROP COLUMN IF EXISTS tier_expires_at;
ALTER TABLE users DROP COLUMN IF EXISTS tier_activated_at;
ALTER TABLE users DROP COLUMN IF EXISTS tier_activated_by;

-- Drop unused licenses table (not used in current implementation)
DROP TABLE IF EXISTS licenses;

-- Drop indexes if they exist
DROP INDEX IF EXISTS idx_users_tier;
DROP INDEX IF EXISTS idx_users_tier_expires_at;
