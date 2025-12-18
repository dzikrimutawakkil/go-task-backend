package projects

import (
	"gotask-backend/models"
	"gotask-backend/modules/tasks"

	"gorm.io/gorm"
)

type ProjectRepository interface {
	FindAllByOrg(orgID string) ([]Project, error)
	FindByIDAndOrg(id string, orgID string) (*Project, error)
	Create(project *Project) error
	Delete(project *Project) error

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
func (r *projectRepository) FindAllByOrg(orgID string) ([]Project, error) {
	var projects []Project
	err := r.db.
		Scopes(models.ByOrg(orgID)).
		Find(&projects).Error
	return projects, err
}

// Find a specific project
func (r *projectRepository) FindByIDAndOrg(id string, orgID string) (*Project, error) {
	var project Project
	err := r.db.
		Where("id = ? AND organization_id = ?", id, orgID).
		First(&project).Error

	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) Create(project *Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) Delete(project *Project) error {
	return r.db.Delete(project).Error
}

func (r *projectRepository) ClearTaskAssignees(projectID uint) error {
	return r.db.Exec("DELETE FROM task_users WHERE task_id IN (SELECT id FROM tasks WHERE project_id = ?)", projectID).Error
}

func (r *projectRepository) DeleteTasksByProject(projectID uint) error {
	return r.db.Where("project_id = ?", projectID).Delete(&tasks.Task{}).Error
}
