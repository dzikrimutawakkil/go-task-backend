package organizations

import (
	"errors"
	"gotask-backend/config"
	"gotask-backend/internal/interfaces"
	"gotask-backend/models"
	"gotask-backend/utils"
	"time"
)

type OrganizationService interface {
	CreateOrganization(name string, ownerID uint) (*Organization, error)
	CheckAccess(userID uint, orgID uint) (bool, error)
	InviteMember(orgID uint, email string, requesterID uint) error
	GetMembers(orgID uint) ([]interfaces.MinimalUser, error)
	RemoveMember(orgID uint, targetUserID uint, requesterID uint) error
	UpdateMemberRole(orgID uint, targetUserID uint, newRole models.Role, requesterID uint) error
	GetUserOrganizations(userID uint) ([]Organization, error)
	// M5: Subscription Tiers
	ActivateForModelsTier(userID uint, tier string, durationMonths int, activatedBy uint) (*ActivateTierResult, error)
	GetTierInfoForModels(userID uint) (*TierInfoResponse, error)
	GetTierPlans() ([]TierPlanWithLimits, error)
}

// ActivateTierResult is returned when a tier is activated successfully.
type ActivateTierResult struct {
	UserID          uint     `json:"user_id"`
	Tier            string   `json:"tier"`
	TierExpiresAt   string   `json:"tier_expires_at"`
	TierActivatedAt string   `json:"tier_activated_at"`
	AffectedOrgs    []string `json:"affected_organizations"`
}

// TierInfoResponse represents the user's tier info for API response.
type TierInfoResponse struct {
	Tier          string     `json:"tier"`
	EffectiveTier string     `json:"effective_tier"`
	IsActive      bool       `json:"is_active"`
	ActivatedAt   string     `json:"activated_at,omitempty"`
	ExpiresAt     string     `json:"expires_at,omitempty"`
	DaysRemaining int        `json:"days_remaining"`
	Limits        LimitsInfo `json:"limits"`
	Features      Features   `json:"features"`
	Usage         UsageInfo  `json:"usage"`
}

// LimitsInfo represents tier limits for API response.
type LimitsInfo struct {
	MaxWorkspaces       int `json:"max_workspaces"`
	MaxProjectsPerWS    int `json:"max_projects_per_workspace"`
	MaxTasksPerProject  int `json:"max_tasks_per_project"`
	MaxMembersPerWS     int `json:"max_members_per_workspace"`
	MaxClients          int `json:"max_clients"`
	MaxInvoicesPerMonth int `json:"max_invoices_per_month"`
}

// UsageInfo represents current usage stats for API response.
type UsageInfo struct {
	OwnedWorkspaces   int `json:"owned_workspaces"`
	Projects          int `json:"projects"`
	Members           int `json:"members"`
	Clients           int `json:"clients"`
	InvoicesThisMonth int `json:"invoices_this_month"`
}

type organizationService struct {
	repo         OrganizationRepository
	authService  interfaces.AuthService
	tierPlanRepo TierPlanRepository
}

// NewOrganizationService creates a new organization service.
// M5: Subscription Tiers — uses interfaces.AuthService for quota checks.
func NewOrganizationService(repo OrganizationRepository, authS interfaces.AuthService, tierPlanRepo TierPlanRepository) OrganizationService {
	return &organizationService{
		repo:         repo,
		authService:  authS,
		tierPlanRepo: tierPlanRepo,
	}
}

func (s *organizationService) GetUserOrganizations(userID uint) ([]Organization, error) {
	return s.repo.FindOrganizationsByUserID(userID)
}

func (s *organizationService) CreateOrganization(name string, ownerID uint) (*Organization, error) {
	// M5: Quota check — check user's workspace limit before creating
	effectiveTier := "free"
	limits := utils.GetTierLimits("free")

	if user, err := s.authService.FindByID(ownerID); err == nil {
		effectiveTier = utils.GetEffectiveTier(user.Tier, user.TierExpiresAt)
		limits = utils.GetTierLimits(effectiveTier)
	}

	// Count owned orgs
	ownedOrgs, err := s.repo.CountByOwner(ownerID)
	if err != nil {
		return nil, err
	}

	if ownedOrgs >= limits.MaxWorkspaces {
		return nil, utils.ErrQuotaExceeded("workspace", limits.MaxWorkspaces, effectiveTier)
	}

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

	user, err := s.authService.GetMinimalUserByEmail(email)
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

	// M5: Quota check — check member limit per org
	org, err := s.repo.FindByID(orgID)
	if err != nil {
		return err
	}

	effectiveTier := "free"
	limits := utils.GetTierLimits("free")

	if owner, err := s.authService.FindByID(org.OwnerID); err == nil {
		effectiveTier = utils.GetEffectiveTier(owner.Tier, owner.TierExpiresAt)
		limits = utils.GetTierLimits(effectiveTier)
	}

	if limits.MaxMembers != -1 {
		count, err := s.repo.CountMembers(orgID)
		if err != nil {
			return err
		}
		if count >= limits.MaxMembers {
			return utils.ErrQuotaExceeded("member", limits.MaxMembers, effectiveTier)
		}
	}

	return s.repo.AddMember(orgID, user.ID, models.RoleMember)
}

