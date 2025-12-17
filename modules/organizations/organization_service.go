package organizations

import (
	"errors"
	"gotask-backend/models"
)

type OrganizationService interface {
	CreateOrganization(name string, ownerID uint) (*models.Organization, error)
	CheckAccess(userID uint, orgID uint) (bool, error)
	InviteMember(orgID uint, email string) error
	GetMembers(orgID uint) ([]models.User, error)
}

type organizationService struct {
	repo OrganizationRepository
}

func NewOrganizationService(repo OrganizationRepository) OrganizationService {
	return &organizationService{repo}
}

func (s *organizationService) CreateOrganization(name string, ownerID uint) (*models.Organization, error) {
	// Buat Object Organization
	org := models.Organization{
		Name:    name,
		OwnerID: ownerID,
	}

	// Simpan ke DB
	if err := s.repo.Create(&org); err != nil {
		return nil, err
	}

	// Tambahkan Owner sebagai Member (Manual Call)
	if err := s.repo.AddMember(org.ID, ownerID); err != nil {
		return nil, err
	}

	return &org, nil
}

func (s *organizationService) CheckAccess(userID uint, orgID uint) (bool, error) {
	return s.repo.IsMember(userID, orgID)
}

// Add/Update this method in your Service interface & implementation
func (s *organizationService) InviteMember(orgID uint, email string) error {
	// 1. Find the User by Email
	// (We need the repo to support this lookup)
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		return errors.New("user with this email not found")
	}

	// 2. Check if already a member
	isMember, err := s.repo.IsMember(user.ID, orgID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("user is already a member")
	}

	// 3. Add Member
	return s.repo.AddMember(orgID, user.ID)
}

func (s *organizationService) GetMembers(orgID uint) ([]models.User, error) {
	return s.repo.FindMembers(orgID)
}
