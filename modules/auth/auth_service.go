package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"gotask-backend/internal/interfaces"
	"gotask-backend/models"
	"gotask-backend/modules/workspaces"
	"gotask-backend/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Signup(input SignupInput) (*User, error)
	Login(input LoginInput) (string, error)
	ForgotPassword(email string) error
	ResetPassword(token string, newPassword string) error
	GetUsersByIDs(ids []uint) ([]User, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id uint) (*User, error)
	UpdateUserProfile(userID uint, name string, phone string, address string) (*User, error)
	ChangePassword(userID uint, currentPassword string, newPassword string) error
	GetMinimalUserByEmail(email string) (*interfaces.MinimalUser, error)
	GetMinimalUsersByIDs(ids []uint) ([]interfaces.MinimalUser, error)
	FindByID(id uint) (*interfaces.MinimalUser, error)
	// M11: Workspace Switch Endpoint
	CheckWorkspaceMembership(userID uint, workspaceID uint) (bool, error)
}

type authService struct {
	repo   AuthRepository
	wsRepo workspaces.WorkspaceRepository
}

func NewAuthService(repo AuthRepository, wsRepo workspaces.WorkspaceRepository) AuthService {
	return &authService{repo: repo, wsRepo: wsRepo}
}

// DTOs
type SignupInput struct {
	Email    string
	Password string
	Name     string
	Phone    string
	Address  string
}

type LoginInput struct {
	Email    string
	Password string
}

// interfaces.AuthService implementation — returns MinimalUser to avoid import cycles
func (s *authService) GetMinimalUserByEmail(email string) (*interfaces.MinimalUser, error) {
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		return nil, err
	}
	return &interfaces.MinimalUser{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Phone:     user.Phone,
		Address:   user.Address,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *authService) GetMinimalUsersByIDs(ids []uint) ([]interfaces.MinimalUser, error) {
	users, err := s.repo.FindUsersByIDs(ids)
	if err != nil {
		return nil, err
	}
	result := make([]interfaces.MinimalUser, len(users))
	for i, u := range users {
		result[i] = interfaces.MinimalUser{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Phone:     u.Phone,
			Address:   u.Address,
			Password:  u.Password,
			CreatedAt: u.CreatedAt,
		}
	}
	return result, nil
}

func (s *authService) Signup(input SignupInput) (*User, error) {
	// 1. Hash Password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 2. Create User
	user := User{
		Email:    input.Email,
		Password: string(hash),
		Name:     input.Name,
		Phone:    input.Phone,
		Address:  input.Address,
	}
	if err := s.repo.CreateUser(&user); err != nil {
		return nil, errors.New("email already registered")
	}

	// 3. Auto-create personal workspace (M-MIGRATION: renamed from organization)
	personalWorkspace := workspaces.Workspace{
		Name:          fmt.Sprintf("%s's Workspace", input.Name),
		OwnerID:       user.ID,
		WorkspaceType: workspaces.WorkspaceTypePersonal,
		Tier:          "free", // M-MIGRATION: New workspace always starts with free tier
	}
	if err := s.wsRepo.Create(&personalWorkspace); err != nil {
		return nil, errors.New("failed to create personal workspace")
	}

	// 4. Add owner as member with RoleOwner
	if err := s.wsRepo.AddMember(personalWorkspace.ID, user.ID, workspaces.RoleOwner); err != nil {
		return nil, errors.New("failed to setup personal workspace")
	}

	return &user, nil
}

