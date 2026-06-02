-- Rollback: Drop tier_plans and tier_limits tables
-- M5: Subscription Tiers — Phase 1: Database & Models

DROP TABLE IF EXISTS tier_limits;
DROP TABLE IF EXISTS tier_plans;