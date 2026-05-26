package auth

import "time"

type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Email         string    `gorm:"unique" json:"email"`
	Name          string    `json:"name"`
	Phone         string    `json:"phone"`
	Address       string    `json:"address"`
	Password      string    `gorm:"column:password_hash" json:"-"`
	Plan          string    `gorm:"default:free" json:"plan"`
	LicenseKey    *string   `gorm:"column:license_key" json:"license_key,omitempty"`
	LicenseStatus string    `gorm:"default:inactive" json:"license_status"`
	CreatedAt     time.Time `json:"created_at"`
}
