-- 000001_init_schema.down.sql
-- Rollback initial schema

DROP TABLE IF EXISTS task_users;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS priorities;
DROP TABLE IF EXISTS statuses;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS organization_users;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;