package middlewares

import (
	"gotask-backend/config"
	"gotask-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireTierFeature returns a middleware that checks if the workspace's tier allows a specific feature.
// Returns 403 Forbidden if the feature is not available in the workspace's tier.
// M5: Subscription Tiers — Phase 6: Feature Gate Middleware
// M-MIGRATION: Tier is now per-workspace, reads from workspace table.
func RequireTierFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get workspace ID from context (set by RequireAuth middleware)
		workspaceIDStr, exists := c.Get("workspace_id")
		if !exists {
			utils.SendError(c, http.StatusInternalServerError, "Workspace context not found")
			c.Abort()
			return
		}
		workspaceID, err := parseUint(workspaceIDStr.(string))
		if err != nil {
			utils.SendError(c, http.StatusInternalServerError, "Invalid workspace ID")
			c.Abort()
			return
		}

		// Read workspace tier from DB
		var ws struct {
			Tier          string
			TierExpiresAt *string
		}
		if err := config.DB.Table("workspaces").
			Select("tier, tier_expires_at").
			Where("id = ?", workspaceID).
			Scan(&ws).Error; err != nil {
			utils.SendError(c, http.StatusInternalServerError, "Failed to read workspace tier")
			c.Abort()
			return
		}

		// Parse expires at for effective tier calculation
		// (just use tier name directly for now since GetEffectiveTierFromString handles parsing)
		_ = ws.TierExpiresAt

		effectiveTier := utils.GetEffectiveTierFromString(ws.Tier, ws.TierExpiresAt)
		limits := utils.GetTierLimits(effectiveTier)

		allowed := false
		switch feature {
		case "comment":
			allowed = limits.CanComment
		case "sse":
			allowed = limits.CanSSE
		case "audit_log":
			allowed = limits.CanAuditLog
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unknown feature: " + feature})
			return
		}

		if !allowed {
			utils.SendError(c, http.StatusForbidden, "This feature requires a higher tier.")
			return
		}

		c.Next()
	}
}

// parseUint converts a string to uint.
func parseUint(s string) (uint, error) {
	var v uint
	_, err := parseUintFmt(s, &v)
	return v, err
}

// parseUintFmt parses a uint from string (simplified).
func parseUintFmt(s string, v *uint) (uint, error) {
	var parsed uint
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		parsed = parsed*10 + uint(c-'0')
	}
	*v = parsed
	return parsed, nil
}
