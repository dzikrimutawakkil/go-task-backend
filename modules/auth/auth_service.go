package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gotask-backend/internal/interfaces"
	"gotask-backend/modules/organizations"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Signup(input SignupInput) (*User, error)
	Login(input LoginInput) (string, error)
	GetUsersByIDs(ids []uint) ([]User, error)
	GetUserByEmail(email string) (*User, error)
	GetMinimalUserByEmail(email string) (*interfaces.MinimalUser, error)
	GetMinimalUsersByIDs(ids []uint) ([]interfaces.MinimalUser, error)
}

// compile-time interface satisfaction check
var _ interfaces.AuthService = (*authService)(nil)

type authService struct {
	repo    AuthRepository
	orgRepo organizations.OrganizationRepository
}

func NewAuthService(repo AuthRepository, orgRepo organizations.OrganizationRepository) AuthService {
	return &authService{repo: repo, orgRepo: orgRepo}
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

	// 3. Auto-create personal workspace
	personalOrg := organizations.Organization{
		Name:    fmt.Sprintf("%s's Workspace", input.Name),
		OwnerID: user.ID,
		OrgType: organizations.OrgTypePersonal,
	}
	if err := s.orgRepo.Create(&personalOrg); err != nil {
		return nil, errors.New("failed to create personal workspace")
	}

	// 4. Add owner as member with RoleOwner
	if err := s.orgRepo.AddMember(personalOrg.ID, user.ID, organizations.RoleOwner); err != nil {
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
