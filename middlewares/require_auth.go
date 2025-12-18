package middlewares

import (
	"fmt"
	"gotask-backend/config"
	"gotask-backend/modules/auth"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(c *gin.Context) {
	// 1. Get the token from the header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		return
	}

	// Header format is usually "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			return
		}

		// 4. Find the user
		var user auth.User
		config.DB.First(&user, claims["sub"])

		if user.ID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// 5. Attach User to request
		c.Set("user", user)

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
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "Access denied: You are not a member of the organization specified in X-Organization-ID",
				})
				return
			}

			// If valid, save it to Context so controllers can use it
			c.Set("org_id", orgIDHeader)
		}

		c.Next()
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	}
}
