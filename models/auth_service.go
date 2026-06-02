package models

import "time"

// MinimalUser is a minimal user representation for cross-package interfaces.
// Used to avoid import cycles between auth and organizations packages.
// M5: Subscription Tiers — replaced Plan/License with Tier fields.
type MinimalUser struct {
	ID              uint
	Email           string
	Name            string
	Phone           string
	Address         string
	Password        string     `json:"-"`
	Tier            string     `json:"tier"`
	TierExpiresAt   *time.Time `json:"tier_expires_at,omitempty"`
	TierActivatedAt *time.Time `json:"tier_activated_at,omitempty"`
	TierActivatedBy *uint      `json:"tier_activated_by,omitempty"`
	CreatedAt       time.Time
}

// AuthService defines the authentication operations needed by other packages.
// Defined here (neutral package) to avoid import cycles between auth and organizations packages.
type AuthService interface {
	GetUserByEmail(email string) (*MinimalUser, error)
	GetUsersByIDs(ids []uint) ([]MinimalUser, error)
	FindByID(id uint) (*MinimalUser, error)
}
