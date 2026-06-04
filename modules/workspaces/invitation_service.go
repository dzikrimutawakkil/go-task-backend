package workspaces

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
	CreateInvitation(workspaceID uint, email string, role models.Role, requesterID uint) (*WorkspaceInvitation, error)
	AcceptInvitation(token string) (*Workspace, error)
	GetPendingInvitations(workspaceID uint) ([]WorkspaceInvitation, error)
	ResendInvitation(invitationID uint) error
	RevokeInvitation(token string, requesterID uint, workspaceID uint) error
}

type invitationService struct {
	repo        InvitationRepository
	wsRepo      WorkspaceRepository
	authService interfaces.AuthService
}

func NewInvitationService(
	repo InvitationRepository,
	wsRepo WorkspaceRepository,
	authS interfaces.AuthService,
) InvitationService {
	return &invitationService{
		repo:        repo,
		wsRepo:      wsRepo,
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

func (s *invitationService) CreateInvitation(workspaceID uint, email string, role models.Role, requesterID uint) (*WorkspaceInvitation, error) {
	// Check permission
	requesterRole, err := s.wsRepo.GetMemberRole(requesterID, workspaceID)
	if err != nil {
		return nil, errors.New("cannot determine requester role")
	}
	if !requesterRole.CanInvite() {
		return nil, errors.New("insufficient permission to invite members")
	}

	// Check if email already in workspace
	user, err := s.authService.GetMinimalUserByEmail(email)
	if err != nil {
		return nil, errors.New("user with this email does not exist")
	}

	isMember, err := s.wsRepo.IsMember(user.ID, workspaceID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, errors.New("user is already a member of this workspace")
	}

	// Get expiry hours from env (default 168 hours = 7 days)
	expiryHours := 168
	if envHours := os.Getenv("INVITE_EXPIRY_HOURS"); envHours != "" {
		if h, err := strconv.Atoi(envHours); err == nil && h > 0 {
			expiryHours = h
		}
	}

	invitation := &WorkspaceInvitation{
		WorkspaceID:  workspaceID,
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

	// Get workspace name for email
	ws, _ := s.wsRepo.FindByID(workspaceID)
	wsName := "the workspace"
	if ws != nil {
		wsName = ws.Name
	}

	// Send email (async, don't block on email failure)
	go func() {
		_ = utils.SendInviteEmail(email, wsName, inviteURL)
	}()

	return invitation, nil
}

func (s *invitationService) AcceptInvitation(token string) (*Workspace, error) {
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
	isMember, err := s.wsRepo.IsMember(user.ID, invitation.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, errors.New("you are already a member of this workspace")
	}

	// Add user to workspace with the invited role
	role := models.Role(invitation.Role)
	if !role.IsValid() {
		role = models.RoleMember
	}

	if err := s.wsRepo.AddMember(invitation.WorkspaceID, user.ID, role); err != nil {
		return nil, err
	}

	// Mark invitation as accepted
	now := time.Now()
	invitation.AcceptedAt = &now
	invitation.Status = "accepted"
	_ = s.repo.Update(invitation)

	// Return the workspace
	return s.wsRepo.FindByID(invitation.WorkspaceID)
}

func (s *invitationService) GetPendingInvitations(workspaceID uint) ([]WorkspaceInvitation, error) {
	return s.repo.FindPendingByWorkspace(workspaceID)
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

	// Get workspace name
	ws, _ := s.wsRepo.FindByID(invitation.WorkspaceID)
	wsName := "the workspace"
	if ws != nil {
		wsName = ws.Name
	}

	// Resend email
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "https://journey-rpg-lotus-collect.trycloudflare.com"
	}
	inviteURL := appURL + "/invite/" + invitation.Token

	go func() {
		_ = utils.SendInviteEmail(invitation.InvitedEmail, wsName, inviteURL)
	}()

	return nil
}

func (s *invitationService) RevokeInvitation(token string, requesterID uint, workspaceID uint) error {
	// Check permission
	requesterRole, err := s.wsRepo.GetMemberRole(requesterID, workspaceID)
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

	if invitation.WorkspaceID != workspaceID {
		return errors.New("invitation does not belong to this workspace")
	}

	invitation.Status = "revoked"
	return s.repo.Update(invitation)
}
