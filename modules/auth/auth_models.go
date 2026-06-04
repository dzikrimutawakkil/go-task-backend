package auth

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"unique" json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Password  string    `gorm:"column:password_hash" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
