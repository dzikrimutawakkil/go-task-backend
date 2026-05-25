package middlewares

import (
	"gotask-backend/models"
	"gotask-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
	"log/slog"
)

// StructuredLoggerMiddleware logs every HTTP request in structured JSON format.
// It captures: timestamp, level, request_id, method, path, user_id, org_id,
// duration_ms, status_code, and error (if any).
// Sensitive data (passwords, tokens) are NEVER logged.
func StructuredLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Capture response details after handlers run
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		requestID := GetRequestID(c)
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Extract user_id and org_id from gin context (set by RequireAuth)
		var userID uint
		var orgID string
		if user, exists := c.Get("user"); exists {
			u := user.(models.MinimalUser)
			userID = u.ID
		}
		if oid, exists := c.Get("org_id"); exists {
			orgID = oid.(string)
		}

		// Determine log level based on status code
		level := slog.LevelInfo
		if statusCode >= 500 {
			level = slog.LevelError
		} else if statusCode >= 400 {
			level = slog.LevelWarn
		}

		// Build structured log args as []any (slog convention)
		args := []any{
			"request_id", requestID,
			"method", method,
			"path", path,
			"query", query,
			"user_id", userID,
			"org_id", orgID,
			"status_code", statusCode,
			"duration_ms", duration.Milliseconds(),
		}

		// Add client IP
		clientIP := c.ClientIP()
		if clientIP != "" {
			args = append(args, "client_ip", clientIP)
		}

		// Add User-Agent (trimmed, non-sensitive)
		userAgent := c.Request.UserAgent()
		if userAgent != "" && len(userAgent) <= 200 {
			args = append(args, "user_agent", userAgent)
		}

		// Add error message if status >= 400
		if statusCode >= 400 {
			errMsg := c.Errors.ByType(gin.ErrorTypePrivate).String()
			if errMsg != "" && len(errMsg) <= 500 {
				args = append(args, "error", errMsg)
			}
		}

		// Log with the appropriate level
		l := utils.GetLogger()
		l.Log(c.Request.Context(), level, "HTTP Request", args...)
	}
}
