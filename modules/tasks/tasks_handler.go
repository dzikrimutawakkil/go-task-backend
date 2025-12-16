package tasks

import (
	"gotask-backend/models"
	"gotask-backend/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service TaskService
}

func NewHandler(service TaskService) *Handler {
	return &Handler{service: service}
}

// GET /projects/:id/tasks
func (h *Handler) FindTasksByProject(c *gin.Context) {
	projectID := c.Param("id")

	tasks, err := h.service.GetTasksByProject(projectID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch tasks")
		return
	}

	utils.SendSuccess(c, "success", tasks)
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

// POST /tasks/:id/take (Assign self)
func (h *Handler) TakeTask(c *gin.Context) {
	taskID := c.Param("id")

	userContext, _ := c.Get("user")
	user := userContext.(models.User)

	// Reuse the AssignUsersByEmail service logic
	_, err := h.service.AssignUsersByEmail(taskID, []string{user.Email})
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, "Task assigned to you successfully")
}

// POST /tasks/:id/assign_users
func (h *Handler) AssignUsers(c *gin.Context) {
	taskID := c.Param("id")

	var req struct {
		Emails []string `json:"emails" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.AssignUsersByEmail(taskID, req.Emails)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	msg := "Users assigned successfully"
	if len(result.MissingEmails) > 0 {
		msg = "Some users were assigned, but some emails were not found"
	}

	utils.SendSuccess(c, msg, result)
}
