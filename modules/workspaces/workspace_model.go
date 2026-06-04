package workspaces

import (
	"gotask-backend/models"
	"time"
)

// WorkspaceType constants
const (
	WorkspaceTypePersonal = "personal"
	WorkspaceTypeTeam     = "team"
)

type Workspace struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"unique" json:"name"`
	OwnerID         uint       `json:"owner_id"`
	WorkspaceType   string     `gorm:"column:workspace_type;type:varchar(20);default:'personal'" json:"workspace_type"`
	Tier            string     `gorm:"default:'free'" json:"tier"`
	TierExpiresAt   *time.Time `json:"tier_expires_at,omitempty"`
	TierActivatedAt *time.Time `json:"tier_activated_at,omitempty"`
	TierActivatedBy *uint      `json:"tier_activated_by,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type WorkspaceMember struct {
	WorkspaceID uint        `gorm:"primaryKey"`
	UserID      uint        `gorm:"primaryKey"`
	Role        models.Role `gorm:"type:varchar(20);default:member" json:"role"`
	JoinedAt    time.Time   `json:"joined_at"`
}

// Role constants for convenience
const (
	RoleOwner  = models.RoleOwner
	RoleAdmin  = models.RoleAdmin
	RoleMember = models.RoleMember
)
