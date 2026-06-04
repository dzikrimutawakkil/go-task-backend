package workspaces

import (
	"errors"
	"gotask-backend/config"
	"gotask-backend/internal/interfaces"
	"gotask-backend/models"
	"gotask-backend/utils"
	"time"
)

type WorkspaceService interface {
	CreateWorkspace(name string, ownerID uint) (*Workspace, error)
	CheckAccess(userID uint, workspaceID uint) (bool, error)
	InviteMember(workspaceID uint, email string, requesterID uint) error
	GetMembers(workspaceID uint) ([]interfaces.MinimalUser, error)
	RemoveMember(workspaceID uint, targetUserID uint, requesterID uint) error
	UpdateMemberRole(workspaceID uint, targetUserID uint, newRole models.Role, requesterID uint) error
	GetUserWorkspaces(userID uint) ([]Workspace, error)
	// M5: Subscription Tiers
	ActivateTier(workspaceID uint, tier string, durationMonths int, activatedBy uint) (*ActivateTierResult, error)
	GetTierInfoForWorkspace(workspaceID uint) (*TierInfoResponse, error)
	GetTierPlans() ([]TierPlanWithLimits, error)
}

// ActivateTierResult is returned when a tier is activated successfully.
type ActivateTierResult struct {
	WorkspaceID     uint   `json:"workspace_id"`
	Tier            string `json:"tier"`
	TierExpiresAt   string `json:"tier_expires_at"`
	TierActivatedAt string `json:"tier_activated_at"`
	WorkspaceName   string `json:"workspace_name"`
}

// TierInfoResponse represents the workspace's tier info for API response.
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

type workspaceService struct {
	repo         WorkspaceRepository
	authService  interfaces.AuthService
	tierPlanRepo TierPlanRepository
}

// NewWorkspaceService creates a new workspace service.
// M5: Subscription Tiers — uses interfaces.AuthService for quota checks.
func NewWorkspaceService(repo WorkspaceRepository, authS interfaces.AuthService, tierPlanRepo TierPlanRepository) WorkspaceService {
	return &workspaceService{
		repo:         repo,
		authService:  authS,
		tierPlanRepo: tierPlanRepo,
	}
}

func (s *workspaceService) GetUserWorkspaces(userID uint) ([]Workspace, error) {
	return s.repo.FindWorkspacesByUserID(userID)
}

func (s *workspaceService) CreateWorkspace(name string, ownerID uint) (*Workspace, error) {
	// M5: Quota check — check user's workspace limit based on user's owned workspaces
	// Note: User tier is no longer used; quota is based on workspace's tier
	// For new workspace creation, we default to free tier and check overall limit
	effectiveTier := "free"
	limits := utils.GetTierLimits("free")

	// Check user's owned workspaces
	ownedCount, err := s.repo.CountByOwner(ownerID)
	if err != nil {
		return nil, err
	}

	// Find any existing workspace owned by this user to get effective tier
	workspaces, _ := s.repo.FindWorkspacesByUserID(ownerID)
	for _, ws := range workspaces {
		wsTier := utils.GetEffectiveTier(ws.Tier, ws.TierExpiresAt)
		if wsTier != "free" && wsTier != effectiveTier {
			effectiveTier = wsTier
			limits = utils.GetTierLimits(effectiveTier)
			break
		}
	}

	if ownedCount >= limits.MaxWorkspaces {
		return nil, utils.ErrQuotaExceeded("workspace", limits.MaxWorkspaces, effectiveTier)
	}

	ws := Workspace{
		Name:          name,
		OwnerID:       ownerID,
		WorkspaceType: WorkspaceTypePersonal,
		Tier:          "free", // New workspace always starts with free tier
	}

	if err := s.repo.Create(&ws); err != nil {
		return nil, err
	}

	// Add owner as member with owner role
	if err := s.repo.AddMember(ws.ID, ownerID, models.RoleOwner); err != nil {
		return nil, err
	}

	return &ws, nil
}

