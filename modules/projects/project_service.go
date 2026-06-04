package projects

import (
	"errors"
	"gotask-backend/internal/interfaces"
	"gotask-backend/modules/clients"
	"gotask-backend/modules/tasks"
	"gotask-backend/modules/workspaces"
	"gotask-backend/utils"
	"strconv"
)

// M-MIGRATION: Updated to use workspace-based tier
type ProjectService interface {
	GetProjects(workspaceID string) ([]Project, error)
	GetProject(id string) (*Project, error)
	CreateProject(input CreateProjectInput, userID uint) (*Project, error)
	UpdateProject(id string, input UpdateProjectInput, workspaceID string) (*Project, error)
	DeleteProject(id string, workspaceID string, requesterID uint) error
}

type projectService struct {
	repo        ProjectRepository
	taskService tasks.TaskService
	wsRepo      workspaces.WorkspaceRepository
	authService interfaces.AuthService
}

// M-MIGRATION: Updated to use workspace-based tier for quota checks
func NewProjectService(repo ProjectRepository, taskService tasks.TaskService, wsRepo workspaces.WorkspaceRepository, authS interfaces.AuthService) ProjectService {
	return &projectService{repo: repo, taskService: taskService, wsRepo: wsRepo, authService: authS}
}

// Input DTOs
type CreateProjectInput struct {
	Name        string
	Description string
	WorkspaceID uint
	ClientID    *uint
	NewClient   *InlineCreateClientRequest
}

type InlineCreateClientRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email"`
	WhatsApp string `json:"whatsapp"`
	Phone    string `json:"phone"`
	Company  string `json:"company"`
}

type UpdateProjectInput struct {
	Name        *string
	Description *string
	StatusID    *int
	Priority    *string
	Budget      *float64
	Deadline    *string
	Progress    *int
	ClientID    *uint
}

func (s *projectService) GetProjects(workspaceID string) ([]Project, error) {
	return s.repo.FindAllByWorkspace(workspaceID)
}

func (s *projectService) GetProject(id string) (*Project, error) {
	return s.repo.FindByID(id)
}

func (s *projectService) UpdateProject(id string, input UpdateProjectInput, workspaceID string) (*Project, error) {
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
	if input.ClientID != nil {
		if *input.ClientID == 0 {
			// Unlink client
			project.ClientID = nil
			project.Client = nil
		} else {
			// Validate client belongs to workspace
			client, err := s.repo.FindClientByID(*input.ClientID)
			if err != nil || client.WorkspaceID != project.WorkspaceID {
				return nil, errors.New("client not found or does not belong to this workspace")
			}
			project.ClientID = input.ClientID
		}
	}

	if err := s.repo.Update(project); err != nil {
		return nil, err
	}
	return s.repo.FindByID(strconv.FormatUint(uint64(project.ID), 10))
}

func (s *projectService) CreateProject(input CreateProjectInput, userID uint) (*Project, error) {
	// M-MIGRATION: Quota check — check project limit based on workspace's tier
	effectiveTier := "free"
	limits := utils.GetTierLimits("free")

	wsInfo, err := s.wsRepo.FindWorkspaceInfoByID(input.WorkspaceID)
	if err == nil {
		effectiveTier = utils.GetEffectiveTier(wsInfo.Tier, wsInfo.TierExpiresAt)
		limits = utils.GetTierLimits(effectiveTier)
	}

	// Validate: cannot have both ClientID and NewClient
	if input.ClientID != nil && input.NewClient != nil {
		return nil, errors.New("cannot specify both client_id and new_client")
	}

	// Handle inline client creation
	var clientID *uint
	if input.NewClient != nil {
		// M-MIGRATION: Inline create client - need to create via client repository
		newClient := clients.Client{
			WorkspaceID: input.WorkspaceID,
			Name:        input.NewClient.Name,
		}
		if input.NewClient.Email != "" {
			newClient.Email = &input.NewClient.Email
		}
		if input.NewClient.WhatsApp != "" {
			newClient.WhatsApp = &input.NewClient.WhatsApp
		}
		if input.NewClient.Phone != "" {
			newClient.Phone = &input.NewClient.Phone
		}
		if input.NewClient.Company != "" {
			newClient.Company = &input.NewClient.Company
		}
		// Use workspace client quota check
		if limits.MaxClients != -1 {
			count, err := s.repo.CountClientsByWorkspace(input.WorkspaceID)
			if err != nil {
				return nil, err
			}
			if int(count) >= limits.MaxClients {
				return nil, utils.ErrQuotaExceeded("client", limits.MaxClients, effectiveTier)
			}
		}
		if err := s.repo.CreateClient(&newClient); err != nil {
			return nil, errors.New("failed to create client")
		}
		clientID = &newClient.ID
	} else {
		clientID = input.ClientID
	}

	// Validate client_id belongs to workspace
	if clientID != nil {
		client, err := s.repo.FindClientByID(*clientID)
		if err != nil || client.WorkspaceID != input.WorkspaceID {
			return nil, errors.New("client not found or does not belong to this workspace")
		}
	}

	// Check project limit per workspace
	if limits.MaxProjects != -1 {
		count, err := s.repo.CountByWorkspace(strconv.FormatUint(uint64(input.WorkspaceID), 10))
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
		Name:        input.Name,
		Description: input.Description,
		WorkspaceID: input.WorkspaceID,
		ClientID:    clientID,
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

func (s *projectService) DeleteProject(id string, workspaceID string, requesterID uint) error {
	project, err := s.repo.FindByIDAndWorkspace(id, workspaceID)
	if err != nil {
		return errors.New("project not found or access denied")
	}

	workspaceIDUint, _ := strconv.ParseUint(workspaceID, 10, 64)
	requesterRole, err := s.wsRepo.GetMemberRole(requesterID, uint(workspaceIDUint))
	if err != nil {
		return errors.New("you are not a member of this workspace")
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
