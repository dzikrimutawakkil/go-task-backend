-- Create tier_plans and tier_limits tables with seeders
-- M5: Subscription Tiers — Phase 1: Database & Models

-- Tier Plans: pricing information for each tier
CREATE TABLE tier_plans (
  id SERIAL PRIMARY KEY,
  tier VARCHAR(20) NOT NULL UNIQUE,
  name VARCHAR(50) NOT NULL,
  description TEXT,
  price_monthly INTEGER NOT NULL DEFAULT 0,
  price_yearly INTEGER NOT NULL DEFAULT 0,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tier Limits: feature quotas for each tier
CREATE TABLE tier_limits (
  id SERIAL PRIMARY KEY,
  tier VARCHAR(20) NOT NULL UNIQUE,
  max_workspaces INTEGER NOT NULL DEFAULT 1,
  max_projects INTEGER NOT NULL DEFAULT 3,
  max_tasks_per_project INTEGER NOT NULL DEFAULT 50,
  max_members INTEGER NOT NULL DEFAULT 1,
  max_clients INTEGER NOT NULL DEFAULT 5,
  max_invoices_per_month INTEGER NOT NULL DEFAULT 10,
  can_comment BOOLEAN NOT NULL DEFAULT false,
  can_sse BOOLEAN NOT NULL DEFAULT false,
  can_audit_log BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed tier_plans
INSERT INTO tier_plans (tier, name, description, price_monthly, price_yearly) VALUES
  ('free', 'Free', 'Untuk freelancer solo yang baru mulai', 0, 0),
  ('pro', 'Pro', 'Untuk freelancer yang berkembang dengan tim kecil', 79000, 790000),
  ('ultimate', 'Ultimate', 'Untuk agensi dan komunitas dengan tim besar', 199000, 1990000);

-- Seed tier_limits (unlimited = -1)
INSERT INTO tier_limits (tier, max_workspaces, max_projects, max_tasks_per_project, max_members, max_clients, max_invoices_per_month, can_comment, can_sse, can_audit_log) VALUES
  ('free', 1, 3, 50, 1, 5, 10, false, false, false),
  ('pro', 2, -1, -1, 3, -1, -1, true, true, false),
  ('ultimate', 4, -1, -1, 15, -1, -1, true, true, true);