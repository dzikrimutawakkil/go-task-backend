package tasks

import (
	"gotask-backend/utils"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateTaskRequest represents the request body for creating a task.
// @Description Request body for creating a task
type CreateTaskRequest struct {
	Title      string     `json:"title" binding:"required" example:"Implement login feature"`
	ProjectID  uint       `json:"project_id" binding:"required" example:"1"`
	StatusID   uint       `json:"status_id" example:"1"`
	PriorityID uint       `json:"priority_id" example:"1"`
	StartDate  *time.Time `json:"start_date" example:"2026-05-20T00:00:00Z"`
	EndDate    *time.Time `json:"end_date" example:"2026-05-25T00:00:00Z"`
	LabelIDs   []uint     `json:"label_ids"`
}

// UpdateTaskRequest represents the request body for updating a task.
// @Description Request body for updating a task
type UpdateTaskRequest struct {
	Title           *string    `json:"title" example:"Updated task title"`
	Description     *string    `json:"description" example:"Updated task description"`
	StatusID        *uint      `json:"status_id" example:"2"`
	PriorityID      *uint      `json:"priority_id" example:"2"`
	AssigneeIDs     []uint     `json:"assignee_ids"`
	StartDate       *time.Time `json:"start_date" example:"2026-05-20T00:00:00Z"`
	EndDate         *time.Time `json:"end_date" example:"2026-05-30T00:00:00Z"`
	LabelIDs        []uint     `json:"label_ids"`
	ExpectedVersion *uint      `json:"expected_version" example:"1"`
}

// CreateStatusRequest represents the request body for creating a status.
// @Description Request body for creating a status
type CreateStatusRequest struct {
	Name  string `json:"name" binding:"required" example:"In Review"`
	Index int    `json:"index" example:"2"`
}

// UpdateStatusRequest represents the request body for updating a status.
// @Description Request body for updating a status
type UpdateStatusRequest struct {
	Name  *string `json:"name" example:"Done"`
	Index *int    `json:"index" example:"3"`
}

type Handler struct {
	service TaskService
}

func NewTaskHandler(service TaskService) *Handler {
	return &Handler{service: service}
}

// FindTasksByProject godoc
// @Summary     List tasks by project
// @Description Retrieve all tasks belonging to a specific project with pagination.
// @Tags        Tasks
// @Produce     json
// @Param       id path int true "Project ID"
// @Param       page query int false "Page number" default(1)
// @Param       limit query int false "Items per page" default(50)
// @Param       X-Organization-ID header string true "Organization ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "success"
// @Failure     400 {object} utils.APIResponse "Missing organization header"
// @Failure     500 {object} utils.APIResponse "Failed to fetch tasks"
// @Router      /projects/{id}/tasks [get]
func (h *Handler) FindTasksByProject(c *gin.Context) {
	projectID := c.Param("id")

	orgIDInterface, exists := c.Get("org_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Organization-ID header is required")
		return
	}
	orgID := orgIDInterface.(string)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	tasks, total, err := h.service.GetTasksByProject(projectID, orgID, page, limit)

	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch tasks")
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	utils.SendSuccess(c, "success", gin.H{
		"tasks": tasks,
		"meta": gin.H{
			"current_page": page,
			"limit":        limit,
			"total_data":   total,
			"total_pages":  totalPages,
		},
	})
}

