package main

import (
	"gotask-backend/config"
	"gotask-backend/middlewares"
	"log"

	"gotask-backend/modules/auth"
	"gotask-backend/modules/organizations"
	"gotask-backend/modules/projects"
	"gotask-backend/modules/tasks"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env variables")
	}

	config.ConnectDatabase()
	r := gin.Default()

	// Apply Middleware (First thing!)
	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.EnsureJSON())

	// Dependency Injection for Auth
	authRepo := auth.NewAuthRepository(config.DB)
	authService := auth.NewAuthService(authRepo)
	authHandler := auth.NewAuthHandler(authService)

	// Dependency Injection for Organization
	orgRepo := organizations.NewOrganizationRepository(config.DB)
	orgService := organizations.NewOrganizationService(orgRepo, authService)
	orgHandler := organizations.NewOrganizationHandler(orgService)

	// Dependency Injection for Projects
	projectRepo := projects.NewProjectRepository(config.DB)
	projectService := projects.NewProjectService(projectRepo)
	projectHandler := projects.NewProjectHandler(projectService)

	// Dependency Injection for Tasks
	taskRepo := tasks.NewTaskRepository(config.DB)
	taskService := tasks.NewTaskService(taskRepo, authService)
	taskHandler := tasks.NewTaskHandler(taskService)

	// PUBLIC ROUTES
	r.POST("/signup", authHandler.Signup)
	r.POST("/login", authHandler.Login)

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

		protected.POST("/organizations", orgHandler.CreateOrganization)
		protected.POST("/organizations/invite", orgHandler.InviteMember)
		protected.GET("/organizations/members", orgHandler.GetMembers)
	}

	r.Run(":8080")
}
