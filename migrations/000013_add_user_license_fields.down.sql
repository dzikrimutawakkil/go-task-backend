-- 000013_add_user_license_fields.down.sql
-- Remove license-related fields from users table

ALTER TABLE users DROP COLUMN IF EXISTS license_status;
ALTER TABLE users DROP COLUMN IF EXISTS license_key;
ALTER TABLE users DROP COLUMN IF EXISTS plan;
