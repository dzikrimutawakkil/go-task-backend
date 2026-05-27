package utils

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the standard API response envelope.
type APIResponse struct {
	Success        bool            `json:"success"`
	Message        string          `json:"message"`
	Data           interface{}     `json:"data,omitempty"`
	LicenseWarning *LicenseWarning `json:"license_warning,omitempty"`
}

// LicenseWarning represents license expiry information in the API response.
// Q17: License Expiry Soft Warning Banner
type LicenseWarning struct {
	Expired       bool   `json:"expired"`
	ExpiredAt     string `json:"expired_at,omitempty"`
	DaysRemaining int    `json:"days_remaining"`
	Message       string `json:"message"`
}

// SendSuccess sends a standard success response.
func SendSuccess(c *gin.Context, message string, data ...interface{}) {
	var responseData interface{}

	// Check if any data was passed
	if len(data) > 0 {
		responseData = data[0] // Take the first item
	} else {
		responseData = nil
	}

	c.JSON(http.StatusOK, APIResponse{
		Success:        true,
		Message:        message,
		Data:           responseData,
		LicenseWarning: getLicenseWarning(c),
	})
}

// SendError sends a standard error response.
func SendError(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Success:        false,
		Message:        message,
		Data:           nil,
		LicenseWarning: getLicenseWarning(c),
	})
}

// getLicenseWarning retrieves license warning from Gin context (set by RequireAuth middleware).
// Returns nil if no warning is set or license is valid.
func getLicenseWarning(c *gin.Context) *LicenseWarning {
	if warning, exists := c.Get("license_warning"); exists {
		if w, ok := warning.(*LicenseWarning); ok {
			return w
		}
	}
	return nil
}

// SetLicenseWarning sets the license warning in Gin context.
func SetLicenseWarning(c *gin.Context, warning *LicenseWarning) {
	c.Set("license_warning", warning)
}

// BuildLicenseWarning creates a LicenseWarning based on expiry time and plan.
func BuildLicenseWarning(expiresAt *time.Time, plan string) *LicenseWarning {
	warning := &LicenseWarning{}

	if plan == "free" || plan == "" {
		warning.Expired = true
		warning.DaysRemaining = 0
		warning.Message = "You are on a free plan. Upgrade to access premium features."
		return warning
	}

	if expiresAt == nil {
		// No expiry set, assume valid
		return nil
	}

	daysRemaining := int(time.Until(*expiresAt).Hours() / 24)

	if daysRemaining < 0 {
		warning.Expired = true
		warning.DaysRemaining = daysRemaining
		warning.ExpiredAt = expiresAt.Format(time.RFC3339)
		warning.Message = "License expired. Please upgrade to continue premium features."
	} else if daysRemaining <= 7 {
		warning.Expired = false
		warning.DaysRemaining = daysRemaining
		warning.Message = "License expires in " + strconv.Itoa(daysRemaining) + " days. Consider upgrading."
	}

	return warning
}
