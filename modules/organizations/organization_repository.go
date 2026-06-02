package organizations

import (
	"gotask-backend/internal/interfaces"
	"gotask-backend/models"
	"strconv"

	"gorm.io/gorm"
)

type OrganizationRepository interface {
	Create(org *Organization) error
	FindByID(id uint) (*Organization, error)
	FindPersonalOrgByOwnerID(ownerID uint) (*Organization, error)
	AddMember(orgID uint, userID uint, role models.Role) error
	IsMember(userID uint, orgID uint) (bool, error)
	CheckMembership(userID uint, orgID uint) (bool, error) // M11: alias for IsMember
	FindMemberIDs(orgID uint) ([]uint, error)
	GetMemberRole(userID uint, orgID uint) (models.Role, error)
	UpdateMemberRole(orgID uint, userID uint, newRole models.Role) error
	RemoveMember(orgID uint, userID uint) error
	GetMemberRoleByOrgID(userID uint, orgID string) (models.Role, error)
	FindOrganizationsByUserID(userID uint) ([]Organization, error)
	// M5: Quota helpers
	CountByOwner(ownerID uint) (int, error)
	CountMembers(orgID uint) (int, error)

	// M5: Implements interfaces.OrgFinder for quota checks
	FindOrgInfoByID(id uint) (*interfaces.OrgInfo, error)
}

type organizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &organizationRepository{db}
}

func (r *organizationRepository) Create(org *Organization) error {
	return r.db.Create(org).Error
}

func (r *organizationRepository) FindByID(id uint) (*Organization, error) {
	var org Organization
	err := r.db.First(&org, id).Error
	return &org, err
}

func (r *organizationRepository) FindPersonalOrgByOwnerID(ownerID uint) (*Organization, error) {
	var org Organization
	err := r.db.Where("owner_id = ? AND org_type = ?", ownerID, OrgTypePersonal).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// AddMember adds a user to the organization with a given role.
func (r *organizationRepository) AddMember(orgID uint, userID uint, role models.Role) error {
	if role == "" {
		role = models.RoleMember
	}
	return r.db.Table("organization_users").Create(map[string]interface{}{
		"organization_id": orgID,
		"user_id":         userID,
		"role":            string(role),
	}).Error
}

func (r *organizationRepository) IsMember(userID uint, orgID uint) (bool, error) {
	var count int64
	err := r.db.Table("organization_users").
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Count(&count).Error
	return count > 0, err
}

// CheckMembership is an alias for IsMember, used by M11: Workspace Switch Endpoint
func (r *organizationRepository) CheckMembership(userID uint, orgID uint) (bool, error) {
	return r.IsMember(userID, orgID)
}

func (r *organizationRepository) FindMemberIDs(orgID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.Table("organization_users").
		Where("organization_id = ?", orgID).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}

// GetMemberRole returns the role of a user in an organization.
func (r *organizationRepository) GetMemberRole(userID uint, orgID uint) (models.Role, error) {
	var orgUser OrganizationUser
	err := r.db.Table("organization_users").
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		First(&orgUser).Error
	return orgUser.Role, err
}

// GetMemberRoleByOrgID is an overloaded version that takes orgID as string (from X-Organization-ID header).
func (r *organizationRepository) GetMemberRoleByOrgID(userID uint, orgID string) (models.Role, error) {
	orgIDUint, err := parseUint(orgID)
	if err != nil {
		return "", err
	}
	return r.GetMemberRole(userID, orgIDUint)
}

// UpdateMemberRole updates the role of a member in the organization.
func (r *organizationRepository) UpdateMemberRole(orgID uint, userID uint, newRole models.Role) error {
	return r.db.Table("organization_users").
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Update("role", string(newRole)).Error
}

// RemoveMember removes a user from the organization.
func (r *organizationRepository) RemoveMember(orgID uint, userID uint) error {
	return r.db.Table("organization_users").
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Delete(&OrganizationUser{}).Error
}

// parseUint converts a string to uint.
func parseUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

// FindOrganizationsByUserID get all organization that followed by userID
func (r *organizationRepository) FindOrganizationsByUserID(userID uint) ([]Organization, error) {
	var orgs []Organization
	err := r.db.
		Joins("JOIN organization_users ON organization_users.organization_id = organizations.id").
		Where("organization_users.user_id = ?", userID).
		Find(&orgs).Error

	return orgs, err
}

// CountByOwner returns the number of organizations owned by a specific user.
// M5: Subscription Tiers — Phase 5: Service Layer — Quota check for workspace limit.
func (r *organizationRepository) CountByOwner(ownerID uint) (int, error) {
	var count int64
	err := r.db.Model(&Organization{}).Where("owner_id = ?", ownerID).Count(&count).Error
	return int(count), err
}

// CountMembers returns the number of members in an organization.
// M5: Subscription Tiers — Phase 5: Service Layer — Quota check for member limit.
func (r *organizationRepository) CountMembers(orgID uint) (int, error) {
	var count int64
	err := r.db.Table("organization_users").
		Where("organization_id = ?", orgID).
		Count(&count).Error
	return int(count), err
}

// FindOrgInfoByID implements interfaces.OrgFinder for quota checks.
// M5: Subscription Tiers — Phase 5: Service Layer — Returns minimal org info for quota checks.
func (r *organizationRepository) FindOrgInfoByID(id uint) (*interfaces.OrgInfo, error) {
	var org Organization
	err := r.db.Select("id, owner_id").First(&org, id).Error
	if err != nil {
		return nil, err
	}
	return &interfaces.OrgInfo{
		ID:      org.ID,
		OwnerID: org.OwnerID,
	}, nil
}
