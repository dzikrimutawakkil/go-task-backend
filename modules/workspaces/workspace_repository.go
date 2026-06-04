package workspaces

import (
	"gotask-backend/internal/interfaces"
	"gotask-backend/models"
	"strconv"

	"gorm.io/gorm"
)

type WorkspaceRepository interface {
	Create(ws *Workspace) error
	FindByID(id uint) (*Workspace, error)
	FindPersonalWorkspaceByOwnerID(ownerID uint) (*Workspace, error)
	AddMember(workspaceID uint, userID uint, role models.Role) error
	IsMember(userID uint, workspaceID uint) (bool, error)
	CheckMembership(userID uint, workspaceID uint) (bool, error) // M11: alias for IsMember
	FindMemberIDs(workspaceID uint) ([]uint, error)
	GetMemberRole(userID uint, workspaceID uint) (models.Role, error)
	UpdateMemberRole(workspaceID uint, userID uint, newRole models.Role) error
	RemoveMember(workspaceID uint, userID uint) error
	GetMemberRoleByWorkspaceID(userID uint, workspaceID string) (models.Role, error)
	FindWorkspacesByUserID(userID uint) ([]Workspace, error)
	// M5: Quota helpers
	CountByOwner(ownerID uint) (int, error)
	CountMembers(workspaceID uint) (int, error)

	// M5: Implements interfaces.WorkspaceFinder for quota checks
	FindWorkspaceInfoByID(id uint) (*interfaces.WorkspaceInfo, error)
}

type workspaceRepository struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) WorkspaceRepository {
	return &workspaceRepository{db}
}

func (r *workspaceRepository) Create(ws *Workspace) error {
	return r.db.Create(ws).Error
}

func (r *workspaceRepository) FindByID(id uint) (*Workspace, error) {
	var ws Workspace
	err := r.db.First(&ws, id).Error
	return &ws, err
}

func (r *workspaceRepository) FindPersonalWorkspaceByOwnerID(ownerID uint) (*Workspace, error) {
	var ws Workspace
	err := r.db.Where("owner_id = ? AND workspace_type = ?", ownerID, WorkspaceTypePersonal).First(&ws).Error
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

// AddMember adds a user to the workspace with a given role.
func (r *workspaceRepository) AddMember(workspaceID uint, userID uint, role models.Role) error {
	if role == "" {
		role = models.RoleMember
	}
	return r.db.Table("workspace_members").Create(map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      userID,
		"role":         string(role),
	}).Error
}

func (r *workspaceRepository) IsMember(userID uint, workspaceID uint) (bool, error) {
	var count int64
	err := r.db.Table("workspace_members").
		Where("user_id = ? AND workspace_id = ?", userID, workspaceID).
		Count(&count).Error
	return count > 0, err
}

// CheckMembership is an alias for IsMember, used by M11: Workspace Switch Endpoint
func (r *workspaceRepository) CheckMembership(userID uint, workspaceID uint) (bool, error) {
	return r.IsMember(userID, workspaceID)
}

func (r *workspaceRepository) FindMemberIDs(workspaceID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.Table("workspace_members").
		Where("workspace_id = ?", workspaceID).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}

// GetMemberRole returns the role of a user in a workspace.
func (r *workspaceRepository) GetMemberRole(userID uint, workspaceID uint) (models.Role, error) {
	var member WorkspaceMember
	err := r.db.Table("workspace_members").
		Where("user_id = ? AND workspace_id = ?", userID, workspaceID).
		First(&member).Error
	return member.Role, err
}

// GetMemberRoleByWorkspaceID is an overloaded version that takes workspaceID as string (from X-Workspace-ID header).
func (r *workspaceRepository) GetMemberRoleByWorkspaceID(userID uint, workspaceID string) (models.Role, error) {
	wsID, err := parseUint(workspaceID)
	if err != nil {
		return "", err
	}
	return r.GetMemberRole(userID, wsID)
}

// UpdateMemberRole updates the role of a member in the workspace.
func (r *workspaceRepository) UpdateMemberRole(workspaceID uint, userID uint, newRole models.Role) error {
	return r.db.Table("workspace_members").
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Update("role", string(newRole)).Error
}

// RemoveMember removes a user from the workspace.
func (r *workspaceRepository) RemoveMember(workspaceID uint, userID uint) error {
	return r.db.Table("workspace_members").
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Delete(&WorkspaceMember{}).Error
}

// parseUint converts a string to uint.
func parseUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

// FindWorkspacesByUserID get all workspaces that followed by userID
func (r *workspaceRepository) FindWorkspacesByUserID(userID uint) ([]Workspace, error) {
	var workspaces []Workspace
	err := r.db.
		Joins("JOIN workspace_members ON workspace_members.workspace_id = workspaces.id").
		Where("workspace_members.user_id = ?", userID).
		Find(&workspaces).Error

	return workspaces, err
}

// CountByOwner returns the number of workspaces owned by a specific user.
// M5: Subscription Tiers — Phase 5: Service Layer — Quota check for workspace limit.
func (r *workspaceRepository) CountByOwner(ownerID uint) (int, error) {
	var count int64
	err := r.db.Model(&Workspace{}).Where("owner_id = ?", ownerID).Count(&count).Error
	return int(count), err
}

// CountMembers returns the number of members in a workspace.
// M5: Subscription Tiers — Phase 5: Service Layer — Quota check for member limit.
func (r *workspaceRepository) CountMembers(workspaceID uint) (int, error) {
	var count int64
	err := r.db.Table("workspace_members").
		Where("workspace_id = ?", workspaceID).
		Count(&count).Error
	return int(count), err
}

// FindWorkspaceInfoByID implements interfaces.WorkspaceFinder for quota checks.
// M5: Subscription Tiers — Phase 5: Service Layer — Returns minimal workspace info for quota checks.
func (r *workspaceRepository) FindWorkspaceInfoByID(id uint) (*interfaces.WorkspaceInfo, error) {
	var ws Workspace
	err := r.db.Select("id, owner_id, tier, tier_expires_at, tier_activated_at, tier_activated_by").First(&ws, id).Error
	if err != nil {
		return nil, err
	}
	return &interfaces.WorkspaceInfo{
		ID:              ws.ID,
		OwnerID:         ws.OwnerID,
		Tier:            ws.Tier,
		TierExpiresAt:   ws.TierExpiresAt,
		TierActivatedAt: ws.TierActivatedAt,
	}, nil
}
