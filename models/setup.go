package models

import (
	"time"
)

// 1. New Status Model
type Status struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique" json:"name"` // e.g., "Todo"
	Slug string `gorm:"unique" json:"slug"` // e.g., "todo"
}

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"unique" json:"email"`
	Password  string    `json:"-"`
	Projects  []Project `gorm:"many2many:project_users;" json:"projects,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Users       []User    `gorm:"many2many:project_users;" json:"users,omitempty"`
	Tasks       []Task    `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// 2. Updated Task Model
type Task struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `json:"title"`
	StatusID  uint      `json:"status_id"`
	Status    Status    `gorm:"foreignKey:StatusID" json:"status"`
	ProjectID uint      `json:"project_id"`
	Assignees []User    `gorm:"many2many:task_users;" json:"assignees"`
	CreatedAt time.Time `json:"created_at"`
}
