-- 000009_add_clients.up.sql
-- Create clients table for contact management

CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    whatsapp VARCHAR(50),
    phone VARCHAR(50),
    company VARCHAR(255),
    website VARCHAR(255),
    address TEXT,
    notes TEXT,
    total_revenue DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clients_org ON clients(organization_id);