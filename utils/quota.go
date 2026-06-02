package utils

import (
	"fmt"
	"time"
)

// TierLimits represents quota limits for each tier.
// M5: Subscription Tiers — Phase 2: Quota Utilities
type TierLimits struct {
	MaxWorkspaces       int
	MaxProjects         int
	MaxTasksPerProject  int
	MaxMembers          int
	MaxClients          int
	MaxInvoicesPerMonth int
	CanComment          bool
	CanSSE              bool
	CanAuditLog         bool
}

// GetTierLimits returns default limits for a given tier.
// Falls back to free tier limits if tier is not recognized.
func GetTierLimits(tier string) TierLimits {
	defaults := map[string]TierLimits{
		"free": {
			MaxWorkspaces: 1, MaxProjects: 3, MaxTasksPerProject: 50,
			MaxMembers: 1, MaxClients: 5, MaxInvoicesPerMonth: 10,
			CanComment: false, CanSSE: false, CanAuditLog: false,
		},
		"pro": {
			MaxWorkspaces: 2, MaxProjects: -1, MaxTasksPerProject: -1,
			MaxMembers: 3, MaxClients: -1, MaxInvoicesPerMonth: -1,
			CanComment: true, CanSSE: true, CanAuditLog: false,
		},
		"ultimate": {
			MaxWorkspaces: 4, MaxProjects: -1, MaxTasksPerProject: -1,
			MaxMembers: 15, MaxClients: -1, MaxInvoicesPerMonth: -1,
			CanComment: true, CanSSE: true, CanAuditLog: true,
		},
	}

	if limits, ok := defaults[tier]; ok {
		return limits
	}
	return defaults["free"]
}

// IsTierActive checks if a tier is still valid (not expired).
// Free tier never expires. Paid tiers expire at tier_expires_at.
func IsTierActive(tier string, expiresAt *time.Time) bool {
	if tier == "free" {
		return true // free never expires
	}
	if expiresAt == nil {
		return false
	}
	return time.Now().Before(*expiresAt)
}

// GetEffectiveTier returns the actual tier, falling back to free if expired.
// M5: Subscription Tiers — Phase 6: Response & Middleware
func GetEffectiveTier(tier string, expiresAt *time.Time) string {
	if IsTierActive(tier, expiresAt) {
		return tier
	}
	return "free"
}

// QuotaError is returned when a resource limit is exceeded.
type QuotaError struct {
	Resource    string
	Limit       int
	CurrentTier string
}

// Error implements the error interface for QuotaError.
func (e *QuotaError) Error() string {
	return fmt.Sprintf("quota exceeded: %s limit is %d on %s tier. Please upgrade.", e.Resource, e.Limit, e.CurrentTier)
}

// ErrQuotaExceeded creates a new QuotaError for the given resource, limit, and tier.
func ErrQuotaExceeded(resource string, limit int, tier string) *QuotaError {
	return &QuotaError{
		Resource:    resource,
		Limit:       limit,
		CurrentTier: tier,
	}
}

// DaysRemaining returns the number of days remaining until expiry.
// Returns negative value if already expired.
func DaysRemaining(expiresAt *time.Time) int {
	if expiresAt == nil {
		return 0
	}
	return int(time.Until(*expiresAt).Hours() / 24)
}
