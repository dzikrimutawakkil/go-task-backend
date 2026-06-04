-- 000021: Clean up project_statuses - change UUID to int
-- This migration changes project_statuses.id from UUID to SERIAL (int)

-- 1. Drop the foreign key constraint from projects
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_status_id_fkey;

-- 2. Drop the existing column from projects (old UUID type)
ALTER TABLE projects DROP COLUMN IF EXISTS status_id;

-- 3. Drop the project_statuses table and recreate with int ID
DROP TABLE IF EXISTS project_statuses;

CREATE TABLE project_statuses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    color VARCHAR(7) NOT NULL DEFAULT '#6B7280',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Add new int column to projects
ALTER TABLE projects ADD COLUMN status_id INT REFERENCES project_statuses(id);

-- 5. Seed default project statuses
INSERT INTO project_statuses (name, color) VALUES
    ('Active', '#22C55E'),
    ('On Hold', '#F59E0B'),
    ('Completed', '#3B82F6'),
    ('Archived', '#6B7280')
ON CONFLICT (name) DO NOTHING;

-- 6. Update existing projects to have Active status
UPDATE projects SET status_id = ps.id
FROM project_statuses ps
WHERE ps.name = 'Active' AND projects.status_id IS NULL;