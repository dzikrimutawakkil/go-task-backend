package middlewares

import (
	"gotask-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header name for the request ID.
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the gin context key for the request ID.
	RequestIDKey = "request_id"
)

// RequestIDMiddleware generates a unique UUID request ID for every request.
// If the client already sends X-Request-ID, use that; otherwise generate one.
// The request ID is stored in gin.Context and also returned in the response header.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if client already provided a request ID
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		// Store in gin context for downstream handlers
		c.Set(RequestIDKey, requestID)

		// Also set in go-context so utils/logger can access it
		ctx := utils.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// Set response header so client can correlate logs
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID extracts the request ID from gin.Context.
func GetRequestID(c *gin.Context) string {
	if v, exists := c.Get(RequestIDKey); exists {
		return v.(string)
	}
	return ""
}
