-- Rollback: Remove tier fields from users table
-- M5: Subscription Tiers — Phase 1: Database & Models

ALTER TABLE users DROP COLUMN IF EXISTS tier;
ALTER TABLE users DROP COLUMN IF EXISTS tier_expires_at;
ALTER TABLE users DROP COLUMN IF EXISTS tier_activated_at;
ALTER TABLE users DROP COLUMN IF EXISTS tier_activated_by;

DROP INDEX IF EXISTS idx_users_tier;
DROP INDEX IF EXISTS idx_users_tier_expires_at;