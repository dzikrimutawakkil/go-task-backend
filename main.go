package main

import (
	"context"
	"gotask-backend/config"
	"gotask-backend/docs"
	"gotask-backend/middlewares"
	"gotask-backend/modules/auth"
	"gotask-backend/modules/clients"
	"gotask-backend/modules/health"
	"gotask-backend/modules/invoices"
	"gotask-backend/modules/projects"
	"gotask-backend/modules/tasks"
	"gotask-backend/modules/workspaces"
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

// @title          GoTask API
// @version        1.0.0
// @description    Task Management RESTful API Backend with authentication, workspaces, projects, tasks, statuses, and labels management.
// @BasePath       /
// @schemes        https
// @securityDefinitions.apikey BearerAuth
// @in             header
// @name           Authorization
// @description    JWT Bearer token. Format: 'Bearer {token}'

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
	healthHandler := health.NewHandler()
	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)

	// Swagger documentation endpoint
	docs.SwaggerInfo.Title = "GoTask API"
	docs.SwaggerInfo.Description = "Task Management RESTful API Backend with authentication, workspaces, projects, tasks, statuses, and labels management."
	docs.SwaggerInfo.Version = "1.0.0"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"https"}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Dependency Injection for Auth
	authRepo := auth.NewAuthRepository(config.DB)
	wsRepo := workspaces.NewWorkspaceRepository(config.DB)
	authService := auth.NewAuthService(authRepo, wsRepo)
	authHandler := auth.NewAuthHandler(authService)

	// Dependency Injection for Tier Plans (needed by workspaceService)
	tierPlanRepo := workspaces.NewTierPlanRepository(config.DB)

	// Dependency Injection for Workspaces
	invitationRepo := workspaces.NewInvitationRepository(config.DB)
	wsService := workspaces.NewWorkspaceService(wsRepo, authService, tierPlanRepo)
	invitationService := workspaces.NewInvitationService(invitationRepo, wsRepo, authService)
	wsHandler := workspaces.NewWorkspaceHandler(wsService)
	invitationHandler := workspaces.NewInvitationHandler(invitationService, wsRepo)

	// Dependency Injection for Tasks (M5: uses workspace-based quota)
	taskRepo := tasks.NewTaskRepository(config.DB)
	labelRepo := tasks.NewLabelRepository(config.DB)
	taskService := tasks.NewTaskService(taskRepo, authService, labelRepo, wsRepo)
	taskHandler := tasks.NewTaskHandler(taskService)
	labelHandler := tasks.NewLabelHandler(tasks.NewLabelService(labelRepo))

	// Dependency Injection for Projects (M5: uses workspace-based quota)
	projectRepo := projects.NewProjectRepository(config.DB)
	projectService := projects.NewProjectService(projectRepo, taskService, wsRepo, authService)
	projectHandler := projects.NewProjectHandler(projectService)

	// Dependency Injection for Clients (M5: uses workspace-based quota)
	clientRepo := clients.NewClientRepository(config.DB)
	clientService := clients.NewClientService(clientRepo, wsRepo, authService)
	clientHandler := clients.NewClientHandler(clientService)

	// Dependency Injection for Invoices (M5: uses workspace-based quota)
	invoiceRepo := invoices.NewInvoiceRepository(config.DB)
	invoiceService := invoices.NewInvoiceService(invoiceRepo, clientRepo, wsRepo, authService)
	invoiceHandler := invoices.NewInvoiceHandler(invoiceService)

	// PUBLIC ROUTES
	r.POST("/signup", authHandler.Signup)
	r.POST("/login", authHandler.Login)
	r.POST("/forgot-password", authHandler.ForgotPassword)
	r.POST("/reset-password", authHandler.ResetPassword)
	r.GET("/tier/plans", wsHandler.GetTierPlans) // Public: pricing page

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
		protected.GET("/api/auth/me", authHandler.Me)
		protected.PATCH("/api/users/me", authHandler.UpdateProfile)
		protected.PATCH("/api/users/me/password", authHandler.ChangePassword)
		protected.POST("/api/users/me/switch-workspace", authHandler.SwitchWorkspace) // M11: Workspace Switch
		protected.GET("/projects", projectHandler.FindProjects)
		protected.GET("/projects/:id", projectHandler.GetProject)
		protected.POST("/projects", projectHandler.CreateProject)
		protected.PATCH("/projects/:id", projectHandler.UpdateProject)
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

		// M-MIGRATION: Updated to /workspaces endpoints
		protected.GET("/workspaces", wsHandler.GetUserWorkspaces)
		protected.POST("/workspaces", wsHandler.CreateWorkspace)
		protected.POST("/workspaces/invite", wsHandler.InviteMember)
		protected.GET("/workspaces/members", wsHandler.GetMembers)
		protected.DELETE("/workspaces/members/:user_id", wsHandler.RemoveMember)
		protected.PATCH("/workspaces/members/:user_id", wsHandler.UpdateMemberRole)
		protected.GET("/workspaces/invitations", invitationHandler.GetInvitations)

		protected.GET("/clients", clientHandler.ListClients)
		protected.GET("/clients/stats", clientHandler.GetClientStats)
		protected.POST("/clients", clientHandler.CreateClient)
		protected.GET("/clients/:id", clientHandler.GetClient)
		protected.PATCH("/clients/:id", clientHandler.UpdateClient)
		protected.DELETE("/clients/:id", clientHandler.DeleteClient)

		protected.GET("/invoices", invoiceHandler.ListInvoices)
		protected.POST("/invoices", invoiceHandler.CreateInvoice)
		protected.GET("/invoices/:id", invoiceHandler.GetInvoice)
		protected.PATCH("/invoices/:id", invoiceHandler.UpdateInvoice)
		protected.DELETE("/invoices/:id", invoiceHandler.DeleteInvoice)
		protected.PATCH("/invoices/:id/mark-paid", invoiceHandler.MarkPaid)

		// M5: Subscription Tiers — M-MIGRATION: Updated to workspace endpoints
		protected.GET("/users/me/tier", wsHandler.GetMyTierInfo)
		protected.PATCH("/admin/workspaces/:id/tier", wsHandler.ActivateTier)
	}

	// Public invitation endpoints (accept doesn't require workspace context)
	r.POST("/invite/accept", invitationHandler.AcceptInvitation)
	r.POST("/invite/resend", invitationHandler.ResendInvitation)
	r.DELETE("/invite/:token", invitationHandler.RevokeInvitation)

	// Q11: Graceful Shutdown
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
