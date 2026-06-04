package workspaces

import (
	"errors"

	"gorm.io/gorm"
)

type InvitationRepository interface {
	Create(invitation *WorkspaceInvitation) error
	FindByToken(token string) (*WorkspaceInvitation, error)
	FindByID(id uint) (*WorkspaceInvitation, error)
	FindPendingByWorkspace(workspaceID uint) ([]WorkspaceInvitation, error)
	Update(invitation *WorkspaceInvitation) error
	Delete(invitation *WorkspaceInvitation) error
	DeleteByToken(token string) error
}

type invitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) InvitationRepository {
	return &invitationRepository{db}
}

func (r *invitationRepository) Create(invitation *WorkspaceInvitation) error {
	return r.db.Create(invitation).Error
}

func (r *invitationRepository) FindByToken(token string) (*WorkspaceInvitation, error) {
	var invitation WorkspaceInvitation
	err := r.db.Where("token = ?", token).First(&invitation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) FindByID(id uint) (*WorkspaceInvitation, error) {
	var invitation WorkspaceInvitation
	err := r.db.First(&invitation, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) FindPendingByWorkspace(workspaceID uint) ([]WorkspaceInvitation, error) {
	var invitations []WorkspaceInvitation
	err := r.db.Where("workspace_id = ? AND status = 'pending'", workspaceID).
		Order("created_at desc").
		Find(&invitations).Error
	return invitations, err
}

func (r *invitationRepository) Update(invitation *WorkspaceInvitation) error {
	return r.db.Save(invitation).Error
}

func (r *invitationRepository) Delete(invitation *WorkspaceInvitation) error {
	return r.db.Delete(invitation).Error
}

func (r *invitationRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&WorkspaceInvitation{}).Error
}
