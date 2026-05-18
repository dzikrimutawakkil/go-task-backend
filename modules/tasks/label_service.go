package tasks

import (
	"errors"
	"strconv"
)

type LabelService interface {
	CreateLabel(projectID uint, name string, color string, requesterID uint, orgID string) (*Label, error)
	GetLabels(projectID uint, orgID string) ([]Label, error)
	UpdateLabel(labelID uint, name *string, color *string, orgID string) (*Label, error)
	DeleteLabel(labelID uint, orgID string) error
}

type labelService struct {
	repo LabelRepository
}

func NewLabelService(repo LabelRepository) LabelService {
	return &labelService{repo: repo}
}

// CreateLabel creates a new label for a project.
func (s *labelService) CreateLabel(projectID uint, name string, color string, requesterID uint, orgID string) (*Label, error) {
	// Verify project access
	hasAccess, err := s.repo.CheckProjectAccess(uintToString(projectID), orgID)
	if err != nil || !hasAccess {
		return nil, errors.New("project not found or access denied")
	}

	label := Label{
		ProjectID: projectID,
		Name:      name,
		Color:     color,
	}

	if err := s.repo.Create(&label); err != nil {
		return nil, err
	}

	return &label, nil
}

// GetLabels retrieves all labels for a project.
func (s *labelService) GetLabels(projectID uint, orgID string) ([]Label, error) {
	// Verify project access
	hasAccess, err := s.repo.CheckProjectAccess(uintToString(projectID), orgID)
	if err != nil || !hasAccess {
		return nil, errors.New("project not found or access denied")
	}

	return s.repo.FindByProjectID(projectID)
}

// UpdateLabel updates a label's name and/or color.
func (s *labelService) UpdateLabel(labelID uint, name *string, color *string, orgID string) (*Label, error) {
	label, err := s.repo.FindByID(labelID)
	if err != nil {
		return nil, errors.New("label not found")
	}

	// Verify project access
	hasAccess, err := s.repo.CheckProjectAccess(uintToString(label.ProjectID), orgID)
	if err != nil || !hasAccess {
		return nil, errors.New("project not found or access denied")
	}

	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if color != nil {
		updates["color"] = *color
	}

	if err := s.repo.Update(label, updates); err != nil {
		return nil, err
	}

	return s.repo.FindByID(labelID)
}

// DeleteLabel removes a label from a project.
func (s *labelService) DeleteLabel(labelID uint, orgID string) error {
	label, err := s.repo.FindByID(labelID)
	if err != nil {
		return errors.New("label not found")
	}

	// Verify project access
	hasAccess, err := s.repo.CheckProjectAccess(uintToString(label.ProjectID), orgID)
	if err != nil || !hasAccess {
		return errors.New("project not found or access denied")
	}

	return s.repo.Delete(label)
}

// Helper: convert uint to string
func uintToString(n uint) string {
	return strconv.FormatUint(uint64(n), 10)
}
