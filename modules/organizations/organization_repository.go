package organizations

import (
	"gotask-backend/models"

	"gorm.io/gorm"
)

type OrganizationRepository interface {
	Create(org *models.Organization) error
	FindByID(id uint) (*models.Organization, error)
	AddMember(orgID uint, userID uint) error
	IsMember(userID uint, orgID uint) (bool, error)
	FindUserByEmail(email string) (*models.User, error)
}

type organizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &organizationRepository{db}
}

func (r *organizationRepository) Create(org *models.Organization) error {
	return r.db.Create(org).Error
}

func (r *organizationRepository) FindByID(id uint) (*models.Organization, error) {
	var org models.Organization
	// Preload Users to see members, or omit if not needed for basic checks
	err := r.db.Preload("Users").First(&org, id).Error
	return &org, err
}

func (r *organizationRepository) AddMember(orgID uint, userID uint) error {
	// Use GORM Association to append user to the organization
	var org models.Organization
	org.ID = orgID

	var user models.User
	user.ID = userID

	return r.db.Model(&org).Association("Users").Append(&user)
}

func (r *organizationRepository) IsMember(userID uint, orgID uint) (bool, error) {
	var count int64
	err := r.db.Table("organization_users").
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Count(&count).Error
	return count > 0, err
}

func (r *organizationRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}
