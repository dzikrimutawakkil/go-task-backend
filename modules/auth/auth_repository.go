package auth

import (
	"gotask-backend/models"

	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateUser(user *models.User) error
	FindUserByEmail(email string) (*models.User, error)
	FindUserByID(id uint) (*models.User, error)

	// Organization checks
	CreateOrganization(org *models.Organization) error
	CheckUserInOrg(userID uint, orgID string) (bool, error)
	AddUserToOrg(orgID uint, userID uint) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *authRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *authRepository) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *authRepository) CreateOrganization(org *models.Organization) error {
	return r.db.Create(org).Error
}

func (r *authRepository) CheckUserInOrg(userID uint, orgID string) (bool, error) {
	var count int64
	err := r.db.Table("organization_users").
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Count(&count).Error
	return count > 0, err
}

func (r *authRepository) AddUserToOrg(orgID uint, userID uint) error {
	// We use a raw SQL execution or a struct insert for the join table
	// Using a struct is cleaner if we have the model, but raw is fine for join tables
	// Let's use the association way which is safest in GORM

	var org models.Organization
	org.ID = orgID
	var user models.User
	user.ID = userID

	return r.db.Model(&org).Association("Users").Append(&user)
}
