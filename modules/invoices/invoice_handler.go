package invoices

import (
	"net/http"
	"strconv"
	"time"

	"gotask-backend/utils"

	"github.com/gin-gonic/gin"
)

// Request DTOs
type CreateInvoiceRequest struct {
	ClientID  *uint                 `json:"client_id"`
	ProjectID *uint                 `json:"project_id"`
	Title     *string               `json:"title"`
	Amount    float64               `json:"amount" binding:"required" example:"500000"`
	Tax       *float64              `json:"tax" example:"50000"`
	Discount  *float64              `json:"discount" example:"0"`
	DueDate   *string               `json:"due_date" example:"2026-06-30T00:00:00Z"`
	Notes     *string               `json:"notes" example:"Payment due within 30 days"`
	Items     *[]InvoiceItemRequest `json:"items"`
}

type InvoiceItemRequest struct {
	Description string  `json:"description" binding:"required" example:"Web Development"`
	Quantity    float64 `json:"quantity" binding:"required" example:"1"`
	UnitPrice   float64 `json:"unit_price" binding:"required" example:"500000"`
	Total       float64 `json:"total" binding:"required" example:"500000"`
}

type UpdateInvoiceRequest struct {
	ClientID  *uint                 `json:"client_id"`
	ProjectID *uint                 `json:"project_id"`
	Title     *string               `json:"title"`
	Amount    *float64              `json:"amount" example:"500000"`
	Tax       *float64              `json:"tax" example:"50000"`
	Discount  *float64              `json:"discount" example:"0"`
	Status    *string               `json:"status" example:"paid"`
	DueDate   *string               `json:"due_date" example:"2026-06-30T00:00:00Z"`
	Notes     *string               `json:"notes"`
	Items     *[]InvoiceItemRequest `json:"items"`
}

type MarkPaidRequest struct {
	AmountPaid float64 `json:"amount_paid" binding:"required,min=0" example:"500000"`
	PaidAt     *string `json:"paid_at" example:"2026-05-20T10:00:00Z"`
}

type Handler struct {
	service InvoiceService
}

func NewInvoiceHandler(service InvoiceService) *Handler {
	return &Handler{service: service}
}

// ListInvoices godoc
// @Summary     List invoices
// @Description Get all invoices for the current organization.
// @Tags        Invoices
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "success"
// @Failure     500 {object} utils.APIResponse "Failed to fetch invoices"
// @Router      /invoices [get]
func (h *Handler) ListInvoices(c *gin.Context) {
	orgID := c.MustGet("org_id").(string)

	invoices, err := h.service.GetInvoices(orgID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch invoices")
		return
	}

	utils.SendSuccess(c, "success", gin.H{
		"invoices": invoices,
	})
}

// GetInvoice godoc
// @Summary     Get an invoice
// @Description Get a single invoice by ID.
// @Tags        Invoices
// @Produce     json
// @Param       id path int true "Invoice ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "success"
// @Failure     404 {object} utils.APIResponse "Invoice not found"
// @Router      /invoices/{id} [get]
func (h *Handler) GetInvoice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	invoice, err := h.service.GetInvoice(uint(id))
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Invoice not found")
		return
	}

	utils.SendSuccess(c, "success", invoice)
}

// CreateInvoice godoc
// @Summary     Create an invoice
// @Description Create a new invoice with auto-generated invoice number (INV-YYYY-XXX).
// @Tags        Invoices
// @Accept      json
// @Produce     json
// @Param       body body CreateInvoiceRequest true "Invoice payload"
// @Security    BearerAuth
// @Success     201 {object} utils.APIResponse "Invoice created successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     500 {object} utils.APIResponse "Failed to create invoice"
// @Router      /invoices [post]
func (h *Handler) CreateInvoice(c *gin.Context) {
	orgIDStr := c.MustGet("org_id").(string)
	orgID, _ := strconv.ParseUint(orgIDStr, 10, 64)

	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		t, err := time.Parse(time.RFC3339, *req.DueDate)
		if err == nil {
			dueDate = &t
		}
	}

	var items InvoiceItems
	if req.Items != nil {
		for _, item := range *req.Items {
			items = append(items, InvoiceItem{
				Description: item.Description,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				Total:       item.Total,
			})
		}
	}

	tax := 0.0
	if req.Tax != nil {
		tax = *req.Tax
	}
	discount := 0.0
	if req.Discount != nil {
		discount = *req.Discount
	}

	input := CreateInvoiceInput{
		OrganizationID: uint(orgID),
		ClientID:       req.ClientID,
		ProjectID:      req.ProjectID,
		Title:          req.Title,
		Amount:         req.Amount,
		Tax:            tax,
		Discount:       discount,
		DueDate:        dueDate,
		Notes:          req.Notes,
		Items:          items,
	}

	invoice, err := h.service.CreateInvoice(input)
	if err != nil {
		if quotaErr, ok := err.(*utils.QuotaError); ok {
			utils.SendError(c, http.StatusForbidden, quotaErr.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "Failed to create invoice")
		return
	}

	utils.SendSuccess(c, "Invoice created successfully", invoice)
}

