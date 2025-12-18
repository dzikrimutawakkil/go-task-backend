package tasks

import (
	"gotask-backend/utils"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service TaskService
}

func NewTaskHandler(service TaskService) *Handler {
	return &Handler{service: service}
}

// GET /projects/:id/tasks
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

// POST /tasks
func (h *Handler) CreateTask(c *gin.Context) {
	var req struct {
		Title      string     `json:"title" binding:"required"`
		ProjectID  uint       `json:"project_id" binding:"required"`
		StatusID   uint       `json:"status_id"`
		PriorityID uint       `json:"priority_id"`
		StartDate  *time.Time `json:"start_date"`
		EndDate    *time.Time `json:"end_date"`
	}

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
	}

	task, err := h.service.CreateTask(input)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to create task")
		return
	}

	utils.SendSuccess(c, "Task created successfully", task)
}

// PATCH /tasks/:id
func (h *Handler) UpdateTask(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Title       *string    `json:"title"`
		StatusID    *uint      `json:"status_id"`
		PriorityID  *uint      `json:"priority_id"`
		AssigneeIDs []uint     `json:"assignee_ids"`
		StartDate   *time.Time `json:"start_date"`
		EndDate     *time.Time `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	input := UpdateTaskInput{
		Title:       req.Title,
		StatusID:    req.StatusID,
		PriorityID:  req.PriorityID,
		AssigneeIDs: req.AssigneeIDs,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	task, err := h.service.UpdateTask(id, input)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccess(c, "Task updated successfully", task)
}

// DELETE /tasks/:id
func (h *Handler) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteTask(id); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	utils.SendSuccess(c, "Task deleted successfully")
}

func (h *Handler) FindStatusesByProject(c *gin.Context) {
	projectID := c.Param("id")

	statuses, err := h.service.GetStatuses(projectID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch statuses")
		return
	}

	utils.SendSuccess(c, "success", statuses)
}