func (s *authService) Login(input LoginInput) (string, error) {
	// 1. Find User
	user, err := s.repo.FindUserByEmail(input.Email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// 2. Check Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	// 3. Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 Days
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return tokenString, nil
}

func (s *authService) GetUsersByIDs(ids []uint) ([]User, error) {
	return s.repo.FindUsersByIDs(ids)
}

func (s *authService) GetUserByEmail(email string) (*User, error) {
	return s.repo.FindUserByEmail(email)
}

func (s *authService) ForgotPassword(email string) error {
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		// Silently ignore — prevent email enumeration attacks
		return nil
	}

	// Generate secure random token (32 bytes = 64 hex chars)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return errors.New("failed to generate reset token")
	}
	token := hex.EncodeToString(tokenBytes)

	// Save token to database with 1 hour expiry
	prt := PasswordResetToken{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if err := s.repo.SavePasswordResetToken(&prt); err != nil {
		return errors.New("failed to save reset token")
	}

	// Send reset email (uses SMTP if configured)
	resetURL := fmt.Sprintf("%s/reset-password?token=%s",
		os.Getenv("APP_URL"), token)

	err = utils.SendEmail(user.Email, "Password Reset Request",
		fmt.Sprintf("Click the link to reset your password (expires in 1 hour):\n%s\n\nIf you didn't request this, ignore this email.", resetURL))
	if err != nil {
		utils.GetLogger().Warn("Failed to send reset email", "error", err)
		// Don't fail — token is saved, user can request again
	}
	return nil
}

func (s *authService) ResetPassword(token string, newPassword string) error {
	// Find the token
	prt, err := s.repo.FindPasswordResetToken(token)
	if err != nil {
		return errors.New("invalid token")
	}

	// Check if token is used
	if prt.UsedAt != nil {
		return errors.New("token already used")
	}

	// Check if token is expired
	if time.Now().After(prt.ExpiresAt) {
		return errors.New("token has expired")
	}

	// Validate password length
	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update user password
	if err := s.repo.UpdateUserPassword(prt.UserID, string(hash)); err != nil {
		return errors.New("failed to update password")
	}

	// Mark token as used (single-use)
	if err := s.repo.MarkTokenUsed(prt); err != nil {
		utils.GetLogger().Warn("Failed to mark token as used", "error", err)
	}

	return nil
}

func (s *authService) GetUserByID(id uint) (*User, error) {
	return s.repo.FindUserByID(id)
}

func (s *authService) UpdateUserProfile(userID uint, name string, phone string, address string) (*User, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Update fields (email is intentionally excluded from this method)
	if name != "" {
		user.Name = name
	}
	if phone != "" {
		user.Phone = phone
	}
	if address != "" {
		user.Address = address
	}

	// Note: We use repo.Update without transaction since it's a simple update
	if err := s.repo.UpdateUserProfileFields(user); err != nil {
		return nil, errors.New("failed to update profile")
	}

	return user, nil
}

func (s *authService) ChangePassword(userID uint, currentPassword string, newPassword string) error {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Check that new password is different
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newPassword)); err == nil {
		return errors.New("new password must be different")
	}

	// Validate new password length
	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	// Hash and save new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return errors.New("failed to hash password")
	}

	if err := s.repo.UpdateUserPassword(userID, string(hash)); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

// CheckWorkspaceMembership checks if a user is a member of a given workspace.
// M11: Workspace Switch Endpoint, M-MIGRATION: renamed from CheckOrganizationMembership
func (s *authService) CheckWorkspaceMembership(userID uint, workspaceID uint) (bool, error) {
	return s.wsRepo.CheckMembership(userID, workspaceID)
}

// FindByID returns a minimal user by ID for quota checks.
// M5: Subscription Tiers — Phase 5: Service Layer
// M-MIGRATION: Returns minimal user without tier fields
func (s *authService) FindByID(id uint) (*interfaces.MinimalUser, error) {
	user, err := s.repo.FindUserByID(id)
	if err != nil {
		return nil, err
	}
	return &interfaces.MinimalUser{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Phone:     user.Phone,
		Address:   user.Address,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
	}, nil
}

// FindByIDForModels returns a minimal user for models package (cross-package interface).
// M-MIGRATION: Returns minimal user without tier fields
func (s *authService) FindByIDForModels(id uint) (*models.MinimalUser, error) {
	user, err := s.repo.FindUserByID(id)
	if err != nil {
		return nil, err
	}
	return &models.MinimalUser{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Phone:     user.Phone,
		Address:   user.Address,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
	}, nil
}
