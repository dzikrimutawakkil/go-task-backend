-- Q19: Add project_statuses table for project-level status management
CREATE TABLE IF NOT EXISTS project_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6B7280',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add status_id foreign key to projects table
ALTER TABLE projects ADD COLUMN IF NOT EXISTS status_id UUID REFERENCES project_statuses(id);

-- Set default status for existing projects (will be updated by seeder)
-- We don't set a default here to avoid circular dependency; seeder will handle it