// SearchTasks godoc
// @Summary     Search tasks
// @Description Full-text search across task titles with optional filters for assignee, status, priority, project, and date ranges.
// @Tags        Tasks
// @Produce     json
// @Param       q query string false "Search query"
// @Param       page query int false "Page number" default(1)
// @Param       limit query int false "Items per page (max 100)" default(20)
// @Param       assignee_id query int false "Filter by assignee"
// @Param       status_id query int false "Filter by status"
// @Param       priority_id query int false "Filter by priority"
// @Param       project_id query int false "Filter by project"
// @Param       due_from query string false "Filter by due date start (YYYY-MM-DD)"
// @Param       due_to query string false "Filter by due date end (YYYY-MM-DD)"
// @Param       created_from query string false "Filter by created date start (RFC3339)"
// @Param       created_to query string false "Filter by created date end (RFC3339)"
// @Param       label_ids query string false "Filter by labels (comma-separated IDs)"
// @Param       X-Organization-ID header string true "Organization ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "success"
// @Failure     400 {object} utils.APIResponse "Missing organization header"
// @Failure     500 {object} utils.APIResponse "Failed to search tasks"
// @Router      /tasks/search [get]
func (h *Handler) SearchTasks(c *gin.Context) {
	orgIDInterface, exists := c.Get("org_id")
	if !exists {
		utils.SendError(c, http.StatusBadRequest, "X-Organization-ID header is required")
		return
	}
	orgID := orgIDInterface.(string)

	// Parse query params
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit > 100 {
		limit = 100
	}
	if page < 1 {
		page = 1
	}

	// Parse filters
	filters := TaskFilters{}

	if assigneeIDStr := c.Query("assignee_id"); assigneeIDStr != "" {
		if id, err := strconv.ParseUint(assigneeIDStr, 10, 64); err == nil {
			aid := uint(id)
			filters.AssigneeID = &aid
		}
	}

	if statusIDStr := c.Query("status_id"); statusIDStr != "" {
		if id, err := strconv.ParseUint(statusIDStr, 10, 64); err == nil {
			sid := uint(id)
			filters.StatusID = &sid
		}
	}

	if priorityIDStr := c.Query("priority_id"); priorityIDStr != "" {
		if id, err := strconv.ParseUint(priorityIDStr, 10, 64); err == nil {
			pid := uint(id)
			filters.PriorityID = &pid
		}
	}

	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		if id, err := strconv.ParseUint(projectIDStr, 10, 64); err == nil {
			pid := uint(id)
			filters.ProjectID = &pid
		}
	}

	if dueFromStr := c.Query("due_from"); dueFromStr != "" {
		if t, err := time.Parse("2006-01-02", dueFromStr); err == nil {
			filters.DueFrom = &t
		}
	}

	if dueToStr := c.Query("due_to"); dueToStr != "" {
		if t, err := time.Parse("2006-01-02", dueToStr); err == nil {
			filters.DueTo = &t
		}
	}

	if createdFromStr := c.Query("created_from"); createdFromStr != "" {
		if t, err := time.Parse(time.RFC3339, createdFromStr); err == nil {
			filters.CreatedFrom = &t
		}
	}

	if createdToStr := c.Query("created_to"); createdToStr != "" {
		if t, err := time.Parse(time.RFC3339, createdToStr); err == nil {
			filters.CreatedTo = &t
		}
	}

	if labelIDsStr := c.Query("label_ids"); labelIDsStr != "" {
		var labelIDs []uint
		for _, s := range splitAndTrim(labelIDsStr, ",") {
			if id, err := strconv.ParseUint(s, 10, 64); err == nil {
				labelIDs = append(labelIDs, uint(id))
			}
		}
		filters.LabelIDs = labelIDs
	}

	tasks, total, err := h.service.SearchTasks(orgID, query, filters, page, limit)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to search tasks")
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	utils.SendSuccess(c, "success", gin.H{
		"tasks": tasks,
		"meta": gin.H{
			"current_page": page,
			"limit":        limit,
			"total_data":   total,
			"total_pages":  totalPages,
		},
	})
}

