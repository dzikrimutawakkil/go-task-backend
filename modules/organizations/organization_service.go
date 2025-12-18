package organizations

import (
	"errors"
	"gotask-backend/modules/auth"
)

type OrganizationService interface {
	CreateOrganization(name string, ownerID uint) (*Organization, error)
	CheckAccess(userID uint, orgID uint) (bool, error)
	InviteMember(orgID uint, email string) error
	GetMembers(orgID uint) ([]auth.User, error)
}

type organizationService struct {
	repo        OrganizationRepository
	authService auth.AuthService
}

func NewOrganizationService(repo OrganizationRepository, authS auth.AuthService) OrganizationService {
	return &organizationService{
		repo:        repo,
		authService: authS,
	}
}

func (s *organizationService) CreateOrganization(name string, ownerID uint) (*Organization, error) {
	// Buat Object Organization
	org := Organization{
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

func (s *organizationService) InviteMember(orgID uint, email string) error {
	// Panggil Auth Service (Komunikasi antar module)
	user, err := s.authService.GetUserByEmail(email)
	if err != nil {
		return errors.New("user with this email not found")
	}

	// Cek logic membership di repo sendiri
	isMember, err := s.repo.IsMember(user.ID, orgID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("user is already a member")
	}

	return s.repo.AddMember(orgID, user.ID)
}

func (s *organizationService) GetMembers(orgID uint) ([]auth.User, error) {
	// Ambil ID Member dari database sendiri (Organization)
	memberIDs, err := s.repo.FindMemberIDs(orgID)
	if err != nil {
		return nil, err
	}

	if len(memberIDs) == 0 {
		return []auth.User{}, nil
	}

	// Ambil Detail User dari Service Tetangga (Auth)
	return s.authService.GetUsersByIDs(memberIDs)
}
