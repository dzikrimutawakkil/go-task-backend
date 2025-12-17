package organizations

import (
	"gotask-backend/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 1. Mock Repository
type MockOrgRepo struct {
	mock.Mock
}

func (m *MockOrgRepo) Create(org *models.Organization) error {
	args := m.Called(org)
	org.ID = 100
	return args.Error(0)
}
func (m *MockOrgRepo) FindByID(id uint) (*models.Organization, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}
func (m *MockOrgRepo) AddMember(orgID uint, userID uint) error {
	args := m.Called(orgID, userID)
	return args.Error(0)
}
func (m *MockOrgRepo) IsMember(userID uint, orgID uint) (bool, error) {
	args := m.Called(userID, orgID)
	return args.Bool(0), args.Error(1)
}
func (m *MockOrgRepo) FindUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockOrgRepo) FindMembers(orgID uint) ([]models.User, error) {
	args := m.Called(orgID)
	return args.Get(0).([]models.User), args.Error(1)
}

// 2. Test Cases

func TestCreateOrganization(t *testing.T) {
	mockRepo := new(MockOrgRepo)
	service := NewOrganizationService(mockRepo)

	mockRepo.On("Create", mock.Anything).Return(nil)

	org, err := service.CreateOrganization("New Corp", 1)

	assert.NoError(t, err)
	assert.Equal(t, "New Corp", org.Name)
	assert.Equal(t, uint(1), org.Users[0].ID) // Owner added as member
}

func TestInviteMember_Success(t *testing.T) {
	mockRepo := new(MockOrgRepo)
	service := NewOrganizationService(mockRepo)

	userToInvite := &models.User{ID: 5, Email: "new@employee.com"}

	// 1. Find User
	mockRepo.On("FindUserByEmail", "new@employee.com").Return(userToInvite, nil)
	// 2. Check if already member (Return False)
	mockRepo.On("IsMember", uint(5), uint(10)).Return(false, nil)
	// 3. Add Member
	mockRepo.On("AddMember", uint(10), uint(5)).Return(nil)

	err := service.InviteMember(10, "new@employee.com")
	assert.NoError(t, err)
}

func TestInviteMember_AlreadyMember(t *testing.T) {
	mockRepo := new(MockOrgRepo)
	service := NewOrganizationService(mockRepo)

	userToInvite := &models.User{ID: 5, Email: "existing@employee.com"}

	mockRepo.On("FindUserByEmail", "existing@employee.com").Return(userToInvite, nil)
	// Return TRUE (Already in org)
	mockRepo.On("IsMember", uint(5), uint(10)).Return(true, nil)

	err := service.InviteMember(10, "existing@employee.com")

	assert.Error(t, err)
	assert.Equal(t, "user is already a member", err.Error())

	// Ensure AddMember was NEVER called
	mockRepo.AssertNotCalled(t, "AddMember")
}
