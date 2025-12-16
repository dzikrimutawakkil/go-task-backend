package auth

import (
	"gotask-backend/config"
	"gotask-backend/models"
	"gotask-backend/utils"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// POST /signup
func Signup(c *gin.Context) {
	var body struct {
		Email    string  `json:"email" binding:"required"`
		Password string  `json:"password" binding:"required"`
		OrgName  *string `json:"org_name"`
	}

	if c.ShouldBindJSON(&body) != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to read body")
		return
	}

	// Hash Password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to hash password")
		return
	}

	// Start Transaction
	tx := config.DB.Begin()

	// Create User
	user := models.User{Email: body.Email, Password: string(hash)}
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		utils.SendError(c, http.StatusBadRequest, "Failed to create user (Email might exist)")
		return
	}

	// Create Organization (OPTIONAL)
	var org *models.Organization

	// Logic: Only create if OrgName is NOT nil AND NOT empty
	if body.OrgName != nil {
		trimmedName := strings.TrimSpace(*body.OrgName)

		if trimmedName != "" {
			newOrg := models.Organization{
				Name:    trimmedName,
				OwnerID: user.ID,
				Users:   []models.User{user},
			}

			if err := tx.Create(&newOrg).Error; err != nil {
				tx.Rollback()
				utils.SendError(c, http.StatusBadRequest, "Failed to create organization")
				return
			}
			org = &newOrg
		}
	}

	tx.Commit()

	utils.SendSuccess(c, "Signup successful", gin.H{
		"user":         user,
		"organization": org,
	})
}

// Helper: Login to get the JWT
func Login(c *gin.Context) {
	var body struct {
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to read body")
		return
	}

	// 1. Look up requested user
	var user models.User
	config.DB.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		utils.SendError(c, http.StatusBadRequest, "Invalid email or password")
		return
	}

	// 2. Compare sent password with saved user password hash
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid email or password")
		return
	}

	// 3. Generate JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,                                    // Subject (User ID)
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(), // Expiration (30 days)
	})

	// Sign and get the complete encoded token as a string using the secret
	// Note: In production, store "SECRET_KEY" in your .env file!
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))

	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to create token")
		return
	}

	// 4. Send it back
	utils.SendSuccess(c, "Login successful", gin.H{
		"token": tokenString,
	})
}
