package controllers

import (
	"gotask-backend/config"
	"gotask-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// POST /tasks - Create a task
func CreateTask(c *gin.Context) {
	var input models.Task
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that the Project exists before creating the task
	var project models.Project
	if err := config.DB.First(&project, input.ProjectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	config.DB.Create(&input)
	c.JSON(http.StatusOK, gin.H{"data": input})
}

// PATCH /tasks/:id - Update task status
func UpdateTask(c *gin.Context) {
	var task models.Task
	id := c.Param("id")

	if err := config.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var input models.Task
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&task).Updates(input)
	c.JSON(http.StatusOK, gin.H{"data": task})
}
