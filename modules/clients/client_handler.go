package clients

import (
	"net/http"
	"strconv"

	"gotask-backend/utils"

	"github.com/gin-gonic/gin"
)

// Request DTOs
type CreateClientRequest struct {
	Name     string  `json:"name" binding:"required" example:"John Doe"`
	Email    *string `json:"email" example:"john@example.com"`
	WhatsApp *string `json:"whatsapp" example:"+6281234567890"`
	Phone    *string `json:"phone" example:"+6281234567890"`
	Company  *string `json:"company" example:"Acme Corp"`
	Website  *string `json:"website" example:"https://acme.com"`
	Address  *string `json:"address" example:"Jakarta, Indonesia"`
	Notes    *string `json:"notes" example:"Referral from friend"`
}

type UpdateClientRequest struct {
	Name     *string `json:"name" example:"John Doe"`
	Email    *string `json:"email" example:"john@example.com"`
	WhatsApp *string `json:"whatsapp" example:"+6281234567890"`
	Phone    *string `json:"phone" example:"+6281234567890"`
	Company  *string `json:"company" example:"Acme Corp"`
	Website  *string `json:"website" example:"https://acme.com"`
	Address  *string `json:"address" example:"Jakarta, Indonesia"`
	Notes    *string `json:"notes" example:"Updated notes"`
}

type Handler struct {
	service ClientService
}

func NewClientHandler(service ClientService) *Handler {
	return &Handler{service: service}
}

// ListClients godoc
// @Summary     List clients
// @Description Get all clients for the current organization.
// @Tags        Clients
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "success"
// @Failure     500 {object} utils.APIResponse "Failed to fetch clients"
// @Router      /clients [get]
func (h *Handler) ListClients(c *gin.Context) {
	orgID := c.MustGet("org_id").(string)

	clients, err := h.service.GetClients(orgID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch clients")
		return
	}

	utils.SendSuccess(c, "success", gin.H{
		"clients": clients,
	})
}

// GetClient godoc
// @Summary     Get a client
// @Description Get a single client by ID.
// @Tags        Clients
// @Produce     json
// @Param       id path int true "Client ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "success"
// @Failure     404 {object} utils.APIResponse "Client not found"
// @Router      /clients/{id} [get]
func (h *Handler) GetClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid client ID")
		return
	}

	client, err := h.service.GetClient(uint(id))
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Client not found")
		return
	}

	utils.SendSuccess(c, "success", client)
}

// CreateClient godoc
// @Summary     Create a client
// @Description Create a new client for the current organization.
// @Tags        Clients
// @Accept      json
// @Produce     json
// @Param       body body CreateClientRequest true "Client payload"
// @Security    BearerAuth
// @Success     201 {object} utils.APIResponse "Client created successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     500 {object} utils.APIResponse "Failed to create client"
// @Router      /clients [post]
func (h *Handler) CreateClient(c *gin.Context) {
	orgIDStr := c.MustGet("org_id").(string)
	orgID, _ := strconv.ParseUint(orgIDStr, 10, 64)

	var req CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	input := CreateClientInput{
		Name:           req.Name,
		Email:          req.Email,
		WhatsApp:       req.WhatsApp,
		Phone:          req.Phone,
		Company:        req.Company,
		Website:        req.Website,
		Address:        req.Address,
		Notes:          req.Notes,
		OrganizationID: uint(orgID),
	}

	client, err := h.service.CreateClient(input)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to create client")
		return
	}

	utils.SendSuccess(c, "Client created successfully", client)
}

// UpdateClient godoc
// @Summary     Update a client
// @Description Update a client's fields.
// @Tags        Clients
// @Accept      json
// @Produce     json
// @Param       id path int true "Client ID"
// @Param       body body UpdateClientRequest true "Client update payload"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Client updated successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     404 {object} utils.APIResponse "Client not found"
// @Router      /clients/{id} [patch]
func (h *Handler) UpdateClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid client ID")
		return
	}

	var req UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	input := UpdateClientInput{
		Name:     req.Name,
		Email:    req.Email,
		WhatsApp: req.WhatsApp,
		Phone:    req.Phone,
		Company:  req.Company,
		Website:  req.Website,
		Address:  req.Address,
		Notes:    req.Notes,
	}

	client, err := h.service.UpdateClient(uint(id), input)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Client not found")
		return
	}

	utils.SendSuccess(c, "Client updated successfully", client)
}

// DeleteClient godoc
// @Summary     Delete a client
// @Description Permanently delete a client.
// @Tags        Clients
// @Produce     json
// @Param       id path int true "Client ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Client deleted successfully"
// @Failure     404 {object} utils.APIResponse "Client not found"
// @Router      /clients/{id} [delete]
func (h *Handler) DeleteClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid client ID")
		return
	}

	if err := h.service.DeleteClient(uint(id)); err != nil {
		utils.SendError(c, http.StatusNotFound, "Client not found")
		return
	}

	utils.SendSuccess(c, "Client deleted successfully")
}

// GetClientStats godoc
// @Summary     Get client statistics
// @Description Get total count, total revenue, and average revenue for clients.
// @Tags        Clients
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "success"
// @Failure     500 {object} utils.APIResponse "Failed to fetch stats"
// @Router      /clients/stats [get]
func (h *Handler) GetClientStats(c *gin.Context) {
	orgID := c.MustGet("org_id").(string)

	stats, err := h.service.GetClientStats(orgID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch stats")
		return
	}

	utils.SendSuccess(c, "success", stats)
}