package middlewares

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	allowOrigins := []string{
		"http://localhost:3000",
		"http://localhost:5173",
		"http://localhost:8080",
		"https://localhost:8080",
		"https://127.0.0.1:8080",
		"http://127.0.0.1:8080",
	}

	return cors.New(cors.Config{
		// Allow specific origins (Frontend URL)
		// For development, you can allow all, but for prod, be specific.
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Organization-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
