package projects

import (
	"gotask-backend/models"

	"gorm.io/gorm"
)

type ProjectRepository interface {
	FindAllByOrg(orgID string) ([]models.Project, error)
	FindByIDAndOrg(id string, orgID string) (*models.Project, error)
	Create(project *models.Project) error
	Delete(project *models.Project) error

	// Task cleanup helpers
	DeleteTasksByProject(projectID uint) error
	ClearTaskAssignees(projectID uint) error
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db}
}

// Fetch all projects in the Organization
func (r *projectRepository) FindAllByOrg(orgID string) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.
		Preload("Tasks.Status").
		Preload("Tasks.Priority").
		Scopes(models.ByOrg(orgID)).
		Find(&projects).Error
	return projects, err
}

// Find a specific project
func (r *projectRepository) FindByIDAndOrg(id string, orgID string) (*models.Project, error) {
	var project models.Project
	err := r.db.
		Where("id = ? AND organization_id = ?", id, orgID).
		First(&project).Error

	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) Create(project *models.Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) Delete(project *models.Project) error {
	return r.db.Delete(project).Error
}

func (r *projectRepository) ClearTaskAssignees(projectID uint) error {
	return r.db.Exec("DELETE FROM task_users WHERE task_id IN (SELECT id FROM tasks WHERE project_id = ?)", projectID).Error
}

func (r *projectRepository) DeleteTasksByProject(projectID uint) error {
	return r.db.Where("project_id = ?", projectID).Delete(&models.Task{}).Error
}
