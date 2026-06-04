package workspaces

import "time"

// TierPlan represents a subscription tier plan (pricing information).
// M5: Subscription Tiers — Phase 1: Database & Models
type TierPlan struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Tier         string    `gorm:"unique;column:tier" json:"tier"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	PriceMonthly int       `json:"price_monthly"`
	PriceYearly  int       `json:"price_yearly"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TierLimit represents quota limits for each tier.
// M5: Subscription Tiers — Phase 1: Database & Models
type TierLimit struct {
	ID                  uint   `gorm:"primaryKey" json:"id"`
	Tier                string `gorm:"unique;column:tier" json:"tier"`
	MaxWorkspaces       int    `json:"max_workspaces"`
	MaxProjects         int    `json:"max_projects"`
	MaxTasksPerProject  int    `json:"max_tasks_per_project"`
	MaxMembers          int    `json:"max_members"`
	MaxClients          int    `json:"max_clients"`
	MaxInvoicesPerMonth int    `json:"max_invoices_per_month"`
	CanComment          bool   `json:"can_comment"`
	CanSSE              bool   `json:"can_sse"`
	CanAuditLog         bool   `json:"can_audit_log"`
}

// TierPlanWithLimits represents a tier plan with its limits for API response.
// M5: Subscription Tiers — Phase 5: Handler & Endpoint
type TierPlanWithLimits struct {
	Tier         string    `json:"tier"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	PriceMonthly int       `json:"price_monthly"`
	PriceYearly  int       `json:"price_yearly"`
	Limits       TierLimit `json:"limits"`
	Features     Features  `json:"features"`
}

// Features represents the boolean features for a tier.
type Features struct {
	Comments bool `json:"comments"`
	Realtime bool `json:"realtime"`
	AuditLog bool `json:"audit_log"`
}
