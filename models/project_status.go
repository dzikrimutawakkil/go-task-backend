package models

import (
	"time"
)

// ProjectStatus represents a status for a project (e.g., Active, On Hold, Completed, Archived).
type ProjectStatus struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Color     string    `gorm:"size:7;default:'#6B7280'" json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for ProjectStatus.
func (ProjectStatus) TableName() string {
	return "project_statuses"
}

// Project status constants
const (
	ProjectStatusActive    = "Active"
	ProjectStatusOnHold    = "On Hold"
	ProjectStatusCompleted = "Completed"
	ProjectStatusArchived  = "Archived"
)
