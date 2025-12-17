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
	FindMembers(orgID uint) ([]models.User, error)
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
	// HAPUS Preload("Users").
	// Kita tidak ingin mengambil data user secara otomatis saat mengambil data organisasi.
	err := r.db.First(&org, id).Error
	return &org, err
}

func (r *organizationRepository) AddMember(orgID uint, userID uint) error {
	// GANTI Association dengan Insert Manual ke tabel pivot
	// Asumsi tabel pivot bernama "organization_users" dengan kolom "organization_id" dan "user_id"

	// Kita gunakan Exec raw SQL atau Map Create agar tidak butuh struct model pivot
	// Cara aman dengan GORM map creation:
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

func (r *organizationRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *organizationRepository) FindMembers(orgID uint) ([]models.User, error) {
	var users []models.User
	// Fungsi ini masih melakukan JOIN ke tabel Users.
	// Dalam Microservices murni, seharusnya fungsi ini hanya mengembalikan []uint (UserIDs).
	// Tapi untuk tahap transisi ini, kita biarkan dulu karena handler membutuhkannya.
	err := r.db.Table("users").
		Joins("JOIN organization_users ON organization_users.user_id = users.id").
		Where("organization_users.organization_id = ?", orgID).
		Select("users.id, users.email").
		Find(&users).Error
	return users, err
}
