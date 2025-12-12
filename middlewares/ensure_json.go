package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func EnsureJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check if the method is one that carries data
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {

			// 2. Check the Content-Type header
			contentType := c.GetHeader("Content-Type")

			// 3. If it's not JSON, stop the request immediately
			if !strings.Contains(contentType, "application/json") {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error": "Content-Type must be application/json",
				})
				c.Abort() // Important: This prevents the pending handlers (controllers) from running
				return
			}
		}

		// 4. If everything is fine, proceed to the controller
		c.Next()
	}
}
