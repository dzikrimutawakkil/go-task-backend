package tasks

import (
	"gotask-backend/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 1. Define the Mock Repository
// This struct "pretends" to be your database
type MockTaskRepo struct {
	mock.Mock
}

func (m *MockTaskRepo) Create(task *models.Task) error {
	args := m.Called(task)
	task.ID = 1 // Simulate DB setting ID
	return args.Error(0)
}

func (m *MockTaskRepo) FindByID(id string) (*models.Task, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepo) FindByProjectID(projectID string) ([]models.Task, error) {
	args := m.Called(projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskRepo) Update(task *models.Task, updates map[string]interface{}) error {
	args := m.Called(task, updates)
	return args.Error(0)
}

func (m *MockTaskRepo) Delete(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepo) ClearAssignees(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepo) AssignUsers(task *models.Task, users []models.User) error {
	args := m.Called(task, users)
	return args.Error(0)
}

func (m *MockTaskRepo) FindUsersByIDs(ids []uint) ([]models.User, error) {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockTaskRepo) CheckProjectAccess(projectID string, orgID string) (bool, error) {
	args := m.Called(projectID, orgID)
	return args.Bool(0), args.Error(1)
}

// --- THE TESTS ---

func TestCreateTask_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockTaskRepo)
	service := NewTaskService(mockRepo)

	// Expectation: "Create" will be called with ANY task, and should return nil (no error)
	mockRepo.On("Create", mock.Anything).Return(nil)
	// Expectation: "FindByID" will be called immediately after (to return the full object)
	mockRepo.On("FindByID", "1").Return(&models.Task{Title: "New Task", ID: 1}, nil)

	// Execution
	input := CreateTaskInput{
		Title:     "New Task",
		ProjectID: 10,
	}
	result, err := service.CreateTask(input)

	// Validation
	assert.NoError(t, err)
	assert.Equal(t, "New Task", result.Title)
	assert.Equal(t, uint(1), result.ID)

	// Ensure defaults were set (Business Logic Check)
	// We didn't send Priority, so it should be 2 (Medium)
	// Note: In a real mock, you'd inspect the arguments passed to Create to verify this.
}

func TestGetTasksByProject_AccessDenied(t *testing.T) {
	// Setup
	mockRepo := new(MockTaskRepo)
	service := NewTaskService(mockRepo)

	// Expectation: CheckAccess returns FALSE
	mockRepo.On("CheckProjectAccess", "10", "5").Return(false, nil)

	// Execution
	tasks, err := service.GetTasksByProject("10", "5")

	// Validation
	assert.Error(t, err)
	assert.Nil(t, tasks)
	assert.Equal(t, "project not found or access denied", err.Error())
}

func TestUpdateTask_AssignUsers(t *testing.T) {
	// Setup
	mockRepo := new(MockTaskRepo)
	service := NewTaskService(mockRepo)

	// 1. Mock Data
	existingTask := &models.Task{ID: 1, Title: "Old Title"}
	usersToAssign := []models.User{{ID: 5, Email: "dev@test.com"}}

	// 2. Expectations
	// First, it finds the task
	mockRepo.On("FindByID", "1").Return(existingTask, nil)

	// Then it looks up the new assignees
	mockRepo.On("FindUsersByIDs", []uint{5}).Return(usersToAssign, nil)

	// Then it clears old assignees
	mockRepo.On("ClearAssignees", existingTask).Return(nil)

	// Then it assigns new ones
	mockRepo.On("AssignUsers", existingTask, usersToAssign).Return(nil)

	// Finally it updates the title (and any other fields)
	// We use mock.Anything for the map to save time, or you can match the specific map
	mockRepo.On("Update", existingTask, mock.Anything).Return(nil)

	// 3. Execution
	newTitle := "New Title"
	input := UpdateTaskInput{
		Title:       &newTitle,
		AssigneeIDs: []uint{5},
	}
	result, err := service.UpdateTask("1", input)

	// 4. Validation
	assert.NoError(t, err)
	assert.Equal(t, uint(1), result.ID)

	// Verify that all our mocks were actually called
	mockRepo.AssertExpectations(t)
}
