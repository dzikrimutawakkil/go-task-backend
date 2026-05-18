-- 000005_add_invitations.up.sql
-- Email invitation system (Q4)

CREATE TABLE IF NOT EXISTS organization_invitations (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    invited_email VARCHAR(255) NOT NULL,
    token VARCHAR(36) UNIQUE NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'member',  -- owner/admin/member
    expires_at TIMESTAMP NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    accepted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invitations_token ON organization_invitations(token);
CREATE INDEX IF NOT EXISTS idx_invitations_org_id ON organization_invitations(organization_id);
CREATE INDEX IF NOT EXISTS idx_invitations_email ON organization_invitations(invited_email);