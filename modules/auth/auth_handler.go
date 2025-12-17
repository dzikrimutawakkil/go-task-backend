package auth

import (
	"gotask-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService AuthService
}

func NewAuthHandler(authS AuthService) *Handler {
	return &Handler{authService: authS}
}

// POST /signup
func (h *Handler) Signup(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Create User
	user, err := h.authService.Signup(SignupInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Signup successful", gin.H{
		"user": user,
	})
}

// POST /login
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.authService.Login(LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, err.Error())
		return
	}

	utils.SendSuccess(c, "Login successful", gin.H{"token": token})
}
