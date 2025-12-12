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
