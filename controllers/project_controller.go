package controllers

import (
	"gotask-backend/config"
	"gotask-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /projects - List all projects (and their tasks)
func FindProjects(c *gin.Context) {
	var projects []models.Project
	// Preload("Tasks") fetches the related tasks automatically
	config.DB.Preload("Tasks").Find(&projects)
	c.JSON(http.StatusOK, gin.H{"data": projects})
}

// POST /projects - Create a new project
func CreateProject(c *gin.Context) {
	var input models.Project
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config.DB.Create(&input)
	c.JSON(http.StatusOK, gin.H{"data": input})
}
