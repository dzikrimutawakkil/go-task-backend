package organizations

import (
	"gotask-backend/models"
	"gotask-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service OrganizationService
}

func NewOrganizationHandler(service OrganizationService) *Handler {
	return &Handler{service: service}
}

// POST /organizations
func (h *Handler) CreateOrganization(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get Current User from Context
	user := c.MustGet("user").(models.User)

	org, err := h.service.CreateOrganization(req.Name, user.ID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to create organization")
		return
	}

	utils.SendSuccess(c, "Organization created successfully", org)
}

// POST /organizations/invite
func (h *Handler) InviteMember(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}

	// 1. Get Org ID from Context (It is a String)
	orgIDInterface, exists := c.Get("org_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Organization-ID header is required")
		return
	}
	orgIDStr := orgIDInterface.(string)

	// 2. CONVERT String -> Uint
	orgID64, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid Organization ID format")
		return
	}
	orgID := uint(orgID64)

	// 3. Bind JSON Body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 4. Call Service (Now passing uint)
	err = h.service.InviteMember(orgID, req.Email)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Member added successfully")
}

func (h *Handler) GetMembers(c *gin.Context) {
	// 1. Get Org ID from Header
	orgIDStr := c.MustGet("org_id").(string)
	orgID, _ := strconv.ParseUint(orgIDStr, 10, 64)

	// 2. Fetch
	users, err := h.service.GetMembers(uint(orgID))
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch members")
		return
	}

	utils.SendSuccess(c, "Success", users)
}
