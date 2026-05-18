package organizations

import (
	"errors"

	"gorm.io/gorm"
)

type InvitationRepository interface {
	Create(invitation *OrganizationInvitation) error
	FindByToken(token string) (*OrganizationInvitation, error)
	FindByID(id uint) (*OrganizationInvitation, error)
	FindPendingByOrg(orgID uint) ([]OrganizationInvitation, error)
	Update(invitation *OrganizationInvitation) error
	Delete(invitation *OrganizationInvitation) error
	DeleteByToken(token string) error
}

type invitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) InvitationRepository {
	return &invitationRepository{db}
}

func (r *invitationRepository) Create(invitation *OrganizationInvitation) error {
	return r.db.Create(invitation).Error
}

func (r *invitationRepository) FindByToken(token string) (*OrganizationInvitation, error) {
	var invitation OrganizationInvitation
	err := r.db.Where("token = ?", token).First(&invitation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) FindByID(id uint) (*OrganizationInvitation, error) {
	var invitation OrganizationInvitation
	err := r.db.First(&invitation, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) FindPendingByOrg(orgID uint) ([]OrganizationInvitation, error) {
	var invitations []OrganizationInvitation
	err := r.db.Where("org_id = ? AND status = 'pending'", orgID).
		Order("created_at desc").
		Find(&invitations).Error
	return invitations, err
}

func (r *invitationRepository) Update(invitation *OrganizationInvitation) error {
	return r.db.Save(invitation).Error
}

func (r *invitationRepository) Delete(invitation *OrganizationInvitation) error {
	return r.db.Delete(invitation).Error
}

func (r *invitationRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&OrganizationInvitation{}).Error
}
