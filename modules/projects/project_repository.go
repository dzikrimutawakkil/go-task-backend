package projects

import (
	"gotask-backend/models"
	"gotask-backend/modules/tasks"

	"gorm.io/gorm"
)

type ProjectRepository interface {
	FindAllByOrg(orgID string) ([]Project, error)
	FindByIDAndOrg(id string, orgID string) (*Project, error)
	FindByID(id string) (*Project, error)
	Create(project *Project) error
	Update(project *Project) error
	Delete(project *Project) error

	// Q19: Project status helper
	GetDefaultStatusID() (string, error)
	SetProjectStatus(projectID uint, statusID string) error

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

// Fetch all projects in the Organization with their status
func (r *projectRepository) FindAllByOrg(orgID string) ([]Project, error) {
	var projects []Project
	err := r.db.
		Scopes(models.ByOrg(orgID)).
		Find(&projects).Error
	if err != nil {
		return nil, err
	}

	// Populate embedded ProjectStatus field from project_statuses table
	for i := range projects {
		if projects[i].StatusID != nil {
			var statusData struct {
				ID    string
				Name  string
				Color string
			}
			err := r.db.Table("project_statuses").
				Select("id, name, color").
				Where("id = ?", *projects[i].StatusID).
				Scan(&statusData).Error
			if err == nil {
				projects[i].ProjectStatus = &ProjectStatus{
					ID:    statusData.ID,
					Name:  statusData.Name,
					Color: statusData.Color,
				}
			}
		}
	}
	return projects, nil
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

func (r *projectRepository) FindByID(id string) (*Project, error) {
	var project Project
	if err := r.db.First(&project, id).Error; err != nil {
		return nil, err
	}

	// Populate embedded ProjectStatus field
	if project.StatusID != nil {
		var statusData struct {
			ID    string
			Name  string
			Color string
		}
		err := r.db.Table("project_statuses").
			Select("id, name, color").
			Where("id = ?", *project.StatusID).
			Scan(&statusData).Error
		if err == nil {
			project.ProjectStatus = &ProjectStatus{
				ID:    statusData.ID,
				Name:  statusData.Name,
				Color: statusData.Color,
			}
		}
	}
	return &project, nil
}

func (r *projectRepository) Update(project *Project) error {
	return r.db.Save(project).Error
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

// GetDefaultStatusID retrieves the default project status ID (Active) as string
// Q19: Project Status Workflow
func (r *projectRepository) GetDefaultStatusID() (string, error) {
	var result struct {
		ID string
	}
	err := r.db.Table("project_statuses").
		Select("id").
		Where("name = 'Active'").
		First(&result).Error
	return result.ID, err
}

// SetProjectStatus sets the status_id for a project
func (r *projectRepository) SetProjectStatus(projectID uint, statusID string) error {
	return r.db.Exec("UPDATE projects SET status_id = ? WHERE id = ?", statusID, projectID).Error
}
