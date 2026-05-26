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

// personalOrgCache stores userID -> orgID mappings to avoid DB hit on every request.
var (
	personalOrgCache = make(map[uint]uint)
	personalOrgMu    sync.RWMutex
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

		// 5. Attach User to request
		minimalUser := models.MinimalUser{
			ID:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			Phone:         user.Phone,
			Address:       user.Address,
			Plan:          user.Plan,
			LicenseKey:    user.LicenseKey,
			LicenseStatus: user.LicenseStatus,
			CreatedAt:     user.CreatedAt,
		}

		c.Set("user", minimalUser)

		// ---------------------------------------------------------
		// NEW: Handle Organization Context Header (X-Organization-ID)
		// ---------------------------------------------------------
		orgIDHeader := c.GetHeader("X-Organization-ID")

		if orgIDHeader != "" {
			// If the header is present, we MUST validate membership immediately.
			var count int64
			config.DB.Table("organization_users").
				Where("user_id = ? AND organization_id = ?", user.ID, orgIDHeader).
				Count(&count)

			if count == 0 {
				// Stop the request here! Security Block.
				utils.SendError(c, http.StatusForbidden, "Access denied: You are not a member of the organization specified in X-Organization-ID")
				c.Abort()
				return
			}

			// If valid, save it to Context so controllers can use it
			c.Set("org_id", orgIDHeader)
		} else {
			// No X-Organization-ID header → auto-resolve to user's personal workspace
			orgID, err := resolvePersonalOrgID(user.ID)
			if err != nil {
				utils.SendError(c, http.StatusInternalServerError, "Personal workspace not found")
				c.Abort()
				return
			}
			c.Set("org_id", strconv.FormatUint(uint64(orgID), 10))
		}

		c.Next()
	} else {
		utils.SendError(c, http.StatusUnauthorized, err.Error())
		c.Abort()
	}
}

// resolvePersonalOrgID looks up the personal org ID for a user, with in-memory cache.
func resolvePersonalOrgID(userID uint) (uint, error) {
	// Fast path: check cache first
	personalOrgMu.RLock()
	orgID, cached := personalOrgCache[userID]
	personalOrgMu.RUnlock()
	if cached {
		return orgID, nil
	}

	// Cache miss: query DB
	var org struct {
		ID uint
	}
	err := config.DB.Table("organizations").
		Select("id").
		Where("owner_id = ? AND org_type = 'personal'", userID).
		First(&org).Error
	if err != nil {
		return 0, err
	}

	// Store in cache
	personalOrgMu.Lock()
	personalOrgCache[userID] = org.ID
	personalOrgMu.Unlock()

	return org.ID, nil
}
