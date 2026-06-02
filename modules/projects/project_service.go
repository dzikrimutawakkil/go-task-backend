package projects

import (
	"errors"
	"gotask-backend/internal/interfaces"
	"gotask-backend/modules/organizations"
	"gotask-backend/modules/tasks"
	"gotask-backend/utils"
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
	authService interfaces.AuthService
}

// M5: Subscription Tiers — added authService for quota checks.
func NewProjectService(repo ProjectRepository, taskService tasks.TaskService, orgRepo organizations.OrganizationRepository, authS interfaces.AuthService) ProjectService {
	return &projectService{repo: repo, taskService: taskService, orgRepo: orgRepo, authService: authS}
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
	StatusID    *string
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
	if input.StatusID != nil {
		if err := s.repo.SetProjectStatus(project.ID, *input.StatusID); err != nil {
			return nil, err
		}
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
	return s.repo.FindByID(id)
}

func (s *projectService) CreateProject(input CreateProjectInput, userID uint) (*Project, error) {
	// M5: Quota check — check project limit based on user's tier
	effectiveTier := "free"
	limits := utils.GetTierLimits("free")

	if user, err := s.authService.FindByID(userID); err == nil {
		effectiveTier = utils.GetEffectiveTier(user.Tier, user.TierExpiresAt)
		limits = utils.GetTierLimits(effectiveTier)
	}

	// Check project limit per org
	if limits.MaxProjects != -1 {
		count, err := s.repo.CountByOrg(strconv.FormatUint(uint64(input.OrganizationID), 10))
		if err != nil {
			return nil, err
		}
		if count >= limits.MaxProjects {
			return nil, utils.ErrQuotaExceeded("project", limits.MaxProjects, effectiveTier)
		}
	}

	// Q19: Auto-assign default "Active" status
	defaultStatusID, err := s.repo.GetDefaultStatusID()
	if err != nil {
		return nil, errors.New("failed to get default project status")
	}

	// Create the project first
	project := Project{
		Name:           input.Name,
		Description:    input.Description,
		OrganizationID: input.OrganizationID,
	}

	if err := s.repo.Create(&project); err != nil {
		return nil, err
	}

	// Set status_id
	if err := s.repo.SetProjectStatus(project.ID, defaultStatusID); err != nil {
		return nil, err
	}

	// Q18: Create default task statuses and labels
	if err := s.taskService.CreateDefaultStatuses(project.ID); err != nil {
		return nil, err
	}
	if err := s.taskService.CreateDefaultLabels(project.ID); err != nil {
		return nil, err
	}

	// Re-fetch to get ProjectStatus populated
	return s.repo.FindByID(strconv.FormatUint(uint64(project.ID), 10))
}

func (s *projectService) DeleteProject(id string, orgID string, requesterID uint) error {
	project, err := s.repo.FindByIDAndOrg(id, orgID)
	if err != nil {
		return errors.New("project not found or access denied")
	}

	orgIDUint, _ := strconv.ParseUint(orgID, 10, 64)
	requesterRole, err := s.orgRepo.GetMemberRole(requesterID, uint(orgIDUint))
	if err != nil {
		return errors.New("you are not a member of this organization")
	}

	if !requesterRole.CanDeleteProject() {
		return errors.New("insufficient permission to delete project")
	}

	if err := s.repo.ClearTaskAssignees(project.ID); err != nil {
		return err
	}
	if err := s.repo.DeleteTasksByProject(project.ID); err != nil {
		return err
	}

	return s.repo.Delete(project)
}
