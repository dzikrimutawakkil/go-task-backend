package auth

import "time"

type User struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Email           string     `gorm:"unique" json:"email"`
	Name            string     `json:"name"`
	Phone           string     `json:"phone"`
	Address         string     `json:"address"`
	Password        string     `gorm:"column:password_hash" json:"-"`
	Tier            string     `gorm:"default:free" json:"tier"`
	TierExpiresAt   *time.Time `json:"tier_expires_at"`
	TierActivatedAt *time.Time `json:"tier_activated_at"`
	TierActivatedBy *uint      `json:"tier_activated_by"`
	CreatedAt       time.Time  `json:"created_at"`
}
