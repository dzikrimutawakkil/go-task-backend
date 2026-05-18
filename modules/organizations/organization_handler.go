package organizations

import (
	"errors"
	"gotask-backend/models"
	"gotask-backend/modules/auth"
	"gotask-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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

	user := c.MustGet("user").(auth.User)

	org, err := h.service.CreateOrganization(req.Name, user.ID)
	if err != nil {
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

	user := c.MustGet("user").(auth.User)

	err = h.service.InviteMember(orgID, req.Email, user.ID)
	if err != nil {
		if isPermissionError(err) {
			utils.SendError(c, http.StatusForbidden, "Insufficient permission")
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

	requester := c.MustGet("user").(auth.User)

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

	requester := c.MustGet("user").(auth.User)

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
