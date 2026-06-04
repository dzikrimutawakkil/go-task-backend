package workspaces

import (
	"time"
)

// WorkspaceInvitation represents an email invitation to join a workspace.
type WorkspaceInvitation struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	WorkspaceID  uint       `gorm:"index;not null" json:"workspace_id"`
	InvitedEmail string     `gorm:"size:255;not null" json:"invited_email"`
	Token        string     `gorm:"size:36;uniqueIndex;not null" json:"token"`
	Role         string     `gorm:"size:20;default:'member'" json:"role"`
	ExpiresAt    time.Time  `gorm:"not null" json:"expires_at"`
	CreatedBy    uint       `gorm:"not null" json:"created_by"`
	AcceptedAt   *time.Time `json:"accepted_at,omitempty"`
	Status       string     `gorm:"size:20;default:'pending'" json:"status"` // pending, accepted, expired, revoked
	CreatedAt    time.Time  `json:"created_at"`
}
