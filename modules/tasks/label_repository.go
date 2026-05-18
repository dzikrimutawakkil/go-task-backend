package tasks

import (
	"gorm.io/gorm"
)

type LabelRepository interface {
	Create(label *Label) error
	FindByID(id uint) (*Label, error)
	FindByProjectID(projectID uint) ([]Label, error)
	Update(label *Label, updates map[string]interface{}) error
	Delete(label *Label) error

	// Task-Label associations
	GetLabelsByTaskID(taskID uint) ([]Label, error)
	SetTaskLabels(taskID uint, labelIDs []uint) error
	ClearTaskLabels(taskID uint) error

	// Search by project access
	CheckProjectAccess(projectID string, orgID string) (bool, error)
}

type labelRepository struct {
	db *gorm.DB
}

func NewLabelRepository(db *gorm.DB) LabelRepository {
	return &labelRepository{db}
}

func (r *labelRepository) Create(label *Label) error {
	return r.db.Create(label).Error
}

func (r *labelRepository) FindByID(id uint) (*Label, error) {
	var label Label
	err := r.db.First(&label, id).Error
	return &label, err
}

func (r *labelRepository) FindByProjectID(projectID uint) ([]Label, error) {
	var labels []Label
	err := r.db.Where("project_id = ?", projectID).Order("name asc").Find(&labels).Error
	return labels, err
}

func (r *labelRepository) Update(label *Label, updates map[string]interface{}) error {
	return r.db.Model(label).Updates(updates).Error
}

func (r *labelRepository) Delete(label *Label) error {
	// First clear task_labels associations
	r.db.Exec("DELETE FROM task_labels WHERE label_id = ?", label.ID)
	return r.db.Delete(label).Error
}

func (r *labelRepository) GetLabelsByTaskID(taskID uint) ([]Label, error) {
	var labels []Label
	err := r.db.Table("task_labels").
		Select("labels.*").
		Joins("JOIN labels ON labels.id = task_labels.label_id").
		Where("task_labels.task_id = ?", taskID).
		Find(&labels).Error
	return labels, err
}

func (r *labelRepository) SetTaskLabels(taskID uint, labelIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Clear existing
		if err := tx.Exec("DELETE FROM task_labels WHERE task_id = ?", taskID).Error; err != nil {
			return err
		}
		// Insert new associations
		if len(labelIDs) == 0 {
			return nil
		}
		var records []map[string]interface{}
		for _, lid := range labelIDs {
			records = append(records, map[string]interface{}{
				"task_id":  taskID,
				"label_id": lid,
			})
		}
		return tx.Table("task_labels").Create(&records).Error
	})
}

func (r *labelRepository) ClearTaskLabels(taskID uint) error {
	return r.db.Exec("DELETE FROM task_labels WHERE task_id = ?", taskID).Error
}

func (r *labelRepository) CheckProjectAccess(projectID string, orgID string) (bool, error) {
	var count int64
	err := r.db.Table("projects").
		Where("id = ? AND organization_id = ?", projectID, orgID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
