package models

import (
	"time"
)

type Organization struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"unique" json:"name"`
	OwnerID   uint      `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Priority struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `gorm:"unique" json:"name"`
	Level int    `json:"level"`
	Color string `json:"color"`
}

type Status struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique" json:"name"`
	Slug string `gorm:"unique" json:"slug"`
}

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"unique" json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Project struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	OrganizationID uint      `json:"organization_id"`
	CreatedAt      time.Time `json:"created_at"`
}

type Task struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Title string `json:"title"`

	StatusID uint   `json:"status_id"`
	Status   Status `gorm:"foreignKey:StatusID" json:"status"` // Status & Priority masih aman karena dianggap "Shared Value Objects" atau "Enums"

	PriorityID uint     `json:"priority_id"`
	Priority   Priority `gorm:"foreignKey:PriorityID" json:"priority"`

	ProjectID uint `json:"project_id"`

	AssigneeIDs []uint `json:"assignee_ids" gorm:"-"`

	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	CreatedAt time.Time  `json:"created_at"`
}
