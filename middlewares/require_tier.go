package middlewares

import (
	"gotask-backend/models"
	"gotask-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireTierFeature returns a middleware that checks if the user's tier allows a specific feature.
// Returns 403 Forbidden if the feature is not available in the user's tier.
// M5: Subscription Tiers — Phase 6: Feature Gate Middleware
func RequireTierFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(models.MinimalUser)
		effectiveTier := utils.GetEffectiveTier(user.Tier, user.TierExpiresAt)
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
			utils.SendError(c, http.StatusForbidden, "This feature requires Pro or Ultimate tier.")
			return
		}

		c.Next()
	}
}
