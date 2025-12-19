package tasks

import (
	"errors"
	"gotask-backend/modules/auth"
	"strconv"
	"time"
)

type TaskService interface {
	CreateTask(input CreateTaskInput) (*Task, error)
	GetTasksByProject(projectID string, orgID string, page int, limit int) ([]Task, int64, error)
	UpdateTask(id string, input UpdateTaskInput) (*Task, error)
	DeleteTask(id string) error

	CreateDefaultStatuses(projectID uint) error
	GetStatuses(projectID string) ([]Status, error)
	CreateNewStatus(projectID uint, name string, index int) (*Status, error)
	UpdateStatus(id string, name *string, index *int) (*Status, error)
	DeleteStatus(id string) error
}

type taskService struct {
	repo        TaskRepository
	authService auth.AuthService
}

func NewTaskService(repo TaskRepository, authS auth.AuthService) TaskService {
	return &taskService{
		repo:        repo,
		authService: authS,
	}
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

func (s *taskService) CreateTask(input CreateTaskInput) (*Task, error) {
	// 1. Set Defaults (Business Logic)
	if input.StatusID == 0 {
		input.StatusID = 1 // Assuming ID 1 is "Todo"
	}
	if input.PriorityID == 0 {
		input.PriorityID = 2 // Assuming ID 2 is "Medium"
	}

	task := Task{
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

func (s *taskService) GetTasksByProject(projectID string, orgID string, page int, limit int) ([]Task, int64, error) {
	// Call Repo to check Security
	hasAccess, err := s.repo.CheckProjectAccess(projectID, orgID)
	if err != nil {
		return nil, 0, err
	}
	if !hasAccess {
		return nil, 0, errors.New("project not found or access denied")
	}

	// Fetch Tasks (Safe now)
	return s.repo.FindByProjectID(projectID, page, limit)
}

func (s *taskService) UpdateTask(id string, input UpdateTaskInput) (*Task, error) {
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
		// Cek apakah user-user ini valid dengan nanya ke Auth Service
		users, err := s.authService.GetUsersByIDs(input.AssigneeIDs)
		if err != nil {
			return nil, err
		}

		// Ambil ID-nya saja untuk disimpan di tabel task_users
		var validIDs []uint
		for _, u := range users {
			validIDs = append(validIDs, u.ID)
		}

		s.repo.ClearAssignees(task)
		s.repo.AssignUsers(task, validIDs)
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

func (s *taskService) CreateDefaultStatuses(projectID uint) error {
	defaults := []string{"Todo", "On Progress", "Done", "Pending", "Cancel"}

	for i, name := range defaults {
		status := Status{
			Name:      name,
			Index:     i,
			ProjectID: int(projectID),
		}
		if err := s.repo.CreateStatus(&status); err != nil {
			return err
		}
	}
	return nil
}

func (s *taskService) GetStatuses(projectID string) ([]Status, error) {
	return s.repo.GetStatusesByProjectID(projectID)
}

func (s *taskService) CreateNewStatus(projectID uint, name string, index int) (*Status, error) {
	getMaxIndex, err := s.repo.GetMaxIndex(strconv.Itoa(int(projectID)))
	if err != nil {
		return nil, err
	}

	status := Status{
		Name:      name,
		Index:     getMaxIndex + 1,
		ProjectID: int(projectID),
	}

	if err := s.repo.CreateStatus(&status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (s *taskService) UpdateStatus(id string, name *string, newIndexPtr *int) (*Status, error) {
	targetStatus, err := s.repo.FindStatusByID(id)
	if err != nil {
		return nil, errors.New("status not found")
	}

	if newIndexPtr != nil {
		newIndex := *newIndexPtr
		oldIndex := targetStatus.Index

		projectStatuses, err := s.repo.GetStatusesByProjectID(strconv.Itoa(targetStatus.ProjectID))
		if err != nil {
			return nil, errors.New("project statuses not found")
		}

		for i := range projectStatuses {
			if projectStatuses[i].ID == targetStatus.ID {
				projectStatuses[i].Index = newIndex
				if name != nil {
					projectStatuses[i].Name = *name
					targetStatus.Name = *name
				}
				targetStatus.Index = newIndex
				continue
			}

			// B. Logika Geser (Shifting)

			// KASUS 1: Pindah ke ATAS/KIRI (Misal: Index 4 -> 1)
			// Item yang ada di posisi 1, 2, 3 harus GESER KANAN (+1)
			if oldIndex > newIndex {
				if projectStatuses[i].Index >= newIndex && projectStatuses[i].Index < oldIndex {
					projectStatuses[i].Index += 1
				}
			}

			// KASUS 2: Pindah ke BAWAH/KANAN (Misal: Index 1 -> 4)
			// Item yang ada di posisi 2, 3, 4 harus GESER KIRI (-1)
			if oldIndex < newIndex {
				if projectStatuses[i].Index > oldIndex && projectStatuses[i].Index <= newIndex {
					projectStatuses[i].Index -= 1
				}
			}
		}
		if err := s.repo.BulkUpdateStatuses(projectStatuses); err != nil {
			return nil, err
		}

	} else {
		if name != nil {
			updates := map[string]interface{}{
				"name": name,
			}
			if err := s.repo.UpdateStatus(targetStatus, updates); err != nil {
				return nil, err
			}
			targetStatus.Name = *name
		} else {
			return targetStatus, nil // Nothing to update
		}
	}

	return targetStatus, nil
}

func (s *taskService) DeleteStatus(id string) error {
	status, err := s.repo.FindStatusByID(id)
	if err != nil {
		return errors.New("status not found")
	}
	return s.repo.DeleteStatus(status)
}
