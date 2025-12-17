package auth

import (
	"gotask-backend/models"
	"gotask-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// POST /signup
func (h *AuthHandler) Signup(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		OrgName  string `json:"org_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	input := SignupInput{
		Email:    req.Email,
		Password: req.Password,
		OrgName:  req.OrgName,
	}

	user, org, err := h.service.Signup(input)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Signup successful", gin.H{
		"user":         user,
		"organization": org,
	})
}

// POST /login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.service.Login(LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, err.Error())
		return
	}

	utils.SendSuccess(c, "Login successful", gin.H{"token": token})
}

// POST /organizations/invite
func (h *AuthHandler) AddUserToOrg(c *gin.Context) {
	// 1. Get User
	user := c.MustGet("user").(models.User)

	// 2. Get Org ID from Header (validated by middleware)
	orgIDStr := c.MustGet("org_id").(string)

	var req struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Call Service
	// Note: We pass orgIDStr directly. In the Service, update the logic to parse it or use string.
	// For this code to work with the Service provided above, ensure Service accepts string or converts.
	// I'll define the final Service Method below to match.

	err := h.service.AddUserToOrg(user.ID, orgIDStr, req.Email)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "User added to organization successfully")
}
