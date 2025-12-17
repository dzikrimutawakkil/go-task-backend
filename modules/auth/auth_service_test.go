package auth

import (
	"gotask-backend/models"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// 1. Define Mock Repository
type MockAuthRepo struct {
	mock.Mock
}

func (m *MockAuthRepo) CreateUser(user *models.User) error {
	args := m.Called(user)
	user.ID = 1 // Simulate DB ID
	return args.Error(0)
}

func (m *MockAuthRepo) FindUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthRepo) FindUserByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// 2. Test Cases

func TestSignup_Success(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	service := NewAuthService(mockRepo)

	// Expect CreateUser to be called
	// We use mock.MatchedBy to verify the password was hashed!
	mockRepo.On("CreateUser", mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "test@example.com" && u.Password != "password123"
	})).Return(nil)

	input := SignupInput{Email: "test@example.com", Password: "password123"}
	user, err := service.Signup(input)

	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NotEqual(t, "password123", user.Password) // Ensure hashing occurred
	mockRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	// Set secret for JWT generation
	os.Setenv("SECRET_KEY", "testsecret")

	mockRepo := new(MockAuthRepo)
	service := NewAuthService(mockRepo)

	// Create a REAL hash so the service's bcrypt comparison works
	password := "password123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	mockUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashed),
	}

	mockRepo.On("FindUserByEmail", "test@example.com").Return(mockUser, nil)

	input := LoginInput{Email: "test@example.com", Password: password}
	token, err := service.Login(input)

	assert.NoError(t, err)
	assert.NotEmpty(t, token) // Ensure we got a JWT string
	mockRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	mockRepo := new(MockAuthRepo)
	service := NewAuthService(mockRepo)

	// Use a hash for "correct-password"
	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), 10)
	mockUser := &models.User{ID: 1, Email: "test@example.com", Password: string(hashed)}

	mockRepo.On("FindUserByEmail", "test@example.com").Return(mockUser, nil)

	// Try logging in with WRONG password
	input := LoginInput{Email: "test@example.com", Password: "wrong-password"}
	token, err := service.Login(input)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "invalid email or password", err.Error())
}
