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

	r.Use(middlewares.EnsureJSON())

	// PUBLIC ROUTES
	r.POST("/signup", auth.Signup)
	r.POST("/login", auth.Login)

	// PROTECTED ROUTES
	protected := r.Group("/")
	protected.Use(middlewares.RequireAuth)
	{
		protected.GET("/projects", projects.FindProjects)
		protected.POST("/projects", projects.CreateProject)
		protected.POST("/projects/:id/invite", projects.InviteUser)
		protected.DELETE("/projects/:id", projects.DeleteProject)
		protected.GET("/projects/:id/users", projects.FindProjectMembers)

		protected.GET("/projects/:id/tasks", tasks.FindTasksByProject)
		protected.POST("/tasks", tasks.CreateTask)
		protected.PATCH("/tasks/:id", tasks.UpdateTask)
		protected.DELETE("/tasks/:id", tasks.DeleteTask)
		protected.POST("/tasks/:id/take", tasks.TakeTask)
		protected.POST("/tasks/:id/assign_users", tasks.AssignUsers)
	}

	r.Run(":8080")
}
