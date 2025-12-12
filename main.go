package main

import (
	"gotask-backend/config"
	"gotask-backend/controllers"
	"gotask-backend/middlewares"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set a default secret if not in .env (FOR DEV ONLY)
	if os.Getenv("SECRET_KEY") == "" {
		os.Setenv("SECRET_KEY", "supersecretkey")
	}

	config.ConnectDatabase()
	r := gin.Default()

	r.Use(middlewares.EnsureJSON())

	// PUBLIC ROUTES
	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)

	// PROTECTED ROUTES
	// We group these so we can apply middleware to all of them at once
	protected := r.Group("/")
	protected.Use(middlewares.RequireAuth)
	{
		protected.GET("/projects", controllers.FindProjects)
		protected.POST("/projects", controllers.CreateProject)
		protected.POST("/tasks", controllers.CreateTask)
		protected.PATCH("/tasks/:id", controllers.UpdateTask)
	}

	r.Run(":8080")
}
