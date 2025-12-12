package models

import (
	"time"
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Email    string `gorm:"unique" json:"email"`
	Password string `json:"-"`

	// NEW: Many-to-Many relationship
	// "project_users" is the name of the hidden join table GORM will create
	Projects []Project `gorm:"many2many:project_users;" json:"projects,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

type Project struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// NEW: The reverse relationship
	Users []User `gorm:"many2many:project_users;" json:"users,omitempty"`

	Tasks     []Task    `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Task struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	ProjectID uint      `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
}
