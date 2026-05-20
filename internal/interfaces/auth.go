package interfaces

import "time"

// MinimalUser is a minimal user representation for cross-package interfaces.
type MinimalUser struct {
	ID        uint
	Email     string
	Name      string
	Phone     string
	Address   string
	Password  string
	CreatedAt time.Time
}

// AuthService defines the authentication operations needed by other packages.
// Defined here (neutral package) to avoid import cycles.
type AuthService interface {
	GetMinimalUserByEmail(email string) (*MinimalUser, error)
	GetMinimalUsersByIDs(ids []uint) ([]MinimalUser, error)
}
