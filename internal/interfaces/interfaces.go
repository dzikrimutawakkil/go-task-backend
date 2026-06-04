package interfaces

import (
	"time"
)

// MinimalUser is a minimal user representation for cross-package interfaces.
// Used to avoid import cycles between auth and workspaces packages.
type MinimalUser struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// WorkspaceInfo is a minimal workspace info for quota checks.
type WorkspaceInfo struct {
	ID              uint       `json:"id"`
	OwnerID         uint       `json:"owner_id"`
	Tier            string     `json:"tier"`
	TierExpiresAt   *time.Time `json:"tier_expires_at,omitempty"`
	TierActivatedAt *time.Time `json:"tier_activated_at,omitempty"`
}

// AuthService defines the authentication operations needed by other packages.
type AuthService interface {
	GetMinimalUserByEmail(email string) (*MinimalUser, error)
	GetMinimalUsersByIDs(ids []uint) ([]MinimalUser, error)
	FindByID(id uint) (*MinimalUser, error)
}

// WorkspaceFinder defines the workspace lookup operations needed by other packages.
type WorkspaceFinder interface {
	FindWorkspaceInfoByID(id uint) (*WorkspaceInfo, error)
}
