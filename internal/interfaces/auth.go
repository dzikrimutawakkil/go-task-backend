package interfaces

import "time"

// MinimalUser is a minimal user representation for cross-package interfaces.
// M5: Subscription Tiers — added Tier fields for quota checks.
type MinimalUser struct {
	ID              uint
	Email           string
	Name            string
	Phone           string
	Address         string
	Password        string
	Tier            string
	TierExpiresAt   *time.Time
	TierActivatedAt *time.Time
	TierActivatedBy *uint
	CreatedAt       time.Time
}

// OrgInfo represents minimal org info for quota checks.
type OrgInfo struct {
	ID      uint
	OwnerID uint
}

// OrgFinder defines methods for finding organization info.
// M5: Subscription Tiers — used by other packages to avoid import cycles.
type OrgFinder interface {
	FindOrgInfoByID(id uint) (*OrgInfo, error)
}

// AuthService defines the authentication operations needed by other packages.
type AuthService interface {
	GetMinimalUserByEmail(email string) (*MinimalUser, error)
	GetMinimalUsersByIDs(ids []uint) ([]MinimalUser, error)
	FindByID(id uint) (*MinimalUser, error)
	// Note: GetUsersByIDs (returning User) is NOT in this interface due to Go's
	// limitation of not supporting method overloading. Use GetMinimalUsersByIDs instead.
}
