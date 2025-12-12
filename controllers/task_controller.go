package controllers

import (
	"gotask-backend/config"
	"gotask-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// POST /tasks
// POST /tasks
func CreateTask(c *gin.Context) {
	// ... (Input binding and Project validation stay the same) ...
	var input struct {
		Title     string `json:"title" binding:"required"`
		ProjectID uint   `json:"project_id" binding:"required"`
		StatusID  uint   `json:"status_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var project models.Project
	if err := config.DB.First(&project, input.ProjectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	if input.StatusID == 0 {
		var todoStatus models.Status
		if err := config.DB.Where("slug = ?", "todo").First(&todoStatus).Error; err == nil {
			input.StatusID = todoStatus.ID
		}
	}

	// 3. Create the Task
	task := models.Task{
		Title:     input.Title,
		ProjectID: input.ProjectID,
		StatusID:  input.StatusID,
	}
	config.DB.Create(&task)

	config.DB.Preload("Status").First(&task, task.ID)

	c.JSON(http.StatusOK, gin.H{"data": task})
}

// PATCH /tasks/:id (Don't forget to update this too!)
func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var task models.Task

	if err := config.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// We accept StatusID now
	var input struct {
		Title    string `json:"title"`
		StatusID uint   `json:"status_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&task).Updates(input)

	// Preload Status so the response looks nice
	config.DB.Preload("Status").First(&task, task.ID)

	c.JSON(http.StatusOK, gin.H{"data": task})
}

// DELETE /tasks/:id
func DeleteTask(c *gin.Context) {
	id := c.Param("id")

	// 1. Check if task exists
	var task models.Task
	if err := config.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// 2. Delete the task
	config.DB.Delete(&task)

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// GET /projects/:id/tasks - Get all tasks for a specific project
func FindTasksByProject(c *gin.Context) {
	projectID := c.Param("id")

	// 1. Get logged-in user
	userContext, _ := c.Get("user")
	user := userContext.(models.User)

	// 2. Security Check: Does this user have access to this project?
	// We query the "project_users" join table via the User model
	var project models.Project
	err := config.DB.Model(&user).Association("Projects").Find(&project, projectID)

	// If project.ID is 0, it means the relation doesn't exist (User isn't in the project)
	if err != nil || project.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found or access denied"})
		return
	}

	// 3. Fetch Tasks with Status
	var tasks []models.Task
	result := config.DB.Preload("Status").Preload("Assignees").Where("project_id = ?", projectID).Find(&tasks)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tasks})
}

// POST /tasks/:id/take - Logged-in user assigns themselves
func TakeTask(c *gin.Context) {
	taskID := c.Param("id")

	// 1. Get logged-in user
	userContext, _ := c.Get("user")
	user := userContext.(models.User)

	// 2. Find Task
	var task models.Task
	if err := config.DB.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// 3. Check if already assigned (Prevent duplicates)
	// We count how many times this user is assigned to this task (should be 0 or 1)
	var count int64
	config.DB.Table("task_users").Where("task_id = ? AND user_id = ?", task.ID, user.ID).Count(&count)

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are already assigned to this task"})
		return
	}

	// 4. Assign the user
	config.DB.Model(&task).Association("Assignees").Append(&user)

	c.JSON(http.StatusOK, gin.H{"message": "Task assigned to you successfully"})
}

// POST /tasks/:id/assign_users - Assign multiple users by email
func AssignUsers(c *gin.Context) {
	taskID := c.Param("id")

	// 1. Define input structure for an Array of emails
	var input struct {
		Emails []string `json:"emails" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Find the Task
	var task models.Task
	if err := config.DB.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// 3. Find All Users matching the emails
	var users []models.User
	// This SQL equivalent is: SELECT * FROM users WHERE email IN ('a@test.com', 'b@test.com')
	if err := config.DB.Where("email IN ?", input.Emails).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	if len(users) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching users found"})
		return
	}

	// 4. Append these users to the Task's Assignees
	// GORM handles duplicates automatically (it won't add the same user twice to the join table)
	err := config.DB.Model(&task).Association("Assignees").Append(&users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Users assigned successfully",
		"count":   len(users),
	})
}
