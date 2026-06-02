-- Add tier fields to users table
-- M5: Subscription Tiers — Phase 1: Database & Models

ALTER TABLE users
  ADD COLUMN tier VARCHAR(20) NOT NULL DEFAULT 'free',
  ADD COLUMN tier_expires_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN tier_activated_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN tier_activated_by BIGINT REFERENCES users(id);

-- Add index for faster tier lookups
CREATE INDEX idx_users_tier ON users(tier);
CREATE INDEX idx_users_tier_expires_at ON users(tier_expires_at);