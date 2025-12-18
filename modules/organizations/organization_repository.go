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
	FindMemberIDs(orgID uint) ([]uint, error)
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
	err := r.db.First(&org, id).Error
	return &org, err
}

func (r *organizationRepository) AddMember(orgID uint, userID uint) error {
	return r.db.Table("organization_users").Create(map[string]interface{}{
		"organization_id": orgID,
		"user_id":         userID,
	}).Error
}

func (r *organizationRepository) IsMember(userID uint, orgID uint) (bool, error) {
	var count int64
	err := r.db.Table("organization_users").
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Count(&count).Error
	return count > 0, err
}

func (r *organizationRepository) FindMemberIDs(orgID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.Table("organization_users").
		Where("organization_id = ?", orgID).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}
