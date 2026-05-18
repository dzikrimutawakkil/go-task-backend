package tasks

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(task *Task) error
	FindByID(id string) (*Task, error)
	FindByProjectID(projectID string, page int, limit int) ([]Task, int64, error)
	Update(task *Task, updates map[string]interface{}) error
	UpdateWithVersion(task *Task, updates map[string]interface{}, expectedVersion uint) (bool, error)
	Delete(task *Task) error

	ClearAssignees(task *Task) error
	AssignUsers(task *Task, userIDs []uint) error

	CheckProjectAccess(projectID string, orgID string) (bool, error)

	// Search & Filter
	SearchTasks(orgID string, query string, filters TaskFilters, page int, limit int) ([]Task, int64, error)

	CreateStatus(status *Status) error
	GetStatusesByProjectID(projectID string) ([]Status, error)
	FindStatusByID(id string) (*Status, error)
	UpdateStatus(status *Status, updates map[string]interface{}) error
	DeleteStatus(status *Status) error
	BulkUpdateStatuses(statuses []Status) error
	GetMaxIndex(projectID string) (int, error)
}

// TaskFilters holds the filter parameters for searching tasks.
type TaskFilters struct {
	ProjectID   *uint
	AssigneeID  *uint
	StatusID    *uint
	PriorityID  *uint
	DueFrom     *time.Time
	DueTo       *time.Time
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	LabelIDs    []uint
}

type repository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &repository{db}
}

// Helper internal untuk mengambil AssigneeIDs
func (r *repository) fetchAssigneeIDs(task *Task) error {
	var userIDs []uint
	// Query manual ke tabel penghubung
	err := r.db.Table("task_users").
		Where("task_id = ?", task.ID).
		Pluck("user_id", &userIDs).Error

	if err == nil {
		task.AssigneeIDs = userIDs
	}
	return err
}

func (r *repository) Create(task *Task) error {
	return r.db.Create(task).Error
}

func (r *repository) FindByID(id string) (*Task, error) {
	var task Task
	err := r.db.Preload("Status").
		Preload("Priority").
		First(&task, id).Error

	if err != nil {
		return nil, err
	}

	// Manual Fetch IDs
	_ = r.fetchAssigneeIDs(&task)
	return &task, nil
}

func (r *repository) FindByProjectID(projectID string, page int, limit int) ([]Task, int64, error) {
	var tasks []Task
	var total int64

	offset := (page - 1) * limit

	if err := r.db.Model(&Task{}).Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Status").
		Preload("Priority").
		Where("project_id = ?", projectID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error

	if err != nil {
		return nil, 0, err
	}

	// Populate IDs for each task (Looping query is N+1 problem, but acceptable for MVP microservice separation)
	for i := range tasks {
		_ = r.fetchAssigneeIDs(&tasks[i])
	}
	return tasks, total, nil
}

func (r *repository) Update(task *Task, updates map[string]interface{}) error {
	return r.db.Model(task).Updates(updates).Error
}

// UpdateWithVersion performs an optimistic lock update.
// It increments the version and only updates if the current version matches expectedVersion.
// Returns (true, nil) if updated, (false, nil) if version mismatch, or (false, err) on DB error.
func (r *repository) UpdateWithVersion(task *Task, updates map[string]interface{}, expectedVersion uint) (bool, error) {
	// Increment version
	updates["version"] = gorm.Expr("version + 1")

	result := r.db.Model(task).
		Where("id = ? AND version = ?", task.ID, expectedVersion).
		Updates(updates)

	if result.Error != nil {
		return false, result.Error
	}

	if result.RowsAffected == 0 {
		// Version mismatch — concurrent modification detected
		return false, nil
	}

	return true, nil
}

func (r *repository) Delete(task *Task) error {
	return r.db.Delete(task).Error
}

func (r *repository) ClearAssignees(task *Task) error {
	// Manual Delete dari tabel penghubung
	return r.db.Exec("DELETE FROM task_users WHERE task_id = ?", task.ID).Error
}

func (r *repository) AssignUsers(task *Task, userIDs []uint) error {
	// Manual Insert ke tabel penghubung
	// Kita buat struct temporary atau insert map
	var records []map[string]interface{}
	for _, uid := range userIDs {
		records = append(records, map[string]interface{}{
			"task_id": task.ID,
			"user_id": uid,
		})
	}

	if len(records) > 0 {
		return r.db.Table("task_users").Create(&records).Error
	}
	return nil
}

func (r *repository) CheckProjectAccess(projectID string, orgID string) (bool, error) {
	var count int64
	err := r.db.Table("projects").
		Where("id = ? AND organization_id = ?", projectID, orgID).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) CreateStatus(status *Status) error {
	return r.db.Create(status).Error
}

func (r *repository) GetStatusesByProjectID(projectID string) ([]Status, error) {
	var statuses []Status
	err := r.db.Where("project_id = ?", projectID).Order("index asc").Find(&statuses).Error
	return statuses, err
}

func (r *repository) FindStatusByID(id string) (*Status, error) {
	var status Status
	err := r.db.First(&status, id).Error
	return &status, err
}

func (r *repository) UpdateStatus(status *Status, updates map[string]interface{}) error {
	return r.db.Model(status).Updates(updates).Error
}

func (r *repository) DeleteStatus(status *Status) error {
	// Validasi opsional: Cek apakah status sedang dipakai oleh Task lain
	var count int64
	r.db.Model(&Task{}).Where("status_id = ?", status.ID).Count(&count)
	if count > 0 {
		return gorm.ErrForeignKeyViolated // Jangan hapus jika masih ada task
	}

	return r.db.Delete(status).Error
}

func (r *repository) BulkUpdateStatuses(statuses []Status) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, status := range statuses {
			updates := map[string]interface{}{
				"index": status.Index,
				"name":  status.Name,
			}

			if err := tx.Model(&Status{}).Where("id = ?", status.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repository) GetMaxIndex(projectID string) (int, error) {
	var status Status

	err := r.db.Where("project_id = ?", projectID).
		Order("index desc").
		First(&status).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return -1, nil
		}
		return 0, err
	}

	return status.Index, nil
}

