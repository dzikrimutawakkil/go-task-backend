package workspaces

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
// @Description Get the authenticated user's tier information for the active workspace including usage statistics.
// @Tags        Tiers
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=TierInfoResponse} "Success"
// @Failure     401 {object} utils.APIResponse "Unauthorized"
// @Failure     500 {object} utils.APIResponse "Failed to fetch tier info"
// @Router      /users/me/tier [get]
func (h *Handler) GetMyTierInfo(c *gin.Context) {
	workspaceIDStr := c.MustGet("workspace_id").(string)
	workspaceID, err := strconv.ParseUint(workspaceIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Invalid workspace ID")
		return
	}

	info, err := h.service.GetTierInfoForWorkspace(uint(workspaceID))
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch tier info")
		return
	}
	utils.SendSuccess(c, "OK", info)
}

// ActivateTierRequest represents the request body for activating a tier.
type ActivateTierRequest struct {
	Tier           string `json:"tier" binding:"required" example:"starter"`
	DurationMonths int    `json:"duration_months" binding:"required,min=1,max=24" example:"12"`
}

// ActivateTier godoc
// @Summary     Activate tier for workspace (Admin only)
// @Description Activate a subscription tier for a workspace. Requires admin role.
// @Tags        Tiers
// @Accept      json
// @Produce     json
// @Param       id path int true "Workspace ID"
// @Param       body body ActivateTierRequest true "Tier activation payload"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=ActivateTierResult} "Tier activated successfully"
// @Failure     400 {object} utils.APIResponse "Invalid request"
// @Failure     403 {object} utils.APIResponse "Insufficient permission"
// @Failure     404 {object} utils.APIResponse "Workspace not found"
// @Router      /admin/workspaces/{id}/tier [patch]
func (h *Handler) ActivateTier(c *gin.Context) {
	workspaceIDStr := c.Param("id")
	workspaceID64, err := strconv.ParseUint(workspaceIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid workspace ID format")
		return
	}
	workspaceID := uint(workspaceID64)

	var req ActivateTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Validate tier value - new tier names (free/pro/ultimate)
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

	result, err := h.service.ActivateTier(workspaceID, req.Tier, req.DurationMonths, requester.ID)
	if err != nil {
		if err.Error() == "workspace not found" {
			utils.SendError(c, http.StatusNotFound, "Workspace not found")
			return
		}
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Tier activated successfully", result)
}

// GetUserWorkspaces godoc
// @Summary     Get user workspaces
// @Description Retrieve a list of all workspaces the authenticated user belongs to.
// @Tags        Workspaces
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=[]Workspace} "Success"
// @Failure     401 {object} utils.APIResponse "Unauthorized"
// @Failure     500 {object} utils.APIResponse "Failed to fetch workspaces"
// @Router      /workspaces [get]
func (h *Handler) GetUserWorkspaces(c *gin.Context) {
	user := c.MustGet("user").(models.MinimalUser)

	workspaces, err := h.service.GetUserWorkspaces(user.ID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch workspaces")
		return
	}

	utils.SendSuccess(c, "Success", workspaces)
}

// CreateWorkspaceRequest represents the request body for creating a workspace.
type CreateWorkspaceRequest struct {
	Name string `json:"name" binding:"required" example:"My Workspace"`
}

// InviteMemberRequest represents the request body for inviting a member.
type InviteMemberRequest struct {
	Email string `json:"email" binding:"required" example:"member@example.com"`
}

// UpdateMemberRoleRequest represents the request body for updating a member's role.
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required" example:"admin"`
}

type Handler struct {
	service WorkspaceService
}

func NewWorkspaceHandler(service WorkspaceService) *Handler {
	return &Handler{service: service}
}