// splitAndTrim splits a string by separator and trims whitespace.
func splitAndTrim(s, sep string) []string {
	var result []string
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i = start - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// CreateTask godoc
// @Summary     Create a task
// @Description Create a new task within a project.
// @Tags        Tasks
// @Accept      json
// @Produce     json
// @Param       body body CreateTaskRequest true "Task payload"
// @Security    BearerAuth
// @Success     201 {object} utils.APIResponse{data=Task} "Task created successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     500 {object} utils.APIResponse "Failed to create task"
// @Router      /tasks [post]
func (h *Handler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	input := CreateTaskInput{
		Title:      req.Title,
		ProjectID:  req.ProjectID,
		StatusID:   req.StatusID,
		PriorityID: req.PriorityID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		LabelIDs:   req.LabelIDs,
	}

	task, err := h.service.CreateTask(input)
	if err != nil {
		if quotaErr, ok := err.(*utils.QuotaError); ok {
			utils.SendError(c, http.StatusForbidden, quotaErr.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "Failed to create task")
		return
	}

	utils.SendSuccess(c, "Task created successfully", task)
}

// UpdateTask godoc
// @Summary     Update a task
// @Description Update a task's fields. Supports optimistic locking via expected_version.
// @Tags        Tasks
// @Accept      json
// @Produce     json
// @Param       id path int true "Task ID"
// @Param       body body UpdateTaskRequest true "Task update payload"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=Task} "Task updated successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     500 {object} utils.APIResponse "Failed to update task"
// @Router      /tasks/{id} [patch]
func (h *Handler) UpdateTask(c *gin.Context) {
	id := c.Param("id")

	var req UpdateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	input := UpdateTaskInput{
		Title:           req.Title,
		StatusID:        req.StatusID,
		PriorityID:      req.PriorityID,
		AssigneeIDs:     req.AssigneeIDs,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		LabelIDs:        req.LabelIDs,
		ExpectedVersion: req.ExpectedVersion,
	}

	task, err := h.service.UpdateTask(id, input)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccess(c, "Task updated successfully", task)
}

// DeleteTask godoc
// @Summary     Delete a task
// @Description Permanently delete a task.
// @Tags        Tasks
// @Produce     json
// @Param       id path int true "Task ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Task deleted successfully"
// @Failure     500 {object} utils.APIResponse "Failed to delete task"
// @Router      /tasks/{id} [delete]
func (h *Handler) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteTask(id); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	utils.SendSuccess(c, "Task deleted successfully")
}

// FindStatusesByProject godoc
// @Summary     List project statuses
// @Description Retrieve all task statuses for a specific project.
// @Tags        Statuses
// @Produce     json
// @Param       id path int true "Project ID"
// @Param       X-Organization-ID header string true "Organization ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "success"
// @Failure     500 {object} utils.APIResponse "Failed to fetch statuses"
// @Router      /projects/{id}/status [get]
func (h *Handler) FindStatusesByProject(c *gin.Context) {
	projectID := c.Param("id")

	statuses, err := h.service.GetStatuses(projectID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch statuses")
		return
	}

	utils.SendSuccess(c, "success", statuses)
}

// CreateStatus godoc
// @Summary     Create a status
// @Description Create a new status (e.g., "To Do", "In Progress", "Done") within a project.
// @Tags        Statuses
// @Accept      json
// @Produce     json
// @Param       id path int true "Project ID"
// @Param       body body CreateStatusRequest true "Status payload"
// @Param       X-Organization-ID header string true "Organization ID"
// @Security    BearerAuth
// @Success     201 {object} utils.APIResponse{data=Status} "Status created"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     500 {object} utils.APIResponse "Failed to create status"
// @Router      /projects/{id}/status [post]
func (h *Handler) CreateStatus(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, _ := strconv.Atoi(projectIDStr)

	var req CreateStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	status, err := h.service.CreateNewStatus(uint(projectID), req.Name, req.Index)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to create status")
		return
	}

	utils.SendSuccess(c, "Status created", status)
}

// UpdateStatus godoc
// @Summary     Update a status
// @Description Update a status name or position (index). Useful for reordering statuses in a Kanban board.
// @Tags        Statuses
// @Accept      json
// @Produce     json
// @Param       id path int true "Status ID"
// @Param       body body UpdateStatusRequest true "Status update payload"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse{data=Status} "Status updated successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     500 {object} utils.APIResponse "Failed to update status"
// @Router      /status/{id} [patch]
func (h *Handler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	status, err := h.service.UpdateStatus(id, req.Name, req.Index)

	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccess(c, "Status updated successfully", status)
}

// DeleteStatus godoc
// @Summary     Delete a status
// @Description Delete a status. Fails if the status is still in use by any task.
// @Tags        Statuses
// @Produce     json
// @Param       id path int true "Status ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Status deleted successfully"
// @Failure     400 {object} utils.APIResponse "Failed to delete (status might be in use)"
// @Router      /status/{id} [delete]
func (h *Handler) DeleteStatus(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteStatus(id); err != nil {
		// Cek jika error karena masih dipakai task
		utils.SendError(c, http.StatusBadRequest, "Failed to delete (status might be in use)")
		return
	}
	utils.SendSuccess(c, "Status deleted successfully")
}
