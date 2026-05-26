package auth

import (
	"time"

	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateUser(user *User) error
	FindUserByEmail(email string) (*User, error)
	FindUserByID(id uint) (*User, error)
	FindUsersByIDs(ids []uint) ([]User, error)
	// Password reset token operations
	SavePasswordResetToken(token *PasswordResetToken) error
	FindPasswordResetToken(token string) (*PasswordResetToken, error)
	MarkTokenUsed(token *PasswordResetToken) error
	UpdateUserPassword(userID uint, passwordHash string) error
	// User profile operations
	UpdateUserProfileFields(user *User) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) CreateUser(user *User) error {
	return r.db.Create(user).Error
}

func (r *authRepository) FindUserByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *authRepository) FindUserByID(id uint) (*User, error) {
	var user User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *authRepository) FindUsersByIDs(ids []uint) ([]User, error) {
	var users []User
	err := r.db.Find(&users, ids).Error
	return users, err
}

func (r *authRepository) SavePasswordResetToken(token *PasswordResetToken) error {
	return r.db.Create(token).Error
}

func (r *authRepository) FindPasswordResetToken(token string) (*PasswordResetToken, error) {
	var prt PasswordResetToken
	err := r.db.Where("token = ?", token).First(&prt).Error
	return &prt, err
}

func (r *authRepository) MarkTokenUsed(token *PasswordResetToken) error {
	now := time.Now()
	token.UsedAt = &now
	return r.db.Save(token).Error
}

func (r *authRepository) UpdateUserPassword(userID uint, passwordHash string) error {
	return r.db.Model(&User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error
}

func (r *authRepository) UpdateUserProfileFields(user *User) error {
	return r.db.Save(user).Error
}
