package projects

import (
	"errors"
	"gotask-backend/models"
)

type ProjectService interface {
	GetProjects(orgID string) ([]models.Project, error)
	CreateProject(input CreateProjectInput, userID uint) (*models.Project, error)
	DeleteProject(id string, orgID string) error
}

type projectService struct {
	repo ProjectRepository
}

func NewProjectService(repo ProjectRepository) ProjectService {
	return &projectService{repo}
}

// Input DTO
type CreateProjectInput struct {
	Name           string
	Description    string
	OrganizationID uint
}

func (s *projectService) GetProjects(orgID string) ([]models.Project, error) {
	return s.repo.FindAllByOrg(orgID)
}

func (s *projectService) CreateProject(input CreateProjectInput, userID uint) (*models.Project, error) {
	project := models.Project{
		Name:           input.Name,
		Description:    input.Description,
		OrganizationID: input.OrganizationID,
		// We add the creator to 'Users' just for record-keeping (e.g. "Project Lead")
		// But it does not control access anymore.
		Users: []models.User{{ID: userID}},
	}

	if err := s.repo.Create(&project); err != nil {
		return nil, err
	}

	// Re-fetch to populate relations (optional)
	return &project, nil
}

func (s *projectService) DeleteProject(id string, orgID string) error {
	// 1. Security: Find Project AND ensure it belongs to the Context Org
	project, err := s.repo.FindByIDAndOrg(id, orgID)
	if err != nil {
		return errors.New("project not found or access denied")
	}

	// 2. Cleanup
	if err := s.repo.ClearTaskAssignees(project.ID); err != nil {
		return err
	}
	if err := s.repo.DeleteTasksByProject(project.ID); err != nil {
		return err
	}

	// 3. Delete
	return s.repo.Delete(project)
}