func (s *workspaceService) CheckAccess(userID uint, workspaceID uint) (bool, error) {
	return s.repo.IsMember(userID, workspaceID)
}

// InviteMember adds an existing user directly to the workspace as member.
func (s *workspaceService) InviteMember(workspaceID uint, email string, requesterID uint) error {
	// Check requester permission
	requesterRole, err := s.repo.GetMemberRole(requesterID, workspaceID)
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

	isMember, err := s.repo.IsMember(user.ID, workspaceID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("user is already a member")
	}

	// M5: Quota check — check member limit per workspace (based on workspace's tier)
	ws, err := s.repo.FindByID(workspaceID)
	if err != nil {
		return err
	}

	effectiveTier := utils.GetEffectiveTier(ws.Tier, ws.TierExpiresAt)
	limits := utils.GetTierLimits(effectiveTier)

	if limits.MaxMembers != -1 {
		count, err := s.repo.CountMembers(workspaceID)
		if err != nil {
			return err
		}
		if count >= limits.MaxMembers {
			return utils.ErrQuotaExceeded("member", limits.MaxMembers, effectiveTier)
		}
	}

	return s.repo.AddMember(workspaceID, user.ID, models.RoleMember)
}

func (s *workspaceService) GetMembers(workspaceID uint) ([]interfaces.MinimalUser, error) {
	memberIDs, err := s.repo.FindMemberIDs(workspaceID)
	if err != nil {
		return nil, err
	}

	if len(memberIDs) == 0 {
		return []interfaces.MinimalUser{}, nil
	}

	return s.authService.GetMinimalUsersByIDs(memberIDs)
}

// RemoveMember removes a member from the workspace.
// Owner cannot be removed.
func (s *workspaceService) RemoveMember(workspaceID uint, targetUserID uint, requesterID uint) error {
	// Get target's role
	targetRole, err := s.repo.GetMemberRole(targetUserID, workspaceID)
	if err != nil {
		return errors.New("target user is not a member of this workspace")
	}

	// Owner cannot be removed
	if targetRole == models.RoleOwner {
		return errors.New("cannot remove the workspace owner")
	}

	// Get requester's role
	requesterRole, err := s.repo.GetMemberRole(requesterID, workspaceID)
	if err != nil {
		return errors.New("requester is not a member of this workspace")
	}

	// Permission check
	if !requesterRole.CanRemoveMember() {
		return errors.New("insufficient permission to remove members")
	}

	return s.repo.RemoveMember(workspaceID, targetUserID)
}

// UpdateMemberRole changes a member's role.
// Owner role cannot be changed. Only owner/admin can change roles.
func (s *workspaceService) UpdateMemberRole(workspaceID uint, targetUserID uint, newRole models.Role, requesterID uint) error {
	if !newRole.IsValid() {
		return errors.New("invalid role value")
	}

	if newRole == models.RoleOwner {
		return errors.New("cannot promote someone to owner")
	}

	// Get target's role
	targetRole, err := s.repo.GetMemberRole(targetUserID, workspaceID)
	if err != nil {
		return errors.New("target user is not a member of this workspace")
	}

	// Owner role cannot be changed
	if targetRole == models.RoleOwner {
		return errors.New("cannot change owner role")
	}

	// Get requester's role
	requesterRole, err := s.repo.GetMemberRole(requesterID, workspaceID)
	if err != nil {
		return errors.New("requester is not a member of this workspace")
	}

	// Permission check
	if !requesterRole.CanUpdateMemberRole() {
		return errors.New("insufficient permission to update member roles")
	}

	return s.repo.UpdateMemberRole(workspaceID, targetUserID, newRole)
}

