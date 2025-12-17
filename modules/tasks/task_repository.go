package tasks

import (
	"gotask-backend/models"

	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(task *models.Task) error
	FindByID(id string) (*models.Task, error)
	FindByProjectID(projectID string) ([]models.Task, error)
	Update(task *models.Task, updates map[string]interface{}) error
	Delete(task *models.Task) error

	ClearAssignees(task *models.Task) error
	AssignUsers(task *models.Task, userIDs []uint) error // Ubah parameter jadi []uint

	FindUsersByIDs(ids []uint) ([]models.User, error)
	CheckProjectAccess(projectID string, orgID string) (bool, error)
}

type repository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &repository{db}
}

// Helper internal untuk mengambil AssigneeIDs
func (r *repository) fetchAssigneeIDs(task *models.Task) error {
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

func (r *repository) Create(task *models.Task) error {
	return r.db.Create(task).Error
}

func (r *repository) FindByID(id string) (*models.Task, error) {
	var task models.Task
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

func (r *repository) FindByProjectID(projectID string) ([]models.Task, error) {
	var tasks []models.Task
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

func (r *repository) Update(task *models.Task, updates map[string]interface{}) error {
	return r.db.Model(task).Updates(updates).Error
}

func (r *repository) Delete(task *models.Task) error {
	return r.db.Delete(task).Error
}

func (r *repository) ClearAssignees(task *models.Task) error {
	// Manual Delete dari tabel penghubung
	return r.db.Exec("DELETE FROM task_users WHERE task_id = ?", task.ID).Error
}

func (r *repository) AssignUsers(task *models.Task, userIDs []uint) error {
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

func (r *repository) FindUsersByIDs(ids []uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Find(&users, ids).Error
	return users, err
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