// UpdateInvoice godoc
// @Summary     Update an invoice
// @Description Update an invoice. Setting status to "paid" triggers revenue sync to client.
// @Tags        Invoices
// @Accept      json
// @Produce     json
// @Param       id path int true "Invoice ID"
// @Param       body body UpdateInvoiceRequest true "Invoice update payload"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Invoice updated successfully"
// @Failure     400 {object} utils.APIResponse "Validation error"
// @Failure     404 {object} utils.APIResponse "Invoice not found"
// @Router      /invoices/{id} [patch]
func (h *Handler) UpdateInvoice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	var req UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		t, err := time.Parse(time.RFC3339, *req.DueDate)
		if err == nil {
			dueDate = &t
		}
	}

	var items InvoiceItems
	if req.Items != nil {
		for _, item := range *req.Items {
			items = append(items, InvoiceItem{
				Description: item.Description,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				Total:       item.Total,
			})
		}
	}

	input := UpdateInvoiceInput{
		ClientID:  req.ClientID,
		ProjectID: req.ProjectID,
		Title:     req.Title,
		Amount:    req.Amount,
		Tax:       req.Tax,
		Discount:  req.Discount,
		Status:    req.Status,
		DueDate:   dueDate,
		Notes:     req.Notes,
		Items:     items,
	}

	invoice, err := h.service.UpdateInvoice(uint(id), input)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Invoice not found")
		return
	}

	utils.SendSuccess(c, "Invoice updated successfully", invoice)
}

// DeleteInvoice godoc
// @Summary     Delete an invoice
// @Description Permanently delete an invoice.
// @Tags        Invoices
// @Produce     json
// @Param       id path int true "Invoice ID"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Invoice deleted successfully"
// @Failure     404 {object} utils.APIResponse "Invoice not found"
// @Router      /invoices/{id} [delete]
func (h *Handler) DeleteInvoice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	if err := h.service.DeleteInvoice(uint(id)); err != nil {
		utils.SendError(c, http.StatusNotFound, "Invoice not found")
		return
	}

	utils.SendSuccess(c, "Invoice deleted successfully")
}

// MarkPaid godoc
// @Summary     Mark invoice as paid
// @Description Marks an invoice as paid and syncs revenue to client atomically.
// @Tags        Invoices
// @Accept      json
// @Produce     json
// @Param       id path int true "Invoice ID"
// @Param       body body MarkPaidRequest true "Mark paid payload"
// @Security    BearerAuth
// @Success     200 {object} utils.APIResponse "Invoice marked as paid"
// @Failure     400 {object} utils.APIResponse "Validation error or invoice cancelled"
// @Failure     404 {object} utils.APIResponse "Invoice not found"
// @Router      /invoices/{id}/mark-paid [patch]
func (h *Handler) MarkPaid(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	var req MarkPaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	input := MarkAsPaidInput{
		AmountPaid: req.AmountPaid,
		PaidAt:     req.PaidAt,
	}

	invoice, err := h.service.MarkAsPaid(uint(id), input)
	if err != nil {
		if err.Error() == "cannot mark a cancelled invoice as paid" {
			utils.SendError(c, http.StatusBadRequest, err.Error())
			return
		}
		if err.Error() == "invoice is already paid" {
			utils.SendError(c, http.StatusBadRequest, err.Error())
			return
		}
		utils.SendError(c, http.StatusNotFound, "Invoice not found")
		return
	}

	utils.SendSuccess(c, "Invoice marked as paid", invoice)
}
