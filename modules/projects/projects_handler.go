package projects

import (
	"gotask-backend/config"
	"gotask-backend/models"
	"gotask-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /projects
func FindProjects(c *gin.Context) {
	// 1. Get Org ID from Context (Saved by Middleware)
	// We check if it exists because the middleware only sets it if the header was sent.
	orgIDInterface, exists := c.Get("org_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Organization-ID header is required for this endpoint")
		return
	}
	orgID := orgIDInterface.(string)

	// 2. Fetch Projects
	var projects []models.Project
	err := config.DB.
		Preload("Tasks.Status").
		Preload("Tasks.Assignees").
		Preload("Tasks.Priority").
		Scopes(models.ByOrg(orgID)).
		Find(&projects).Error

	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch projects")
		return
	}

	utils.SendSuccess(c, "Success", projects)
}

// POST /projects
func CreateProject(c *gin.Context) {
	// 1. Input now requires OrganizationID
	var input struct {
		Name           string `json:"name" binding:"required"`
		Description    string `json:"description"`
		OrganizationID uint   `json:"organization_id" binding:"required"` // <--- NEW
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Get logged-in user
	userContext, _ := c.Get("user")
	user := userContext.(models.User)

	// 3. SECURITY CHECK: Does this user belong to the requested Org?
	// We query the join table 'organization_users' to see if a link exists
	var count int64
	config.DB.Table("organization_users").
		Where("user_id = ? AND organization_id = ?", user.ID, input.OrganizationID).
		Count(&count)

	if count == 0 {
		utils.SendError(c, http.StatusForbidden, "You are not a member of this organization")
		return
	}

	// 4. Create the Project
	project := models.Project{
		Name:           input.Name,
		Description:    input.Description,
		OrganizationID: input.OrganizationID,
		Users:          []models.User{user},
	}

	if err := config.DB.Create(&project).Error; err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to create project")
		return
	}

	config.DB.Preload("Organization").Preload("Users").First(&project, project.ID)

	utils.SendSuccess(c, "Project created successfully", project)
}

// DELETE /projects/:id
func DeleteProject(c *gin.Context) {
	id := c.Param("id")
	userContext, _ := c.Get("user")
	user := userContext.(models.User)

	// 1. Find Project and verify ownership
	var project models.Project
	err := config.DB.Model(&user).Association("Projects").Find(&project, id)

	if err != nil || project.ID == 0 {
		utils.SendError(c, http.StatusNotFound, "Project not found or access denied")
		return
	}

	// 2. CLEANUP: Remove "Assignees" from the tasks (The invisible blocker!)
	// We need to delete from the 'task_users' join table first.
	// We use a raw query because GORM Many-to-Many delete can be tricky.
	if err := config.DB.Exec("DELETE FROM task_users WHERE task_id IN (SELECT id FROM tasks WHERE project_id = ?)", project.ID).Error; err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to clear task assignees")
		return
	}

	// 3. Delete the Tasks
	// Now safe to delete because assignees are gone
	if err := config.DB.Where("project_id = ?", project.ID).Delete(&models.Task{}).Error; err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to delete project tasks")
		return
	}

	// 4. Clear Project Members (project_users table)
	if err := config.DB.Model(&project).Association("Users").Clear(); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to remove project members")
		return
	}

	// 5. Finally, Delete the Project
	if err := config.DB.Delete(&project).Error; err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	utils.SendSuccess(c, "Project and all its tasks deleted successfully")
}

// POST /projects/:id/invite - Add a user to a project by email
func InviteUser(c *gin.Context) {
	// 1. Get Project ID
	projectID := c.Param("id")

	// 2. Get Email from body
	var body struct {
		Email string `json:"email"`
	}
	if c.ShouldBindJSON(&body) != nil {
		utils.SendError(c, http.StatusBadRequest, "Email is required")
		return
	}

	// 3. Find the Project
	var project models.Project
	if err := config.DB.First(&project, projectID).Error; err != nil {
		utils.SendError(c, http.StatusNotFound, "Project not found")
		return
	}

	// 4. Find the User to invite
	var userToInvite models.User
	if err := config.DB.Where("email = ?", body.Email).First(&userToInvite).Error; err != nil {
		utils.SendError(c, http.StatusNotFound, "User with that email not found")
		return
	}

	// 5. Add relationship
	config.DB.Model(&project).Association("Users").Append(&userToInvite)

	utils.SendSuccess(c, "User added to project!")
}

func FindProjectMembers(c *gin.Context) {
	projectId := c.Param("id")

	useroContext, _ := c.Get("user")
	loggedUser := useroContext.(models.User)

	var project models.Project
	err := config.DB.Model(&loggedUser).Association("Projects").Find(&project, projectId)

	if err != nil || project.ID == 0 {
		utils.SendError(c, http.StatusNotFound, "Project not found or access denied")
		return
	}

	var members []models.User
	if err := config.DB.Model(&project).Association("Users").Find(&members); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to get project members")
	}

	utils.SendSuccess(c, "Project member fetched successfully", members)
}
