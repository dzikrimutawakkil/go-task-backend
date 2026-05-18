-- 000004_add_labels.up.sql
-- Labels/tags system per project (Q8)

-- Labels table
CREATE TABLE IF NOT EXISTS labels (
    id SERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6366F1',
    version INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_labels_project_id ON labels(project_id);

-- Task-Label join table (many-to-many)
CREATE TABLE IF NOT EXISTS task_labels (
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    label_id BIGINT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, label_id)
);

CREATE INDEX IF NOT EXISTS idx_task_labels_label_id ON task_labels(label_id);