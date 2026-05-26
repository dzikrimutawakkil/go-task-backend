-- 000012_add_licenses.down.sql
-- Drop licenses table

DROP INDEX IF EXISTS idx_licenses_key;
DROP INDEX IF EXISTS idx_licenses_status;
DROP INDEX IF EXISTS idx_licenses_activated_by;
DROP TABLE IF EXISTS licenses;
