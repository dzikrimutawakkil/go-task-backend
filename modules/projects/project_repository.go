package projects

import (
	"gotask-backend/models"
	"gotask-backend/modules/clients"
	"gotask-backend/modules/tasks"

	"gorm.io/gorm"
)

// M-MIGRATION: Updated interface to use workspace instead of organization
type ProjectRepository interface {
	FindAllByWorkspace(workspaceID string) ([]Project, error)
	FindByIDAndWorkspace(id string, workspaceID string) (*Project, error)
	FindByID(id string) (*Project, error)
	Create(project *Project) error
	Update(project *Project) error
	Delete(project *Project) error

	// Client helpers for inline create
	FindClientByID(clientID uint) (*clients.Client, error)
	CreateClient(client *clients.Client) error
	CountClientsByWorkspace(workspaceID uint) (int64, error)

	// Q19: Project status helper
	GetDefaultStatusID() (int, error)
	SetProjectStatus(projectID uint, statusID int) error

	// Task cleanup helpers
	DeleteTasksByProject(projectID uint) error
	ClearTaskAssignees(projectID uint) error

	// M5: Quota check helpers
	CountByWorkspace(workspaceID string) (int, error)
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db}
}

// FindAllByWorkspace fetches all projects in the workspace with their status and client.
// M-MIGRATION: Renamed from FindAllByOrg, updated to use workspace_id and preload client
func (r *projectRepository) FindAllByWorkspace(workspaceID string) ([]Project, error) {
	var projects []Project
	err := r.db.
		Preload("Client").
		Scopes(models.ByWorkspace(workspaceID)).
		Find(&projects).Error
	if err != nil {
		return nil, err
	}

	// Populate embedded ProjectStatus field from project_statuses table
	for i := range projects {
		if projects[i].StatusID != nil {
			var statusData struct {
				ID    int
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

// FindByIDAndWorkspace finds a specific project by ID and workspace.
// M-MIGRATION: Renamed from FindByIDAndOrg
func (r *projectRepository) FindByIDAndWorkspace(id string, workspaceID string) (*Project, error) {
	var project Project
	err := r.db.
		Preload("Client").
		Where("id = ? AND workspace_id = ?", id, workspaceID).
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
	if err := r.db.Preload("Client").First(&project, id).Error; err != nil {
		return nil, err
	}

	// Populate embedded ProjectStatus field
	if project.StatusID != nil {
		var statusData struct {
			ID    int
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

// FindClientByID finds a client by ID for validation.
func (r *projectRepository) FindClientByID(clientID uint) (*clients.Client, error) {
	var client clients.Client
	err := r.db.First(&client, clientID).Error
	return &client, err
}

// CreateClient creates a new client (for inline client creation in project).
func (r *projectRepository) CreateClient(client *clients.Client) error {
	return r.db.Create(client).Error
}

// CountClientsByWorkspace counts the number of clients in a workspace.
func (r *projectRepository) CountClientsByWorkspace(workspaceID uint) (int64, error) {
	var count int64
	err := r.db.Model(&clients.Client{}).Where("workspace_id = ?", workspaceID).Count(&count).Error
	return count, err
}

func (r *projectRepository) ClearTaskAssignees(projectID uint) error {
	return r.db.Exec("DELETE FROM task_users WHERE task_id IN (SELECT id FROM tasks WHERE project_id = ?)", projectID).Error
}

func (r *projectRepository) DeleteTasksByProject(projectID uint) error {
	return r.db.Where("project_id = ?", projectID).Delete(&tasks.Task{}).Error
}

// GetDefaultStatusID retrieves the default project status ID (Active) as int
// Q19: Project Status Workflow
func (r *projectRepository) GetDefaultStatusID() (int, error) {
	var result struct {
		ID int
	}
	err := r.db.Table("project_statuses").
		Select("id").
		Where("name = 'Active'").
		First(&result).Error
	return result.ID, err
}

// SetProjectStatus sets the status_id for a project
func (r *projectRepository) SetProjectStatus(projectID uint, statusID int) error {
	return r.db.Exec("UPDATE projects SET status_id = ? WHERE id = ?", statusID, projectID).Error
}

// CountByWorkspace returns the number of projects in a workspace.
// M5: Subscription Tiers — Phase 5: Service Layer — Quota check for project limit.
// M-MIGRATION: Renamed from CountByOrg
func (r *projectRepository) CountByWorkspace(workspaceID string) (int, error) {
	var count int64
	err := r.db.Model(&Project{}).Where("workspace_id = ?", workspaceID).Count(&count).Error
	return int(count), err
}
