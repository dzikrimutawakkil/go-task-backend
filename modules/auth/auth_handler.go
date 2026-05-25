package auth

import (
	"gotask-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gotask-backend/models"
)

// SignupRequest represents the request body for user registration.
// @Description Request body for user signup
type SignupRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"securepassword123"`
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Phone    string `json:"phone" binding:"required" example:"+628123456789"`
	Address  string `json:"address" example:"Jl. Sudirman No.123, Jakarta"`
}

// LoginRequest represents the request body for user login.
// @Description Request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"securepassword123"`
}

// ForgotPasswordRequest represents the request body for password reset.
// @Description Request body for password reset
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required" example:"user@example.com"`
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
// @Description Register a new user with email and password. Returns the created user object and JWT token.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body SignupRequest true "Signup payload"
// @Success     200 {object} utils.APIResponse{data=map[string]interface{}} "Signup successful"
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
		Name:     req.Name,
		Phone:    req.Phone,
		Address:  req.Address,
	})
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Generate JWT token for immediate login after registration
	token, err := h.authService.Login(LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Signup succeeded but token generation failed")
		return
	}

	utils.SendSuccess(c, "Signup successful", gin.H{
		"user":  user,
		"token": token,
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

	// Fetch user to return alongside token (frontend expects { user, token })
	user, err := h.authService.GetUserByEmail(req.Email)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Login succeeded but user fetch failed")
		return
	}

	utils.SendSuccess(c, "Login successful", gin.H{
		"user":  user,
		"token": token,
	})
}

// Me godoc
// @Summary     Get current user
// @Description Returns the authenticated user's profile based on JWT token.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "success"
// @Failure     401 {object} utils.APIResponse "Unauthorized"
// @Router      /api/auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	user := c.MustGet("user").(models.MinimalUser)
	utils.SendSuccess(c, "success", gin.H{
		"user": user,
	})
}

// ForgotPassword godoc
// @Summary     Request password reset
// @Description Sends a password reset email to the user if the email exists.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body ForgotPasswordRequest true "Email payload"
// @Success     200 {object} utils.APIResponse "If the email exists, a reset link has been sent"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Router      /forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Always return 200 to prevent email enumeration attacks
	if err := h.authService.ForgotPassword(req.Email); err != nil {
		utils.SendSuccess(c, "If the email exists, a reset link has been sent")
		return
	}

	utils.SendSuccess(c, "If the email exists, a reset link has been sent")
}
