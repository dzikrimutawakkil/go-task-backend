package auth

import (
	"errors"
	"gotask-backend/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Signup(input SignupInput) (*models.User, error)
	Login(input LoginInput) (string, error)
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
}

type LoginInput struct {
	Email    string
	Password string
}

func (s *authService) Signup(input SignupInput) (*models.User, error) {
	// 1. Hash Password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 2. Create User
	user := models.User{
		Email:    input.Email,
		Password: string(hash),
	}
	if err := s.repo.CreateUser(&user); err != nil {
		return nil, errors.New("email already registered")
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
