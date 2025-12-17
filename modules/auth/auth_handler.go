package auth

import (
	"gotask-backend/models"
	"gotask-backend/modules/organizations"
	"gotask-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService AuthService
	orgService  organizations.OrganizationService
}

// NewHandler now accepts BOTH services to coordinate them
func NewAuthHandler(authS AuthService, orgS organizations.OrganizationService) *Handler {
	return &Handler{authService: authS, orgService: orgS}
}

// POST /signup
func (h *Handler) Signup(c *gin.Context) {
	var req struct {
		Email    string  `json:"email" binding:"required"`
		Password string  `json:"password" binding:"required"`
		OrgName  *string `json:"org_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 1. Create User (Delegated to Auth Service)
	user, err := h.authService.Signup(SignupInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Create Organization (Delegated to Organization Service)
	// This keeps Auth module clean from Org logic
	var org *models.Organization
	if req.OrgName != nil {
		newOrg, err := h.orgService.CreateOrganization(*req.OrgName, user.ID)
		if err != nil {
			// Note: In a real production app, you might want to rollback the User creation here.
			utils.SendError(c, http.StatusInternalServerError, "User created but failed to create Organization")
			return
		}
		org = newOrg
	}

	utils.SendSuccess(c, "Signup successful", gin.H{
		"user":         user,
		"organization": org,
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
