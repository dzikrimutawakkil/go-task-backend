package projects

import (
	"gotask-backend/modules/auth"
	"gotask-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateProjectRequest represents the request body for creating a project.
// @Description Request body for project creation
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required" example:"Website Redesign"`
	Description string `json:"description" example:"Complete redesign of company website"`
}

// UpdateProjectRequest represents the request body for updating a project.
// @Description Request body for project update (all fields optional)
type UpdateProjectRequest struct {
	Name        *string  `json:"name" example:"Updated Project Name"`
	Description *string  `json:"description" example:"Updated description"`
	Status      *string  `json:"status" example:"in_progress"`
	Priority    *string  `json:"priority" example:"high"`
	Budget      *float64 `json:"budget" example:"5000000"`
	Deadline    *string  `json:"deadline" example:"2026-06-30"`
	Progress    *int     `json:"progress" example:"50"`
}

type ProjectHandler struct {
	service ProjectService
}

func NewProjectHandler(service ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// FindProjects godoc
// @Summary     List projects
// @Description Retrieve all projects belonging to the current organization.
// @Tags        Projects
// @Produce     json
// @Param       X-Organization-ID header string true "Organization ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=[]Project} "Success"
// @Failure     400 {object} utils.APIResponse "Missing organization header"
// @Failure     500 {object} utils.APIResponse "Failed to fetch projects"
// @Router      /projects [get]
func (h *ProjectHandler) FindProjects(c *gin.Context) {
	orgID := c.MustGet("org_id").(string)

	projects, err := h.service.GetProjects(orgID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch projects")
		return
	}

	utils.SendSuccess(c, "Success", projects)
}

// CreateProject godoc
// @Summary     Create a project
// @Description Create a new project within the organization.
// @Tags        Projects
// @Accept      json
// @Produce     json
// @Param       body body CreateProjectRequest true "Project payload"
// @Param       X-Organization-ID header string true "Organization ID"
// @Security    BearerAuth
// @Success     201 {object} utils.APIResponse{data=Project} "Project created successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     500 {object} utils.APIResponse "Failed to create project"
// @Router      /projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	orgIDStr := c.MustGet("org_id").(string)
	orgID64, _ := strconv.ParseUint(orgIDStr, 10, 64)

	var jsonInput CreateProjectRequest

	if err := c.ShouldBindJSON(&jsonInput); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	user := c.MustGet("user").(auth.User)

	input := CreateProjectInput{
		Name:           jsonInput.Name,
		Description:    jsonInput.Description,
		OrganizationID: uint(orgID64),
	}

	project, err := h.service.CreateProject(input, user.ID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to create project")
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
// @Param       X-Organization-ID header string true "Organization ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Project deleted successfully"
// @Failure     403 {object} utils.APIResponse "Insufficient permission"
// @Failure     404 {object} utils.APIResponse "Project not found"
// @Router      /projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	orgID := c.MustGet("org_id").(string)

	user := c.MustGet("user").(auth.User)

	if err := h.service.DeleteProject(id, orgID, user.ID); err != nil {
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
// @Param       X-Organization-ID header string true "Organization ID"
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
// @Description Update a project's fields (name, description, status, priority, budget, deadline, progress).
// @Tags        Projects
// @Produce     json
// @Param       id path int true "Project ID"
// @Param       body body UpdateProjectRequest true "Project update payload"
// @Param       X-Organization-ID header string true "Organization ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=Project} "Project updated successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     404 {object} utils.APIResponse "Project not found"
// @Router      /projects/{id} [patch]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")

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
	if req.Status != nil {
		input.Status = req.Status
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

	project, err := h.service.UpdateProject(id, input)
	if err != nil {
		if err.Error() == "project not found" {
			utils.SendError(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "Failed to update project")
		return
	}

	utils.SendSuccess(c, "Project updated successfully", project)
}
