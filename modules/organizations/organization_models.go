package organizations

import (
	"gotask-backend/models"
	"time"
)

// OrganizationType constants
const (
	OrgTypePersonal = "personal"
	OrgTypeTeam     = "team"
)

type Organization struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"unique" json:"name"`
	OwnerID   uint      `json:"owner_id"`
	OrgType   string    `gorm:"type:varchar(20);default:'personal'" json:"org_type"`
	CreatedAt time.Time `json:"created_at"`
}

type OrganizationUser struct {
	OrganizationID uint        `gorm:"primaryKey"`
	UserID         uint        `gorm:"primaryKey"`
	Role           models.Role `gorm:"type:varchar(20);default:member" json:"role"`
	JoinedAt       time.Time   `json:"joined_at"`
}

// Role constants for convenience
const (
	RoleOwner  = models.RoleOwner
	RoleAdmin  = models.RoleAdmin
	RoleMember = models.RoleMember
)
