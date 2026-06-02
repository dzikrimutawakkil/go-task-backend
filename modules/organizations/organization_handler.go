package organizations

import (
	"errors"
	"gotask-backend/models"
	"gotask-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetTierPlans godoc
// @Summary     Get all tier plans
// @Description Retrieve all active subscription tier plans with their limits and features. Public endpoint for pricing page.
// @Tags        Tiers
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=[]TierPlanWithLimits} "Success"
// @Failure     500 {object} utils.APIResponse "Failed to fetch tier plans"
// @Router      /tier/plans [get]
func (h *Handler) GetTierPlans(c *gin.Context) {
	plans, err := h.service.GetTierPlans()
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch tier plans")
		return
	}
	utils.SendSuccess(c, "OK", plans)
}

// GetMyTierInfo godoc
// @Summary     Get my tier info
// @Description Get the authenticated user's tier information including usage statistics.
// @Tags        Tiers
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=TierInfoResponse} "Success"
// @Failure     401 {object} utils.APIResponse "Unauthorized"
// @Failure     500 {object} utils.APIResponse "Failed to fetch tier info"
// @Router      /users/me/tier [get]
func (h *Handler) GetMyTierInfo(c *gin.Context) {
	user := c.MustGet("user").(models.MinimalUser)

	info, err := h.service.GetTierInfoForModels(user.ID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch tier info")
		return
	}
	utils.SendSuccess(c, "OK", info)
}

// ActivateTierRequest represents the request body for activating a tier.
// @Description Request body for activating a tier for a user
type ActivateTierRequest struct {
	Tier           string `json:"tier" binding:"required" example:"pro"`
	DurationMonths int    `json:"duration_months" binding:"required,min=1,max=24" example:"12"`
}

// ActivateTier godoc
// @Summary     Activate tier for user (Admin only)
// @Description Activate a subscription tier for a user. Requires admin role.
// @Tags        Tiers
// @Accept      json
// @Produce     json
// @Param       id path int true "User ID"
// @Param       body body ActivateTierRequest true "Tier activation payload"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=ActivateTierResult} "Tier activated successfully"
// @Failure     400 {object} utils.APIResponse "Invalid request"
// @Failure     403 {object} utils.APIResponse "Insufficient permission"
// @Failure     404 {object} utils.APIResponse "User not found"
// @Router      /admin/users/{id}/tier [patch]
func (h *Handler) ActivateTier(c *gin.Context) {
	// Check if requester is admin (you can add admin check here based on your auth system)
	// For now, we'll assume the route is protected by admin middleware in main.go

	userIDStr := c.Param("id")
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid user ID format")
		return
	}
	userID := uint(userID64)

	var req ActivateTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Validate tier value
	validTiers := map[string]bool{"free": true, "pro": true, "ultimate": true}
	if !validTiers[req.Tier] {
		utils.SendError(c, http.StatusBadRequest, "Invalid tier. Must be: free, pro, or ultimate")
		return
	}

	// Validate duration
	if req.DurationMonths < 1 || req.DurationMonths > 24 {
		utils.SendError(c, http.StatusBadRequest, "duration_months must be between 1 and 24")
		return
	}

	requester := c.MustGet("user").(models.MinimalUser)

	result, err := h.service.ActivateForModelsTier(userID, req.Tier, req.DurationMonths, requester.ID)
	if err != nil {
		if err.Error() == "user not found" {
			utils.SendError(c, http.StatusNotFound, "User not found")
			return
		}
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Tier activated successfully", result)
}

// GetUserOrganizations godoc
// @Summary     Get user organizations
// @Description Retrieve a list of all organizations the authenticated user belongs to.
// @Tags        Organizations
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=[]Organization} "Success"
// @Failure     401 {object} utils.APIResponse "Unauthorized"
// @Failure     500 {object} utils.APIResponse "Failed to fetch organizations"
// @Router      /organizations [get]
func (h *Handler) GetUserOrganizations(c *gin.Context) {
	// Ambil user ID dari context (hasil set middleware RequireAuth)
	user := c.MustGet("user").(models.MinimalUser)

	orgs, err := h.service.GetUserOrganizations(user.ID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch organizations")
		return
	}

	// Jika sukses, kembalikan list organisasinya
	utils.SendSuccess(c, "Success", orgs)
}

