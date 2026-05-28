package projects

import "time"

// Project represents a project within an organization.
type Project struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	OrganizationID uint           `json:"organization_id"`
	Priority       string         `gorm:"default:medium" json:"priority"`
	Progress       int64          `gorm:"default:0" json:"progress"`
	Budget         *float64       `json:"budget"`
	Deadline       *string        `json:"deadline"`
	Version        int            `gorm:"default:1" json:"version"`
	StatusID       *string        `gorm:"column:status_id" json:"status_id"`
	ProjectStatus  *ProjectStatus `gorm:"-" json:"project_status,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

// ProjectStatus represents the status of a project (Active, On Hold, Completed, Archived).
// Q19: Project Status Workflow
type ProjectStatus struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}