// GetTierInfoForWorkspace returns the workspace's tier information and usage statistics.
// M5: Subscription Tiers — Phase 5: Service Layer
func (s *workspaceService) GetTierInfoForWorkspace(workspaceID uint) (*TierInfoResponse, error) {
	ws, err := s.repo.FindByID(workspaceID)
	if err != nil {
		return nil, errors.New("workspace not found")
	}

	effectiveTier := utils.GetEffectiveTier(ws.Tier, ws.TierExpiresAt)
	limits := utils.GetTierLimits(effectiveTier)
	isActive := utils.IsTierActive(ws.Tier, ws.TierExpiresAt)

	// Calculate days remaining
	var daysRemaining int
	if ws.TierExpiresAt != nil {
		daysRemaining = utils.DaysRemaining(ws.TierExpiresAt)
	}

	// Get usage stats
	ownedCount, _ := s.repo.CountByOwner(ws.OwnerID)

	// Count projects via raw table query (avoid import cycles)
	var projectsCount int
	var count int64
	config.DB.Table("projects").Where("workspace_id = ?", ws.ID).Count(&count)
	projectsCount = int(count)

	// Count members
	membersCount, _ := s.repo.CountMembers(ws.ID)

	// Count clients via raw table query
	var clientsCount int
	config.DB.Table("clients").Where("workspace_id = ?", ws.ID).Count(&count)
	clientsCount = int(count)

	// Count invoices this month via raw table query
	var invoicesCount int
	monthStart := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	config.DB.Table("invoices").
		Where("workspace_id = ? AND created_at >= ?", ws.ID, monthStart).
		Count(&count)
	invoicesCount = int(count)

	result := &TierInfoResponse{
		Tier:          ws.Tier,
		EffectiveTier: effectiveTier,
		IsActive:      isActive,
		ActivatedAt:   formatTimePtr(ws.TierActivatedAt),
		ExpiresAt:     formatTimePtr(ws.TierExpiresAt),
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
			OwnedWorkspaces:   ownedCount,
			Projects:          projectsCount,
			Members:           membersCount,
			Clients:           clientsCount,
			InvoicesThisMonth: invoicesCount,
		},
	}

	return result, nil
}

// ActivateTier activates a tier for a workspace.
// M5: Subscription Tiers — Phase 5: Service Layer
func (s *workspaceService) ActivateTier(workspaceID uint, tier string, durationMonths int, activatedBy uint) (*ActivateTierResult, error) {
	// Validate tier - only new tier names (free/pro/ultimate)
	validTiers := map[string]bool{"free": true, "pro": true, "ultimate": true}
	if !validTiers[tier] {
		return nil, errors.New("invalid tier: must be free, pro, or ultimate")
	}

	// Validate duration (1-24 months)
	if durationMonths < 1 || durationMonths > 24 {
		return nil, errors.New("duration_months must be between 1 and 24")
	}

	// Get workspace
	ws, err := s.repo.FindByID(workspaceID)
	if err != nil {
		return nil, errors.New("workspace not found")
	}

	// Calculate new expiry date
	now := time.Now()
	activatedAt := now
	var tierExpiresAt time.Time

	if ws.TierExpiresAt != nil && ws.TierExpiresAt.After(now) {
		// Extend from current expiry
		tierExpiresAt = ws.TierExpiresAt.AddDate(0, durationMonths, 0)
	} else {
		// Start from now
		tierExpiresAt = now.AddDate(0, durationMonths, 0)
	}

	// Update workspace tier in DB
	activatedByUint := activatedBy
	result := config.DB.Table("workspaces").
		Where("id = ?", workspaceID).
		Updates(map[string]interface{}{
			"tier":              tier,
			"tier_activated_at": activatedAt,
			"tier_expires_at":   tierExpiresAt,
			"tier_activated_by": activatedByUint,
		})
	if result.Error != nil {
		return nil, errors.New("failed to update workspace tier")
	}

	return &ActivateTierResult{
		WorkspaceID:     workspaceID,
		Tier:            tier,
		TierExpiresAt:   tierExpiresAt.Format(time.RFC3339),
		TierActivatedAt: activatedAt.Format(time.RFC3339),
		WorkspaceName:   ws.Name,
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
func (s *workspaceService) GetTierPlans() ([]TierPlanWithLimits, error) {
	return s.tierPlanRepo.FindAllPlansWithLimits()
}
