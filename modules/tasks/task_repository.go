package tasks

import (
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(task *Task) error
	FindByID(id string) (*Task, error)
	FindByProjectID(projectID string) ([]Task, error)
	Update(task *Task, updates map[string]interface{}) error
	Delete(task *Task) error

	ClearAssignees(task *Task) error
	AssignUsers(task *Task, userIDs []uint) error

	CheckProjectAccess(projectID string, orgID string) (bool, error)
}

type repository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &repository{db}
}

// Helper internal untuk mengambil AssigneeIDs
func (r *repository) fetchAssigneeIDs(task *Task) error {
	var userIDs []uint
	// Query manual ke tabel penghubung
	err := r.db.Table("task_users").
		Where("task_id = ?", task.ID).
		Pluck("user_id", &userIDs).Error

	if err == nil {
		task.AssigneeIDs = userIDs
	}
	return err
}

func (r *repository) Create(task *Task) error {
	return r.db.Create(task).Error
}

func (r *repository) FindByID(id string) (*Task, error) {
	var task Task
	err := r.db.Preload("Status").
		Preload("Priority").
		First(&task, id).Error

	if err != nil {
		return nil, err
	}

	// Manual Fetch IDs
	_ = r.fetchAssigneeIDs(&task)
	return &task, nil
}

func (r *repository) FindByProjectID(projectID string) ([]Task, error) {
	var tasks []Task
	err := r.db.Preload("Status").
		Preload("Priority").
		Where("project_id = ?", projectID).
		Find(&tasks).Error

	if err != nil {
		return tasks, err
	}

	// Populate IDs for each task (Looping query is N+1 problem, but acceptable for MVP microservice separation)
	for i := range tasks {
		_ = r.fetchAssigneeIDs(&tasks[i])
	}
	return tasks, nil
}

func (r *repository) Update(task *Task, updates map[string]interface{}) error {
	return r.db.Model(task).Updates(updates).Error
}

func (r *repository) Delete(task *Task) error {
	return r.db.Delete(task).Error
}

func (r *repository) ClearAssignees(task *Task) error {
	// Manual Delete dari tabel penghubung
	return r.db.Exec("DELETE FROM task_users WHERE task_id = ?", task.ID).Error
}

func (r *repository) AssignUsers(task *Task, userIDs []uint) error {
	// Manual Insert ke tabel penghubung
	// Kita buat struct temporary atau insert map
	var records []map[string]interface{}
	for _, uid := range userIDs {
		records = append(records, map[string]interface{}{
			"task_id": task.ID,
			"user_id": uid,
		})
	}

	if len(records) > 0 {
		return r.db.Table("task_users").Create(&records).Error
	}
	return nil
}

func (r *repository) CheckProjectAccess(projectID string, orgID string) (bool, error) {
	var count int64
	err := r.db.Table("projects").
		Where("id = ? AND organization_id = ?", projectID, orgID).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
