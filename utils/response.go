package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the standard API response envelope.
// @Description Standard API response envelope used by all endpoints
type APIResponse struct {
	// True if the request was successful
	Success bool `json:"success"`
	// Human-readable message describing the result
	Message string `json:"message"`
	// Response payload, omitted when nil
	Data interface{} `json:"data,omitempty"`
}

func SendSuccess(c *gin.Context, message string, data ...interface{}) {
	var responseData interface{}

	// Check if any data was passed
	if len(data) > 0 {
		responseData = data[0] // Take the first item
	} else {
		responseData = nil
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    responseData,
	})
}

func SendError(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Success: false,
		Message: message,
		Data:    nil,
	})
}
