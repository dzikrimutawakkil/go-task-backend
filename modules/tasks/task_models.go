package tasks

import "time"

type Status struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique" json:"name"`
	Slug string `gorm:"unique" json:"slug"`
}

type Priority struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `gorm:"unique" json:"name"`
	Level int    `json:"level"`
	Color string `json:"color"`
}

type Task struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Title       string     `json:"title"`
	StatusID    uint       `json:"status_id"`
	Status      Status     `gorm:"foreignKey:StatusID" json:"status"`
	PriorityID  uint       `json:"priority_id"`
	Priority    Priority   `gorm:"foreignKey:PriorityID" json:"priority"`
	ProjectID   uint       `json:"project_id"`
	AssigneeIDs []uint     `json:"assignee_ids" gorm:"-"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CreatedAt   time.Time  `json:"created_at"`
}
