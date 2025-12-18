package auth

import (
	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateUser(user *User) error
	FindUserByEmail(email string) (*User, error)
	FindUserByID(id uint) (*User, error)
	FindUsersByIDs(ids []uint) ([]User, error)
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