// CreateOrgRequest represents the request body for creating an organization.
// @Description Request body for organization creation
type CreateOrgRequest struct {
	Name string `json:"name" binding:"required" example:"My Organization"`
}

// InviteMemberRequest represents the request body for inviting a member.
// @Description Request body for inviting a member to organization
type InviteMemberRequest struct {
	Email string `json:"email" binding:"required" example:"member@example.com"`
}

// UpdateMemberRoleRequest represents the request body for updating a member's role.
// @Description Request body for updating member role
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required" example:"admin"`
}

type Handler struct {
	service OrganizationService
}

func NewOrganizationHandler(service OrganizationService) *Handler {
	return &Handler{service: service}
}

// CreateOrganization godoc
// @Summary     Create organization
// @Description Create a new organization. The requesting user automatically becomes the owner.
// @Tags        Organizations
// @Accept      json
// @Produce     json
// @Param       body body CreateOrgRequest true "Organization payload"
// @Security    BearerAuth
// @Success     201 {object} utils.APIResponse{data=Organization} "Organization created successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     401 {object} utils.APIResponse "Unauthorized"
// @Failure     500 {object} utils.APIResponse "Failed to create organization"
// @Router      /organizations [post]
func (h *Handler) CreateOrganization(c *gin.Context) {
	var req CreateOrgRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	user := c.MustGet("user").(models.MinimalUser)

	org, err := h.service.CreateOrganization(req.Name, user.ID)
	if err != nil {
		if quotaErr, ok := err.(*utils.QuotaError); ok {
			utils.SendError(c, http.StatusForbidden, quotaErr.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "Failed to create organization")
		return
	}

	utils.SendSuccess(c, "Organization created successfully", org)
}

// InviteMember godoc
// @Summary     Invite a member
// @Description Send an invitation email to add a new member to the organization. Requires owner or admin role.
// @Tags        Organizations
// @Accept      json
// @Produce     json
// @Param       body body InviteMemberRequest true "Invitation payload"
// @Param       X-Organization-ID header string true "Organization ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Member added successfully"
// @Failure     400 {object} utils.APIResponse "Validation error or invitation failed"
// @Failure     403 {object} utils.APIResponse "Insufficient permission"
// @Router      /organizations/invite [post]
func (h *Handler) InviteMember(c *gin.Context) {
	var req InviteMemberRequest

	orgIDInterface, exists := c.Get("org_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Organization-ID header is required")
		return
	}
	orgIDStr := orgIDInterface.(string)

	orgID64, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid Organization ID format")
		return
	}
	orgID := uint(orgID64)

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	user := c.MustGet("user").(models.MinimalUser)

	err = h.service.InviteMember(orgID, req.Email, user.ID)
	if err != nil {
		if isPermissionError(err) {
			utils.SendError(c, http.StatusForbidden, "Insufficient permission")
			return
		}
		if quotaErr, ok := err.(*utils.QuotaError); ok {
			utils.SendError(c, http.StatusForbidden, quotaErr.Error())
			return
		}
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Member added successfully")
}

// GetMembers godoc
// @Summary     Get organization members
// @Description Retrieve all members of the organization with their roles.
// @Tags        Organizations
// @Produce     json
// @Param       X-Organization-ID header string true "Organization ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Success"
// @Failure     400 {object} utils.APIResponse "Missing organization header"
// @Failure     500 {object} utils.APIResponse "Failed to fetch members"
// @Router      /organizations/members [get]
func (h *Handler) GetMembers(c *gin.Context) {
	orgIDInterface, exists := c.Get("org_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Organization-ID header is required")
		return
	}
	orgIDStr := orgIDInterface.(string)

	orgID64, _ := strconv.ParseUint(orgIDStr, 10, 64)

	users, err := h.service.GetMembers(uint(orgID64))
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch members")
		return
	}

	utils.SendSuccess(c, "Success", users)
}

