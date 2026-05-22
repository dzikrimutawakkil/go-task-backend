package projects

import "time"

type Project struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	OrganizationID uint      `json:"organization_id"`
	Status         string    `gorm:"default:backlog" json:"status"`
	Priority       string    `gorm:"default:medium" json:"priority"`
	Progress       int64     `gorm:"default:0" json:"progress"`
	Budget         *float64  `json:"budget"`
	Deadline       *string   `json:"deadline"`
	Version        int       `gorm:"default:1" json:"version"`
	CreatedAt      time.Time `json:"created_at"`
}
