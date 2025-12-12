package models

import (
	"time"
)

// 1. The Parent Table
type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tasks       []Task    `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"` // One-to-Many relation
	CreatedAt   time.Time `json:"created_at"`
}

// 2. The Child Table
type Task struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`     // e.g., "To Do", "Done"
	ProjectID uint      `json:"project_id"` // Foreign Key
	CreatedAt time.Time `json:"created_at"`
}
