package projects

import (
	"gotask-backend/models"
	"gotask-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateProjectRequest represents the request body for creating a project.
// M-MIGRATION: Added client_id and new_client support
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required" example:"Website Redesign"`
	Description string `json:"description" example:"Complete redesign of company website"`
	ClientID    *uint  `json:"client_id" example:"1" description:"ID of existing client to link to this project"`
	NewClient   *struct {
		Name     string `json:"name" binding:"required" example:"PT Maju Jaya"`
		Email    string `json:"email" example:"contact@majujaya.co.id"`
		WhatsApp string `json:"whatsapp" example:"6281234567890"`
		Phone    string `json:"phone" example:"+62 21 1234567"`
		Company  string `json:"company" example:"PT Maju Jaya"`
	} `json:"new_client" description:"Create a new client inline while creating this project"`
}

// UpdateProjectRequest represents the request body for updating a project.
// M-MIGRATION: Added client_id support for updating/unlinking client
type UpdateProjectRequest struct {
	Name        *string  `json:"name" example:"Updated Project Name"`
	Description *string  `json:"description" example:"Updated description"`
	Status      *string  `json:"status" example:"in_progress"`
	StatusID    *int     `json:"status_id" example:"1" description:"Project status ID from project_statuses table"`
	Priority    *string  `json:"priority" example:"high"`
	Budget      *float64 `json:"budget" example:"5000000"`
	Deadline    *string  `json:"deadline" example:"2026-06-30"`
	Progress    *int     `json:"progress" example:"50"`
	ClientID    *uint    `json:"client_id" example:"1" description:"Set client ID (or 0 to unlink)"`
}

type ProjectHandler struct {
	service ProjectService
}

func NewProjectHandler(service ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// FindProjects godoc
// @Summary     List projects
// @Description Retrieve all projects belonging to the current workspace.
// @Tags        Projects
// @Produce     json
// @Param       X-Workspace-ID header string true "Workspace ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=[]Project} "Success"
// @Failure     400 {object} utils.APIResponse "Missing workspace header"
// @Failure     500 {object} utils.APIResponse "Failed to fetch projects"
// @Router      /projects [get]
func (h *ProjectHandler) FindProjects(c *gin.Context) {
	workspaceID := c.MustGet("workspace_id").(string)

	projects, err := h.service.GetProjects(workspaceID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch projects")
		return
	}

	utils.SendSuccess(c, "Success", projects)
}

// CreateProject godoc
// @Summary     Create a project
// @Description Create a new project within the workspace. Can optionally link to a client or create one inline.
// @Tags        Projects
// @Accept      json
// @Produce     json
// @Param       body body CreateProjectRequest true "Project payload"
// @Param       X-Workspace-ID header string true "Workspace ID"
// @Security    BearerAuth
// @Success     201 {object} utils.APIResponse{data=Project} "Project created successfully"
// @Failure     400 {object} utils.APIResponse "Validation error or cannot use both client_id and new_client"
// @Failure     500 {object} utils.APIResponse "Failed to create project"
// @Router      /projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	workspaceIDStr := c.MustGet("workspace_id").(string)
	workspaceID64, _ := strconv.ParseUint(workspaceIDStr, 10, 64)

	var jsonInput CreateProjectRequest

	if err := c.ShouldBindJSON(&jsonInput); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	user := c.MustGet("user").(models.MinimalUser)

	input := CreateProjectInput{
		Name:        jsonInput.Name,
		Description: jsonInput.Description,
		WorkspaceID: uint(workspaceID64),
		ClientID:    jsonInput.ClientID,
	}

	// Handle inline client creation
	if jsonInput.NewClient != nil {
		input.NewClient = &InlineCreateClientRequest{
			Name:     jsonInput.NewClient.Name,
			Email:    jsonInput.NewClient.Email,
			WhatsApp: jsonInput.NewClient.WhatsApp,
			Phone:    jsonInput.NewClient.Phone,
			Company:  jsonInput.NewClient.Company,
		}
	}

	project, err := h.service.CreateProject(input, user.ID)
	if err != nil {
		if quotaErr, ok := err.(*utils.QuotaError); ok {
			utils.SendError(c, http.StatusForbidden, quotaErr.Error())
			return
		}
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Project created successfully", project)
}

// DeleteProject godoc
// @Summary     Delete a project
// @Description Delete a project and all its associated tasks. Requires owner or admin role.
// @Tags        Projects
// @Produce     json
// @Param       id path int true "Project ID"
// @Param       X-Workspace-ID header string true "Workspace ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Project deleted successfully"
// @Failure     403 {object} utils.APIResponse "Insufficient permission"
// @Failure     404 {object} utils.APIResponse "Project not found"
// @Router      /projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	workspaceID := c.MustGet("workspace_id").(string)

	user := c.MustGet("user").(models.MinimalUser)

	if err := h.service.DeleteProject(id, workspaceID, user.ID); err != nil {
		if err.Error() == "insufficient permission to delete project" {
			utils.SendError(c, http.StatusForbidden, "Insufficient permission")
			return
		}
		utils.SendError(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SendSuccess(c, "Project deleted successfully")
}

// GetProject godoc
// @Summary     Get a project
// @Description Retrieve a single project by ID.
// @Tags        Projects
// @Produce     json
// @Param       id path int true "Project ID"
// @Param       X-Workspace-ID header string true "Workspace ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=Project} "Success"
// @Failure     404 {object} utils.APIResponse "Project not found"
// @Router      /projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	id := c.Param("id")

	project, err := h.service.GetProject(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Project not found")
		return
	}

	utils.SendSuccess(c, "Success", project)
}

// UpdateProject godoc
// @Summary     Update a project
// @Description Update a project's fields including client linking. Use client_id: 0 to unlink.
// @Tags        Projects
// @Produce     json
// @Param       id path int true "Project ID"
// @Param       body body UpdateProjectRequest true "Project update payload"
// @Param       X-Workspace-ID header string true "Workspace ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=Project} "Project updated successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     404 {object} utils.APIResponse "Project not found"
// @Router      /projects/{id} [patch]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	workspaceID := c.MustGet("workspace_id").(string)

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	input := UpdateProjectInput{}
	if req.Name != nil {
		input.Name = req.Name
	}
	if req.Description != nil {
		input.Description = req.Description
	}
	if req.StatusID != nil {
		input.StatusID = req.StatusID
	}
	if req.Priority != nil {
		input.Priority = req.Priority
	}
	if req.Budget != nil {
		input.Budget = req.Budget
	}
	if req.Deadline != nil {
		input.Deadline = req.Deadline
	}
	if req.Progress != nil {
		input.Progress = req.Progress
	}
	if req.ClientID != nil {
		input.ClientID = req.ClientID
	}

	project, err := h.service.UpdateProject(id, input, workspaceID)
	if err != nil {
		if err.Error() == "project not found" {
			utils.SendError(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Project updated successfully", project)
}