// SearchTasks implements full-text search and filtering for tasks.
// It searches by title/description and applies all filters.
func (r *repository) SearchTasks(orgID string, query string, filters TaskFilters, page int, limit int) ([]Task, int64, error) {
	var tasks []Task
	var total int64

	// Base query: join through projects to filter by org
	baseQuery := r.db.Table("tasks").
		Joins("JOIN projects ON tasks.project_id = projects.id").
		Where("projects.organization_id = ?", orgID)

	// Apply text search on title
	if query != "" {
		searchPattern := "%" + strings.ToLower(query) + "%"
		baseQuery = baseQuery.Where("LOWER(tasks.title) LIKE ?", searchPattern)
	}

	// Apply filters
	if filters.ProjectID != nil {
		baseQuery = baseQuery.Where("tasks.project_id = ?", *filters.ProjectID)
	}
	if filters.AssigneeID != nil {
		baseQuery = baseQuery.Joins("JOIN task_users ON tasks.id = task_users.task_id").
			Where("task_users.user_id = ?", *filters.AssigneeID)
	}
	if filters.StatusID != nil {
		baseQuery = baseQuery.Where("tasks.status_id = ?", *filters.StatusID)
	}
	if filters.PriorityID != nil {
		baseQuery = baseQuery.Where("tasks.priority_id = ?", *filters.PriorityID)
	}
	if filters.DueFrom != nil {
		baseQuery = baseQuery.Where("tasks.end_date >= ?", *filters.DueFrom)
	}
	if filters.DueTo != nil {
		baseQuery = baseQuery.Where("tasks.end_date <= ?", *filters.DueTo)
	}
	if filters.CreatedFrom != nil {
		baseQuery = baseQuery.Where("tasks.created_at >= ?", *filters.CreatedFrom)
	}
	if filters.CreatedTo != nil {
		baseQuery = baseQuery.Where("tasks.created_at <= ?", *filters.CreatedTo)
	}
	if len(filters.LabelIDs) > 0 {
		baseQuery = baseQuery.Joins("JOIN task_labels ON tasks.id = task_labels.task_id").
			Where("task_labels.label_id IN ?", filters.LabelIDs)
	}

	// Get total count
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit

	// Fetch tasks with relations
	err := baseQuery.
		Preload("Status").
		Preload("Priority").
		Order("tasks.created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error

	if err != nil {
		return nil, 0, err
	}

	// Populate assignee IDs
	for i := range tasks {
		_ = r.fetchAssigneeIDs(&tasks[i])
	}

	return tasks, total, nil
}
