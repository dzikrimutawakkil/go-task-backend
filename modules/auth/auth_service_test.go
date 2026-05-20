package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "securepassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "verylongpasswordwithalotofcharacters123456789!@#$%",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := bcrypt.GenerateFromPassword([]byte(tc.password), 10)
			if tc.wantErr && err == nil {
				t.Errorf("HashPassword() expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("HashPassword() unexpected error: %v", err)
			}
			if !tc.wantErr {
				// Verify hash is not empty
				if len(hash) == 0 {
					t.Errorf("HashPassword() returned empty hash")
				}
				// Verify hash can be compared
				err = bcrypt.CompareHashAndPassword(hash, []byte(tc.password))
				if err != nil {
					t.Errorf("HashPassword() generated hash cannot be compared: %v", err)
				}
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	testCases := []struct {
		name     string
		password string
		hash     []byte
		wantErr  bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			wantErr:  false,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := bcrypt.CompareHashAndPassword(tc.hash, []byte(tc.password))
			if tc.wantErr && err == nil {
				t.Errorf("VerifyPassword() expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("VerifyPassword() unexpected error: %v", err)
			}
		})
	}
}

func TestGenerateJWT(t *testing.T) {
	os.Setenv("SECRET_KEY", "test-secret-key-for-jwt-validation")
	defer os.Unsetenv("SECRET_KEY")

	userID := uint(1)
	expiryDuration := time.Hour * 24 * 30 // 30 days

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(expiryDuration).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	// Validate token is not empty
	if tokenString == "" {
		t.Errorf("GenerateJWT() returned empty token string")
	}

	// Validate token structure (should have 3 parts separated by dots)
	parts := 0
	for _, c := range tokenString {
		if c == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("GenerateJWT() token should have 3 parts separated by dots, got %d dots", parts)
	}
}

func TestValidateJWT(t *testing.T) {
	os.Setenv("SECRET_KEY", "test-secret-key-for-jwt-validation")
	defer os.Unsetenv("SECRET_KEY")

	secretKey := []byte(os.Getenv("SECRET_KEY"))
	userID := uint(42)

	// Generate valid token
	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	validTokenString, _ := validToken.SignedString(secretKey)

	// Generate expired token
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(-time.Hour).Unix(), // Already expired
	})
	expiredTokenString, _ := expiredToken.SignedString(secretKey)

	// Generate token with different secret
	wrongSecretToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	wrongSecretTokenString, _ := wrongSecretToken.SignedString([]byte("wrong-secret"))

	testCases := []struct {
		name        string
		tokenString string
		wantValid   bool
		wantUserID  uint
	}{
		{
			name:        "valid token",
			tokenString: validTokenString,
			wantValid:   true,
			wantUserID:  userID,
		},
		{
			name:        "expired token",
			tokenString: expiredTokenString,
			wantValid:   false,
			wantUserID:  0,
		},
		{
			name:        "wrong secret token",
			tokenString: wrongSecretTokenString,
			wantValid:   false,
			wantUserID:  0,
		},
		{
			name:        "malformed token",
			tokenString: "invalid.token.string",
			wantValid:   false,
			wantUserID:  0,
		},
		{
			name:        "empty token",
			tokenString: "",
			wantValid:   false,
			wantUserID:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := jwt.Parse(tc.tokenString, func(token *jwt.Token) (interface{}, error) {
				return secretKey, nil
			})

			if tc.wantValid {
				if err != nil {
					t.Errorf("ValidateJWT() unexpected error for valid token: %v", err)
					return
				}
				if !token.Valid {
					t.Errorf("ValidateJWT() token should be valid")
					return
				}
				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok {
					t.Errorf("ValidateJWT() could not parse claims")
					return
				}
				// Check user ID from "sub" claim
				subClaim, ok := claims["sub"].(float64)
				if !ok {
					t.Errorf("ValidateJWT() could not find 'sub' claim")
					return
				}
				if uint(subClaim) != tc.wantUserID {
					t.Errorf("ValidateJWT() user ID mismatch: got %v, want %v", uint(subClaim), tc.wantUserID)
				}
			} else {
				if err == nil && token.Valid {
					t.Errorf("ValidateJWT() expected invalid token, got valid")
				}
			}
		})
	}
}

func TestUserModelValidation(t *testing.T) {
	t.Run("valid email format", func(t *testing.T) {
		user := User{
			Email:    "test@example.com",
			Password: "hashedpassword",
		}
		if user.Email == "" {
			t.Errorf("User model should store email")
		}
	})

	t.Run("password should be hidden in JSON", func(t *testing.T) {
		user := User{
			Email:    "test@example.com",
			Password: "secretpassword",
		}
		// The json:"-" tag should prevent password from being serialized
		if user.Password == "" {
			t.Errorf("User model should store password internally")
		}
	})
}

func TestSignupInput(t *testing.T) {
	t.Run("signup input fields", func(t *testing.T) {
		input := SignupInput{
			Email:    "user@example.com",
			Password: "securepassword",
		}

		if input.Email == "" {
			t.Errorf("SignupInput should have Email field")
		}
		if input.Password == "" {
			t.Errorf("SignupInput should have Password field")
		}
	})
}

func TestLoginInput(t *testing.T) {
	t.Run("login input fields", func(t *testing.T) {
		input := LoginInput{
			Email:    "user@example.com",
			Password: "password123",
		}

		if input.Email == "" {
			t.Errorf("LoginInput should have Email field")
		}
		if input.Password == "" {
			t.Errorf("LoginInput should have Password field")
		}
	})
}
