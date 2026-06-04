package workspaces

import (
	"net/http"
	"strconv"

	"gotask-backend/models"
	"gotask-backend/utils"

	"github.com/gin-gonic/gin"
)

// AcceptInvitationRequest represents the request body for accepting an invitation.
type AcceptInvitationRequest struct {
	Token string `json:"token" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ResendInvitationRequest represents the request body for resending an invitation.
type ResendInvitationRequest struct {
	InvitationID uint `json:"invitation_id" binding:"required" example:"1"`
}

type InvitationHandler struct {
	service InvitationService
	wsRepo  WorkspaceRepository
}

func NewInvitationHandler(service InvitationService, wsRepo WorkspaceRepository) *InvitationHandler {
	return &InvitationHandler{
		service: service,
		wsRepo:  wsRepo,
	}
}

// AcceptInvitation godoc
// @Summary     Accept an invitation
// @Description Accept a workspace invitation using the token received via email. The authenticated user must match the invited email.
// @Tags        Invitations
// @Accept      json
// @Produce     json
// @Param       body body AcceptInvitationRequest true "Accept invitation payload"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=Workspace} "Successfully joined the workspace"
// @Failure     400 {object} utils.APIResponse "Invalid token or already used"
// @Failure     401 {object} utils.APIResponse "User email does not match invitation"
// @Failure     410 {object} utils.APIResponse "Invitation has expired or been revoked"
// @Router      /invite/accept [post]
func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	var req AcceptInvitationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	ws, err := h.service.AcceptInvitation(req.Token)
	if err != nil {
		switch err.Error() {
		case "invitation has expired":
			utils.SendError(c, http.StatusGone, "Invitation has expired")
		case "invitation has already been used":
			utils.SendError(c, http.StatusConflict, "Invitation has already been used")
		case "invitation has been revoked":
			utils.SendError(c, http.StatusGone, "Invitation has been revoked")
		case "cannot identify accepting user":
			utils.SendError(c, http.StatusUnauthorized, "Please login with the invited email to accept")
		default:
			utils.SendError(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	utils.SendSuccess(c, "Successfully joined the workspace", ws)
}

// GetInvitations godoc
// @Summary     List pending invitations
// @Description Retrieve all pending invitations for the workspace. Requires owner or admin role.
// @Tags        Invitations
// @Produce     json
// @Param       X-Workspace-ID header string true "Workspace ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=[]WorkspaceInvitation} "Success"
// @Failure     400 {object} utils.APIResponse "Missing workspace header"
// @Failure     403 {object} utils.APIResponse "Insufficient permission"
// @Failure     500 {object} utils.APIResponse "Failed to fetch invitations"
// @Router      /workspaces/invitations [get]
func (h *InvitationHandler) GetInvitations(c *gin.Context) {
	workspaceIDInterface, exists := c.Get("workspace_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Workspace-ID header is required")
		return
	}
	workspaceIDStr := workspaceIDInterface.(string)
	workspaceID, err := strconv.ParseUint(workspaceIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid workspace ID")
		return
	}

	// Check permission
	user := c.MustGet("user")

	var requesterID uint
	if u, ok := user.(interface{ GetID() uint }); ok {
		requesterID = u.GetID()
	}

	requesterRole, err := h.getRequesterRole(requesterID, uint(workspaceID))
	if err != nil || !requesterRole.CanInvite() {
		utils.SendError(c, http.StatusForbidden, "Insufficient permission")
		return
	}

	invitations, err := h.service.GetPendingInvitations(uint(workspaceID))
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch invitations")
		return
	}

	utils.SendSuccess(c, "success", invitations)
}

// ResendInvitation godoc
// @Summary     Resend an invitation
// @Description Resend an invitation email to the previously invited email address. Extends the expiry time.
// @Tags        Invitations
// @Accept      json
// @Produce     json
// @Param       body body ResendInvitationRequest true "Resend invitation payload"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Invitation resent successfully"
// @Failure     400 {object} utils.APIResponse "Invalid invitation ID"
// @Router      /invite/resend [post]
func (h *InvitationHandler) ResendInvitation(c *gin.Context) {
	var req ResendInvitationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.ResendInvitation(req.InvitationID)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Invitation resent successfully")
}

// RevokeInvitation godoc
// @Summary     Revoke an invitation
// @Description Revoke an existing invitation token, preventing it from being accepted. Requires owner or admin role.
// @Tags        Invitations
// @Produce     json
// @Param       token path string true "Invitation token to revoke"
// @Param       X-Workspace-ID header string true "Workspace ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Invitation revoked successfully"
// @Failure     400 {object} utils.APIResponse "Token is required"
// @Failure     403 {object} utils.APIResponse "Insufficient permission"
// @Router      /invite/{token} [delete]
func (h *InvitationHandler) RevokeInvitation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.SendError(c, http.StatusBadRequest, "Token is required")
		return
	}

	workspaceIDInterface, exists := c.Get("workspace_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Workspace-ID header is required")
		return
	}
	workspaceIDStr := workspaceIDInterface.(string)
	workspaceID, err := strconv.ParseUint(workspaceIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid workspace ID")
		return
	}

	user := c.MustGet("user")

	var requesterID uint
	if u, ok := user.(interface{ GetID() uint }); ok {
		requesterID = u.GetID()
	}

	err = h.service.RevokeInvitation(token, requesterID, uint(workspaceID))
	if err != nil {
		if isPermissionError(err) {
			utils.SendError(c, http.StatusForbidden, "Insufficient permission")
			return
		}
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Invitation revoked successfully")
}

// getRequesterRole is a helper to get the requester's role from the repository.
func (h *InvitationHandler) getRequesterRole(userID, workspaceID uint) (models.Role, error) {
	return h.wsRepo.GetMemberRole(userID, workspaceID)
}
