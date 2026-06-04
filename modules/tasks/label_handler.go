package tasks

import (
	"net/http"
	"strconv"

	"gotask-backend/utils"

	"github.com/gin-gonic/gin"
)

// CreateLabelRequest represents the request body for creating a label.
// @Description Request body for creating a label
type CreateLabelRequest struct {
	Name  string `json:"name" binding:"required" example:"Bug"`
	Color string `json:"color" example:"#EF4444"`
}

// UpdateLabelRequest represents the request body for updating a label.
// @Description Request body for updating a label
type UpdateLabelRequest struct {
	Name  *string `json:"name" example:"Feature"`
	Color *string `json:"color" example:"#10B981"`
}

type LabelHandler struct {
	service LabelService
}

func NewLabelHandler(service LabelService) *LabelHandler {
	return &LabelHandler{service: service}
}

// CreateLabel godoc
// @Summary     Create a label
// @Description Create a new label for categorizing tasks within a project.
// @Tags        Labels
// @Accept      json
// @Produce     json
// @Param       id path int true "Project ID"
// @Param       body body CreateLabelRequest true "Label payload"
// @Param       X-Workspace-ID header string false "Workspace ID (optional - auto-resolved if not provided)"
// @Security    BearerAuth
// @Success     201 {object} utils.APIResponse{data=Label} "Label created successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     500 {object} utils.APIResponse "Failed to create label"
// @Router      /projects/{id}/labels [post]
func (h *LabelHandler) CreateLabel(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	wsIDInterface, exists := c.Get("workspace_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "Workspace context not found")
		return
	}
	wsID := wsIDInterface.(string)

	var req CreateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.Color == "" {
		req.Color = "#6366F1"
	}

	user := c.MustGet("user")

	var requesterID uint
	if u, ok := user.(interface{ GetID() uint }); ok {
		requesterID = u.GetID()
	}

	label, err := h.service.CreateLabel(uint(projectID), req.Name, req.Color, requesterID, wsID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccess(c, "Label created successfully", label)
}

// GetLabels godoc
// @Summary     List project labels
// @Description Retrieve all labels belonging to a specific project.
// @Tags        Labels
// @Produce     json
// @Param       id path int true "Project ID"
// @Param       X-Workspace-ID header string false "Workspace ID (optional - auto-resolved if not provided)"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "success"
// @Failure     400 {object} utils.APIResponse "Missing workspace context"
// @Failure     500 {object} utils.APIResponse "Failed to fetch labels"
// @Router      /projects/{id}/labels [get]
func (h *LabelHandler) GetLabels(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	wsIDInterface, exists := c.Get("workspace_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "Workspace context not found")
		return
	}
	wsID := wsIDInterface.(string)

	labels, err := h.service.GetLabels(uint(projectID), wsID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccess(c, "success", labels)
}

// UpdateLabel godoc
// @Summary     Update a label
// @Description Update a label's name or color.
// @Tags        Labels
// @Accept      json
// @Produce     json
// @Param       id path int true "Label ID"
// @Param       body body UpdateLabelRequest true "Label update payload"
// @Param       X-Workspace-ID header string false "Workspace ID (optional - auto-resolved if not provided)"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=Label} "Label updated successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     500 {object} utils.APIResponse "Failed to update label"
// @Router      /labels/{id} [patch]
func (h *LabelHandler) UpdateLabel(c *gin.Context) {
	labelIDStr := c.Param("id")
	labelID, err := strconv.ParseUint(labelIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid label ID")
		return
	}

	wsIDInterface, exists := c.Get("workspace_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "Workspace context not found")
		return
	}
	wsID := wsIDInterface.(string)

	var req UpdateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	label, err := h.service.UpdateLabel(uint(labelID), req.Name, req.Color, wsID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccess(c, "Label updated successfully", label)
}

// DeleteLabel godoc
// @Summary     Delete a label
// @Description Delete a label. The label will be removed from all associated tasks.
// @Tags        Labels
// @Produce     json
// @Param       id path int true "Label ID"
// @Param       X-Workspace-ID header string false "Workspace ID (optional - auto-resolved if not provided)"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Label deleted successfully"
// @Failure     500 {object} utils.APIResponse "Failed to delete label"
// @Router      /labels/{id} [delete]
func (h *LabelHandler) DeleteLabel(c *gin.Context) {
	labelIDStr := c.Param("id")
	labelID, err := strconv.ParseUint(labelIDStr, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid label ID")
		return
	}

	wsIDInterface, exists := c.Get("workspace_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "Workspace context not found")
		return
	}
	wsID := wsIDInterface.(string)

	err = h.service.DeleteLabel(uint(labelID), wsID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccess(c, "Label deleted successfully")
}
