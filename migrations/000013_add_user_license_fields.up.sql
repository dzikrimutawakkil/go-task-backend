-- 000013_add_user_license_fields.up.sql
-- Add license-related fields to users table

ALTER TABLE users ADD COLUMN IF NOT EXISTS plan VARCHAR(50) NOT NULL DEFAULT 'free';
ALTER TABLE users ADD COLUMN IF NOT EXISTS license_key VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS license_status VARCHAR(20) NOT NULL DEFAULT 'inactive';
