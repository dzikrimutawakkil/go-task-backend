package middlewares

import (
	"fmt"
	"gotask-backend/config"
	"gotask-backend/models"
	"gotask-backend/modules/auth"
	"gotask-backend/utils"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// personalWorkspaceCache stores userID -> workspaceID mappings to avoid DB hit on every request.
var (
	personalWorkspaceCache = make(map[uint]uint)
	personalWorkspaceMu    sync.RWMutex
)

func RequireAuth(c *gin.Context) {
	// 1. Get the token from the header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.SendError(c, http.StatusUnauthorized, "Authorization header missing")
		c.Abort()
		return
	}

	// Header format is usually "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 {
		utils.SendError(c, http.StatusUnauthorized, "Invalid token format")
		c.Abort()
		return
	}
	tokenString := tokenParts[1]

	// 2. Parse and Validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 3. Check expiration
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			utils.SendError(c, http.StatusUnauthorized, "Token expired")
			c.Abort()
			return
		}

		// 4. Find the user
		var user auth.User
		config.DB.First(&user, claims["sub"])

		if user.ID == 0 {
			utils.SendError(c, http.StatusUnauthorized, "User not found")
			c.Abort()
			return
		}

		// 5. Attach User to request (without Tier fields - tier is now per-workspace)
		minimalUser := models.MinimalUser{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Phone:     user.Phone,
			Address:   user.Address,
			CreatedAt: user.CreatedAt,
		}

		c.Set("user", minimalUser)

		// ---------------------------------------------------------
		// M-MIGRATION: Handle Workspace Context Header (X-Workspace-ID)
		// ---------------------------------------------------------
		workspaceIDHeader := c.GetHeader("X-Workspace-ID")

		if workspaceIDHeader != "" {
			// If the header is present, we MUST validate membership immediately.
			var count int64
			config.DB.Table("workspace_members").
				Where("user_id = ? AND workspace_id = ?", user.ID, workspaceIDHeader).
				Count(&count)

			if count == 0 {
				// Stop the request here! Security Block.
				utils.SendError(c, http.StatusForbidden, "Access denied: You are not a member of the workspace specified in X-Workspace-ID")
				c.Abort()
				return
			}

			// If valid, save it to Context so controllers can use it
			c.Set("workspace_id", workspaceIDHeader)
		} else {
			// No X-Workspace-ID header → auto-resolve to user's personal workspace
			workspaceID, err := resolvePersonalWorkspaceID(user.ID)
			if err != nil {
				utils.SendError(c, http.StatusInternalServerError, "Personal workspace not found")
				c.Abort()
				return
			}
			c.Set("workspace_id", strconv.FormatUint(uint64(workspaceID), 10))
		}

		c.Next()
	} else {
		utils.SendError(c, http.StatusUnauthorized, err.Error())
		c.Abort()
	}
}

// resolvePersonalWorkspaceID looks up the personal workspace ID for a user, with in-memory cache.
func resolvePersonalWorkspaceID(userID uint) (uint, error) {
	// Fast path: check cache first
	personalWorkspaceMu.RLock()
	workspaceID, cached := personalWorkspaceCache[userID]
	personalWorkspaceMu.RUnlock()
	if cached {
		return workspaceID, nil
	}

	// Cache miss: query DB
	var ws struct {
		ID uint
	}
	err := config.DB.Table("workspaces").
		Select("id").
		Where("owner_id = ? AND workspace_type = 'personal'", userID).
		First(&ws).Error
	if err != nil {
		return 0, err
	}

	// Store in cache
	personalWorkspaceMu.Lock()
	personalWorkspaceCache[userID] = ws.ID
	personalWorkspaceMu.Unlock()

	return ws.ID, nil
}
