package auth

import (
	"time"
)

type PasswordResetToken struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Token     string     `gorm:"unique;column:token" json:"token"`
	UserID    uint       `gorm:"column:user_id" json:"user_id"`
	ExpiresAt time.Time  `gorm:"column:expires_at" json:"expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at" json:"used_at"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
}
