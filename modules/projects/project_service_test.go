package projects

import (
	"errors"
	"gotask-backend/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 1. Define the Mock Repository
type MockProjectRepo struct {
	mock.Mock
}

func (m *MockProjectRepo) FindAllByOrg(orgID string) ([]models.Project, error) {
	args := m.Called(orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectRepo) FindByIDAndOrg(id string, orgID string) (*models.Project, error) {
	args := m.Called(id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectRepo) Create(project *models.Project) error {
	args := m.Called(project)
	project.ID = 1 // Simulate DB ID assignment
	return args.Error(0)
}

func (m *MockProjectRepo) Delete(project *models.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockProjectRepo) DeleteTasksByProject(projectID uint) error {
	args := m.Called(projectID)
	return args.Error(0)
}

func (m *MockProjectRepo) ClearTaskAssignees(projectID uint) error {
	args := m.Called(projectID)
	return args.Error(0)
}

// 2. Test Cases

func TestCreateProject_Success(t *testing.T) {
	mockRepo := new(MockProjectRepo)
	service := NewProjectService(mockRepo)

	// Expectation: Repository.Create is called
	mockRepo.On("Create", mock.MatchedBy(func(p *models.Project) bool {
		return p.Name == "New App" && p.OrganizationID == 10
	})).Return(nil)

	input := CreateProjectInput{
		Name:           "New App",
		Description:    "Test Description",
		OrganizationID: 10,
	}

	// Logic: UserID is passed but logic doesn't explicitly rely on it in current service
	result, err := service.CreateProject(input, 1)

	assert.NoError(t, err)
	assert.Equal(t, "New App", result.Name)
	assert.Equal(t, uint(1), result.ID)
	mockRepo.AssertExpectations(t)
}

func TestGetProjects_Success(t *testing.T) {
	mockRepo := new(MockProjectRepo)
	service := NewProjectService(mockRepo)

	mockData := []models.Project{
		{ID: 1, Name: "Project A"},
		{ID: 2, Name: "Project B"},
	}

	mockRepo.On("FindAllByOrg", "10").Return(mockData, nil)

	projects, err := service.GetProjects("10")

	assert.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.Equal(t, "Project A", projects[0].Name)
}

func TestDeleteProject_Success(t *testing.T) {
	mockRepo := new(MockProjectRepo)
	service := NewProjectService(mockRepo)

	existingProject := &models.Project{ID: 5, Name: "To Delete", OrganizationID: 10}

	// 1. Security Check (Find Project)
	mockRepo.On("FindByIDAndOrg", "5", "10").Return(existingProject, nil)

	// 2. Cleanup 1: Clear Assignees
	mockRepo.On("ClearTaskAssignees", uint(5)).Return(nil)

	// 3. Cleanup 2: Delete Tasks
	mockRepo.On("DeleteTasksByProject", uint(5)).Return(nil)

	// 4. Delete Project
	mockRepo.On("Delete", existingProject).Return(nil)

	err := service.DeleteProject("5", "10")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteProject_AccessDenied(t *testing.T) {
	mockRepo := new(MockProjectRepo)
	service := NewProjectService(mockRepo)

	// Simulate "Project Not Found" or "Belongs to another Org"
	mockRepo.On("FindByIDAndOrg", "5", "999").Return(nil, errors.New("record not found"))

	err := service.DeleteProject("5", "999")

	assert.Error(t, err)
	assert.Equal(t, "project not found or access denied", err.Error())

	// Ensure delete logic was NOT triggered
	mockRepo.AssertNotCalled(t, "Delete")
	mockRepo.AssertNotCalled(t, "ClearTaskAssignees")
}
