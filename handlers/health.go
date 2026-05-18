package handlers

import (
	"gotask-backend/config"
	"gotask-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints.
// GET /health — basic liveness check (always 200 if process is alive)
// GET /ready  — readiness probe (200 if DB connected, 503 otherwise)
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health godoc
// @Summary     Liveness check
// @Description Returns 200 OK when the server process is alive. Used for basic liveness and Docker HEALTHCHECK.
// @Tags        Health
// @Produce     json
// @Success     200 {object} map[string]string "server is alive"
// @Router      /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// Ready godoc
// @Summary     Readiness check
// @Description Checks database connectivity. Returns 200 when DB is reachable, 503 when not. Used for load balancer readiness probes.
// @Tags        Health
// @Produce     json
// @Success     200 {object} map[string]string "server is ready"
// @Failure     503 {object} map[string]string "server is not ready"
// @Router      /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	// Ping the database with a short timeout
	sqlDB, err := config.DB.DB()
	if err != nil {
		utils.GetLogger().Warn("Health check: DB handle error", "error", err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"db":     "error",
			"error":  "Database handle unavailable",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		utils.GetLogger().Warn("Health check: DB ping failed", "error", err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"db":     "disconnected",
			"error":  "Cannot ping database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"db":     "connected",
	})
}
