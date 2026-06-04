package models

import "time"

// MinimalUser is a minimal user representation for cross-package interfaces.
// Used to avoid import cycles between auth and workspaces packages.
// M-MIGRATION: Removed Tier fields - tier is now per-workspace, not per-user.
type MinimalUser struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthService defines the authentication operations needed by other packages.
// Defined here (neutral package) to avoid import cycles between auth and workspaces packages.
type AuthService interface {
	GetUserByEmail(email string) (*MinimalUser, error)
	GetUsersByIDs(ids []uint) ([]MinimalUser, error)
	FindByID(id uint) (*MinimalUser, error)
}
