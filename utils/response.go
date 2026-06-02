package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the standard API response envelope.
type APIResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
	TierInfo *TierInfo   `json:"tier_info,omitempty"`
}

// TierInfo represents tier status in the API response.
// M5: Subscription Tiers — Phase 6: Response & Middleware
type TierInfo struct {
	Tier          string `json:"tier"`
	IsActive      bool   `json:"is_active"`
	ExpiresAt     string `json:"expires_at,omitempty"`
	DaysRemaining int    `json:"days_remaining"`
	Warning       string `json:"warning,omitempty"`
}

// SendSuccess sends a standard success response.
func SendSuccess(c *gin.Context, message string, data ...interface{}) {
	var responseData interface{}
	if len(data) > 0 {
		responseData = data[0]
	} else {
		responseData = nil
	}

	c.JSON(http.StatusOK, APIResponse{
		Success:  true,
		Message:  message,
		Data:     responseData,
		TierInfo: getTierInfo(c),
	})
}

// SendError sends a standard error response.
func SendError(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Success:  false,
		Message:  message,
		Data:     nil,
		TierInfo: getTierInfo(c),
	})
}

// getTierInfo retrieves tier info from Gin context (set by RequireAuth middleware).
func getTierInfo(c *gin.Context) *TierInfo {
	if info, exists := c.Get("tier_info"); exists {
		if t, ok := info.(*TierInfo); ok {
			return t
		}
	}
	return nil
}

// SetTierInfo sets the tier info in Gin context.
func SetTierInfo(c *gin.Context, info *TierInfo) {
	c.Set("tier_info", info)
}

// BuildTierInfo creates a TierInfo based on user's tier fields.
func BuildTierInfo(tier string, tierExpiresAt *time.Time) *TierInfo {
	info := &TierInfo{
		Tier:     tier,
		IsActive: IsTierActive(tier, tierExpiresAt),
	}

	if tierExpiresAt != nil {
		info.ExpiresAt = tierExpiresAt.Format(time.RFC3339)
		info.DaysRemaining = DaysRemaining(tierExpiresAt)
	}

	if tier == "free" {
		info.IsActive = true
		info.Warning = "Upgrade to access premium features."
	} else if !info.IsActive {
		info.Warning = "Tier expired. Your account has been downgraded to Free."
	} else if info.DaysRemaining <= 7 && info.DaysRemaining >= 0 {
		info.Warning = "Tier expires in " + string(rune(info.DaysRemaining+'0')) + " days. Consider renewing."
	}

	return info
}
