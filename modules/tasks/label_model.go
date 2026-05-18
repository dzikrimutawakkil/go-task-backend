package tasks

import (
	"time"
)

// Label represents a tag/label for categorizing tasks within a project.
type Label struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"index" json:"project_id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Color     string    `gorm:"size:7;default:'#6366F1'" json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskLabel is the join table for many-to-many relationship between Task and Label.
type TaskLabel struct {
	TaskID  uint `gorm:"primaryKey"`
	LabelID uint `gorm:"primaryKey"`
}

// TaskWithLabels embeds Task with an additional Labels field.
type TaskWithLabels struct {
	Task
	Labels []Label `json:"labels" gorm:"-"`
}
