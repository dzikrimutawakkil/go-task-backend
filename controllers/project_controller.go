package controllers

import (
	"gotask-backend/config"
	"gotask-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /projects - List projects the logged-in user belongs to
func FindProjects(c *gin.Context) {
	// 1. Get logged-in user
	userContext, _ := c.Get("user")
	user := userContext.(models.User)

	var projects []models.Project

	// 2. Use GORM Association Mode to find projects for this user
	// This generates SQL like: SELECT * FROM projects JOIN project_users ... WHERE user_id = ?
	err := config.DB.Model(&user).Association("Projects").Find(&projects)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": projects})
}

// POST /projects - Create a project and add the creator as a member
func CreateProject(c *gin.Context) {
	var input models.Project
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Get logged-in user
	userContext, _ := c.Get("user")
	user := userContext.(models.User)

	// 2. Add the current user to the project's Users list
	input.Users = []models.User{user}

	// 3. Save (GORM will insert into 'projects' AND 'project_users')
	config.DB.Create(&input)

	c.JSON(http.StatusOK, gin.H{"data": input})
}

// DELETE /projects/:id
func DeleteProject(c *gin.Context) {
	id := c.Param("id")

	// 1. Get logged-in user (Security check)
	userContext, _ := c.Get("user")
	user := userContext.(models.User)

	// 2. Find the project AND ensure the user is actually a member/owner
	var project models.Project
	// We verify membership by joining the user's projects
	err := config.DB.Model(&user).Association("Projects").Find(&project, id)

	if err != nil || project.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found or access denied"})
		return
	}

	// 3. Delete all tasks associated with this project (Manual Cascade)
	config.DB.Where("project_id = ?", project.ID).Delete(&models.Task{})

	// 4. Delete the relationship in the join table (Remove users from project)
	config.DB.Model(&project).Association("Users").Clear()

	// 5. Delete the project itself
	result := config.DB.Delete(&project)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project and all its tasks deleted"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	// 3. Find the Project
	var project models.Project
	if err := config.DB.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// 4. Find the User to invite
	var userToInvite models.User
	if err := config.DB.Where("email = ?", body.Email).First(&userToInvite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User with that email not found"})
		return
	}

	// 5. Add relationship
	config.DB.Model(&project).Association("Users").Append(&userToInvite)

	c.JSON(http.StatusOK, gin.H{"message": "User added to project!"})
}