// CreateWorkspace godoc
// @Summary     Create workspace
// @Description Create a new workspace. The requesting user automatically becomes the owner.
// @Tags        Workspaces
// @Accept      json
// @Produce     json
// @Param       body body CreateWorkspaceRequest true "Workspace payload"
// @Security    BearerAuth
// @Success     201 {object} utils.APIResponse{data=Workspace} "Workspace created successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     401 {object} utils.APIResponse "Unauthorized"
// @Failure     500 {object} utils.APIResponse "Failed to create workspace"
// @Router      /workspaces [post]
func (h *Handler) CreateWorkspace(c *gin.Context) {
	var req CreateWorkspaceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	user := c.MustGet("user").(models.MinimalUser)

	ws, err := h.service.CreateWorkspace(req.Name, user.ID)
	if err != nil {
		if quotaErr, ok := err.(*utils.QuotaError); ok {
			utils.SendError(c, http.StatusForbidden, quotaErr.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "Failed to create workspace")
		return
	}

	utils.SendSuccess(c, "Workspace created successfully", ws)
}

// InviteMember godoc
// @Summary     Invite a member
// @Description Send an invitation email to add a new member to the workspace. Requires owner or admin role.
// @Tags        Workspaces
// @Accept      json
// @Produce     json
// @Param       body body InviteMemberRequest true "Invitation payload"
// @Param       X-Workspace-ID header string true "Workspace ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Member added successfully"
// @Failure     400 {object} utils.APIResponse "Validation error or invitation failed"
// @Failure     403 {object} utils.APIResponse "Insufficient permission"
// @Router      /workspaces/invite [post]
func (h *Handler) InviteMember(c *gin.Context) {
	var req InviteMemberRequest

	workspaceIDInterface, exists := c.Get("workspace_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Workspace-ID header is required")
		return
	}
	workspaceIDStr := workspaceIDInterface.(string)

	workspaceID64, err := strconv.ParseUint(workspaceIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid Workspace ID format")
		return
	}
	workspaceID := uint(workspaceID64)

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	user := c.MustGet("user").(models.MinimalUser)

	err = h.service.InviteMember(workspaceID, req.Email, user.ID)
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
// @Summary     Get workspace members
// @Description Retrieve all members of the workspace with their roles.
// @Tags        Workspaces
// @Produce     json
// @Param       X-Workspace-ID header string true "Workspace ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Success"
// @Failure     400 {object} utils.APIResponse "Missing workspace header"
// @Failure     500 {object} utils.APIResponse "Failed to fetch members"
// @Router      /workspaces/members [get]
func (h *Handler) GetMembers(c *gin.Context) {
	workspaceIDInterface, exists := c.Get("workspace_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Workspace-ID header is required")
		return
	}
	workspaceIDStr := workspaceIDInterface.(string)

	workspaceID64, _ := strconv.ParseUint(workspaceIDStr, 10, 64)

	users, err := h.service.GetMembers(uint(workspaceID64))
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch members")
		return
	}

	utils.SendSuccess(c, "Success", users)
}

// RemoveMember godoc
// @Summary     Remove a member
// @Description Remove a member from the workspace. Requires owner or admin role. Owner cannot be removed.
// @Tags        Workspaces
// @Produce     json
// @Param       user_id path int true "User ID to remove"
// @Param       X-Workspace-ID header string true "Workspace ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Member removed successfully"
// @Failure     400 {object} utils.APIResponse "Invalid user ID format"
// @Failure     403 {object} utils.APIResponse "Insufficient permission or cannot remove owner"
// @Router      /workspaces/members/{user_id} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	workspaceIDInterface, exists := c.Get("workspace_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Workspace-ID header is required")
		return
	}
	workspaceIDStr := workspaceIDInterface.(string)

	workspaceID64, err := strconv.ParseUint(workspaceIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid Workspace ID format")
		return
	}
	workspaceID := uint(workspaceID64)

	targetUserIDStr := c.Param("user_id")
	targetUserID64, err := strconv.ParseUint(targetUserIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid user ID format")
		return
	}
	targetUserID := uint(targetUserID64)

	requester := c.MustGet("user").(models.MinimalUser)

	err = h.service.RemoveMember(workspaceID, targetUserID, requester.ID)
	if err != nil {
		if isPermissionError(err) {
			utils.SendError(c, http.StatusForbidden, "Insufficient permission")
			return
		}
		if isOwnerError(err) {
			utils.SendError(c, http.StatusForbidden, "Cannot remove workspace owner")
			return
		}
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Member removed successfully")
}

// UpdateMemberRole godoc
// @Summary     Update member role
// @Description Update a member's role within the workspace. Requires owner or admin role. Cannot change owner's role.
// @Tags        Workspaces
// @Accept      json
// @Produce     json
// @Param       user_id path int true "User ID to update"
// @Param       body body UpdateMemberRoleRequest true "Role update payload"
// @Param       X-Workspace-ID header string true "Workspace ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Role updated successfully"
// @Failure     400 {object} utils.APIResponse "Invalid role or user ID"
// @Failure     403 {object} utils.APIResponse "Insufficient permission or cannot change owner role"
// @Router      /workspaces/members/{user_id} [patch]
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	workspaceIDInterface, exists := c.Get("workspace_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Workspace-ID header is required")
		return
	}
	workspaceIDStr := workspaceIDInterface.(string)

	workspaceID64, err := strconv.ParseUint(workspaceIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid Workspace ID format")
		return
	}
	workspaceID := uint(workspaceID64)

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

	err = h.service.UpdateMemberRole(workspaceID, targetUserID, newRole, requester.ID)
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
	return err != nil && (errors.Is(err, errors.New("cannot remove the workspace owner")) ||
		errors.Is(err, errors.New("cannot change owner role")) ||
		errors.Is(err, errors.New("cannot promote someone to owner")))
}