func (s *organizationService) GetMembers(orgID uint) ([]interfaces.MinimalUser, error) {
	memberIDs, err := s.repo.FindMemberIDs(orgID)
	if err != nil {
		return nil, err
	}

	if len(memberIDs) == 0 {
		return []interfaces.MinimalUser{}, nil
	}

	return s.authService.GetMinimalUsersByIDs(memberIDs)
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

// GetTierInfoForModels returns the user's tier information and usage statistics.
// M5: Subscription Tiers — Phase 5: Service Layer
func (s *organizationService) GetTierInfoForModels(userID uint) (*TierInfoResponse, error) {
	user, err := s.authService.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	effectiveTier := utils.GetEffectiveTier(user.Tier, user.TierExpiresAt)
	limits := utils.GetTierLimits(effectiveTier)
	isActive := utils.IsTierActive(user.Tier, user.TierExpiresAt)

	// Calculate days remaining
	var daysRemaining int
	if user.TierExpiresAt != nil {
		daysRemaining = utils.DaysRemaining(user.TierExpiresAt)
	}

	// Get usage stats
	ownedOrgs, _ := s.repo.CountByOwner(userID)
	orgs, _ := s.repo.FindOrganizationsByUserID(userID)

	// Count projects via raw table query (avoid import cycles)
	var projectsCount int
	for _, org := range orgs {
		var count int64
		config.DB.Table("projects").Where("organization_id = ?", org.ID).Count(&count)
		projectsCount += int(count)
	}

	// Count members across all orgs
	var totalMembers int
	for _, org := range orgs {
		count, _ := s.repo.CountMembers(org.ID)
		totalMembers += count
	}

	// Count clients via raw table query
	var clientsCount int
	for _, org := range orgs {
		var count int64
		config.DB.Table("clients").Where("organization_id = ?", org.ID).Count(&count)
		clientsCount += int(count)
	}

	// Count invoices this month via raw table query
	var invoicesCount int
	monthStart := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	for _, org := range orgs {
		var count int64
		config.DB.Table("invoices").
			Where("organization_id = ? AND created_at >= ?", org.ID, monthStart).
			Count(&count)
		invoicesCount += int(count)
	}

	result := &TierInfoResponse{
		Tier:          user.Tier,
		EffectiveTier: effectiveTier,
		IsActive:      isActive,
		ActivatedAt:   formatTimePtr(user.TierActivatedAt),
		ExpiresAt:     formatTimePtr(user.TierExpiresAt),
		DaysRemaining: daysRemaining,
		Limits: LimitsInfo{
			MaxWorkspaces:       limits.MaxWorkspaces,
			MaxProjectsPerWS:    limits.MaxProjects,
			MaxTasksPerProject:  limits.MaxTasksPerProject,
			MaxMembersPerWS:     limits.MaxMembers,
			MaxClients:          limits.MaxClients,
			MaxInvoicesPerMonth: limits.MaxInvoicesPerMonth,
		},
		Features: Features{
			Comments: limits.CanComment,
			Realtime: limits.CanSSE,
			AuditLog: limits.CanAuditLog,
		},
		Usage: UsageInfo{
			OwnedWorkspaces:   ownedOrgs,
			Projects:          projectsCount,
			Members:           totalMembers,
			Clients:           clientsCount,
			InvoicesThisMonth: invoicesCount,
		},
	}

	return result, nil
}

// ActivateForModelsTier activates a tier for a user.
// M5: Subscription Tiers — Phase 5: Service Layer
func (s *organizationService) ActivateForModelsTier(userID uint, tier string, durationMonths int, activatedBy uint) (*ActivateTierResult, error) {
	// Validate tier
	validTiers := map[string]bool{"free": true, "pro": true, "ultimate": true}
	if !validTiers[tier] {
		return nil, errors.New("invalid tier: must be free, pro, or ultimate")
	}

	// Validate duration (1-24 months)
	if durationMonths < 1 || durationMonths > 24 {
		return nil, errors.New("duration_months must be between 1 and 24")
	}

	// Get user
	user, err := s.authService.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Calculate new expiry date
	now := time.Now()
	activatedAt := now
	var tierExpiresAt time.Time

	if user.TierExpiresAt != nil && user.TierExpiresAt.After(now) {
		// Extend from current expiry
		tierExpiresAt = user.TierExpiresAt.AddDate(0, durationMonths, 0)
	} else {
		// Start from now
		tierExpiresAt = now.AddDate(0, durationMonths, 0)
	}

	// Update user in DB (use table name to avoid import cycle)
	activatedByUint := activatedBy
	result := config.DB.Table("users").
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"tier":              tier,
			"tier_activated_at": activatedAt,
			"tier_expires_at":   tierExpiresAt,
			"tier_activated_by": activatedByUint,
		})
	if result.Error != nil {
		return nil, errors.New("failed to update user tier")
	}

	// Get affected orgs
	orgs, _ := s.repo.FindOrganizationsByUserID(userID)
	var orgNames []string
	for _, org := range orgs {
		orgNames = append(orgNames, org.Name)
	}

	return &ActivateTierResult{
		UserID:          userID,
		Tier:            tier,
		TierExpiresAt:   tierExpiresAt.Format(time.RFC3339),
		TierActivatedAt: activatedAt.Format(time.RFC3339),
		AffectedOrgs:    orgNames,
	}, nil
}

// formatTimePtr converts a time pointer to RFC3339 string or empty string.
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// GetTierPlans returns all active tier plans with their limits.
// M5: Subscription Tiers — Phase 5: Handler & Endpoint
func (s *organizationService) GetTierPlans() ([]TierPlanWithLimits, error) {
	return s.tierPlanRepo.FindAllPlansWithLimits()
}
