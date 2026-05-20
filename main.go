package main

import (
	"context"
	"gotask-backend/config"
	"gotask-backend/docs"
	"gotask-backend/handlers"
	"gotask-backend/middlewares"
	"gotask-backend/modules/auth"
	"gotask-backend/modules/organizations"
	"gotask-backend/modules/projects"
	"gotask-backend/modules/tasks"
	"gotask-backend/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           GoTask API
// @version         1.0.0
// @description     Task Management RESTful API Backend with authentication, organizations, projects, tasks, statuses, and labels management.
// @host            localhost:8080
// @BasePath        /
// @schemes         http https
// @securityDefinition.BearerAuth Security API Key
// @in              header
// @name            Authorization
// @description     JWT Bearer token. Format: 'Bearer {token}'

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env variables")
	}

	// Q1: Initialize structured JSON logger
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	utils.InitLogger(logLevel)

	config.ConnectDatabase()
	r := gin.New()

	// Q1: Apply structured logging middlewares FIRST
	// Order: RequestID → StructuredLogger → CORSMiddleware → EnsureJSON → RequireAuth
	r.Use(middlewares.RequestIDMiddleware())
	r.Use(middlewares.StructuredLoggerMiddleware())
	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.EnsureJSON())

	// Q2: Health check endpoints (NO auth required — must be before RequireAuth)
	healthHandler := handlers.NewHealthHandler()
	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)

	// Swagger documentation endpoint
	docs.SwaggerInfo.Title = "GoTask API"
	docs.SwaggerInfo.Description = "Task Management RESTful API Backend with authentication, organizations, projects, tasks, statuses, and labels management."
	docs.SwaggerInfo.Version = "1.0.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Dependency Injection for Auth
	authRepo := auth.NewAuthRepository(config.DB)
	orgRepo := organizations.NewOrganizationRepository(config.DB)
	authService := auth.NewAuthService(authRepo, orgRepo)
	authHandler := auth.NewAuthHandler(authService)

	// Dependency Injection for Organization
	invitationRepo := organizations.NewInvitationRepository(config.DB)
	orgService := organizations.NewOrganizationService(orgRepo, authService)
	invitationService := organizations.NewInvitationService(invitationRepo, orgRepo, authService)
	orgHandler := organizations.NewOrganizationHandler(orgService)
	invitationHandler := organizations.NewInvitationHandler(invitationService, orgRepo)

	// Dependency Injection for Tasks
	taskRepo := tasks.NewTaskRepository(config.DB)
	labelRepo := tasks.NewLabelRepository(config.DB)
	taskService := tasks.NewTaskService(taskRepo, authService, labelRepo)
	taskHandler := tasks.NewTaskHandler(taskService)
	labelHandler := tasks.NewLabelHandler(tasks.NewLabelService(labelRepo))

	// Dependency Injection for Projects
	projectRepo := projects.NewProjectRepository(config.DB)
	projectService := projects.NewProjectService(projectRepo, taskService, orgRepo)
	projectHandler := projects.NewProjectHandler(projectService)

	// PUBLIC ROUTES
	r.POST("/signup", authHandler.Signup)
	r.POST("/login", authHandler.Login)

	// PROTECTED ROUTES
	// Q12: Rate limiting — per IP for unauthenticated, per user for authenticated
	r.Use(middlewares.RateLimiterMiddleware(middlewares.RateLimiterConfig{
		RequestsPerWindow: 100,
		WindowDuration:    1 * time.Minute,
		KeyFunc:           middlewares.IPKeyFunc,
	}))

	protected := r.Group("/")
	protected.Use(middlewares.RequireAuth)
	protected.Use(middlewares.RateLimiterMiddleware(middlewares.RateLimiterConfig{
		RequestsPerWindow: 500,
		WindowDuration:    1 * time.Minute,
		KeyFunc:           middlewares.UserKeyFunc,
	}))
	{
		protected.GET("/projects", projectHandler.FindProjects)
		protected.POST("/projects", projectHandler.CreateProject)
		protected.DELETE("/projects/:id", projectHandler.DeleteProject)

		protected.GET("/projects/:id/tasks", taskHandler.FindTasksByProject)
		protected.GET("/tasks/search", taskHandler.SearchTasks)
		protected.POST("/tasks", taskHandler.CreateTask)
		protected.PATCH("/tasks/:id", taskHandler.UpdateTask)
		protected.DELETE("/tasks/:id", taskHandler.DeleteTask)

		protected.GET("/projects/:id/status", taskHandler.FindStatusesByProject)
		protected.POST("/projects/:id/status", taskHandler.CreateStatus)
		protected.PATCH("/status/:id", taskHandler.UpdateStatus)
		protected.DELETE("/status/:id", taskHandler.DeleteStatus)

		protected.GET("/projects/:id/labels", labelHandler.GetLabels)
		protected.POST("/projects/:id/labels", labelHandler.CreateLabel)
		protected.PATCH("/labels/:id", labelHandler.UpdateLabel)
		protected.DELETE("/labels/:id", labelHandler.DeleteLabel)

		protected.POST("/organizations", orgHandler.CreateOrganization)
		protected.POST("/organizations/invite", orgHandler.InviteMember)
		protected.GET("/organizations/members", orgHandler.GetMembers)
		protected.DELETE("/organizations/members/:user_id", orgHandler.RemoveMember)
		protected.PATCH("/organizations/members/:user_id", orgHandler.UpdateMemberRole)
		protected.GET("/organizations/invitations", invitationHandler.GetInvitations)
	}

	// Public invitation endpoints (accept doesn't require org context)
	r.POST("/invite/accept", invitationHandler.AcceptInvitation)
	r.POST("/invite/resend", invitationHandler.ResendInvitation)
	r.DELETE("/invite/:token", invitationHandler.RevokeInvitation)

	// Q11: Graceful Shutdown (placeholder — full implementation in Q11)
	srv := &http.Server{Addr: ":8080", Handler: r}

	go func() {
		utils.GetLogger().Info("Starting server", "addr", ":8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.GetLogger().Error("Server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.GetLogger().Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		utils.GetLogger().Error("Server forced to shutdown", "error", err)
	}
	utils.GetLogger().Info("Server exited gracefully")
}
