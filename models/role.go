package models

// Role represents a user's role within an organization.
// Used for RBAC permission checks.
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

// IsValid checks if the role value is valid.
func (r Role) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	}
	return false
}

// CanInvite returns true if the role can invite new members.
func (r Role) CanInvite() bool {
	return r == RoleOwner || r == RoleAdmin
}

// CanRemoveMember returns true if the role can remove members.
func (r Role) CanRemoveMember() bool {
	return r == RoleOwner || r == RoleAdmin
}

// CanUpdateMemberRole returns true if the role can update member roles.
func (r Role) CanUpdateMemberRole() bool {
	return r == RoleOwner || r == RoleAdmin
}

// CanDeleteOrganization returns true if the role can delete the organization.
func (r Role) CanDeleteOrganization() bool {
	return r == RoleOwner
}

// CanDeleteProject returns true if the role can delete projects.
func (r Role) CanDeleteProject() bool {
	return r == RoleOwner || r == RoleAdmin
}

// CanManageLabels returns true if the role can create/update/delete labels.
func (r Role) CanManageLabels() bool {
	return r == RoleOwner || r == RoleAdmin
}
