package organizations

import (
	"errors"
	"gotask-backend/models"
	"gotask-backend/modules/auth"
)

type OrganizationService interface {
	CreateOrganization(name string, ownerID uint) (*Organization, error)
	CheckAccess(userID uint, orgID uint) (bool, error)
	InviteMember(orgID uint, email string, requesterID uint) error
	GetMembers(orgID uint) ([]auth.User, error)
	RemoveMember(orgID uint, targetUserID uint, requesterID uint) error
	UpdateMemberRole(orgID uint, targetUserID uint, newRole models.Role, requesterID uint) error
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
	org := Organization{
		Name:    name,
		OwnerID: ownerID,
	}

	if err := s.repo.Create(&org); err != nil {
		return nil, err
	}

	// Add owner as member with owner role
	if err := s.repo.AddMember(org.ID, ownerID, models.RoleOwner); err != nil {
		return nil, err
	}

	return &org, nil
}

func (s *organizationService) CheckAccess(userID uint, orgID uint) (bool, error) {
	return s.repo.IsMember(userID, orgID)
}

// InviteMember adds an existing user directly to the organization as member.
func (s *organizationService) InviteMember(orgID uint, email string, requesterID uint) error {
	// Check requester permission
	requesterRole, err := s.repo.GetMemberRole(requesterID, orgID)
	if err != nil {
		return errors.New("cannot determine requester role")
	}

	if !requesterRole.CanInvite() {
		return errors.New("insufficient permission to invite members")
	}

	user, err := s.authService.GetUserByEmail(email)
	if err != nil {
		return errors.New("user with this email not found")
	}

	isMember, err := s.repo.IsMember(user.ID, orgID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("user is already a member")
	}

	return s.repo.AddMember(orgID, user.ID, models.RoleMember)
}

func (s *organizationService) GetMembers(orgID uint) ([]auth.User, error) {
	memberIDs, err := s.repo.FindMemberIDs(orgID)
	if err != nil {
		return nil, err
	}

	if len(memberIDs) == 0 {
		return []auth.User{}, nil
	}

	return s.authService.GetUsersByIDs(memberIDs)
}

// RemoveMember removes a member from the organization.
// Owner cannot be removed.
func (s *organizationService) RemoveMember(orgID uint, targetUserID uint, requesterID uint) error {
	// Get target's role
	targetRole, err := s.repo.GetMemberRole(targetUserID, orgID)
	if err != nil {
		return errors.New("target user is not a member of this organization")
	}

	// Owner cannot be removed
	if targetRole == models.RoleOwner {
		return errors.New("cannot remove the organization owner")
	}

	// Get requester's role
	requesterRole, err := s.repo.GetMemberRole(requesterID, orgID)
	if err != nil {
		return errors.New("requester is not a member of this organization")
	}

	// Permission check
	if !requesterRole.CanRemoveMember() {
		return errors.New("insufficient permission to remove members")
	}

	return s.repo.RemoveMember(orgID, targetUserID)
}

// UpdateMemberRole changes a member's role.
// Owner role cannot be changed. Only owner/admin can change roles.
func (s *organizationService) UpdateMemberRole(orgID uint, targetUserID uint, newRole models.Role, requesterID uint) error {
	if !newRole.IsValid() {
		return errors.New("invalid role value")
	}

	if newRole == models.RoleOwner {
		return errors.New("cannot promote someone to owner")
	}

	// Get target's role
	targetRole, err := s.repo.GetMemberRole(targetUserID, orgID)
	if err != nil {
		return errors.New("target user is not a member of this organization")
	}

	// Owner role cannot be changed
	if targetRole == models.RoleOwner {
		return errors.New("cannot change owner role")
	}

	// Get requester's role
	requesterRole, err := s.repo.GetMemberRole(requesterID, orgID)
	if err != nil {
		return errors.New("requester is not a member of this organization")
	}

	// Permission check
	if !requesterRole.CanUpdateMemberRole() {
		return errors.New("insufficient permission to update member roles")
	}

	return s.repo.UpdateMemberRole(orgID, targetUserID, newRole)
}
