-- 000011_add_password_reset_tokens.down.sql
-- Drop password_reset_tokens table

DROP INDEX IF EXISTS idx_password_reset_tokens_token;
DROP INDEX IF EXISTS idx_password_reset_tokens_user_id;
DROP TABLE IF EXISTS password_reset_tokens;
