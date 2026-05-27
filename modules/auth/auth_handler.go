package auth

import (
	"net/http"

	"gotask-backend/utils"

	"github.com/gin-gonic/gin"
	"gotask-backend/models"
)

// Request DTOs
type SignupRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"securepassword123"`
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Phone    string `json:"phone" binding:"required" example:"+628123456789"`
	Address  string `json:"address" example:"Jl. Sudirman No.123, Jakarta"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"securepassword123"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required" example:"user@example.com"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required" example:"abc123..."`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newpassword123"`
}

type UpdateProfileRequest struct {
	Name    *string `json:"name" example:"John Doe"`
	Phone   *string `json:"phone" example:"+628123456789"`
	Address *string `json:"address" example:"Jl. Sudirman No.123, Jakarta"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"oldpassword123"`
	NewPassword     string `json:"new_password" binding:"required,min=8" example:"newpassword123"`
}

// SwitchOrganizationRequest represents the request body for switching active organization.
// M11: Workspace Switch Endpoint
type SwitchOrganizationRequest struct {
	OrganizationID uint `json:"organization_id" binding:"required" example:"1"`
}

type AuthResponse struct {
	User  any    `json:"user,omitempty"`
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

// ResetPassword godoc
// @Summary     Reset password with token
// @Description Validates the reset token and sets a new password. Token is single-use.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body ResetPasswordRequest true "Reset password payload"
// @Success     200 {object} utils.APIResponse "Password reset successful"
// @Failure     400 {object} utils.APIResponse "Invalid, expired, or already used token"
// @Router      /reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Password reset successful")
}

// UpdateProfile godoc
// @Summary     Update current user profile
// @Description Updates the authenticated user's profile (name, phone, address). Email cannot be changed.
// @Tags        Profile
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body UpdateProfileRequest true "Profile update payload"
// @Success     200 {object} utils.APIResponse "Profile updated successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     401 {object} utils.APIResponse "Unauthorized"
// @Router      /api/users/me [patch]
func (h *Handler) UpdateProfile(c *gin.Context) {
	user := c.MustGet("user").(models.MinimalUser)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Validate name is not empty string
	if req.Name != nil && *req.Name == "" {
		utils.SendError(c, http.StatusBadRequest, "name cannot be empty")
		return
	}

	updatedUser, err := h.authService.UpdateUserProfile(user.ID, "", "", "")
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Apply updates if provided
	if req.Name != nil {
		updatedUser.Name = *req.Name
	}
	if req.Phone != nil {
		updatedUser.Phone = *req.Phone
	}
	if req.Address != nil {
		updatedUser.Address = *req.Address
	}

	updatedUser, err = h.authService.UpdateUserProfile(user.ID, updatedUser.Name, updatedUser.Phone, updatedUser.Address)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccess(c, "Profile updated successfully", gin.H{
		"user": updatedUser,
	})
}

// ChangePassword godoc
// @Summary     Change password
// @Description Changes the authenticated user's password after validating the current one.
// @Tags        Profile
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body ChangePasswordRequest true "Password change payload"
// @Success     200 {object} utils.APIResponse "Password changed successfully"
// @Failure     400 {object} utils.APIResponse "Validation error or incorrect current password"
// @Failure     401 {object} utils.APIResponse "Unauthorized"
// @Router      /api/users/me/password [patch]
func (h *Handler) ChangePassword(c *gin.Context) {
	user := c.MustGet("user").(models.MinimalUser)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.authService.ChangePassword(user.ID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Password changed successfully")
}

// SwitchOrganization godoc
// @Summary     Switch active organization
// @Description Switch the user's active organization context. User must be a member of the target organization.
// @Tags        Profile
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body SwitchOrganizationRequest true "Organization switch payload"
// @Success     200 {object} utils.APIResponse "Organization switched successfully"
// @Failure     400 {object} utils.APIResponse "Invalid organization ID"
// @Failure     403 {object} utils.APIResponse "Not a member of the organization"
// @Router      /api/users/me/switch-organization [post]
// M11: Workspace Switch Endpoint
func (h *Handler) SwitchOrganization(c *gin.Context) {
	var req SwitchOrganizationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	user := c.MustGet("user").(models.MinimalUser)

	// Validate membership
	valid, err := h.authService.CheckOrganizationMembership(user.ID, req.OrganizationID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to verify membership")
		return
	}
	if !valid {
		utils.SendError(c, http.StatusForbidden, "You are not a member of this organization")
		return
	}

	// Update org context in middleware cache (this is handled by the middleware on next request)
	// For now, just return success - the middleware will resolve the new org on subsequent requests
	// Or we can update the user's context directly

	utils.SendSuccess(c, "Organization switched successfully", gin.H{
		"organization_id": req.OrganizationID,
		"message":         "Switched organization. Use X-Organization-ID header for subsequent requests.",
	})
}
