package licenses

import (
	"net/http"
	"os"

	"gotask-backend/models"
	"gotask-backend/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service LicenseService
}

func NewLicenseHandler(service LicenseService) *Handler {
	return &Handler{service: service}
}

// Request DTOs
type ValidateLicenseRequest struct {
	Key string `json:"key" binding:"required" example:"ABCD-1234-EFGH-5678"`
}

type ActivateLicenseRequest struct {
	Key string `json:"key" binding:"required" example:"ABCD-1234-EFGH-5678"`
}

type CreateLicenseRequest struct {
	Keys []LicenseInput `json:"keys" binding:"required"`
}

// ValidateLicense godoc
// @Summary     Validate a license key
// @Description Checks if a license key is valid without requiring authentication.
// @Tags        Licenses
// @Accept      json
// @Produce     json
// @Param       body body ValidateLicenseRequest true "License key to validate"
// @Success     200 {object} utils.APIResponse "Validation result"
// @Failure     400 {object} utils.APIResponse "Invalid key format"
// @Router      /api/licenses/validate [post]
func (h *Handler) ValidateLicense(c *gin.Context) {
	var req ValidateLicenseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.ValidateKey(req.Key)
	if err != nil {
		// Format error -> 400
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "validation complete", result)
}

// ActivateLicense godoc
// @Summary     Activate a license key
// @Description Activates a license key for the authenticated user.
// @Tags        Licenses
// @Accept      json
// @Produce     json
// @Param       body body ActivateLicenseRequest true "License key"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "License activated"
// @Failure     400 {object} utils.APIResponse "Invalid or already used key"
// @Failure     401 {object} utils.APIResponse "Unauthorized"
// @Router      /api/licenses/activate [post]
func (h *Handler) ActivateLicense(c *gin.Context) {
	var req ActivateLicenseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get user from context (set by RequireAuth middleware)
	user := c.MustGet("user")
	if user == nil {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get userID from MinimalUser
	minimalUser, ok := user.(models.MinimalUser)
	if !ok {
		utils.SendError(c, http.StatusInternalServerError, "Invalid user context")
		return
	}

	license, err := h.service.ActivateKey(req.Key, minimalUser.ID)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "License activated successfully", gin.H{
		"license": license,
	})
}

// CreateLicenses godoc
// @Summary     Create license keys (admin only)
// @Description Bulk creates license keys. Requires x-admin-secret header.
// @Tags        Licenses
// @Accept      json
// @Produce     json
// @Param       x-admin-secret header string true "Admin secret key"
// @Param       body body CreateLicenseRequest true "License keys to create"
// @Success     200 {object} utils.APIResponse "Licenses created"
// @Failure     401 {object} utils.APIResponse "Invalid admin secret"
// @Router      /api/licenses [post]
func (h *Handler) CreateLicenses(c *gin.Context) {
	// Verify admin secret
	adminSecret := c.GetHeader("x-admin-secret")
	expectedSecret := os.Getenv("ADMIN_SECRET")

	if adminSecret == "" || adminSecret != expectedSecret {
		utils.SendError(c, http.StatusUnauthorized, "Invalid admin secret")
		return
	}

	var req CreateLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	if len(req.Keys) == 0 {
		utils.SendError(c, http.StatusBadRequest, "No keys provided")
		return
	}

	results, err := h.service.CreateLicenseKeys(req.Keys)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to create licenses")
		return
	}

	utils.SendSuccess(c, "Licenses created", gin.H{
		"results": results,
	})
}
