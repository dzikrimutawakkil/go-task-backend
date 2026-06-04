package health

import (
	"gotask-backend/config"
	"gotask-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler handles health check endpoints.
type Handler struct{}

// NewHandler creates a new HealthHandler.
func NewHandler() *Handler {
	return &Handler{}
}

// Health godoc
// @Summary     Liveness check
// @Description Returns 200 OK when the server process is alive. Used for basic liveness and Docker HEALTHCHECK.
// @Tags        Health
// @Produce     json
// @Success     200 {object} map[string]string "server is alive"
// @Router      /health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready godoc
// @Summary     Readiness check
// @Description Checks database connectivity. Returns 200 when DB is reachable, 503 when not.
// @Tags        Health
// @Produce     json
// @Success     200 {object} map[string]string "server is ready"
// @Failure     503 {object} map[string]string "server is not ready"
// @Router      /ready [get]
func (h *Handler) Ready(c *gin.Context) {
	sqlDB, err := config.DB.DB()
	if err != nil {
		utils.GetLogger().Warn("Health check: DB handle error", "error", err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready", "db": "error", "error": "Database handle unavailable",
		})
		return
	}
	if err := sqlDB.Ping(); err != nil {
		utils.GetLogger().Warn("Health check: DB ping failed", "error", err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready", "db": "disconnected", "error": "Cannot ping database",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready", "db": "connected"})
}
