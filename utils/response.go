package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Define the standard structure for ALL responses
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // "omitempty" hides this field if it's null
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
