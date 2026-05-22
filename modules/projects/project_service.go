package projects

import (
	"errors"
	"gotask-backend/modules/organizations"
	"gotask-backend/modules/tasks"
	"strconv"
)

type ProjectService interface {
	GetProjects(orgID string) ([]Project, error)
	GetProject(id string) (*Project, error)
	CreateProject(input CreateProjectInput, userID uint) (*Project, error)
	UpdateProject(id string, input UpdateProjectInput) (*Project, error)
	DeleteProject(id string, orgID string, requesterID uint) error
}

type projectService struct {
	repo        ProjectRepository
	taskService tasks.TaskService
	orgRepo     organizations.OrganizationRepository
}

func NewProjectService(repo ProjectRepository, taskService tasks.TaskService, orgRepo organizations.OrganizationRepository) ProjectService {
	return &projectService{repo, taskService, orgRepo}
}

// Input DTOs
type CreateProjectInput struct {
	Name           string
	Description    string
	OrganizationID uint
}

type UpdateProjectInput struct {
	Name        *string
	Description *string
	Status      *string
	Priority    *string
	Budget      *float64
	Deadline    *string
	Progress    *int
}

func (s *projectService) GetProjects(orgID string) ([]Project, error) {
	return s.repo.FindAllByOrg(orgID)
}

func (s *projectService) GetProject(id string) (*Project, error) {
	return s.repo.FindByID(id)
}

func (s *projectService) UpdateProject(id string, input UpdateProjectInput) (*Project, error) {
	project, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("project not found")
	}

	if input.Name != nil {
		project.Name = *input.Name
	}
	if input.Description != nil {
		project.Description = *input.Description
	}
	if input.Status != nil {
		project.Status = *input.Status
	}
	if input.Priority != nil {
		project.Priority = *input.Priority
	}
	if input.Budget != nil {
		project.Budget = input.Budget
	}
	if input.Deadline != nil {
		project.Deadline = input.Deadline
	}
	if input.Progress != nil {
		project.Progress = int64(*input.Progress)
	}

	if err := s.repo.Update(project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *projectService) CreateProject(input CreateProjectInput, userID uint) (*Project, error) {
	project := Project{
		Name:           input.Name,
		Description:    input.Description,
		OrganizationID: input.OrganizationID,
	}

	if err := s.repo.Create(&project); err != nil {
		return nil, err
	}

	if err := s.taskService.CreateDefaultStatuses(project.ID); err != nil {
		return nil, err
	}

	// Re-fetch to populate relations (optional)
	return &project, nil
}

func (s *projectService) DeleteProject(id string, orgID string, requesterID uint) error {
	// 1. Security: Find Project AND ensure it belongs to the Context Org
	project, err := s.repo.FindByIDAndOrg(id, orgID)
	if err != nil {
		return errors.New("project not found or access denied")
	}

	// 2. RBAC: Check permission
	orgIDUint, _ := strconv.ParseUint(orgID, 10, 64)
	requesterRole, err := s.orgRepo.GetMemberRole(requesterID, uint(orgIDUint))
	if err != nil {
		return errors.New("you are not a member of this organization")
	}

	if !requesterRole.CanDeleteProject() {
		return errors.New("insufficient permission to delete project")
	}

	// 3. Cleanup
	if err := s.repo.ClearTaskAssignees(project.ID); err != nil {
		return err
	}
	if err := s.repo.DeleteTasksByProject(project.ID); err != nil {
		return err
	}

	// 3. Delete
	return s.repo.Delete(project)
}
