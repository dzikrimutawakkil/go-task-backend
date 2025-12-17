package tasks

import (
	"errors"
	"gotask-backend/models"
	"strconv"
	"time"
)

type TaskService interface {
	CreateTask(input CreateTaskInput) (*models.Task, error)
	GetTasksByProject(projectID string, orgID string) ([]models.Task, error)
	UpdateTask(id string, input UpdateTaskInput) (*models.Task, error)
	DeleteTask(id string) error
}

type taskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) TaskService {
	return &taskService{repo}
}

type CreateTaskInput struct {
	Title      string
	ProjectID  uint
	StatusID   uint
	PriorityID uint
	StartDate  *time.Time
	EndDate    *time.Time
}

type UpdateTaskInput struct {
	Title       *string
	StatusID    *uint
	PriorityID  *uint
	AssigneeIDs []uint
	StartDate   *time.Time
	EndDate     *time.Time
}

func (s *taskService) CreateTask(input CreateTaskInput) (*models.Task, error) {
	// 1. Set Defaults (Business Logic)
	if input.StatusID == 0 {
		input.StatusID = 1 // Assuming ID 1 is "Todo"
	}
	if input.PriorityID == 0 {
		input.PriorityID = 2 // Assuming ID 2 is "Medium"
	}

	task := models.Task{
		Title:      input.Title,
		ProjectID:  input.ProjectID,
		StatusID:   input.StatusID,
		PriorityID: input.PriorityID,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
	}

	if err := s.repo.Create(&task); err != nil {
		return nil, err
	}

	// Return fully loaded task
	return s.repo.FindByID(interfaceToString(task.ID))
}

func (s *taskService) GetTasksByProject(projectID string, orgID string) ([]models.Task, error) {
	// Call Repo to check Security
	hasAccess, err := s.repo.CheckProjectAccess(projectID, orgID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.New("project not found or access denied")
	}

	// Fetch Tasks (Safe now)
	return s.repo.FindByProjectID(projectID)
}

func (s *taskService) UpdateTask(id string, input UpdateTaskInput) (*models.Task, error) {
	task, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("task not found")
	}

	updates := make(map[string]interface{})
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.StatusID != nil {
		updates["status_id"] = *input.StatusID
	}
	if input.PriorityID != nil {
		updates["priority_id"] = *input.PriorityID
	}
	if input.StartDate != nil {
		updates["start_date"] = *input.StartDate
	}
	if input.EndDate != nil {
		updates["end_date"] = *input.EndDate
	}

	if len(updates) > 0 {
		if err := s.repo.Update(task, updates); err != nil {
			return nil, err
		}
	}

	// Handle Assignees Sync
	if input.AssigneeIDs != nil {
		users, _ := s.repo.FindUsersByIDs(input.AssigneeIDs)
		// Clear old and set new (Replace)
		s.repo.ClearAssignees(task)
		s.repo.AssignUsers(task, users)
	}

	return s.repo.FindByID(id)
}

func (s *taskService) DeleteTask(id string) error {
	task, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("task not found")
	}
	return s.repo.Delete(task)
}

// Helper function to convert uint ID to string
func interfaceToString(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
