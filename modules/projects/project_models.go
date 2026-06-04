package projects

import (
	"gotask-backend/modules/clients"
	"time"
)

// Project represents a project within a workspace.
// M-MIGRATION: Renamed organization_id to workspace_id, added client_id
type Project struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	WorkspaceID   uint            `json:"workspace_id"`
	ClientID      *uint           `json:"client_id,omitempty" gorm:"index"`
	Client        *clients.Client `json:"client,omitempty" gorm:"foreignKey:ClientID"`
	Priority      string          `gorm:"default:medium" json:"priority"`
	Progress      int64           `gorm:"default:0" json:"progress"`
	Budget        *float64        `json:"budget"`
	Deadline      *string         `json:"deadline"`
	Version       int             `gorm:"default:1" json:"version"`
	StatusID      *int            `gorm:"column:status_id" json:"status_id"`
	ProjectStatus *ProjectStatus  `gorm:"-" json:"project_status,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// ProjectStatus represents the status of a project (Active, On Hold, Completed, Archived).
// Q19: Project Status Workflow
type ProjectStatus struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}
