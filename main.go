package main

import (
	"gotask-backend/config"
	"gotask-backend/middlewares"

	"gotask-backend/modules/auth"
	"gotask-backend/modules/projects"
	"gotask-backend/modules/tasks"

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

	// Dependency Injection for Projects
	projectRepo := projects.NewProjectRepository(config.DB)
	projectService := projects.NewProjectService(projectRepo)
	projectHandler := projects.NewProjectHandler(projectService)

	// Dependency Injection for Tasks
	taskRepo := tasks.NewTaskRepository(config.DB)
	taskService := tasks.NewTaskService(taskRepo)
	taskHandler := tasks.NewTaskHandler(taskService)

	r.Use(middlewares.EnsureJSON())

	// PUBLIC ROUTES
	r.POST("/signup", auth.Signup)
	r.POST("/login", auth.Login)

	// PROTECTED ROUTES
	protected := r.Group("/")
	protected.Use(middlewares.RequireAuth)
	{
		protected.GET("/projects", projectHandler.FindProjects)
		protected.POST("/projects", projectHandler.CreateProject)
		protected.DELETE("/projects/:id", projectHandler.DeleteProject)

		protected.GET("/projects/:id/tasks", taskHandler.FindTasksByProject)
		protected.POST("/tasks", taskHandler.CreateTask)
		protected.PATCH("/tasks/:id", taskHandler.UpdateTask)
		protected.DELETE("/tasks/:id", taskHandler.DeleteTask)
		protected.POST("/tasks/:id/take", taskHandler.TakeTask)
		protected.POST("/tasks/:id/assign_users", taskHandler.AssignUsers)

		protected.POST("/organizations/invite", auth.AddUserToOrg)
	}

	r.Run(":8080")
}
