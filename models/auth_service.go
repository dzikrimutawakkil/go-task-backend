package models

import "time"

// MinimalUser is a minimal user representation for cross-package interfaces.
// Used to avoid import cycles between auth and organizations packages.
type MinimalUser struct {
	ID        uint
	Email     string
	Name      string
	Phone     string
	Address   string
	Password  string `json:"-"`
	CreatedAt time.Time
}

// AuthService defines the authentication operations needed by other packages.
// Defined here (neutral package) to avoid import cycles between auth and organizations packages.
type AuthService interface {
	GetUserByEmail(email string) (*MinimalUser, error)
	GetUsersByIDs(ids []uint) ([]MinimalUser, error)
}
