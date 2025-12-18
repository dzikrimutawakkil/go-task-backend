package projects

import (
	"gotask-backend/modules/auth"
	"gotask-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service ProjectService
}

func NewProjectHandler(service ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// GET /projects
func (h *ProjectHandler) FindProjects(c *gin.Context) {
	// Get Org ID from Context (Header: X-Organization-ID)
	orgID := c.MustGet("org_id").(string)

	projects, err := h.service.GetProjects(orgID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch projects")
		return
	}

	utils.SendSuccess(c, "Success", projects)
}

// POST /projects
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	orgIDStr := c.MustGet("org_id").(string)
	orgID64, _ := strconv.ParseUint(orgIDStr, 10, 64)

	var jsonInput struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

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

// DELETE /projects/:id
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	orgID := c.MustGet("org_id").(string)

	if err := h.service.DeleteProject(id, orgID); err != nil {
		utils.SendError(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SendSuccess(c, "Project deleted successfully")
}
