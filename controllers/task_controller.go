package controllers

import (
	"gotask-backend/config"
	"gotask-backend/models"
	"gotask-backend/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// POST /tasks
// POST /tasks
func CreateTask(c *gin.Context) {
	// ... (Input binding and Project validation stay the same) ...
	var input struct {
		Title      string     `json:"title" binding:"required"`
		ProjectID  uint       `json:"project_id" binding:"required"`
		StatusID   uint       `json:"status_id"`
		PriorityID uint       `json:"priority_id"`
		StartDate  *time.Time `json:"start_date"`
		EndDate    *time.Time `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	var project models.Project
	if err := config.DB.First(&project, input.ProjectID).Error; err != nil {
		utils.SendError(c, http.StatusNotFound, "Project not found")
		return
	}

	if input.StatusID == 0 {
		var todoStatus models.Status
		if err := config.DB.Where("slug = ?", "todo").First(&todoStatus).Error; err == nil {
			input.StatusID = todoStatus.ID
		}
	}

	if input.PriorityID == 0 {
		var medium models.Priority
		if err := config.DB.Where("name = ?", "Medium").First(&medium).Error; err == nil {
			input.PriorityID = medium.ID
		}
	}

	// 3. Create the Task
	task := models.Task{
		Title:      input.Title,
		ProjectID:  input.ProjectID,
		StatusID:   input.StatusID,
		PriorityID: input.PriorityID,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
	}
	config.DB.Create(&task)

	config.DB.Preload("Status").Preload("Priority").First(&task, task.ID)

	utils.SendSuccess(c, "Task created successfully", task)
}

// PATCH /tasks/:id (Don't forget to update this too!)
// PATCH /tasks/:id - Update any task detail (Title, Status, Assignees)
func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var task models.Task

	// 1. Find Task
	if err := config.DB.First(&task, id).Error; err != nil {
		utils.SendError(c, http.StatusNotFound, "Task not found")
		return
	}

	// 2. Define Input with Pointers
	// Pointers allow us to distinguish between "missing field" (nil) and "empty value"
	var input struct {
		Title       *string    `json:"title"`
		StatusID    *uint      `json:"status_id"`
		PriorityID  *uint      `json:"priority_id"`
		AssigneeIDs []uint     `json:"assignee_ids"`
		StartDate   *time.Time `json:"start_date"`
		EndDate     *time.Time `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Prepare Updates for Basic Fields
	// We use a map so GORM only updates the non-nil fields
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

	// Apply Basic Updates
	if len(updates) > 0 {
		config.DB.Model(&task).Updates(updates)
	}

	// 4. Handle Assignees (If provided)
	// This REPLACES the current list with the new list (Sync)
	if input.AssigneeIDs != nil {
		var users []models.User
		if len(input.AssigneeIDs) > 0 {
			config.DB.Find(&users, input.AssigneeIDs)
		}
		// "Replace" the current assignees with the new set
		config.DB.Model(&task).Association("Assignees").Replace(&users)
	}

	// 5. Reload Task with all details (Status + Assignees)
	config.DB.Preload("Status").Preload("Assignees").Preload("Priority").First(&task, task.ID)

	utils.SendSuccess(c, "Task updated successfully", task)
}

// DELETE /tasks/:id
func DeleteTask(c *gin.Context) {
	id := c.Param("id")

	// 1. Check if task exists
	var task models.Task
	if err := config.DB.First(&task, id).Error; err != nil {
		utils.SendError(c, http.StatusNotFound, "Task not found")
		return
	}

	// 2. Delete the task
	config.DB.Delete(&task)

	utils.SendSuccess(c, "Task deleted successfully")
}

// GET /projects/:id/tasks - Get all tasks for a specific project
func FindTasksByProject(c *gin.Context) {
	projectID := c.Param("id")

	// 1. Get logged-in user
	userContext, _ := c.Get("user")
	user := userContext.(models.User)

	// 2. Security Check: Does this user have access to this project?
	// We query the "project_users" join table via the User model
	var project models.Project
	err := config.DB.Model(&user).Association("Projects").Find(&project, projectID)

	// If project.ID is 0, it means the relation doesn't exist (User isn't in the project)
	if err != nil || project.ID == 0 {
		utils.SendError(c, http.StatusNotFound, "Project not found or access denied")
		return
	}

	// 3. Fetch Tasks with Status
	var tasks []models.Task
	result := config.DB.Preload("Status").Preload("Assignees").Preload("Priority").Where("project_id = ?", projectID).Find(&tasks)

	if result.Error != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch tasks")
		return
	}

	utils.SendSuccess(c, "success", tasks)
}

// POST /tasks/:id/take - Logged-in user assigns themselves
func TakeTask(c *gin.Context) {
	taskID := c.Param("id")

	// 1. Get logged-in user
	userContext, _ := c.Get("user")
	user := userContext.(models.User)

	// 2. Find Task
	var task models.Task
	if err := config.DB.First(&task, taskID).Error; err != nil {
		utils.SendError(c, http.StatusNotFound, "Task not found")
		return
	}

	// 3. Check if already assigned (Prevent duplicates)
	// We count how many times this user is assigned to this task (should be 0 or 1)
	var count int64
	config.DB.Table("task_users").Where("task_id = ? AND user_id = ?", task.ID, user.ID).Count(&count)

	if count > 0 {
		utils.SendError(c, http.StatusBadRequest, "You are already assigned to this task")
		return
	}

	// 4. Assign the user
	config.DB.Model(&task).Association("Assignees").Append(&user)

	utils.SendSuccess(c, "Task assigned to you successfully")
}

// POST /tasks/:id/assign_users - Assign multiple users by email
func AssignUsers(c *gin.Context) {
	taskID := c.Param("id")

	var input struct {
		Emails []string `json:"emails" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	var task models.Task
	if err := config.DB.First(&task, taskID).Error; err != nil {
		utils.SendError(c, http.StatusNotFound, "Task not found")
		return
	}

	// Find users
	var users []models.User
	if err := config.DB.Where("email IN ?", input.Emails).Find(&users).Error; err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	if len(users) == 0 {
		utils.SendError(c, http.StatusNotFound, "No matching users found")
		return
	}

	foundEmails := make(map[string]bool)
	for _, u := range users {
		foundEmails[u.Email] = true
	}

	var missingEmails []string
	for _, email := range input.Emails {
		if !foundEmails[email] {
			missingEmails = append(missingEmails, email)
		}
	}

	// Assign existing users
	err := config.DB.Model(&task).Association("Assignees").Append(&users)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to assign users")
		return
	}

	// Return success but also mention if anyone was skipped
	responseMsg := "Users assigned successfully"
	if len(missingEmails) > 0 {
		responseMsg = "Some users were assigned, but some were not found"
	}

	utils.SendSuccess(c, responseMsg, gin.H{
		"assigned_count": len(users),
		"assigned_users": users,
		"missing_emails": missingEmails,
	})
}
