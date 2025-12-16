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

	// Assignee Management
	ClearAssignees(task *models.Task) error
	AssignUsers(task *models.Task, users []models.User) error

	// Helper to find users by IDs (for assignment)
	FindUsersByIDs(ids []uint) ([]models.User, error)
	// Helper to find users by Emails (for bulk assign)
	FindUsersByEmails(emails []string) ([]models.User, error)
	CheckProjectAccess(projectID string, orgID string) (bool, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) TaskRepository {
	return &repository{db}
}

func (r *repository) Create(task *models.Task) error {
	return r.db.Create(task).Error
}

func (r *repository) FindByID(id string) (*models.Task, error) {
	var task models.Task
	err := r.db.Preload("Status").
		Preload("Priority").
		Preload("Assignees").
		First(&task, id).Error
	return &task, err
}

func (r *repository) FindByProjectID(projectID string) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Preload("Status").
		Preload("Priority").
		Preload("Assignees").
		Where("project_id = ?", projectID).
		Find(&tasks).Error
	return tasks, err
}

func (r *repository) Update(task *models.Task, updates map[string]interface{}) error {
	return r.db.Model(task).Updates(updates).Error
}

func (r *repository) Delete(task *models.Task) error {
	return r.db.Delete(task).Error
}

func (r *repository) ClearAssignees(task *models.Task) error {
	return r.db.Model(task).Association("Assignees").Clear()
}

func (r *repository) AssignUsers(task *models.Task, users []models.User) error {
	return r.db.Model(task).Association("Assignees").Append(users) // Append or Replace based on need
}

func (r *repository) FindUsersByIDs(ids []uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Find(&users, ids).Error
	return users, err
}

func (r *repository) FindUsersByEmails(emails []string) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("email IN ?", emails).Find(&users).Error
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
