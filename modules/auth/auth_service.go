package auth

import (
	"errors"
	"gotask-backend/models"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Signup(input SignupInput) (*models.User, *models.Organization, error)
	Login(input LoginInput) (string, error)
	AddUserToOrg(currentUserID uint, orgIDStr string, emailToAdd string) error
}

type authService struct {
	repo AuthRepository
}

func NewAuthService(repo AuthRepository) AuthService {
	return &authService{repo}
}

// DTOs
type SignupInput struct {
	Email    string
	Password string
	OrgName  string
}

type LoginInput struct {
	Email    string
	Password string
}

// --- Implementation ---

func (s *authService) Signup(input SignupInput) (*models.User, *models.Organization, error) {
	// 1. Hash Password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return nil, nil, errors.New("failed to hash password")
	}

	// 2. Create User
	user := models.User{
		Email:    input.Email,
		Password: string(hash),
	}
	if err := s.repo.CreateUser(&user); err != nil {
		return nil, nil, errors.New("email already registered")
	}

	// 3. Create Org (Optional)
	var org *models.Organization
	if input.OrgName != "" {
		newOrg := models.Organization{
			Name:    input.OrgName,
			OwnerID: user.ID,
			Users:   []models.User{user}, // Automatically adds to join table
		}
		if err := s.repo.CreateOrganization(&newOrg); err != nil {
			// If org fails, we might want to return error or just ignore.
			// For strictness, let's return error.
			return &user, nil, errors.New("failed to create organization")
		}
		org = &newOrg
	}

	return &user, org, nil
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

func (s *authService) AddUserToOrg(currentUserID uint, orgIDStr string, emailToAdd string) error {
	// 1. Security Check
	inOrg, err := s.repo.CheckUserInOrg(currentUserID, orgIDStr)
	if err != nil || !inOrg {
		return errors.New("access denied")
	}

	// 2. Find User
	userToAdd, err := s.repo.FindUserByEmail(emailToAdd)
	if err != nil {
		return errors.New("user not found")
	}

	// 3. Parse ID
	orgIDUint, _ := strconv.ParseUint(orgIDStr, 10, 64)

	// 4. Add
	return s.repo.AddUserToOrg(uint(orgIDUint), userToAdd.ID)
}
