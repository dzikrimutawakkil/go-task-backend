package auth

import (
	"gotask-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SignupRequest represents the request body for user registration.
// @Description Request body for user signup
type SignupRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"securepassword123"`
}

// LoginRequest represents the request body for user login.
// @Description Request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"securepassword123"`
}

// AuthResponse represents the data payload for auth endpoints.
// @Description Auth data payload containing user info or token
type AuthResponse struct {
	User  *User  `json:"user,omitempty"`
	Token string `json:"token,omitempty"`
}

type Handler struct {
	authService AuthService
}

func NewAuthHandler(authS AuthService) *Handler {
	return &Handler{authService: authS}
}

// Signup godoc
// @Summary     User registration
// @Description Register a new user with email and password. Returns the created user object.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body SignupRequest true "Signup payload"
// @Success     200 {object} utils.APIResponse{data=User} "Signup successful"
// @Failure     400 {object} utils.APIResponse "Validation error or email already exists"
// @Router      /signup [post]
func (h *Handler) Signup(c *gin.Context) {
	var req SignupRequest

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

// Login godoc
// @Summary     User login
// @Description Authenticate a user with email and password. Returns a JWT token on success.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body LoginRequest true "Login payload"
// @Success     200 {object} utils.APIResponse{data=map[string]string} "Login successful"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     401 {object} utils.APIResponse "Invalid credentials"
// @Router      /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

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