// RemoveMember godoc
// @Summary     Remove a member
// @Description Remove a member from the organization. Requires owner or admin role. Owner cannot be removed.
// @Tags        Organizations
// @Produce     json
// @Param       user_id path int true "User ID to remove"
// @Param       X-Organization-ID header string true "Organization ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Member removed successfully"
// @Failure     400 {object} utils.APIResponse "Invalid user ID format"
// @Failure     403 {object} utils.APIResponse "Insufficient permission or cannot remove owner"
// @Router      /organizations/members/{user_id} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	orgIDInterface, exists := c.Get("org_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Organization-ID header is required")
		return
	}
	orgIDStr := orgIDInterface.(string)

	orgID64, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid Organization ID format")
		return
	}
	orgID := uint(orgID64)

	targetUserIDStr := c.Param("user_id")
	targetUserID64, err := strconv.ParseUint(targetUserIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid user ID format")
		return
	}
	targetUserID := uint(targetUserID64)

	requester := c.MustGet("user").(models.MinimalUser)

	err = h.service.RemoveMember(orgID, targetUserID, requester.ID)
	if err != nil {
		if isPermissionError(err) {
			utils.SendError(c, http.StatusForbidden, "Insufficient permission")
			return
		}
		if isOwnerError(err) {
			utils.SendError(c, http.StatusForbidden, "Cannot remove organization owner")
			return
		}
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Member removed successfully")
}

// UpdateMemberRole godoc
// @Summary     Update member role
// @Description Update a member's role within the organization. Requires owner or admin role. Cannot change owner's role.
// @Tags        Organizations
// @Accept      json
// @Produce     json
// @Param       user_id path int true "User ID to update"
// @Param       body body UpdateMemberRoleRequest true "Role update payload"
// @Param       X-Organization-ID header string true "Organization ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Role updated successfully"
// @Failure     400 {object} utils.APIResponse "Invalid role or user ID"
// @Failure     403 {object} utils.APIResponse "Insufficient permission or cannot change owner role"
// @Router      /organizations/members/{user_id} [patch]
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	orgIDInterface, exists := c.Get("org_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Organization-ID header is required")
		return
	}
	orgIDStr := orgIDInterface.(string)

	orgID64, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid Organization ID format")
		return
	}
	orgID := uint(orgID64)

	targetUserIDStr := c.Param("user_id")
	targetUserID64, err := strconv.ParseUint(targetUserIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid user ID format")
		return
	}
	targetUserID := uint(targetUserID64)

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	newRole := models.Role(req.Role)
	if !newRole.IsValid() {
		utils.SendError(c, http.StatusBadRequest, "Invalid role. Must be 'admin' or 'member'")
		return
	}

	requester := c.MustGet("user").(models.MinimalUser)

	err = h.service.UpdateMemberRole(orgID, targetUserID, newRole, requester.ID)
	if err != nil {
		if isPermissionError(err) {
			utils.SendError(c, http.StatusForbidden, "Insufficient permission")
			return
		}
		if isOwnerError(err) {
			utils.SendError(c, http.StatusForbidden, "Cannot change owner role")
			return
		}
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Role updated successfully", gin.H{
		"user_id": targetUserID,
		"role":    newRole,
	})
}

// isPermissionError checks if the error is a permission-related error.
func isPermissionError(err error) bool {
	return err != nil && (errors.Is(err, errors.New("insufficient permission to invite members")) ||
		errors.Is(err, errors.New("insufficient permission to remove members")) ||
		errors.Is(err, errors.New("insufficient permission to update member roles")))
}

// isOwnerError checks if the error is related to owner protection.
func isOwnerError(err error) bool {
	return err != nil && (errors.Is(err, errors.New("cannot remove the organization owner")) ||
		errors.Is(err, errors.New("cannot change owner role")) ||
		errors.Is(err, errors.New("cannot promote someone to owner")))
}
