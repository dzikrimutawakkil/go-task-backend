package projects

import "time"

type Project struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	OrganizationID uint      `json:"organization_id"`
	CreatedAt      time.Time `json:"created_at"`
}
