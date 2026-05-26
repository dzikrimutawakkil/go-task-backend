package licenses

import (
	"time"
)

// License represents a license key in the system
type License struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Key         string     `gorm:"unique;column:key" json:"key"`
	Type        string     `gorm:"column:type" json:"type"`     // free, pro, team, enterprise
	Status      string     `gorm:"column:status" json:"status"` // available, activated, revoked, expired
	ActivatedBy *uint      `gorm:"column:activated_by" json:"activated_by"`
	ActivatedAt *time.Time `gorm:"column:activated_at" json:"activated_at"`
	ExpiresAt   *time.Time `gorm:"column:expires_at" json:"expires_at"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
}

// License status constants
const (
	LicenseStatusAvailable = "available"
	LicenseStatusActivated = "activated"
	LicenseStatusRevoked   = "revoked"
	LicenseStatusExpired   = "expired"
)

// License type constants
const (
	LicenseTypeFree       = "free"
	LicenseTypePro        = "pro"
	LicenseTypeTeam       = "team"
	LicenseTypeEnterprise = "enterprise"
)
