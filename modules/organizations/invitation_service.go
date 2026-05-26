package organizations

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"gotask-backend/internal/interfaces"
	"gotask-backend/models"
	"gotask-backend/utils"
)

type InvitationService interface {
	CreateInvitation(orgID uint, email string, role models.Role, requesterID uint) (*OrganizationInvitation, error)
	AcceptInvitation(token string) (*Organization, error)
	GetPendingInvitations(orgID uint) ([]OrganizationInvitation, error)
	ResendInvitation(invitationID uint) error
	RevokeInvitation(token string, requesterID uint, orgID uint) error
}

type invitationService struct {
	repo        InvitationRepository
	orgRepo     OrganizationRepository
	authService interfaces.AuthService
}

func NewInvitationService(
	repo InvitationRepository,
	orgRepo OrganizationRepository,
	authS interfaces.AuthService,
) InvitationService {
	return &invitationService{
		repo:        repo,
		orgRepo:     orgRepo,
		authService: authS,
	}
}

// generateToken creates a new UUID token for invitation.
func generateToken() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() >> uint(i*8) & 0xFF)
	}
	result := ""
	for i, v := range b {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			result += "-"
		}
		result += fmt.Sprintf("%02x", v)
	}
	return result
}

func (s *invitationService) CreateInvitation(orgID uint, email string, role models.Role, requesterID uint) (*OrganizationInvitation, error) {
	// Check permission
	requesterRole, err := s.orgRepo.GetMemberRole(requesterID, orgID)
	if err != nil {
		return nil, errors.New("cannot determine requester role")
	}
	if !requesterRole.CanInvite() {
		return nil, errors.New("insufficient permission to invite members")
	}

	// Check if email already in org
	user, err := s.authService.GetMinimalUserByEmail(email)
	if err != nil {
		return nil, errors.New("user with this email does not exist")
	}

	isMember, err := s.orgRepo.IsMember(user.ID, orgID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, errors.New("user is already a member of this organization")
	}

	// Get expiry hours from env (default 168 hours = 7 days)
	expiryHours := 168
	if envHours := os.Getenv("INVITE_EXPIRY_HOURS"); envHours != "" {
		if h, err := strconv.Atoi(envHours); err == nil && h > 0 {
			expiryHours = h
		}
	}

	invitation := &OrganizationInvitation{
		OrgID:        orgID,
		InvitedEmail: email,
		Token:        generateToken(),
		Role:         string(role),
		ExpiresAt:    time.Now().Add(time.Duration(expiryHours) * time.Hour),
		CreatedBy:    requesterID,
		Status:       "pending",
	}

	if err := s.repo.Create(invitation); err != nil {
		return nil, err
	}

	// Send email invitation
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "https://journey-rpg-lotus-collect.trycloudflare.com"
	}
	inviteURL := appURL + "/invite/" + invitation.Token

	// Get org name for email
	org, _ := s.orgRepo.FindByID(orgID)
	orgName := "the organization"
	if org != nil {
		orgName = org.Name
	}

	// Send email (async, don't block on email failure)
	go func() {
		_ = utils.SendInviteEmail(email, orgName, inviteURL)
	}()

	return invitation, nil
}

func (s *invitationService) AcceptInvitation(token string) (*Organization, error) {
	invitation, err := s.repo.FindByToken(token)
	if err != nil {
		return nil, errors.New("invitation not found")
	}

	// Check status
	if invitation.Status == "accepted" {
		return nil, errors.New("invitation has already been used")
	}
	if invitation.Status == "expired" || invitation.Status == "revoked" {
		return nil, errors.New("invitation has been revoked")
	}

	// Check expiry
	if time.Now().After(invitation.ExpiresAt) {
		// Mark as expired
		invitation.Status = "expired"
		_ = s.repo.Update(invitation)
		return nil, errors.New("invitation has expired")
	}

	// Get the invited user (must be logged in to accept)
	user, err := s.authService.GetMinimalUserByEmail(invitation.InvitedEmail)
	if err != nil {
		return nil, errors.New("cannot identify accepting user")
	}

	// Check if already a member
	isMember, err := s.orgRepo.IsMember(user.ID, invitation.OrgID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, errors.New("you are already a member of this organization")
	}

	// Add user to organization with the invited role
	role := models.Role(invitation.Role)
	if !role.IsValid() {
		role = models.RoleMember
	}

	if err := s.orgRepo.AddMember(invitation.OrgID, user.ID, role); err != nil {
		return nil, err
	}

	// Mark invitation as accepted
	now := time.Now()
	invitation.AcceptedAt = &now
	invitation.Status = "accepted"
	_ = s.repo.Update(invitation)

	// Return the organization
	return s.orgRepo.FindByID(invitation.OrgID)
}

func (s *invitationService) GetPendingInvitations(orgID uint) ([]OrganizationInvitation, error) {
	return s.repo.FindPendingByOrg(orgID)
}

func (s *invitationService) ResendInvitation(invitationID uint) error {
	invitation, err := s.repo.FindByID(invitationID)
	if err != nil {
		return errors.New("invitation not found")
	}

	if invitation.Status != "pending" {
		return errors.New("can only resend pending invitations")
	}

	// Generate new token and reset expiry
	invitation.Token = generateToken()
	invitation.ExpiresAt = time.Now().Add(168 * time.Hour)
	_ = s.repo.Update(invitation)

	// Get org name
	org, _ := s.orgRepo.FindByID(invitation.OrgID)
	orgName := "the organization"
	if org != nil {
		orgName = org.Name
	}

	// Resend email
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "https://journey-rpg-lotus-collect.trycloudflare.com"
	}
	inviteURL := appURL + "/invite/" + invitation.Token

	go func() {
		_ = utils.SendInviteEmail(invitation.InvitedEmail, orgName, inviteURL)
	}()

	return nil
}

func (s *invitationService) RevokeInvitation(token string, requesterID uint, orgID uint) error {
	// Check permission
	requesterRole, err := s.orgRepo.GetMemberRole(requesterID, orgID)
	if err != nil {
		return errors.New("cannot determine requester role")
	}
	if !requesterRole.CanRemoveMember() {
		return errors.New("insufficient permission to revoke invitations")
	}

	invitation, err := s.repo.FindByToken(token)
	if err != nil {
		return errors.New("invitation not found")
	}

	if invitation.OrgID != orgID {
		return errors.New("invitation does not belong to this organization")
	}

	invitation.Status = "revoked"
	return s.repo.Update(invitation)
}
