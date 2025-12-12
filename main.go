package main

import (
	"gotask-backend/config"
	"gotask-backend/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Connect to DB
	config.ConnectDatabase()

	// 2. Initialize Router
	r := gin.Default()

	// 3. Define Routes
	// Project Routes
	r.GET("/projects", controllers.FindProjects)
	r.POST("/projects", controllers.CreateProject)

	// Task Routes
	r.POST("/tasks", controllers.CreateTask)
	r.PATCH("/tasks/:id", controllers.UpdateTask)

	// 4. Run Server (Default port 8080)
	r.Run(":8080")
}
