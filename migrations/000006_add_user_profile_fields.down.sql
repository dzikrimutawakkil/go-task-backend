-- 000006_add_user_profile_fields.down.sql
-- Remove name, phone, address fields from users table

ALTER TABLE users DROP COLUMN IF EXISTS address;
ALTER TABLE users DROP COLUMN IF EXISTS phone;
ALTER TABLE users DROP COLUMN IF EXISTS name;